package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// searchFresh is how long a site's answer to a search page is shown again
// instead of asking the site.
const searchFresh = 6 * time.Hour

// SearchBatch is part of one site's answer to a search page.
type SearchBatch struct {
	Plugin string
	Refs   []schema.RecordRef
	// Stored: the answer is the one the site gave within searchFresh, not asked now.
	Stored bool
	// Final: the site's last batch for this page.
	Final bool
	// Exhausted: the site has no further pages for this search.
	Exhausted bool
	// Returned: how many results the site gave in this batch, before a filter.
	Returned int
}

// SearchPage is one page of a search: limit items per site, page 0 first.
type SearchPage struct {
	Query string
	Page  int
	Limit int
	// Content, when set, asks only the sites carrying that kind of content
	// (plugins.Content*), and a plugin carrying several only its sites of it.
	Content string
}

// storedAnswer is a site's answer to a search page as the cache keeps it.
type storedAnswer struct {
	Refs      []schema.RecordRef `json:"refs"`
	Exhausted bool               `json:"exhausted"`
}

// SearchStream asks every active plugin that lists entity (and carries the
// page's content, when it names one) for one page of a search and sends each
// site's answer on the returned channel as it comes, in batches when the site
// streams them. A site that answered the same page within searchFresh isn't
// asked; its stored answer is sent instead. Only pluginNames are asked when
// it's set. A filter, when given, narrows what's sent (the stored answers
// stay whole) and its sites narrow who's asked. The channel closes once every
// site answered, failed or timed out.
func (p *Pipeline) SearchStream(ctx context.Context, entity plugins.MediaType, search SearchPage, pluginNames []string, filter *schema.Filter) (<-chan SearchBatch, int) {
	searchers := p.searchers(entity, pluginNames, search.Content)
	if filter != nil && len(filter.Sources) > 0 {
		searchers = slices.DeleteFunc(searchers, func(plugin *plugins.Plugin) bool {
			return !slices.Contains(filter.Sources, plugin.Meta.Name)
		})
	}
	answers := make(chan SearchBatch)
	var group sync.WaitGroup
	for _, plugin := range searchers {
		group.Go(func() {
			p.searchPage(ctx, plugin, entity, search, answers)
		})
	}
	go func() {
		group.Wait()
		close(answers)
	}()
	return p.filtered(ctx, answers, filter), len(searchers)
}

// filtered passes the batches on with the results filter doesn't match
// left out, each counting what its site returned.
func (p *Pipeline) filtered(ctx context.Context, answers <-chan SearchBatch, filter *schema.Filter) <-chan SearchBatch {
	batches := make(chan SearchBatch)
	go func() {
		defer close(batches)
		for batch := range answers {
			batch.Returned = len(batch.Refs)
			if filter != nil && len(batch.Refs) > 0 {
				matching, err := p.library.Matching(ctx, batch.Refs, *filter)
				if err != nil {
					log.Printf("search %s: filter: %v", batch.Plugin, err)
				}
				batch.Refs = matching
			}
			send(ctx, batches, batch)
		}
	}()
	return batches
}

// Search asks every active plugin that lists entity for the first page of
// query, like SearchStream, and returns the results sites taking turns.
func (p *Pipeline) Search(ctx context.Context, entity plugins.MediaType, query string, limit int) []schema.RecordRef {
	batches, _ := p.SearchStream(ctx, entity, SearchPage{Query: query, Limit: limit}, nil, nil)
	bySite := map[string][]schema.RecordRef{}
	order := []string{}
	for batch := range batches {
		if _, seen := bySite[batch.Plugin]; !seen {
			order = append(order, batch.Plugin)
		}
		bySite[batch.Plugin] = append(bySite[batch.Plugin], batch.Refs...)
	}
	lists := make([][]schema.RecordRef, 0, len(order))
	for _, name := range order {
		lists = append(lists, bySite[name])
	}
	return interleave(lists)
}

// searchers returns the active plugins that list entity, only the named ones
// when names is set and only those carrying content when it's set.
func (p *Pipeline) searchers(entity plugins.MediaType, names []string, content string) []*plugins.Plugin {
	all := p.registry.WithCapability(plugins.ListCapability(entity))
	return slices.DeleteFunc(all, func(plugin *plugins.Plugin) bool {
		if len(names) > 0 && !slices.Contains(names, plugin.Meta.Name) {
			return true
		}
		return content != "" && !plugin.Carries(content)
	})
}

// searchPage sends one site's answer to a search page: the stored one when
// it's fresh, else the site's, batch by batch, which it then stores. A site
// that fails sends an empty final batch and stores nothing, so it's asked
// again next time.
func (p *Pipeline) searchPage(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, search SearchPage, batches chan<- SearchBatch) {
	name := plugin.Meta.Name
	key := searchKey(entity, name, search)
	if answer, ok := p.storedSearch(ctx, key); ok {
		send(ctx, batches, SearchBatch{Plugin: name, Refs: answer.Refs, Stored: true, Final: true, Exhausted: answer.Exhausted})
		return
	}
	searchContext, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()
	answer := storedAnswer{Refs: []schema.RecordRef{}}
	args := plugins.ListArgs{Query: search.Query, Limit: search.Limit, Offset: search.Page * search.Limit, Content: search.Content}
	result, err := p.runner.ListStream(searchContext, plugin, entity, args, func(items []plugins.DiscoveredItem) {
		refs := p.storeListed(ctx, name, items)
		answer.Refs = append(answer.Refs, refs...)
		send(ctx, batches, SearchBatch{Plugin: name, Refs: refs})
	})
	if err != nil {
		log.Printf("search %s %q page %d: %v", name, search.Query, search.Page, err)
		send(ctx, batches, SearchBatch{Plugin: name, Final: true})
		return
	}
	answer.Exhausted = len(result.Items) < search.Limit
	send(ctx, batches, SearchBatch{Plugin: name, Final: true, Exhausted: answer.Exhausted})
	p.storeSearch(ctx, key, answer)
}

// send hands a batch over unless the subscriber is gone.
func send(ctx context.Context, batches chan<- SearchBatch, batch SearchBatch) {
	select {
	case batches <- batch:
	case <-ctx.Done():
	}
}

// storedSearch returns a site's answer to a search page from within searchFresh.
func (p *Pipeline) storedSearch(ctx context.Context, key string) (storedAnswer, bool) {
	var answer storedAnswer
	if p.searches == nil {
		return answer, false
	}
	value, err := p.searches.Get(ctx, key)
	if errors.Is(err, cache.ErrMiss) {
		return answer, false
	}
	if err != nil {
		log.Printf("search memory %s: %v", key, err)
		return answer, false
	}
	if err := json.Unmarshal([]byte(value), &answer); err != nil {
		return answer, false
	}
	return answer, true
}

// storeSearch keeps a site's answer to a search page for searchFresh.
func (p *Pipeline) storeSearch(ctx context.Context, key string, answer storedAnswer) {
	if p.searches == nil {
		return
	}
	value, _ := json.Marshal(answer)
	if err := p.searches.Set(ctx, key, string(value), searchFresh); err != nil {
		log.Printf("search memory %s: %v", key, err)
	}
}

// searchKey names a site's answer to a search page; queries differing only
// in case or spacing share it. A content kind asks a plugin carrying several
// for less, so it's part of the name.
func searchKey(entity plugins.MediaType, pluginName string, search SearchPage) string {
	folded := strings.Join(strings.Fields(strings.ToLower(search.Query)), " ")
	key := fmt.Sprintf("search|%s|%s|%d|%d|%s", entity, pluginName, search.Page, search.Limit, folded)
	if search.Content == "" {
		return key
	}
	return key + "|" + search.Content
}
