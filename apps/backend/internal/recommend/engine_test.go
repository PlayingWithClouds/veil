package recommend

import (
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestColdStartServesNewest recommends the newest items, videos and
// galleries, with a score in 0..1 and a reason, before any signal exists.
func TestColdStartServesNewest(t *testing.T) {
	f := newFixture(t)
	videoID := f.video(mediaRow{title: "First video"})
	galleryID := f.galleryRow(mediaRow{title: "First set", images: 12})
	matches := f.recommendations(10)
	if positionOf(matches, videoID) < 0 || positionOf(matches, galleryID) < 0 {
		t.Fatalf("matches = %+v", matches)
	}
	for _, match := range matches {
		if match.Score == nil || *match.Score < 0 || *match.Score > 1 {
			t.Errorf("score of %s = %v", match.Item.Id, match.Score)
		}
		if !slices.Contains(evidenceOf(match), "Newly added") || match.Evidence[0].Plugin != "veil" {
			t.Errorf("evidence of %s = %v", match.Item.Id, evidenceOf(match))
		}
	}
}

// TestBlockedAndTrashedNeverRecommended leaves out blocked media, media
// crediting a blocked performer, studio or tag, and trashed media, also when
// blocked after the ranking was built.
func TestBlockedAndTrashedNeverRecommended(t *testing.T) {
	f := newFixture(t)
	performer := f.named("Performer", "Blocked Performer")
	studio := f.named("Studio", "Blocked Studio")
	tag := f.named("Tag", "blocked-tag")
	kept := f.video(mediaRow{title: "Kept"})
	blocked := []string{
		f.video(mediaRow{title: "Blocked itself"}),
		f.video(mediaRow{title: "With blocked performer", performers: []string{performer}}),
		f.video(mediaRow{title: "With blocked studio", studio: studio}),
		f.galleryRow(mediaRow{title: "With blocked tag", tags: []string{tag}, images: 3}),
	}
	trashed := f.video(mediaRow{title: "Trashed"})
	if err := f.database.Trash(f.ctx, "Video", trashed); err != nil {
		t.Fatal(err)
	}
	f.state(schema.RecordKindVideo, blocked[0], db.Vars{"blocked": true})
	f.state(schema.RecordKindPerformer, performer, db.Vars{"blocked": true})
	f.state(schema.RecordKindStudio, studio, db.Vars{"blocked": true})
	f.state(schema.RecordKindTag, tag, db.Vars{"blocked": true})
	matches := f.recommendations(50)
	if positionOf(matches, kept) < 0 {
		t.Fatalf("kept video missing: %+v", matches)
	}
	for _, id := range append(blocked, trashed) {
		if positionOf(matches, id) >= 0 {
			t.Errorf("%s was recommended", id)
		}
	}
	f.state(schema.RecordKindVideo, kept, db.Vars{"blocked": true})
	if positionOf(f.recommendations(50), kept) >= 0 {
		t.Error("video blocked after ranking was still recommended")
	}
}

// TestLikedPerformerRanksFirst ranks an unwatched video of the performer the
// user watched to the end of a session above videos of strangers, and says
// why.
func TestLikedPerformerRanksFirst(t *testing.T) {
	f := newFixture(t)
	favourite := f.named("Performer", "Favourite One")
	stranger := f.named("Performer", "Someone Else")
	seen := f.video(mediaRow{title: "Seen scene", performers: []string{favourite}, duration: 1200})
	f.watched(seen, 1200, 2*time.Hour)
	f.event(schema.EventKindO, schema.RecordKindVideo, seen, 2*time.Hour, map[string]any{"position": 900})
	candidate := f.video(mediaRow{title: "New with favourite", performers: []string{favourite}, duration: 1200})
	for index := 0; index < 5; index++ {
		f.video(mediaRow{title: "Other " + string(rune('A'+index)), performers: []string{stranger}, duration: 1200})
	}
	matches := f.recommendations(20)
	position := positionOf(matches, candidate)
	if position != 0 {
		t.Fatalf("favourite's video at %d: %+v", position, matches)
	}
	if detail := evidenceOf(matches[0]); len(detail) == 0 || detail[len(detail)-1] == "Newly added" {
		t.Errorf("reason = %v", detail)
	}
	if positionOf(matches, seen) >= 0 && positionOf(matches, seen) < position {
		t.Error("the watched video outranks the unwatched one")
	}
}

// TestDislikedNeverRecommended leaves out a thumbs-down video and one rated
// low.
func TestDislikedNeverRecommended(t *testing.T) {
	f := newFixture(t)
	thumbsDown := f.video(mediaRow{title: "Thumbs down"})
	ratedLow := f.video(mediaRow{title: "Rated low"})
	f.event(schema.EventKindLike, schema.RecordKindVideo, thumbsDown, time.Hour, map[string]any{"value": -1})
	f.state(schema.RecordKindVideo, ratedLow, db.Vars{"rating": 1.0})
	matches := f.recommendations(20)
	if positionOf(matches, thumbsDown) >= 0 || positionOf(matches, ratedLow) >= 0 {
		t.Errorf("disliked recommended: %+v", matches)
	}
}

// TestCopiesAcrossSitesShownOnce keeps one of two copies of the same scene.
func TestCopiesAcrossSitesShownOnce(t *testing.T) {
	f := newFixture(t)
	performer := f.named("Performer", "Copy Star")
	first := f.video(mediaRow{title: "Poolside afternoon with Copy Star", performers: []string{performer}, duration: 1800})
	second := f.video(mediaRow{title: "Poolside afternoon with Copy Star", performers: []string{performer}, duration: 1801})
	other := f.video(mediaRow{title: "Something unrelated", duration: 600})
	matches := f.recommendations(20)
	copies := 0
	for _, id := range []string{first, second} {
		if positionOf(matches, id) >= 0 {
			copies++
		}
	}
	if copies != 1 || positionOf(matches, other) < 0 {
		t.Errorf("copies shown = %d, matches %+v", copies, matches)
	}
}

// TestPagesDoNotOverlap pages through one ranking.
func TestPagesDoNotOverlap(t *testing.T) {
	f := newFixture(t)
	for index := 0; index < 30; index++ {
		f.video(mediaRow{title: "Video " + string(rune('a'+index%26)) + string(rune('a'+index/26))})
	}
	first := f.recommendations(10)
	second, err := f.engine.Recommendations(f.ctx, schema.Filter{}, 10, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range second {
		if positionOf(first, match.Item.Id) >= 0 {
			t.Errorf("%s on both pages", match.Item.Id)
		}
	}
	if len(first) != 10 || len(second) != 10 {
		t.Errorf("page sizes %d, %d", len(first), len(second))
	}
}

// TestRecentSearchProposesMatches recommends what a recent search names, and
// a video that looks like a watched one, with their reasons.
func TestRecentSearchProposesMatches(t *testing.T) {
	f := newFixture(t)
	searched := f.video(mediaRow{title: "Rooftop sunset", duration: 600})
	seen := f.video(mediaRow{title: "Watched one", duration: 600})
	lookalike := f.video(mediaRow{title: "Visually close", duration: 600})
	for index := 0; index < 10; index++ {
		f.video(mediaRow{title: "Filler " + string(rune('a'+index)), duration: 600})
	}
	f.insert("Event", db.Vars{"at": db.FormatTime(f.now.Add(-time.Hour)), "kind": "search", "data": map[string]any{"text": "rooftop"}})
	f.watched(seen, 600, 2*time.Hour)
	f.insert("Embedding", db.Vars{"subjectKind": "video", "subjectId": seen, "kind": "visual", "model": "clip-test", "vector": encodeVector(1, 0)})
	f.insert("Embedding", db.Vars{"subjectKind": "video", "subjectId": lookalike, "kind": "visual", "model": "clip-test", "vector": encodeVector(0.95, 0.05)})
	matches := f.recommendations(20)
	if !hasEvidence(matches, searched, "From your search “rooftop”") {
		t.Errorf("searched video = %+v", matches)
	}
	if !hasEvidence(matches, lookalike, "Because you watched Watched one") {
		t.Errorf("lookalike video = %+v", matches)
	}
}

// TestTasteSummarizesSignals puts the liked performer and tag on top and the
// disliked studio at the bottom.
func TestTasteSummarizesSignals(t *testing.T) {
	f := newFixture(t)
	performer := f.named("Performer", "Liked Performer")
	tag := f.named("Tag", "liked-tag")
	studio := f.named("Studio", "Disliked Studio")
	liked := f.video(mediaRow{title: "Liked", performers: []string{performer}, tags: []string{tag}, duration: 900})
	f.event(schema.EventKindLike, schema.RecordKindVideo, liked, time.Hour, map[string]any{"value": 1})
	f.event(schema.EventKindLike, schema.RecordKindStudio, studio, time.Hour, map[string]any{"value": -1})
	taste, err := f.engine.Taste(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if taste.SignalCount == 0 {
		t.Fatal("no signals counted")
	}
	if len(taste.Performers) == 0 || taste.Performers[0].Id != performer || taste.Performers[0].Name != "Liked Performer" || taste.Performers[0].Weight != 1 {
		t.Errorf("performers = %+v", taste.Performers)
	}
	if len(taste.Tags) == 0 || taste.Tags[0].Id != tag {
		t.Errorf("tags = %+v", taste.Tags)
	}
	if len(taste.Studios) == 0 || taste.Studios[0].Id != studio || taste.Studios[0].Weight >= 0 {
		t.Errorf("studios = %+v", taste.Studios)
	}
}

// TestSessionEndOutweighsFollowedWatch classifies the last watch of a
// session as its strongest signal and one another video followed as weak.
func TestSessionEndOutweighsFollowedWatch(t *testing.T) {
	now := time.Now()
	followed := watch{videoID: "a", startedAt: now.Add(-3 * time.Hour), lastAt: now.Add(-170 * time.Minute), playedSeconds: 200, durationSeconds: 1200}
	last := watch{videoID: "b", startedAt: now.Add(-165 * time.Minute), lastAt: now.Add(-2 * time.Hour), playedSeconds: 400, durationSeconds: 1200}
	flags := followedWatches([]watch{followed, last})
	if !flags[0] || flags[1] {
		t.Fatalf("followed flags = %v", flags)
	}
	first, _ := watchWeight(followed, flags[0], now.Sub(followed.lastAt))
	second, _ := watchWeight(last, flags[1], now.Sub(last.lastAt))
	if first != weightPartialWatch || second != weightSessionEnd {
		t.Errorf("weights %v, %v", first, second)
	}
	abandoned := watch{videoID: "c", startedAt: now.Add(-time.Hour), lastAt: now.Add(-time.Hour), playedSeconds: 5, durationSeconds: 1200}
	if weight, counts := watchWeight(abandoned, false, time.Hour); !counts || weight != weightEarlyAbandon {
		t.Errorf("abandon weight %v counts %v", weight, counts)
	}
}

// TestRerankCapsPerformers keeps a page from filling with one performer.
func TestRerankCapsPerformers(t *testing.T) {
	scored := []scoredItem{}
	for index := 0; index < 10; index++ {
		performer := "star"
		if index >= 6 {
			performer = "other" + string(rune('a'+index))
		}
		scored = append(scored, scoredItem{
			candidate: &candidate{item: schema.RecordRef{Kind: schema.RecordKindVideo, Id: string(rune('a' + index))}, source: sourcePerformer},
			features:  &mediaFeatures{performers: []string{performer}},
			score:     float64(10 - index),
		})
	}
	ranked := rerank(scored, kindQuotas)
	starsInFirstFour := 0
	for _, item := range ranked[:4] {
		if item.features.performers[0] == "star" {
			starsInFirstFour++
		}
	}
	if starsInFirstFour > maxPerPerformerPerPage || len(ranked) != len(scored) {
		t.Errorf("stars in first four %d, ranked %d", starsInFirstFour, len(ranked))
	}
}

// TestLearnFromNeedsEvidence keeps the hand-set values on thin data and keeps
// learned factors and quotas within their bounds.
func TestLearnFromNeedsEvidence(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	if policy := learnFrom(map[string]outcomeCounts{sourceTag: {shows: 10, reward: 5}}, random); policy != nil {
		t.Errorf("learned from 10 showings: %+v", policy)
	}
	policy := learnFrom(map[string]outcomeCounts{
		sourceTag:    {shows: 2000, reward: 200},
		sourceRandom: {shows: 2000, reward: 2},
	}, random)
	if policy == nil {
		t.Fatal("nothing learned")
	}
	if policy.multiplier(sourceTag) <= policy.multiplier(sourceRandom) {
		t.Errorf("tag x%v, random x%v", policy.multiplier(sourceTag), policy.multiplier(sourceRandom))
	}
	for kind, quota := range policy.quotas {
		if quota < kindQuotaMinimums[kind] || float64(quota) > float64(kindQuotas[kind])*learnQuotaMaxFactor {
			t.Errorf("quota %s = %d", kind, quota)
		}
	}
}

// TestOutcomesCreditServedSource credits an opened impression to the source
// its item was served from.
func TestOutcomesCreditServedSource(t *testing.T) {
	f := newFixture(t)
	videoID := f.video(mediaRow{title: "Served"})
	matches := f.recommendations(10)
	if positionOf(matches, videoID) < 0 {
		t.Fatal("video not served")
	}
	f.now = time.Now()
	f.event(schema.EventKindImpression, schema.RecordKindVideo, videoID, 10*time.Hour, map[string]any{"surface": "home", "position": 0})
	f.event(schema.EventKindChoice, schema.RecordKindVideo, videoID, 9*time.Hour, map[string]any{"surface": "home", "position": 0})
	outcomes, err := f.engine.sourceOutcomes(f.ctx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	counts := outcomes[sourceNewest]
	if counts.shows <= 0 || counts.reward <= 0 {
		t.Errorf("outcomes = %+v", outcomes)
	}
}

// TestRecommendationsFollowFilter keeps only the recommended videos a filter
// matches, here the long ones.
func TestRecommendationsFollowFilter(t *testing.T) {
	f := newFixture(t)
	long := f.video(mediaRow{title: "Long one", duration: 3000})
	f.video(mediaRow{title: "Short one", duration: 300})

	minimum := int32(1800)
	matches, err := f.engine.Recommendations(f.ctx, schema.Filter{MinDurationSeconds: &minimum}, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].Item.Id != long {
		t.Errorf("matches = %+v, want only %s", matches, long)
	}
}
