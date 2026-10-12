package sitesuggest

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// testSite returns a plugin named name.
func testSite(name string) *plugins.Plugin {
	return &plugins.Plugin{Meta: plugins.PluginMeta{Name: name}}
}

// newTestService returns a service over the sites, answered by suggest.
func newTestService(suggest suggestFunc, sites ...*plugins.Plugin) *Service {
	return &Service{
		sites:   func() []*plugins.Plugin { return sites },
		suggest: suggest,
		now:     time.Now,
		cache:   map[string]cachedAnswer{},
	}
}

// queries makes completions of kind "query" from texts.
func queries(texts ...string) *plugins.SuggestResult {
	result := &plugins.SuggestResult{}
	for _, text := range texts {
		result.Suggestions = append(result.Suggestions, plugins.SiteSuggestion{Text: text})
	}
	return result
}

// TestSuggestRanksCompletionsSitesAgreeOn checks that a completion offered by
// several sites counts once, lists all of them and comes before single-site ones.
func TestSuggestRanksCompletionsSitesAgreeOn(t *testing.T) {
	answers := map[string]*plugins.SuggestResult{
		"alpha": queries("isla fisher", "Isla Summer", "isla dorian"),
		"beta":  queries("isla summer", "isla moon"),
	}
	service := newTestService(func(_ context.Context, plugin *plugins.Plugin, _ plugins.SuggestArgs) (*plugins.SuggestResult, error) {
		return answers[plugin.Meta.Name], nil
	}, testSite("beta"), testSite("alpha"))

	got := service.Suggest(context.Background(), "isla", 10)

	if len(got) != 4 {
		t.Fatalf("got %d suggestions, want 4: %+v", len(got), got)
	}
	if got[0].Text != "Isla Summer" || len(got[0].Sources) != 2 || got[0].Sources[0] != "alpha" || got[0].Sources[1] != "beta" {
		t.Fatalf("first = %+v, want Isla Summer from alpha and beta", got[0])
	}
	if got[0].Kind != "query" {
		t.Fatalf("kind = %q, want the default query", got[0].Kind)
	}
}

// TestSuggestLeavesOutFailingAndSlowSites checks that a site that errors or
// overruns its time limit doesn't hold back or discard the others.
func TestSuggestLeavesOutFailingAndSlowSites(t *testing.T) {
	service := newTestService(func(ctx context.Context, plugin *plugins.Plugin, _ plugins.SuggestArgs) (*plugins.SuggestResult, error) {
		switch plugin.Meta.Name {
		case "broken":
			return nil, errors.New("site is down")
		case "slow":
			<-ctx.Done()
			return nil, ctx.Err()
		}
		return queries("isla summer"), nil
	}, testSite("broken"), testSite("slow"), testSite("fine"))

	started := time.Now()
	got := service.Suggest(context.Background(), "isla", 10)

	if len(got) != 1 || got[0].Text != "isla summer" {
		t.Fatalf("got %+v, want the working site's completion", got)
	}
	if elapsed := time.Since(started); elapsed > siteTimeout+500*time.Millisecond {
		t.Fatalf("took %s, the slow site was not cut off at %s", elapsed, siteTimeout)
	}
}

// TestSuggestAsksSitesOncePerQuery checks that a repeated query is answered from
// the cache, and that very short queries ask no site at all.
func TestSuggestAsksSitesOncePerQuery(t *testing.T) {
	var asked atomic.Int32
	service := newTestService(func(context.Context, *plugins.Plugin, plugins.SuggestArgs) (*plugins.SuggestResult, error) {
		asked.Add(1)
		return queries("isla summer"), nil
	}, testSite("alpha"))

	service.Suggest(context.Background(), "Isla", 10)
	service.Suggest(context.Background(), "isla", 10)
	if asked.Load() != 1 {
		t.Fatalf("site asked %d times for one query, want 1", asked.Load())
	}
	if got := service.Suggest(context.Background(), "i", 10); len(got) != 0 || asked.Load() != 1 {
		t.Fatalf("one-letter query asked a site: %+v (asked %d)", got, asked.Load())
	}
}

// TestSuggestDropsTheTypedText checks that echoing the query back is not a suggestion.
func TestSuggestDropsTheTypedText(t *testing.T) {
	service := newTestService(func(context.Context, *plugins.Plugin, plugins.SuggestArgs) (*plugins.SuggestResult, error) {
		return queries("Isla", "isla summer"), nil
	}, testSite("alpha"))

	got := service.Suggest(context.Background(), "isla", 10)

	if len(got) != 1 || got[0].Text != "isla summer" {
		t.Fatalf("got %+v, want only the completion", got)
	}
}
