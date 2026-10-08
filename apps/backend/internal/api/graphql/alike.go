package graphql

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/subscriptions"
)

const (
	// alikeSearchLimit caps results pulled per query per plugin.
	alikeSearchLimit = 20
	// alikeMaxCandidates bounds how many unique candidates get a poster fetched
	// and scored in stage 1, keeping the on-demand action responsive.
	alikeMaxCandidates = 40
	// alikeScrapeTopK is how many best stage-1 candidates get scraped in stage 2
	// to read their performers and duration.
	alikeScrapeTopK = 8
	// alikeDefaultLimit is the number of ranked matches returned when the caller
	// does not specify a limit.
	alikeDefaultLimit = 12
	// alikeFetchConcurrency bounds parallel poster/scrape work.
	alikeFetchConcurrency = 8
	// alikePosterMaxBytes caps how much of a poster is read before hashing.
	alikePosterMaxBytes = 4 << 20
)

// errNoAlikeSource is returned when a chosen match yields no scrapable scene.
var errNoAlikeSource = errors.New("attachAlikeSource: no scene found at source URL")

// scoredCandidate carries a discovered item alongside its running match score and
// the metadata gathered while ranking it.
type scoredCandidate struct {
	entry           discoveredItemWithPlugin
	posterHash      uint64
	performers      []string
	durationSeconds int
	score           float64
}

// findAlikeSources gathers and ranks alternate copies of a scene across every
// search-capable plugin. Matching is performer/duration/poster driven, not
// title driven, so copies with unrelated titles still surface. See the two-stage
// ranking: cheap poster+title pre-rank, then a scrape of the top few for the
// authoritative performer+duration score.
func (r *Resolver) findAlikeSources(ctx context.Context, sceneID string, limit *int) ([]*model.AlikeCandidate, error) {
	scene, err := r.repo.GetScene(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	if scene == nil {
		return []*model.AlikeCandidate{}, nil
	}

	target := r.buildAlikeTarget(ctx, scene)

	searchPlugins := adultPlugins(r.registry.WithCapability(plugins.CapabilitySceneList))
	if len(searchPlugins) == 0 {
		return []*model.AlikeCandidate{}, nil
	}

	originPlugin, originURL, _ := r.repo.SceneOrigin(ctx, sceneID)
	candidates := r.gatherAlikeCandidates(ctx, searchPlugins, target, scene, originPlugin, originURL)
	if len(candidates) == 0 {
		return []*model.AlikeCandidate{}, nil
	}

	r.rankAlikeStageOne(ctx, target, candidates)
	sortByScoreDesc(candidates)
	if len(candidates) > alikeScrapeTopK {
		candidates = candidates[:alikeScrapeTopK]
	}

	r.rankAlikeStageTwo(ctx, target, candidates)
	sortByScoreDesc(candidates)

	max := alikeDefaultLimit
	if limit != nil && *limit > 0 {
		max = *limit
	}
	return toAlikeCandidates(candidates, max), nil
}

// buildAlikeTarget projects the current scene onto the matcher's target shape,
// including a perceptual hash of its poster.
func (r *Resolver) buildAlikeTarget(ctx context.Context, scene *model.Scene) alike.Target {
	target := alike.Target{
		Title:      scene.Title,
		Performers: performerNames(scene.Performers),
	}
	if scene.Studio != nil {
		target.Studio = scene.Studio.Name
	}
	if scene.DurationSeconds != nil {
		target.DurationSeconds = *scene.DurationSeconds
	}
	if scene.Date != nil {
		target.Date = *scene.Date
	}
	if scene.PosterPath != nil {
		target.PosterHash = r.hashPoster(ctx, *scene.PosterPath)
	}
	return target
}

// gatherAlikeCandidates runs every performer/studio-driven query, unions the
// results, and drops the scene's own copies (same source URL or same host).
func (r *Resolver) gatherAlikeCandidates(
	ctx context.Context,
	searchPlugins []*plugins.Plugin,
	target alike.Target,
	scene *model.Scene,
	originPlugin, originURL string,
) []*scoredCandidate {
	originHost := hostOf(originURL)
	seen := make(map[string]bool)
	candidates := make([]*scoredCandidate, 0, alikeMaxCandidates)

	for _, query := range alike.BuildQueries(target) {
		for _, entry := range r.fanOutSearch(ctx, searchPlugins, query, alikeSearchLimit) {
			if len(candidates) >= alikeMaxCandidates {
				return candidates
			}
			if seen[entry.item.ExternalID] {
				continue
			}
			if isSameSceneOrigin(entry, scene, originPlugin, originHost) {
				continue
			}
			seen[entry.item.ExternalID] = true
			candidates = append(candidates, &scoredCandidate{entry: entry})
		}
	}
	return candidates
}

// isSameSceneOrigin reports whether a candidate is just the scene's own source
// again — the same URL, the same origin plugin, or the same host.
func isSameSceneOrigin(entry discoveredItemWithPlugin, scene *model.Scene, originPlugin, originHost string) bool {
	if entry.item.SourceURL == scene.SourceURL {
		return true
	}
	if originPlugin != "" && entry.plugin == originPlugin {
		return true
	}
	if originHost != "" && hostOf(entry.item.SourceURL) == originHost {
		return true
	}
	return false
}

// rankAlikeStageOne scores candidates on the signals available before scraping:
// poster similarity, a light title bonus, and any listing-provided runtime.
// Poster hashes are fetched concurrently.
func (r *Resolver) rankAlikeStageOne(ctx context.Context, target alike.Target, candidates []*scoredCandidate) {
	r.eachConcurrently(candidates, func(candidate *scoredCandidate) {
		if poster := candidate.entry.item.PosterPath; poster != "" {
			candidate.posterHash = r.hashPoster(ctx, poster)
		}
	})
	for _, candidate := range candidates {
		candidate.durationSeconds = candidate.entry.item.DurationSeconds
		candidate.score = alike.Score(target, alike.Candidate{
			Title:           candidate.entry.item.Title,
			PosterHash:      candidate.posterHash,
			Date:            candidate.entry.item.Date,
			DurationSeconds: candidate.durationSeconds,
		})
	}
}

// rankAlikeStageTwo scrapes the surviving candidates to obtain their performers
// and real duration, then recomputes the full composite score.
func (r *Resolver) rankAlikeStageTwo(ctx context.Context, target alike.Target, candidates []*scoredCandidate) {
	r.eachConcurrently(candidates, func(candidate *scoredCandidate) {
		scene := r.scrapeSceneDetail(ctx, candidate.entry.plugin, candidate.entry.item.SourceURL)
		if scene == nil {
			return
		}
		candidate.performers = scenePerformerNames(scene.Performers)
		if scene.Duration > 0 {
			candidate.durationSeconds = scene.Duration
		}
	})
	for _, candidate := range candidates {
		candidate.score = alike.Score(target, alike.Candidate{
			Performers:      candidate.performers,
			Title:           candidate.entry.item.Title,
			DurationSeconds: candidate.durationSeconds,
			PosterHash:      candidate.posterHash,
			Date:            candidate.entry.item.Date,
		})
	}
}

// scrapeSceneDetail scrapes one candidate URL and returns its scene payload, or
// nil on failure or if the plugin is not registered.
func (r *Resolver) scrapeSceneDetail(ctx context.Context, pluginName, sourceURL string) *plugins.Scene {
	plugin, ok := r.registry.Get(pluginName)
	if !ok {
		return nil
	}
	results, err := r.runner.Find(ctx, plugin, plugins.MediaTypeScene, sourceURL)
	if err != nil {
		log.Printf("findAlikeSources: scrape %s via %s: %v", sourceURL, pluginName, err)
		return nil
	}
	for _, result := range results {
		if result.Scene != nil {
			return result.Scene
		}
	}
	return nil
}

// attachAlikeSource scrapes a chosen match and attaches its playback sources to
// the existing scene (deduped by URL), then returns the scene's ranked streams.
func (r *Resolver) attachAlikeSource(ctx context.Context, sceneID, pluginName, sourceURL string) ([]*model.Stream, error) {
	scene := r.scrapeSceneDetail(ctx, pluginName, sourceURL)
	if scene == nil {
		return nil, errNoAlikeSource
	}
	if len(scene.Downloads) == 0 {
		return r.loadRankedStreams(ctx, sceneID)
	}

	if err := r.ingestSvc.SyncStreams(ctx, sceneID, pluginName, scene.Downloads); err != nil {
		return nil, err
	}
	if r.hub != nil {
		r.hub.Publish(subscriptions.StreamsTopic(sceneID), nil)
	}

	streams, err := r.loadRankedStreams(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	log.Printf("attachAlikeSource %s: attached %d source(s) from %s (%s)", sceneID, len(scene.Downloads), pluginName, sourceURL)
	return streams, nil
}

// hashPoster fetches an image URL and returns its perceptual hash, or 0 on any
// failure (treated as "no poster signal").
func (r *Resolver) hashPoster(ctx context.Context, rawURL string) uint64 {
	if rawURL == "" {
		return 0
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(fetchCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0
	}
	resp, err := outbound.Client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, alikePosterMaxBytes))
	if err != nil {
		return 0
	}
	hash, err := alike.DecodePosterHash(data)
	if err != nil {
		return 0
	}
	return hash
}

// eachConcurrently runs work over candidates with a bounded worker pool.
func (r *Resolver) eachConcurrently(candidates []*scoredCandidate, work func(*scoredCandidate)) {
	semaphore := make(chan struct{}, alikeFetchConcurrency)
	var wg sync.WaitGroup
	for _, candidate := range candidates {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(candidate *scoredCandidate) {
			defer wg.Done()
			defer func() { <-semaphore }()
			work(candidate)
		}(candidate)
	}
	wg.Wait()
}

func sortByScoreDesc(candidates []*scoredCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
}

// toAlikeCandidates maps the top-scoring candidates to the GraphQL model,
// dropping any with no matching signal at all.
func toAlikeCandidates(candidates []*scoredCandidate, max int) []*model.AlikeCandidate {
	out := make([]*model.AlikeCandidate, 0, max)
	for _, candidate := range candidates {
		if candidate.score <= 0 {
			continue
		}
		if len(out) >= max {
			break
		}
		out = append(out, alikeCandidateModel(candidate))
	}
	return out
}

func alikeCandidateModel(candidate *scoredCandidate) *model.AlikeCandidate {
	item := candidate.entry.item
	result := &model.AlikeCandidate{
		ExternalID:    item.ExternalID,
		Title:         item.Title,
		Plugin:        candidate.entry.plugin,
		SourceURL:     item.SourceURL,
		PreviewImages: item.PreviewImages,
		MatchScore:    candidate.score,
	}
	if result.PreviewImages == nil {
		result.PreviewImages = []string{}
	}
	if item.Date != "" {
		date := item.Date
		result.Date = &date
	}
	if item.PosterPath != "" {
		poster := item.PosterPath
		result.PosterURL = &poster
	}
	if item.PreviewVideo != "" {
		preview := item.PreviewVideo
		result.PreviewVideo = &preview
	}
	if candidate.durationSeconds > 0 {
		duration := candidate.durationSeconds
		result.DurationSeconds = &duration
	}
	return result
}

func performerNames(performers []*model.Performer) []string {
	names := make([]string, 0, len(performers))
	for _, performer := range performers {
		if performer != nil && performer.Name != "" {
			names = append(names, performer.Name)
		}
	}
	return names
}

func scenePerformerNames(performers []plugins.ScenePerformer) []string {
	names := make([]string, 0, len(performers))
	for _, performer := range performers {
		if performer.Name != "" {
			names = append(names, performer.Name)
		}
	}
	return names
}

// hostOf returns the lowercased host of a URL without a leading "www.".
func hostOf(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
}
