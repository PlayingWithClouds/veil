package updater

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
)

const (
	// feedInterval is how often the feeder pulls from the sites on its own.
	feedInterval = 12 * time.Hour
	// feedCheckInterval is how often Start looks whether a run is due.
	feedCheckInterval = 15 * time.Minute
	// topUpCooldown spaces out the runs a short candidate pool asks for.
	topUpCooldown = time.Hour
	// feedRunTimeout bounds one whole run, feedListTimeout one listing in it.
	feedRunTimeout  = 5 * time.Minute
	feedListTimeout = 45 * time.Second
	// catalogItemLimit is how many scenes each plugin's catalog page lists;
	// catalogPages is how deep the runs go before starting over at page one.
	catalogItemLimit = 40
	catalogPages     = 3
	// How many top-affinity entities per dimension get their pages pulled.
	tastePerformerLimit = 3
	tasteStudioLimit    = 2
	tasteTagLimit       = 3
	// feedRetention is how long a find is kept after it was last listed.
	feedRetention = 30 * 24 * time.Hour

	feedLastRunKey     = "feeder:last-run"
	feedCatalogPageKey = "feeder:catalog-page"
)

// TasteSource is what the feeder needs from the recommendation engine.
type TasteSource interface {
	Profile(ctx context.Context) (*recommend.Profile, error)
}

// Feeder keeps the recommendation feed supplied without the user searching:
// it lists each scene plugin's catalog and the site pages of the top-affinity
// performers, studios and tags, ingests the scenes as stubs and records them
// in feed_discovery, where the engine's discover:* sources read them. It runs
// every feedInterval, and sooner when the engine reports a short pool.
type Feeder struct {
	scheduler *Scheduler
	database  *db.DB
	taste     TasteSource
	marks     FetchMarks
	// onFed is called after a run that stored scenes.
	onFed func()

	mutex       sync.Mutex
	running     bool
	lastStarted time.Time
}

// NewFeeder creates a feeder on top of the scheduler's discovery; onFed may be nil.
func NewFeeder(scheduler *Scheduler, database *db.DB, taste TasteSource, marks FetchMarks, onFed func()) *Feeder {
	return &Feeder{scheduler: scheduler, database: database, taste: taste, marks: marks, onFed: onFed}
}

// Start runs whenever the last successful run is older than feedInterval,
// checking every feedCheckInterval until ctx is cancelled.
func (f *Feeder) Start(ctx context.Context) {
	ticker := time.NewTicker(feedCheckInterval)
	defer ticker.Stop()
	for {
		if f.due(ctx) {
			f.runGuarded(ctx)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// TopUp starts a run in the background unless one is running or started
// within topUpCooldown. It is the engine's low-pool callback.
func (f *Feeder) TopUp() {
	f.mutex.Lock()
	cooling := time.Since(f.lastStarted) < topUpCooldown
	f.mutex.Unlock()
	if cooling {
		return
	}
	go f.runGuarded(context.Background())
}

// due reports whether no run has succeeded within feedInterval.
func (f *Feeder) due(ctx context.Context) bool {
	_, err := f.marks.Get(ctx, feedLastRunKey)
	return err != nil
}

// runGuarded runs once under feedRunTimeout, unless a run is in progress.
func (f *Feeder) runGuarded(ctx context.Context) {
	if !f.claim() {
		return
	}
	defer f.release()
	runContext, cancel := context.WithTimeout(ctx, feedRunTimeout)
	defer cancel()
	stored, err := f.Run(runContext)
	if err != nil {
		log.Printf("feeder: %v", err)
	}
	log.Printf("feeder: stored %d scenes", stored)
}

// claim marks a run as started; false when one is already running.
func (f *Feeder) claim() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.running {
		return false
	}
	f.running = true
	f.lastStarted = time.Now()
	return true
}

// release marks the run as finished.
func (f *Feeder) release() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.running = false
}

// Run pulls the catalog and taste pages once and returns how many scenes it
// recorded. A run that stored anything counts as the interval's run; one
// that found nothing (no plugins installed yet, offline) is retried at the
// next check. Listing failures are joined into the error.
func (f *Feeder) Run(ctx context.Context) (int, error) {
	catalogStored, catalogErr := f.feedCatalog(ctx)
	tasteStored, tasteErr := f.feedTaste(ctx)
	stored := catalogStored + tasteStored
	recordContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), recordTimeout)
	defer cancel()
	pruneErr := f.prune(recordContext)
	if stored == 0 {
		return 0, errors.Join(catalogErr, tasteErr, pruneErr)
	}
	markErr := f.marks.Set(recordContext, feedLastRunKey, db.FormatTime(time.Now()), feedInterval)
	if f.onFed != nil {
		f.onFed()
	}
	return stored, errors.Join(catalogErr, tasteErr, pruneErr, markErr)
}

// feedCatalog lists one catalog page of every scene plugin, a page deeper
// than the previous run, and records each plugin's scenes in listing order.
func (f *Feeder) feedCatalog(ctx context.Context) (int, error) {
	catalogPlugins := f.scheduler.discovery.Plugins(nil)
	if len(catalogPlugins) == 0 {
		return 0, nil
	}
	page := f.nextCatalogPage(ctx)
	scenesByPlugin := map[string][]string{}
	listContext, cancel := context.WithTimeout(ctx, feedListTimeout)
	defer cancel()
	options := discovery.Options{Limit: catalogItemLimit, Offset: page * catalogItemLimit, Paginate: true, Entity: plugins.MediaTypeScene}
	listErr := f.scheduler.discovery.Run(listContext, options, catalogPlugins, func(hit discovery.Hit) {
		if strings.HasPrefix(hit.ID, "scene:") {
			scenesByPlugin[hit.Plugin] = append(scenesByPlugin[hit.Plugin], hit.ID)
		}
	})
	stored := 0
	for pluginName, sceneIDs := range scenesByPlugin {
		count, err := f.record(ctx, "catalog", pluginName, sceneIDs)
		if err != nil {
			return stored, errors.Join(listErr, err)
		}
		stored += count
	}
	return stored, listErr
}

// nextCatalogPage returns the catalog page this run lists and advances the
// stored page, wrapping after catalogPages.
func (f *Feeder) nextCatalogPage(ctx context.Context) int {
	page := 0
	if value, err := f.marks.Get(ctx, feedCatalogPageKey); err == nil {
		page, _ = strconv.Atoi(value)
	}
	page = page % catalogPages
	if err := f.marks.Set(ctx, feedCatalogPageKey, strconv.Itoa(page+1), 0); err != nil {
		log.Printf("feeder: store catalog page: %v", err)
	}
	return page
}

// feedTaste pulls the scenes of the profile's top performers, studios and
// tags, one entity after another to stay gentle on the sites.
func (f *Feeder) feedTaste(ctx context.Context) (int, error) {
	profile, err := f.taste.Profile(ctx)
	if err != nil {
		return 0, fmt.Errorf("taste profile: %w", err)
	}
	stored := 0
	var failures []error
	for _, entityID := range tasteEntities(profile) {
		if ctx.Err() != nil {
			break
		}
		count, err := f.feedEntity(ctx, entityID)
		stored += count
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", entityID, err))
		}
	}
	return stored, errors.Join(failures...)
}

// tasteEntities lists the entities whose pages a run pulls, performers first.
func tasteEntities(profile *recommend.Profile) []string {
	entities := recommend.TopEntities(profile.Performers, tastePerformerLimit)
	entities = append(entities, recommend.TopEntities(profile.Studios, tasteStudioLimit)...)
	return append(entities, recommend.TopEntities(profile.Tags, tasteTagLimit)...)
}

// feedEntity fetches one entity's scenes and records them.
func (f *Feeder) feedEntity(ctx context.Context, entityID string) (int, error) {
	fetchContext, cancel := context.WithTimeout(ctx, feedListTimeout)
	defer cancel()
	sceneIDs, fetchErr := f.scheduler.fetchEntity(fetchContext, entityID)
	count, err := f.record(ctx, "taste", entityID, onlyScenes(sceneIDs))
	return count, errors.Join(fetchErr, err)
}

// onlyScenes keeps the scene ids of a mixed media id list.
func onlyScenes(mediaIDs []string) []string {
	var sceneIDs []string
	for _, mediaID := range mediaIDs {
		if strings.HasPrefix(mediaID, "scene:") {
			sceneIDs = append(sceneIDs, mediaID)
		}
	}
	return sceneIDs
}

// record upserts one listing's scenes into feed_discovery, ranked by their
// position, and returns how many it wrote.
func (f *Feeder) record(ctx context.Context, source, entity string, sceneIDs []string) (int, error) {
	unique := uniqueInOrder(sceneIDs)
	if len(unique) == 0 {
		return 0, nil
	}
	_, err := f.database.Exec(ctx,
		`INSERT INTO feed_discovery (scene, source, entity, rank, found_at)
		 SELECT value, $source, $entity, key, $now FROM json_each($scenes) WHERE true
		 ON CONFLICT (source, entity, scene) DO UPDATE SET rank = excluded.rank, found_at = excluded.found_at`,
		db.Vars{"source": source, "entity": entity, "scenes": unique, "now": db.FormatTime(time.Now())})
	if err != nil {
		return 0, fmt.Errorf("record %s finds: %w", source, err)
	}
	return len(unique), nil
}

// prune drops finds not listed again within feedRetention.
func (f *Feeder) prune(ctx context.Context) error {
	_, err := f.database.Exec(ctx, `DELETE FROM feed_discovery WHERE found_at < $before`,
		db.Vars{"before": db.FormatTime(time.Now().Add(-feedRetention))})
	if err != nil {
		return fmt.Errorf("prune finds: %w", err)
	}
	return nil
}

// uniqueInOrder drops repeated ids, keeping first occurrences.
func uniqueInOrder(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
