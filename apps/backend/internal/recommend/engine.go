// Package recommend is the recommendation engine: candidate generation from
// cheap local sources, a transparent weighted ranking against a time-decayed
// taste profile, and diversity-aware re-ranking. Everything runs as a handful
// of indexed SQLite queries per ranking pass, so it works embedded on a phone.
//
// The taste profile is computed on demand rather than cached in a table: the
// signal tables are one row per user action inside a bounded window, so the
// aggregation stays small, and there is no cache to invalidate on every like,
// o-event or watch progress write. The ranked feed itself is memoized for
// pagination (see feedCacheTTL).
package recommend

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/media"
)

// Engine ranks recommendations. Safe for concurrent use.
type Engine struct {
	database *db.DB
	repo     *media.Repository
	// now is the clock, swappable in tests.
	now func() time.Time

	mutex sync.Mutex
	cache *rankedFeed
}

// rankedFeed is one full ranking pass.
type rankedFeed struct {
	builtAt   time.Time
	pluginKey string
	// scored holds every eligible candidate, best score first.
	scored []scoredItem
	// ranked is the re-ranked feed order.
	ranked []Item
	// observingPlugins lists, per ranked scene, every plugin that observed it,
	// for the feed's source filter.
	observingPlugins map[string][]string
	// durations holds each ranked scene's runtime in seconds (absent when
	// unknown), for the duration filter.
	durations map[string]int
}

// New returns an engine over the database; repo supplies the blocklist and
// listing exclusions shared with the rest of the app.
func New(database *db.DB, repo *media.Repository) *Engine {
	return &Engine{database: database, repo: repo, now: time.Now}
}

// Feed returns one page of the ranked feed. Offset 0 re-ranks once the
// ranking is older than feedFirstPageTTL (a reload shows the same feed), or
// always when refresh is set; later pages read the same ranking while it is
// fresh, so pages neither overlap nor skip as impressions shift the scores.
// Non-empty sources keep only scenes some of those plugins observed (the
// same rule as the scene listings' source filter); offset and limit then
// page through that filtered view of the one shared ranking.
func (e *Engine) Feed(ctx context.Context, disabledPlugins []string, sources []string, limit, offset int, refresh bool) ([]Item, error) {
	return e.FeedFiltered(ctx, disabledPlugins, sources, FeedFilter{}, limit, offset, refresh)
}

// FeedFilter narrows a feed beyond its sources; the zero value keeps everything.
type FeedFilter struct {
	// Duration keeps only scenes whose runtime falls inside the window.
	Duration DurationRange
	// ExcludeTagIDs drops scenes carrying any of these tags (own, studio's or
	// a credited performer's).
	ExcludeTagIDs []string
}

// DurationRange narrows a feed to a runtime window in seconds; a zero bound
// is open. Scenes with an unknown runtime never match a set bound.
type DurationRange struct {
	MinSeconds int
	MaxSeconds int
}

// isOpen reports whether the range keeps every scene.
func (window DurationRange) isOpen() bool {
	return window.MinSeconds <= 0 && window.MaxSeconds <= 0
}

// contains reports whether a runtime falls inside the range.
func (window DurationRange) contains(durationSeconds int) bool {
	if window.isOpen() {
		return true
	}
	if durationSeconds <= 0 {
		return false
	}
	if window.MinSeconds > 0 && durationSeconds < window.MinSeconds {
		return false
	}
	return window.MaxSeconds <= 0 || durationSeconds <= window.MaxSeconds
}

// FeedFiltered is Feed additionally narrowed by filter; paging walks the
// narrowed view of the same shared ranking, so pages stay stable under it.
func (e *Engine) FeedFiltered(ctx context.Context, disabledPlugins []string, sources []string, filter FeedFilter, limit, offset int, refresh bool) ([]Item, error) {
	maxAge := feedCacheTTL
	if offset == 0 {
		maxAge = feedFirstPageTTL
	}
	if refresh {
		maxAge = 0
	}
	feed, err := e.rankedFeed(ctx, disabledPlugins, maxAge)
	if err != nil {
		return nil, err
	}
	items, err := e.withoutTags(ctx, feed.within(filter.Duration, feed.observedBy(sources)), filter.ExcludeTagIDs)
	if err != nil {
		return nil, err
	}
	return page(items, limit, offset), nil
}

// within keeps the items whose runtime falls inside the range, in order.
func (feed *rankedFeed) within(window DurationRange, items []Item) []Item {
	if window.isOpen() {
		return items
	}
	out := []Item{}
	for _, item := range items {
		if window.contains(feed.durations[item.SceneID]) {
			out = append(out, item)
		}
	}
	return out
}

// withoutTags drops the items whose scene carries any of tagIDs; items itself
// when there are none to exclude.
func (e *Engine) withoutTags(ctx context.Context, items []Item, tagIDs []string) ([]Item, error) {
	if len(tagIDs) == 0 {
		return items, nil
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.SceneID)
	}
	carrying, err := e.repo.ScenesCarryingTags(ctx, ids, tagIDs)
	if err != nil {
		return nil, err
	}
	kept := make([]Item, 0, len(items))
	for _, item := range items {
		if !carrying[item.SceneID] {
			kept = append(kept, item)
		}
	}
	return kept, nil
}

// observedBy returns the ranked feed narrowed to scenes some of the sources
// plugins observed, keeping the ranked order; the whole feed when sources is
// empty.
func (feed *rankedFeed) observedBy(sources []string) []Item {
	if len(sources) == 0 {
		return feed.ranked
	}
	wanted := toSet(sources)
	out := []Item{}
	for _, item := range feed.ranked {
		if containsAny(feed.observingPlugins[item.SceneID], wanted) {
			out = append(out, item)
		}
	}
	return out
}

// Rows groups the ranked candidates by reason into titled rows ("Because you
// watched X", "More from <performer>", ...), strongest row first.
func (e *Engine) Rows(ctx context.Context, disabledPlugins []string, rowLimit, perRow int) ([]Row, error) {
	feed, err := e.rankedFeed(ctx, disabledPlugins, feedCacheTTL)
	if err != nil {
		return nil, err
	}
	return groupRows(feed.scored, rowLimit, perRow), nil
}

// Pair returns the two best-ranked scenes not compared recently, for the A/B
// choice; empty when there aren't two.
func (e *Engine) Pair(ctx context.Context, disabledPlugins []string) ([]Item, error) {
	feed, err := e.rankedFeed(ctx, disabledPlugins, feedCacheTTL)
	if err != nil {
		return nil, err
	}
	compared, err := e.database.Strings(ctx,
		`SELECT chosen FROM preference_event WHERE created_at >= $since
		 UNION SELECT rejected FROM preference_event WHERE created_at >= $since`,
		db.Vars{"since": db.FormatTime(e.now().Add(-pairCooldown))})
	if err != nil {
		return nil, fmt.Errorf("recently compared: %w", err)
	}
	return pickPair(feed.ranked, toSet(compared)), nil
}

// Categories returns the top-affinity tags, each with its best-scored scenes.
// Empty on a cold start.
func (e *Engine) Categories(ctx context.Context, disabledPlugins []string, categoryLimit, perCategory int) ([]Category, error) {
	state, err := e.takeSnapshot(ctx, disabledPlugins)
	if err != nil {
		return nil, err
	}
	out := []Category{}
	for _, tagID := range TopEntities(state.profile.Tags, categoryLimit) {
		items, err := e.tagCategory(ctx, state, tagID, perCategory)
		if err != nil {
			return nil, err
		}
		if len(items) > 0 {
			out = append(out, Category{TagID: tagID, Items: items})
		}
	}
	return out, nil
}

// TopTags returns the user's highest-affinity tag ids.
func (e *Engine) TopTags(ctx context.Context, limit int) ([]string, error) {
	profile, _, err := e.profileAt(ctx, e.now())
	if err != nil {
		return nil, err
	}
	return TopEntities(profile.Tags, limit), nil
}

// Profile returns the current taste profile.
func (e *Engine) Profile(ctx context.Context) (*Profile, error) {
	profile, _, err := e.profileAt(ctx, e.now())
	return profile, err
}

// rankedFeed returns the memoized ranking, rebuilding it when it is older than
// maxAge or was taken for a different set of inactive plugins.
func (e *Engine) rankedFeed(ctx context.Context, disabledPlugins []string, maxAge time.Duration) (*rankedFeed, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	key := pluginKey(disabledPlugins)
	if e.cache.freshFor(key, e.now(), maxAge) {
		return e.cache, nil
	}
	feed, err := e.buildFeed(ctx, disabledPlugins)
	if err != nil {
		return nil, err
	}
	feed.pluginKey = key
	e.cache = feed
	return feed, nil
}

// freshFor reports whether the cached ranking can serve a request that
// accepts rankings up to maxAge old.
func (feed *rankedFeed) freshFor(key string, now time.Time, maxAge time.Duration) bool {
	return feed != nil && feed.pluginKey == key && now.Sub(feed.builtAt) < maxAge
}

// buildFeed runs the full pipeline: snapshot, candidates, scoring, reasons,
// re-ranking.
func (e *Engine) buildFeed(ctx context.Context, disabledPlugins []string) (*rankedFeed, error) {
	state, err := e.takeSnapshot(ctx, disabledPlugins)
	if err != nil {
		return nil, err
	}
	pool, err := e.generateCandidates(ctx, state)
	if err != nil {
		return nil, err
	}
	scored, err := e.score(ctx, state, pool.list())
	if err != nil {
		return nil, err
	}
	scored, err = e.dropDuplicates(ctx, scored)
	if err != nil {
		return nil, err
	}
	if err := e.describeReasons(ctx, scored); err != nil {
		return nil, err
	}
	observingPlugins, err := e.loadObservingPlugins(ctx, scoredSceneIDs(scored))
	if err != nil {
		return nil, err
	}
	durations, err := e.loadDurations(ctx, scoredSceneIDs(scored))
	if err != nil {
		return nil, err
	}
	return &rankedFeed{builtAt: state.now, scored: scored, ranked: rerank(scored), observingPlugins: observingPlugins, durations: durations}, nil
}

// loadDurations reads the runtime in seconds of each scene that has one.
func (e *Engine) loadDurations(ctx context.Context, sceneIDs []string) (map[string]int, error) {
	out := map[string]int{}
	if len(sceneIDs) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx,
		`SELECT id, duration_seconds FROM scene
		 WHERE duration_seconds IS NOT NULL AND id IN (SELECT value FROM json_each($ids))`,
		db.Vars{"ids": sceneIDs})
	if err != nil {
		return nil, fmt.Errorf("load durations: %w", err)
	}
	for _, row := range rows {
		out[rowString(row, "id")] = rowInt(row, "duration_seconds")
	}
	return out, nil
}

// scoredSceneIDs lists the scene ids of the scored items.
func scoredSceneIDs(scored []scoredItem) []string {
	ids := make([]string, 0, len(scored))
	for _, item := range scored {
		ids = append(ids, item.SceneID)
	}
	return ids
}

// loadObservingPlugins reads every plugin that observed each of the scenes,
// keyed by scene id. Scenes no plugin observed are absent.
func (e *Engine) loadObservingPlugins(ctx context.Context, sceneIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(sceneIDs) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx,
		`SELECT DISTINCT target, plugin FROM observation
		 WHERE target IN (SELECT value FROM json_each($ids))`,
		db.Vars{"ids": sceneIDs})
	if err != nil {
		return nil, fmt.Errorf("load observing plugins: %w", err)
	}
	for _, row := range rows {
		sceneID := rowString(row, "target")
		out[sceneID] = append(out[sceneID], rowString(row, "plugin"))
	}
	return out, nil
}

// score filters the candidates to the eligible ones and scores them, best first.
func (e *Engine) score(ctx context.Context, state *snapshot, candidates []*candidate) ([]scoredItem, error) {
	eligible := state.eligible(candidates)
	ids := make([]string, 0, len(eligible))
	for _, proposed := range eligible {
		ids = append(ids, proposed.sceneID)
	}
	features, err := e.loadFeatures(ctx, ids)
	if err != nil {
		return nil, err
	}
	return scoreCandidates(eligible, features, state), nil
}

// tagCategory scores the newest scenes of one tag and keeps the best.
func (e *Engine) tagCategory(ctx context.Context, state *snapshot, tagID string, limit int) ([]Item, error) {
	ids, err := e.sceneIDs(ctx, entitySceneQueries[ReasonTag], db.Vars{"entity": tagID, "limit": limit * categoryOverfetch})
	if err != nil {
		return nil, err
	}
	prior := sourcePriors[SourceTag] * state.profile.Tags[tagID]
	candidates := sceneCandidates(ids, SourceTag, prior, Reason{Kind: ReasonTag, EntityID: tagID})
	pointers := make([]*candidate, 0, len(candidates))
	for index := range candidates {
		pointers = append(pointers, &candidates[index])
	}
	scored, err := e.score(ctx, state, pointers)
	if err != nil {
		return nil, err
	}
	return itemsOf(scored, limit), nil
}

// describeReasons resolves the entity names behind the reasons and writes
// their text.
func (e *Engine) describeReasons(ctx context.Context, items []scoredItem) error {
	names, err := e.entityNames(ctx, namedEntityIDs(items))
	if err != nil {
		return err
	}
	for index := range items {
		reason := &items[index].Reason
		if name, found := names[reason.EntityID]; found {
			reason.EntityName = name
		}
		reason.describe()
	}
	return nil
}

// namedEntityIDs lists the performer, studio and tag ids the reasons refer to;
// seed credits and subscription queries already carry their names.
func namedEntityIDs(items []scoredItem) []string {
	seen := map[string]bool{}
	for _, item := range items {
		switch item.Reason.Kind {
		case ReasonPerformer, ReasonStudio, ReasonTag, ReasonExplore:
			seen[item.Reason.EntityID] = true
		}
	}
	return setKeys(seen)
}

// entityNames reads the display names of performers, studios and tags by id.
func (e *Engine) entityNames(ctx context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx,
		`SELECT id, name FROM performer WHERE id IN (SELECT value FROM json_each($ids))
		 UNION ALL SELECT id, name FROM studio WHERE id IN (SELECT value FROM json_each($ids))
		 UNION ALL SELECT id, name FROM tag WHERE id IN (SELECT value FROM json_each($ids))`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("entity names: %w", err)
	}
	for _, row := range rows {
		out[rowString(row, "id")] = rowString(row, "name")
	}
	return out, nil
}

// groupRows buckets the best-first items by reason, drops thin and
// newest-only rows, and orders rows by the mean score of their top three.
func groupRows(scored []scoredItem, rowLimit, perRow int) []Row {
	rows := []Row{}
	for _, row := range bucketByReason(scored) {
		if row.Reason.Kind != ReasonNewest && len(row.Items) >= minRowItems {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(left, right int) bool {
		return rowStrength(rows[left]) > rowStrength(rows[right])
	})
	return trimRows(rows, rowLimit, perRow)
}

// bucketByReason groups items by reason key, in order of each group's first item.
func bucketByReason(scored []scoredItem) []Row {
	indexByKey := map[string]int{}
	rows := []Row{}
	for _, item := range scored {
		key := item.Reason.Key()
		index, found := indexByKey[key]
		if !found {
			index = len(rows)
			indexByKey[key] = index
			rows = append(rows, Row{Key: key, Title: item.Reason.Text, Reason: item.Reason})
		}
		rows[index].Items = append(rows[index].Items, item.Item)
	}
	return rows
}

// rowStrength is the mean score of a row's top three items.
func rowStrength(row Row) float64 {
	count := min(3, len(row.Items))
	sum := 0.0
	for _, item := range row.Items[:count] {
		sum += item.Score
	}
	return sum / float64(count)
}

// trimRows caps the number of rows and the items per row.
func trimRows(rows []Row, rowLimit, perRow int) []Row {
	if len(rows) > rowLimit {
		rows = rows[:rowLimit]
	}
	for index := range rows {
		if len(rows[index].Items) > perRow {
			rows[index].Items = rows[index].Items[:perRow]
		}
	}
	return rows
}

// pickPair returns the first two items not in excluded, or nothing.
func pickPair(ranked []Item, excluded map[string]bool) []Item {
	pair := []Item{}
	for _, item := range ranked {
		if excluded[item.SceneID] {
			continue
		}
		pair = append(pair, item)
		if len(pair) == 2 {
			return pair
		}
	}
	return []Item{}
}

// itemsOf strips the features and keeps at most limit items.
func itemsOf(scored []scoredItem, limit int) []Item {
	out := make([]Item, 0, min(limit, len(scored)))
	for _, item := range scored {
		if len(out) >= limit {
			break
		}
		out = append(out, item.Item)
	}
	return out
}

// page slices one page out of the feed.
func page(items []Item, limit, offset int) []Item {
	if offset >= len(items) || limit <= 0 {
		return []Item{}
	}
	end := min(offset+limit, len(items))
	return slices.Clone(items[offset:end])
}

// pluginKey identifies a set of inactive plugins regardless of order.
func pluginKey(disabledPlugins []string) string {
	sorted := slices.Clone(disabledPlugins)
	slices.Sort(sorted)
	return strings.Join(sorted, ",")
}
