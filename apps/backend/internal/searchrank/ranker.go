package searchrank

import (
	"math"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// LibrarySource is the Hit.Source of results from the user's own library.
const LibrarySource = "db"

// Hit is one result as the search saw it, before scoring.
type Hit struct {
	// Source is LibrarySource or the plugin that listed the result.
	Source string
	// ID is the canonical record id, empty when the stub could not be stored.
	ID   string
	Item plugins.DiscoveredItem
	// Performers, Studio and Tags are names the listing does not carry but the
	// library knows; the studio falls back to the item's own StudioRef.
	Performers []string
	Studio     string
	Tags       []string
	// Personal is the affinity in [-1, 1] of the user's taste profile for this
	// result; zero without a profile.
	Personal float64
}

// Verdict says what to do with a hit after scoring.
type Verdict struct {
	// Emit is false when a copy of the same thing that scores at least as well
	// has already been sent: there is nothing new to show.
	Emit  bool
	Score float64
	// Replaces is the SourceURL of an earlier-sent copy this hit beats; the
	// client swaps it out. Empty when nothing is replaced.
	Replaces string
}

// Ranker scores the hits of one search as they stream in. Not safe for
// concurrent use: callers feed it from one goroutine.
type Ranker struct {
	queryWords []string
	// rankOffset shifts every site rank when the search pages through the
	// listings: page two starts where page one ended.
	rankOffset int
	// positions counts the hits seen per list (source and media type): the
	// next hit of a list sits one place lower in that list's own order.
	positions map[string]int
	clusters  []*cluster
	byRecord  map[string]*cluster
	byURL     map[string]*cluster
}

// cluster is the copies of one thing seen so far. best is the copy that was
// (or will be) sent; fused adds up every copy's rank share.
type cluster struct {
	best      Hit
	bestOwn   float64
	bestShare float64
	fused     float64
}

// New starts a ranker for a search query. rankOffset is how many results of
// each site list precede this page.
func New(query string, rankOffset int) *Ranker {
	return &Ranker{
		queryWords: QueryWords(query),
		rankOffset: rankOffset,
		positions:  map[string]int{},
		byRecord:   map[string]*cluster{},
		byURL:      map[string]*cluster{},
	}
}

// Add scores a hit, folds it into the copies seen so far and says whether to
// send it. The hit's place in its list is the order of Add calls with the same
// Source and media type, so feed a list's hits in the order the site gave them.
func (r *Ranker) Add(hit Hit) Verdict {
	share := rankShare(r.nextRank(hit))
	own := r.ownScore(hit, share)

	existing := r.findCluster(hit)
	if existing == nil {
		created := &cluster{best: hit, bestOwn: own, bestShare: share, fused: share}
		r.register(created)
		return Verdict{Emit: true, Score: r.clusterScore(created)}
	}
	existing.fused += share
	if own <= existing.bestOwn {
		return Verdict{}
	}
	previous := existing.best
	existing.best, existing.bestOwn, existing.bestShare = hit, own, share
	r.register(existing)
	return Verdict{Emit: true, Score: r.clusterScore(existing), Replaces: previous.Item.SourceURL}
}

// nextRank advances and returns the 1-based place of a hit in its list.
func (r *Ranker) nextRank(hit Hit) int {
	key := hit.Source + "|" + string(hit.Item.MediaType)
	r.positions[key]++
	return r.rankOffset + r.positions[key]
}

// rankShare is a list place's reciprocal-rank-fusion share, normalized so first
// place is 1.
func rankShare(rank int) float64 {
	return (RankFusionConstant + 1) / (RankFusionConstant + float64(rank))
}

// ownScore is the score of a hit counting only its own list place.
func (r *Ranker) ownScore(hit Hit, share float64) float64 {
	score := WeightText * TextMatch(r.queryWords, textFields(hit))
	score += WeightSiteRank * share
	score += WeightQuality * (Quality(hit.Item.Rating, hit.Item.ViewCount) - NeutralQuality)
	score += WeightPersonal * hit.Personal
	return score
}

// clusterScore is the score of the cluster's best copy with the rank shares of
// all the other copies added in: being listed by several sites counts.
func (r *Ranker) clusterScore(found *cluster) float64 {
	return found.bestOwn + WeightSiteRank*(found.fused-found.bestShare)
}

// textFields gathers what the hit's text match reads.
func textFields(hit Hit) TextFields {
	fields := TextFields{Title: hit.Item.Title, Studio: hit.Studio}
	if fields.Studio == "" && hit.Item.Studio != nil {
		fields.Studio = hit.Item.Studio.Name
	}
	for _, name := range hit.Performers {
		fields.Performers += name + " "
	}
	for _, name := range hit.Tags {
		fields.Tags += name + " "
	}
	return fields
}

// Quality rates a listing's own numbers in [0, 1]: the mean of the rating
// (0..10) and the log-scaled view count, whichever the listing shows;
// NeutralQuality when it shows neither.
func Quality(rating float64, viewCount int) float64 {
	var parts []float64
	if rating > 0 {
		parts = append(parts, min(1, rating/10))
	}
	if viewCount > 0 {
		parts = append(parts, min(1, math.Log10(1+float64(viewCount))/math.Log10(1+ViewCountCeiling)))
	}
	if len(parts) == 0 {
		return NeutralQuality
	}
	total := 0.0
	for _, part := range parts {
		total += part
	}
	return total / float64(len(parts))
}

// findCluster returns the cluster the hit is another copy of: the same stored
// record or page, else the same video on another site (alike.SameScene) or the
// same gallery (alike.SameGallery). Nil when it is new.
func (r *Ranker) findCluster(hit Hit) *cluster {
	if hit.ID != "" && r.byRecord[hit.ID] != nil {
		return r.byRecord[hit.ID]
	}
	if hit.Item.SourceURL != "" && r.byURL[hit.Item.SourceURL] != nil {
		return r.byURL[hit.Item.SourceURL]
	}
	for _, candidate := range r.clusters {
		if sameThing(candidate.best, hit) {
			return candidate
		}
	}
	return nil
}

// register indexes a cluster by the record and page of its best copy and, when
// new, adds it to the list compared against later hits.
func (r *Ranker) register(found *cluster) {
	known := false
	for _, existing := range r.clusters {
		if existing == found {
			known = true
			break
		}
	}
	if !known {
		r.clusters = append(r.clusters, found)
	}
	if found.best.ID != "" {
		r.byRecord[found.best.ID] = found
	}
	if found.best.Item.SourceURL != "" {
		r.byURL[found.best.Item.SourceURL] = found
	}
}

// sameThing reports whether two hits are copies of one scene or one gallery.
// Scenes without a known runtime never match (alike.SameScene is strict
// there), nor do hits of different media types.
func sameThing(first, second Hit) bool {
	if first.Item.MediaType != second.Item.MediaType {
		return false
	}
	switch first.Item.MediaType {
	case plugins.MediaTypeScene:
		return alike.SameScene(candidateOf(first), candidateOf(second))
	case plugins.MediaTypeGallery:
		return alike.SameGallery(candidateOf(first), candidateOf(second), first.Item.Count, second.Item.Count)
	}
	return false
}

// candidateOf describes a hit for alike's matchers.
func candidateOf(hit Hit) alike.Candidate {
	return alike.Candidate{
		Title:           hit.Item.Title,
		Performers:      hit.Performers,
		DurationSeconds: hit.Item.DurationSeconds,
	}
}
