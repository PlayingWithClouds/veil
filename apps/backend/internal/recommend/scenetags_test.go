package recommend

import (
	"math"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// tag inserts a tag keyed by its folded name, as the library stores them.
func (f *fixture) tag(name string) string {
	return f.insert("Tag", db.Vars{"name": name, "nameKey": names.Key(name), "aliases": []string{}})
}

// aiSegment stores a scene model segment of a video.
func (f *fixture) aiSegment(videoID, tagID string, start, end float64) string {
	return f.insert("Segment", db.Vars{
		"subjectKind": "video", "subjectId": videoID, "startSeconds": start, "endSeconds": end,
		"tag": tagID, "performers": []string{}, "origin": "ai",
	})
}

// TestSceneTagSharesCountOverlapsOnce weighs each tag by the runtime it
// covers and drops slivers.
func TestSceneTagSharesCountOverlapsOnce(t *testing.T) {
	segments := []taggedSegment{
		{span: span{start: 0, end: 300}, tagID: "cowgirl"},
		{span: span{start: 200, end: 500}, tagID: "cowgirl"},
		{span: span{start: 600, end: 610}, tagID: "kissing"},
	}
	shares := sceneTagShares(segments, 1000)
	if math.Abs(shares["cowgirl"]-0.5) > 1e-9 {
		t.Errorf("cowgirl share = %v", shares["cowgirl"])
	}
	if _, kept := shares["kissing"]; kept {
		t.Errorf("a 1%% sliver was kept: %v", shares)
	}
	if fallback := sceneTagShares(segments[:1], 0); fallback["cowgirl"] != 1 {
		t.Errorf("without a runtime the last end stands in: %v", fallback)
	}
}

// TestSceneTagsFeedTheProfileAndMatch learns a scene tag from a watched
// video's ai segments and ranks an unseen video carrying it first.
func TestSceneTagsFeedTheProfileAndMatch(t *testing.T) {
	f := newFixture(t)
	cowgirl := f.tag("Cowgirl")
	watched := f.video(mediaRow{title: "Watched", duration: 600})
	f.aiSegment(watched, cowgirl, 0, 400)
	f.watched(watched, 600, time.Hour)
	matching := f.video(mediaRow{title: "Matching", duration: 600})
	f.aiSegment(matching, cowgirl, 100, 500)
	other := f.video(mediaRow{title: "Other", duration: 600})
	taste, _, _, err := f.engine.profileAt(f.ctx, f.now)
	if err != nil {
		t.Fatal(err)
	}
	if taste.tags[cowgirl] <= 0 {
		t.Fatalf("scene tag affinity = %v", taste.tags[cowgirl])
	}
	features, err := f.engine.loadFeatures(f.ctx, []schema.RecordRef{{Kind: schema.RecordKindVideo, Id: matching}, {Kind: schema.RecordKindVideo, Id: other}})
	if err != nil {
		t.Fatal(err)
	}
	matchingVideo := features[schema.RecordRef{Kind: schema.RecordKindVideo, Id: matching}]
	otherVideo := features[schema.RecordRef{Kind: schema.RecordKindVideo, Id: other}]
	if affinityMatch(matchingVideo, taste) <= affinityMatch(otherVideo, taste) {
		t.Errorf("matching %v not above other %v", affinityMatch(matchingVideo, taste), affinityMatch(otherVideo, taste))
	}
}

// TestCoverTagsLiftWhatIsClicked learns that covers showing one label are
// picked more than covers showing another.
func TestCoverTagsLiftWhatIsClicked(t *testing.T) {
	f := newFixture(t)
	doggy := f.tag("Doggystyle")
	kissing := f.tag("Kissing")
	for index := 0; index < 10; index++ {
		video := f.video(mediaRow{title: "Shown"})
		label := "Kissing"
		if index%2 == 0 {
			label = "Doggystyle"
		}
		data := map[string]any{"surface": "home", "position": index, "coverTags": map[string]any{label: 0.9}}
		f.event(schema.EventKindImpression, schema.RecordKindVideo, video, time.Hour, data)
		if label == "Doggystyle" {
			f.event(schema.EventKindChoice, schema.RecordKindVideo, video, time.Hour, data)
		}
	}
	lifts, err := f.engine.loadCoverTagLifts(f.ctx, f.now)
	if err != nil {
		t.Fatal(err)
	}
	if lifts[doggy] <= 0 || lifts[kissing] >= 0 {
		t.Errorf("lifts = doggy %v, kissing %v", lifts[doggy], lifts[kissing])
	}
	taste, _, _, err := f.engine.profileAt(f.ctx, f.now)
	if err != nil {
		t.Fatal(err)
	}
	if taste.tags[doggy] <= taste.tags[kissing] {
		t.Errorf("taste: doggy %v, kissing %v", taste.tags[doggy], taste.tags[kissing])
	}
}

// TestShortsTakeOnlyLikedAISegments plays the model's segments whose tag the
// user liked, and none of the others.
func TestShortsTakeOnlyLikedAISegments(t *testing.T) {
	f := newFixture(t)
	liked := f.tag("Reverse Cowgirl")
	unliked := f.tag("Kissing")
	video := f.video(mediaRow{title: "Scanned", duration: 1200})
	likedSegment := f.aiSegment(video, liked, 300, 420)
	unlikedSegment := f.aiSegment(video, unliked, 30, 90)
	shorts, err := f.engine.Shorts(f.ctx, schema.Filter{}, 20)
	if err != nil || len(shorts) != 0 {
		t.Fatalf("shorts without liked tags = %+v, err %v", shorts, err)
	}
	f.event(schema.EventKindLike, schema.RecordKindTag, liked, time.Hour, map[string]any{"value": 1})
	shorts, err = f.engine.Shorts(f.ctx, schema.Filter{}, 20)
	if err != nil {
		t.Fatal(err)
	}
	byItem := map[schema.RecordRef]Short{}
	for _, short := range shorts {
		byItem[short.Match.Item] = short
	}
	expectShort(t, byItem, schema.RecordRef{Kind: schema.RecordKindSegment, Id: likedSegment}, 300, 300+shortMaxSeconds)
	if _, found := byItem[schema.RecordRef{Kind: schema.RecordKindSegment, Id: unlikedSegment}]; found {
		t.Error("an unliked ai segment became a short")
	}
	detail := evidenceOf(byItem[schema.RecordRef{Kind: schema.RecordKindSegment, Id: likedSegment}].Match)
	if len(detail) == 0 || detail[0] != "Reverse Cowgirl at 5:00" {
		t.Errorf("detail = %v", detail)
	}
}
