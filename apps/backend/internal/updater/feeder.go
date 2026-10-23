package updater

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/media"
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
	// exploreTagLimit is how many of the profile's adjacent tags get their
	// site pages pulled per run (a random pick, so every adjacent tag comes
	// up over a few runs).
	exploreTagLimit = 3
	// popularPagesPerRun is how many of the most-viewed pages are listed per
	// run, taking turns through popularPages.
	popularPagesPerRun = 2
	// feedRetention is how long a find is kept after it was last listed.
	feedRetention = 30 * 24 * time.Hour

	feedLastRunKey = "feeder:last-run"
	// feedCatalogPageKey + plugin name holds the catalog page it lists next.
	feedCatalogPageKey = "feeder:catalog-page:"
	// feedPopularIndexKey holds the position in popularPages the next run
	// starts at.
	feedPopularIndexKey = "feeder:popular-index"
)

// defaultPopularPages are the most-viewed listings the feeder pulls. No plugin
// capability asks a site for its popular videos, so these are pages of the
// sites themselves, listed through scene:list:page by the plugin that serves
// their host (today the kvs sites; eporner's page listing only follows
// channels, pornstars, categories and profiles, so its "best videos" page
// needs a plugin change first). Each was checked to list through kvs.
var defaultPopularPages = []string{
	"https://www.pornhat.com/most-popular/",
	"https://ok.xxx/popular/",
	"https://www.porngo.com/most-popular/",
	"https://www.xozilla.com/most-popular/",
	"https://www.perfectgirls.xxx/most-popular/",
	"https://www.porntrex.com/most-popular/",
	"https://xcafe.com/most-popular/",
}

// TasteSource is what the feeder needs from the recommendation engine.
type TasteSource interface {
	Profile(ctx context.Context) (*recommend.Profile, error)
}

// Feeder keeps the recommendation feed supplied without the user searching:
// it lists each scene plugin's catalog and the site pages of the top-affinity
// performers, studios and tags, ingests the scenes as stubs and records them
// in feed_discovery, where the engine's discover:* and explore:site sources
// read them. Beyond the user's taste it lists the site pages of the profile's
// adjacent tags (exploration from the sites rather than the library) and a few
// most-viewed pages. It runs every feedInterval, and sooner when the engine
// reports a short pool.
type Feeder struct {
	scheduler *Scheduler
	database  *db.DB
	taste     TasteSource
	marks     FetchMarks
	// popularPages are the most-viewed listings taken turns through.
	popularPages []string
	// onFed is called after a run that stored scenes.
	onFed func()

	mutex       sync.Mutex
	running     bool
	lastStarted time.Time
}

// NewFeeder creates a feeder on top of the scheduler's discovery; onFed may be nil.
func NewFeeder(scheduler *Scheduler, database *db.DB, taste TasteSource, marks FetchMarks, onFed func()) *Feeder {
	return &Feeder{scheduler: scheduler, database: database, taste: taste, marks: marks, popularPages: defaultPopularPages, onFed: onFed}
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

// Run pulls the catalog, taste, explore and popular pages once and returns how
// many scenes it recorded. A run that stored anything counts as the
// interval's run; one that found nothing (no plugins installed yet, offline)
// is retried at the next check. Listing failures are joined into the error.
func (f *Feeder) Run(ctx context.Context) (int, error) {
	catalogStored, catalogErr := f.feedCatalog(ctx, f.scheduler.discovery.Plugins(nil))
	profile, profileErr := f.taste.Profile(ctx)
	stored := catalogStored
	failures := []error{catalogErr}
	if profileErr != nil {
		failures = append(failures, fmt.Errorf("taste profile: %w", profileErr))
	} else {
		tasteStored, tasteErr := f.feedTaste(ctx, profile)
		exploreStored, exploreErr := f.feedExplore(ctx, profile)
		stored += tasteStored + exploreStored
		failures = append(failures, tasteErr, exploreErr)
	}
	popularStored, popularErr := f.feedPopular(ctx)
	stored += popularStored
	failures = append(failures, popularErr)
	recordContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), recordTimeout)
	defer cancel()
	failures = append(failures, f.prune(recordContext))
	if stored == 0 {
		return 0, errors.Join(failures...)
	}
	failures = append(failures, f.marks.Set(recordContext, feedLastRunKey, db.FormatTime(time.Now()), feedInterval))
	if f.onFed != nil {
		f.onFed()
	}
	return stored, errors.Join(failures...)
}

// FeedSites lists the next catalog page of the named plugins right away and
// returns how many scenes it recorded: a site-filtered feed that came up
// short. Runs beside the scheduled runs; the upserts don't conflict.
func (f *Feeder) FeedSites(ctx context.Context, pluginNames []string) (int, error) {
	sitePlugins := f.scheduler.discovery.Plugins(pluginNames)
	if len(pluginNames) == 0 || len(sitePlugins) == 0 {
		return 0, nil
	}
	stored, err := f.feedCatalog(ctx, sitePlugins)
	if stored > 0 && f.onFed != nil {
		f.onFed()
	}
	return stored, err
}

// feedCatalog lists one catalog page of each plugin, a page deeper than that
// plugin's previous listing, and records each plugin's scenes in listing order.
func (f *Feeder) feedCatalog(ctx context.Context, catalogPlugins []*plugins.Plugin) (int, error) {
	scenesByPlugin := map[string][]string{}
	listContext, cancel := context.WithTimeout(ctx, feedListTimeout)
	defer cancel()
	var listFailures []error
	for page, pagePlugins := range f.pluginsByCatalogPage(ctx, catalogPlugins) {
		options := discovery.Options{Limit: catalogItemLimit, Offset: page * catalogItemLimit, Paginate: true, Entity: plugins.MediaTypeScene}
		err := f.scheduler.discovery.Run(listContext, options, pagePlugins, func(hit discovery.Hit) {
			if strings.HasPrefix(hit.ID, "scene:") {
				scenesByPlugin[hit.Plugin] = append(scenesByPlugin[hit.Plugin], hit.ID)
			}
		})
		listFailures = append(listFailures, err)
	}
	listErr := errors.Join(listFailures...)
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

// pluginsByCatalogPage groups the plugins by the catalog page each lists
// next, advancing every plugin's stored page (wrapping after catalogPages).
func (f *Feeder) pluginsByCatalogPage(ctx context.Context, catalogPlugins []*plugins.Plugin) map[int][]*plugins.Plugin {
	byPage := map[int][]*plugins.Plugin{}
	for _, plugin := range catalogPlugins {
		page := f.nextCatalogPage(ctx, plugin.Meta.Name)
		byPage[page] = append(byPage[page], plugin)
	}
	return byPage
}

// nextCatalogPage returns the catalog page a plugin lists now and stores the
// one after it.
func (f *Feeder) nextCatalogPage(ctx context.Context, pluginName string) int {
	key := feedCatalogPageKey + pluginName
	page := 0
	if value, err := f.marks.Get(ctx, key); err == nil {
		page, _ = strconv.Atoi(value)
	}
	page = page % catalogPages
	if err := f.marks.Set(ctx, key, strconv.Itoa(page+1), 0); err != nil {
		log.Printf("feeder: store catalog page: %v", err)
	}
	return page
}

// feedTaste pulls the scenes of the profile's top performers, studios and
// tags, one entity after another to stay gentle on the sites.
func (f *Feeder) feedTaste(ctx context.Context, profile *recommend.Profile) (int, error) {
	return f.feedEntities(ctx, "taste", tasteEntities(profile))
}

// feedExplore pulls the site pages of a few of the profile's adjacent tags:
// subjects close to what the user likes that they have not seen much of.
func (f *Feeder) feedExplore(ctx context.Context, profile *recommend.Profile) (int, error) {
	return f.feedEntities(ctx, "explore", exploreTags(profile))
}

// feedEntities fetches each entity's scenes in turn and records them under
// source.
func (f *Feeder) feedEntities(ctx context.Context, source string, entityIDs []string) (int, error) {
	stored := 0
	var failures []error
	for _, entityID := range entityIDs {
		if ctx.Err() != nil {
			break
		}
		count, err := f.feedEntity(ctx, source, entityID)
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

// exploreTags picks up to exploreTagLimit of the profile's adjacent tags at
// random.
func exploreTags(profile *recommend.Profile) []string {
	tagIDs := append([]string(nil), profile.AdjacentTags...)
	rand.Shuffle(len(tagIDs), func(first, second int) { tagIDs[first], tagIDs[second] = tagIDs[second], tagIDs[first] })
	if len(tagIDs) > exploreTagLimit {
		tagIDs = tagIDs[:exploreTagLimit]
	}
	return tagIDs
}

// feedEntity fetches one entity's scenes and records them under source.
func (f *Feeder) feedEntity(ctx context.Context, source, entityID string) (int, error) {
	fetchContext, cancel := context.WithTimeout(ctx, feedListTimeout)
	defer cancel()
	sceneIDs, fetchErr := f.scheduler.fetchEntity(fetchContext, entityID)
	count, err := f.record(ctx, source, entityID, onlyScenes(sceneIDs))
	return count, errors.Join(fetchErr, err)
}

// feedPopular lists the next popularPagesPerRun most-viewed pages, taking
// turns through popularPages across runs, and records each under its host.
func (f *Feeder) feedPopular(ctx context.Context) (int, error) {
	stored := 0
	var failures []error
	for _, pageURL := range f.nextPopularPages(ctx) {
		if ctx.Err() != nil {
			break
		}
		count, err := f.feedPopularPage(ctx, pageURL)
		stored += count
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", pageURL, err))
		}
	}
	return stored, errors.Join(failures...)
}

// nextPopularPages returns the pages this run lists and stores where the next
// run continues.
func (f *Feeder) nextPopularPages(ctx context.Context) []string {
	if len(f.popularPages) == 0 {
		return nil
	}
	start := 0
	if value, err := f.marks.Get(ctx, feedPopularIndexKey); err == nil {
		start, _ = strconv.Atoi(value)
	}
	count := min(popularPagesPerRun, len(f.popularPages))
	pages := make([]string, 0, count)
	for offset := range count {
		pages = append(pages, f.popularPages[(start+offset)%len(f.popularPages)])
	}
	next := (start + count) % len(f.popularPages)
	if err := f.marks.Set(ctx, feedPopularIndexKey, strconv.Itoa(next), 0); err != nil {
		log.Printf("feeder: store popular index: %v", err)
	}
	return pages
}

// feedPopularPage lists one most-viewed page through the plugin serving its
// host and records its scenes under that host. A page no installed plugin
// serves is skipped.
func (f *Feeder) feedPopularPage(ctx context.Context, pageURL string) (int, error) {
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return 0, err
	}
	listContext, cancel := context.WithTimeout(ctx, feedListTimeout)
	defer cancel()
	mediaIDs, listed, listErr := f.scheduler.listEntityPage(listContext, &media.EntityOrigin{SourceURL: pageURL})
	if !listed {
		return 0, listErr
	}
	return f.record(ctx, "popular", parsed.Host, onlyScenes(mediaIDs))
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
