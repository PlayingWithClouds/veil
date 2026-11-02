package recommend

import (
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// replay logs the user playing a stretch of a video several times.
func (f *fixture) replay(videoID string, start, end float64, times int) {
	for index := 0; index < times; index++ {
		f.event(schema.EventKindHeatSpan, schema.RecordKindVideo, videoID, time.Duration(index+1)*time.Hour, map[string]any{"start": start, "end": end})
	}
}

// TestShortsMixesEveryKind returns a saved clip, a segment, a most replayed
// moment and a short vertical video, each with its stretch, and honours the
// filter.
func TestShortsMixesEveryKind(t *testing.T) {
	f := newFixture(t)
	studio := f.named("Studio", "Filtered Out")
	clipVideo := f.video(mediaRow{title: "Clip source", duration: 1200})
	clipID := f.insert("Clip", db.Vars{"video": clipVideo, "startSeconds": 100.0, "endSeconds": 130.0, "title": "Best bit"})
	segmentVideo := f.video(mediaRow{title: "Chaptered", duration: 1800})
	segmentID := f.insert("Segment", db.Vars{"subjectKind": "video", "subjectId": segmentVideo, "startSeconds": 300.0, "endSeconds": 600.0, "title": "Second scene", "performers": []string{}, "origin": "plugin"})
	replayed := f.video(mediaRow{title: "Replayed", duration: 1200})
	f.replay(replayed, 0, 1200, 1)
	f.replay(replayed, 750, 780, 3)
	vertical := f.video(mediaRow{title: "Upright", duration: 45, width: 1080, height: 1920})
	excluded := f.video(mediaRow{title: "Excluded upright", studio: studio, duration: 30, width: 720, height: 1280})
	shorts, err := f.engine.Shorts(f.ctx, schema.Filter{ExcludeTagIds: []string{"none"}}, 20)
	if err != nil {
		t.Fatal(err)
	}
	byItem := map[schema.RecordRef]Short{}
	for _, short := range shorts {
		byItem[short.Match.Item] = short
	}
	expectShort(t, byItem, schema.RecordRef{Kind: schema.RecordKindClip, Id: clipID}, 100, 130)
	expectShort(t, byItem, schema.RecordRef{Kind: schema.RecordKindSegment, Id: segmentID}, 300, 300+shortMaxSeconds)
	expectShort(t, byItem, schema.RecordRef{Kind: schema.RecordKindVideo, Id: replayed}, 750, 780)
	expectShort(t, byItem, schema.RecordRef{Kind: schema.RecordKindVideo, Id: vertical}, 0, 45)
	if detail := evidenceOf(byItem[schema.RecordRef{Kind: schema.RecordKindVideo, Id: replayed}].Match); len(detail) == 0 || detail[0] != "Most replayed at 12:30" {
		t.Errorf("moment detail = %v", detail)
	}
	filtered, err := f.engine.Shorts(f.ctx, schema.Filter{StudioIds: []string{studio}}, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].Match.Item.Id != excluded {
		t.Errorf("filtered shorts = %+v", filtered)
	}
	galleriesOnly, err := f.engine.Shorts(f.ctx, schema.Filter{Kinds: []schema.MediaKind{schema.MediaKindGallery}}, 20)
	if err != nil || len(galleriesOnly) != 0 {
		t.Errorf("gallery shorts = %+v, err %v", galleriesOnly, err)
	}
}

// expectShort checks that a short is there and plays the expected stretch.
func expectShort(t *testing.T, byItem map[schema.RecordRef]Short, item schema.RecordRef, start, end float64) {
	t.Helper()
	short, found := byItem[item]
	if !found {
		t.Errorf("%s %s missing", item.Kind, item.Id)
		return
	}
	if short.StartSeconds == nil || short.EndSeconds == nil || *short.StartSeconds != start || *short.EndSeconds != end {
		t.Errorf("%s plays %v..%v, want %v..%v", item.Kind, *short.StartSeconds, *short.EndSeconds, start, end)
	}
}

// TestBestMomentPrefersReplays finds the most replayed point, falls back to a
// tagged segment, and reports nothing for a video without either.
func TestBestMomentPrefersReplays(t *testing.T) {
	f := newFixture(t)
	replayed := f.video(mediaRow{title: "Replayed", duration: 1200})
	f.replay(replayed, 0, 1200, 1)
	f.replay(replayed, 600, 640, 4)
	seconds, ok, err := f.engine.BestMoment(f.ctx, replayed)
	if err != nil || !ok || seconds != 600 {
		t.Errorf("replayed: %v %v %v", seconds, ok, err)
	}
	tag := f.named("Tag", "highlight")
	segmented := f.video(mediaRow{title: "Segmented", duration: 1200})
	f.insert("Segment", db.Vars{"subjectKind": "video", "subjectId": segmented, "startSeconds": 10.0, "performers": []string{}, "origin": "plugin"})
	f.insert("Segment", db.Vars{"subjectKind": "video", "subjectId": segmented, "startSeconds": 420.0, "tag": tag, "performers": []string{}, "origin": "plugin"})
	seconds, ok, err = f.engine.BestMoment(f.ctx, segmented)
	if err != nil || !ok || seconds != 420 {
		t.Errorf("segmented: %v %v %v", seconds, ok, err)
	}
	plain := f.video(mediaRow{title: "Plain", duration: 1200})
	f.replay(plain, 0, 1200, 1)
	if _, ok, err := f.engine.BestMoment(f.ctx, plain); ok || err != nil {
		t.Errorf("plain: %v %v", ok, err)
	}
}

// TestHeatPeakNeedsReplays finds no peak in a single full watch and finds
// the stretch an o event marks.
func TestHeatPeakNeedsReplays(t *testing.T) {
	once := addHeat(nil, weightedSpan{span: span{start: 0, end: 600}, weight: 1})
	if _, found := heatPeak(once); found {
		t.Error("peak in a single watch")
	}
	withO := addHeat(once, weightedSpan{span: span{start: 270, end: 300}, weight: heatOWeight})
	peak, found := heatPeak(withO)
	if !found || peak.start != 270 || peak.end != 300 {
		t.Errorf("peak = %+v, found %v", peak, found)
	}
}
