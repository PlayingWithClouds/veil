// Package updater re-runs subscriptions on their schedule: saved searches and
// followed studios (channels), performers and tags.
package updater

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const (
	// checkInterval is how often due subscriptions are looked for; each
	// subscription's own interval is in hours.
	checkInterval = 5 * time.Minute
	// resultLimit caps the results each plugin returns per run.
	resultLimit   = 40
	runTimeout    = 5 * time.Minute
	recordTimeout = 15 * time.Second
)

// ErrAlreadyRunning is returned when a subscription's search is in progress.
var ErrAlreadyRunning = errors.New("search subscription is already running")

// errNoSearchPlugin is recorded when no active plugin can run a subscription.
var errNoSearchPlugin = errors.New("no active search plugin matches this subscription")

// Discoverer is what the scheduler needs from discovery.Service: picking
// plugins, keyword search and page listing, each ingesting hits as stubs.
type Discoverer interface {
	Plugins(sources []string) []*plugins.Plugin
	Run(ctx context.Context, options discovery.Options, searchPlugins []*plugins.Plugin, onHit func(discovery.Hit)) error
	ListPage(ctx context.Context, plugin *plugins.Plugin, pageURL string, limit int, onHit func(discovery.Hit)) error
}

// Scheduler runs due subscriptions and records what they find.
type Scheduler struct {
	repo      *media.Repository
	discovery Discoverer
	// Subscription ids with a run in progress.
	running sync.Map
}

// NewScheduler creates a scheduler.
func NewScheduler(repo *media.Repository, discoverySvc Discoverer) *Scheduler {
	return &Scheduler{repo: repo, discovery: discoverySvc}
}

// Start runs overdue subscriptions right away, then keeps checking until ctx
// is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	s.runDue(ctx)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runDue(ctx)
		}
	}
}

// RunInBackground starts one subscription's run without waiting for it.
func (s *Scheduler) RunInBackground(subscription *model.SearchSubscription) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
		defer cancel()
		if err := s.Run(ctx, subscription); err != nil && !errors.Is(err, ErrAlreadyRunning) {
			log.Printf("subscription %q: %v", subscription.Query, err)
		}
	}()
}

// Run collects the subscription's scenes, adds them to its feed and schedules
// the next run. Plugin failures are stored on the subscription, not returned;
// the error is for failures to run or record at all.
func (s *Scheduler) Run(ctx context.Context, subscription *model.SearchSubscription) error {
	if _, alreadyRunning := s.running.LoadOrStore(subscription.ID, struct{}{}); alreadyRunning {
		return ErrAlreadyRunning
	}
	defer s.running.Delete(subscription.ID)

	var mediaIDs []string
	var runErr error
	if subscription.Target == nil {
		mediaIDs, runErr = s.collectSearch(ctx, subscription)
	} else {
		mediaIDs, runErr = s.collectEntity(ctx, subscription.Target.ID)
	}
	// Keep what was found even when the run ran out of time.
	recordContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), recordTimeout)
	defer cancel()
	return s.repo.RecordSearchSubscriptionRun(recordContext, subscription.ID, mediaIDs, runErr)
}

// collectSearch runs a saved search on its plugins.
func (s *Scheduler) collectSearch(ctx context.Context, subscription *model.SearchSubscription) ([]string, error) {
	searchPlugins := s.discovery.Plugins(subscription.Sources)
	if len(searchPlugins) == 0 {
		return nil, errNoSearchPlugin
	}
	return s.search(ctx, subscription.Query, searchPlugins)
}

// collectEntity gathers a followed studio/performer/tag's scenes: its own
// page on its site when a plugin can list that page, otherwise a keyword
// search for its name; plus the library scenes already credited to it.
func (s *Scheduler) collectEntity(ctx context.Context, entityID string) ([]string, error) {
	origin, err := s.repo.EntityOrigin(ctx, entityID)
	if err != nil {
		return nil, err
	}
	mediaIDs, listed, runErr := s.listEntityPage(ctx, origin)
	if !listed {
		searched, searchErr := s.searchEntity(ctx, entityID, origin)
		mediaIDs = searched
		runErr = errors.Join(runErr, searchErr)
	}
	library, err := s.repo.LinkedSceneIDs(ctx, entityID)
	if err != nil {
		return mediaIDs, errors.Join(runErr, err)
	}
	return append(mediaIDs, library...), runErr
}

// listEntityPage lists the entity's page on its site, trying each plugin that
// can list it until one succeeds. listed is false when none could.
func (s *Scheduler) listEntityPage(ctx context.Context, origin *media.EntityOrigin) (mediaIDs []string, listed bool, err error) {
	var failures []error
	for _, plugin := range s.pagePlugins(origin) {
		var found []string
		listErr := s.discovery.ListPage(ctx, plugin, origin.SourceURL, resultLimit, collectHits(&found))
		if listErr == nil {
			return found, true, nil
		}
		failures = append(failures, fmt.Errorf("%s: %w", plugin.Meta.Name, listErr))
	}
	return nil, false, errors.Join(failures...)
}

// pagePlugins returns the active plugins that can list the entity's page:
// they declare scene:list:page and either serve its host or ingested the
// entity themselves.
func (s *Scheduler) pagePlugins(origin *media.EntityOrigin) []*plugins.Plugin {
	if origin.SourceURL == "" {
		return nil
	}
	observers := stringSet(origin.Observers)
	var listers []*plugins.Plugin
	for _, plugin := range s.discovery.Plugins(nil) {
		if !plugin.Meta.Has(plugins.CapabilitySceneListPage) {
			continue
		}
		if plugin.ServesURL(origin.SourceURL) || observers[plugin.Meta.Name] {
			listers = append(listers, plugin)
		}
	}
	return listers
}

// searchEntity searches the entity's name on the plugins it came from (every
// search plugin when none of those is active) and keeps the hits credited to
// it, plus stubs not credited to anything of its kind yet. Stubs lack
// credits until visited, so those are keyword matches: the price of
// following an entity on a site that has no page listing for it.
func (s *Scheduler) searchEntity(ctx context.Context, entityID string, origin *media.EntityOrigin) ([]string, error) {
	allPlugins := s.discovery.Plugins(nil)
	searchPlugins := s.discovery.Plugins(originPluginNames(origin, allPlugins))
	if len(searchPlugins) == 0 {
		searchPlugins = allPlugins
	}
	if len(searchPlugins) == 0 {
		return nil, errNoSearchPlugin
	}
	found, searchErr := s.search(ctx, origin.Name, searchPlugins)
	kept, err := s.repo.KeepLinkedOrUncredited(ctx, entityID, found)
	if err != nil {
		return nil, errors.Join(searchErr, err)
	}
	return kept, searchErr
}

// search runs a keyword search and returns the ids of the ingested hits.
func (s *Scheduler) search(ctx context.Context, query string, searchPlugins []*plugins.Plugin) ([]string, error) {
	var mediaIDs []string
	err := s.discovery.Run(ctx, discovery.Options{Query: query, Limit: resultLimit}, searchPlugins, collectHits(&mediaIDs))
	return mediaIDs, err
}

// runDue runs every due subscription, one after another.
func (s *Scheduler) runDue(ctx context.Context) {
	due, err := s.repo.DueSearchSubscriptions(ctx)
	if err != nil {
		log.Printf("scheduler: list due subscriptions: %v", err)
		return
	}
	for _, subscription := range due {
		if ctx.Err() != nil {
			return
		}
		s.runOne(ctx, subscription)
	}
}

// runOne runs a single subscription under runTimeout.
func (s *Scheduler) runOne(ctx context.Context, subscription *model.SearchSubscription) {
	runContext, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()
	if err := s.Run(runContext, subscription); err != nil && !errors.Is(err, ErrAlreadyRunning) {
		log.Printf("scheduler: subscription %q: %v", subscription.Query, err)
	}
}

// originPluginNames names the plugins an entity came from: those serving its
// source URL's host, those that ingested it, then those whose scenes credit it.
func originPluginNames(origin *media.EntityOrigin, candidates []*plugins.Plugin) []string {
	var names []string
	for _, plugin := range candidates {
		if origin.SourceURL != "" && plugin.ServesURL(origin.SourceURL) {
			names = append(names, plugin.Meta.Name)
		}
	}
	names = append(names, origin.Observers...)
	names = append(names, origin.SceneSources...)
	return names
}

// collectHits returns an onHit callback appending each ingested hit's id.
func collectHits(mediaIDs *[]string) func(discovery.Hit) {
	return func(hit discovery.Hit) {
		if hit.ID != "" {
			*mediaIDs = append(*mediaIDs, hit.ID)
		}
	}
}

// stringSet turns a list into a membership map.
func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
