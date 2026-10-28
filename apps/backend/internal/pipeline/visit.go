package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/subscriptions"
)

const (
	// relatedTarget is how many related scenes a visit aims for; fallback
	// searches only run while the scene has fewer.
	relatedTarget = 24
	// relatedSearchLimit caps the results taken from one fallback search.
	relatedSearchLimit = 12
	// Fallback searches per kind, strongest first (billing / tag order).
	maxPerformerSearches = 2
	maxTagSearches       = 2
	// detailRetryInterval is the least time between detail fetches of a visited
	// scene that still has no streams.
	detailRetryInterval = 10 * time.Minute
	// detailFetchTimeout bounds one detail fetch. It runs detached from the
	// visit that started it (a card's hover prefetch can end before the page
	// that waits for the same fetch), and sites behind FlareSolverr are slow.
	detailFetchTimeout = 2 * time.Minute
	// relatedFillTimeout bounds the background fallback searches of one visit.
	relatedFillTimeout = 2 * time.Minute
)

// relatedSearch is one fallback query for related scenes.
type relatedSearch struct {
	source string
	query  string
}

// VisitScene is called when a scene page is opened. The first visit fetches the
// scene's detail page from its origin plugin (details, streams, and the site's
// related list); every visit then tops up related scenes in the background.
// A visit while that fetch runs waits for it.
func (o *Orchestrator) VisitScene(ctx context.Context, sceneID string) error {
	needsFetch, err := o.needsDetailFetch(ctx, sceneID)
	if err != nil {
		return err
	}
	if needsFetch || o.fetchingDetail(sceneID) {
		if err := o.sharedDetailFetch(ctx, sceneID, o.fetchSceneDetail); err != nil {
			return err
		}
	}
	o.startRelatedFill(sceneID)
	return nil
}

// RefetchScene fetches the scene's detail page again whether or not it was
// visited, sharing a fetch already running for it.
func (o *Orchestrator) RefetchScene(ctx context.Context, sceneID string) error {
	return o.sharedDetailFetch(ctx, sceneID, o.fetchSceneDetail)
}

// needsDetailFetch reports whether a visit should fetch the scene's detail page:
// on the first visit, and again (at most once per detailRetryInterval) while a
// visited scene still has no streams, which is the case when its plugin gained
// the stream since, or the site failed to show one that time.
func (o *Orchestrator) needsDetailFetch(ctx context.Context, sceneID string) (bool, error) {
	fetched, err := o.ingestSvc.DetailFetched(ctx, sceneID)
	if err != nil || !fetched {
		return !fetched, err
	}
	streams, err := o.ingestSvc.StreamCount(ctx, sceneID)
	if err != nil || streams > 0 {
		return false, err
	}
	return o.mayRetryDetail(sceneID), nil
}

// mayRetryDetail reports whether a refetch of the scene is due and, if so,
// starts the interval.
func (o *Orchestrator) mayRetryDetail(sceneID string) bool {
	lastTry, tried := o.detailRetried.Load(sceneID)
	if tried && time.Since(lastTry.(time.Time)) < detailRetryInterval {
		return false
	}
	o.detailRetried.Store(sceneID, time.Now())
	return true
}

// sharedDetailFetch runs fetch for the id unless a fetch of it is running
// already, and waits for that one fetch either way: a visit that arrives
// mid-fetch (the page after a card's hover prefetch) gets its result instead
// of the scene as it was before. The fetch outlives a caller that gives up.
func (o *Orchestrator) sharedDetailFetch(ctx context.Context, id string, fetch func(context.Context, string) error) error {
	result := o.detailFetches.DoChan(id, func() (any, error) {
		o.detailInflight.Store(id, struct{}{})
		defer o.detailInflight.Delete(id)
		fetchContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), detailFetchTimeout)
		defer cancel()
		return nil, fetch(fetchContext, id)
	})
	select {
	case outcome := <-result:
		return outcome.Err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// fetchingDetail reports whether a detail fetch of the id is running.
func (o *Orchestrator) fetchingDetail(id string) bool {
	_, fetching := o.detailInflight.Load(id)
	return fetching
}

// fetchSceneDetail runs the origin plugin's scene:find on the scene's page and
// ingests the result.
func (o *Orchestrator) fetchSceneDetail(ctx context.Context, sceneID string) error {
	plugin, sourceURL, err := o.origin(ctx, sceneID)
	if err != nil {
		return err
	}
	results, err := o.runner.Find(ctx, plugin, plugins.MediaTypeScene, sourceURL)
	if err != nil {
		return fmt.Errorf("find %s via %s: %w", sourceURL, plugin.Meta.Name, err)
	}
	if err := o.ingestScrapeResult(ctx, plugin.Meta.Name, sourceURL, results); err != nil {
		return err
	}
	return o.ingestSvc.MarkDetailFetched(ctx, sceneID)
}

// origin returns the plugin and page URL a record (scene, gallery) was
// discovered from.
func (o *Orchestrator) origin(ctx context.Context, recordID string) (*plugins.Plugin, string, error) {
	row, err := o.ingestSvc.GetByID(ctx, recordID)
	if err != nil {
		return nil, "", err
	}
	if row == nil {
		return nil, "", fmt.Errorf("%s not found", recordID)
	}
	pluginName, sourceURL := o.ingestSvc.PrimarySource(ctx, row)
	if pluginName == "" || sourceURL == "" {
		return nil, "", fmt.Errorf("%s has no origin (plugin=%q url=%q)", recordID, pluginName, sourceURL)
	}
	plugin, ok := o.registry.Get(pluginName)
	if !ok || !o.registry.IsActive(pluginName) {
		return nil, "", fmt.Errorf("origin plugin %q not registered, disabled or unavailable", pluginName)
	}
	return plugin, sourceURL, nil
}

// startRelatedFill tops up sceneID's related scenes in the background, once at
// a time per scene.
func (o *Orchestrator) startRelatedFill(sceneID string) {
	if _, alreadyFilling := o.relatedInflight.LoadOrStore(sceneID, struct{}{}); alreadyFilling {
		return
	}
	go func() {
		defer o.relatedInflight.Delete(sceneID)
		ctx, cancel := context.WithTimeout(context.Background(), relatedFillTimeout)
		defer cancel()
		linked, err := o.fillRelated(ctx, sceneID)
		if err != nil {
			log.Printf("related fill %s: %v", sceneID, err)
		}
		if linked > 0 {
			o.hub.Publish(subscriptions.RelatedTopic(sceneID), nil)
		}
	}()
}

// fillRelated runs fallback searches on the scene's origin plugin until it
// has relatedTarget related scenes or the searches run out. Returns the
// number of edges added.
func (o *Orchestrator) fillRelated(ctx context.Context, sceneID string) (int, error) {
	count, err := o.ingestSvc.RelatedCount(ctx, sceneID)
	if err != nil || count >= relatedTarget {
		return 0, err
	}
	plugin, _, err := o.origin(ctx, sceneID)
	if err != nil {
		return 0, err
	}
	if !plugin.Meta.Has(plugins.ListCapability(plugins.MediaTypeScene)) {
		return 0, nil
	}
	searches, err := o.relatedSearches(ctx, sceneID)
	if err != nil {
		return 0, err
	}
	linked := 0
	for _, search := range searches {
		if count+linked >= relatedTarget {
			break
		}
		result, err := o.runner.List(ctx, plugin, plugins.MediaTypeScene, plugins.ListArgs{Query: search.query, Limit: relatedSearchLimit})
		if err != nil {
			log.Printf("related fill %s: %s search %q: %v", sceneID, search.source, search.query, err)
			continue
		}
		added, err := o.ingestSvc.LinkRelated(ctx, sceneID, plugin.Meta.Name, search.source, result.Items)
		linked += added
		if err != nil {
			return linked, err
		}
	}
	return linked, nil
}

// relatedSearches lists the fallback queries for a scene: its top performers,
// then its top tags, then its title.
func (o *Orchestrator) relatedSearches(ctx context.Context, sceneID string) ([]relatedSearch, error) {
	performers, tags, title, err := o.ingestSvc.SceneSearchTerms(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	searches := []relatedSearch{}
	for _, name := range firstN(performers, maxPerformerSearches) {
		searches = append(searches, relatedSearch{source: ingest.RelatedSourcePerformer, query: name})
	}
	for _, name := range firstN(tags, maxTagSearches) {
		searches = append(searches, relatedSearch{source: ingest.RelatedSourceTag, query: name})
	}
	if title != "" {
		searches = append(searches, relatedSearch{source: ingest.RelatedSourceTitle, query: title})
	}
	return searches, nil
}

func firstN(values []string, limit int) []string {
	if len(values) > limit {
		return values[:limit]
	}
	return values
}
