package recommend

import (
	"slices"
	"strings"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestRelatedFindsCompanionsCreditsAndLookalikes relates a video to its
// linked gallery, the gallery likely shot with it, items sharing its
// performer and studio, and an item whose visual embedding is close, and
// leaves out unrelated and blocked items.
func TestRelatedFindsCompanionsCreditsAndLookalikes(t *testing.T) {
	f := newFixture(t)
	performer := f.named("Performer", "Shared Star")
	studio := f.named("Studio", "Shared Studio")
	blockedPerformer := f.named("Performer", "Blocked Star")
	f.state(schema.RecordKindPerformer, blockedPerformer, db.Vars{"blocked": true})
	target := f.video(mediaRow{title: "Beach shoot with Shared Star", performers: []string{performer}, studio: studio, duration: 1500, date: "2026-05-01"})
	linked := f.galleryRow(mediaRow{title: "Unrelated title", video: target, images: 40})
	likely := f.galleryRow(mediaRow{title: "Beach shoot with Shared Star photos", performers: []string{performer}, date: "2026-05-02", images: 60})
	samePerformer := f.video(mediaRow{title: "Another day", performers: []string{performer}, duration: 900})
	sameStudio := f.video(mediaRow{title: "Studio work", studio: studio, duration: 700})
	lookalike := f.video(mediaRow{title: "Looks alike", duration: 800})
	unrelated := f.video(mediaRow{title: "Nothing in common", duration: 800})
	blocked := f.video(mediaRow{title: "Blocked co-star", performers: []string{performer, blockedPerformer}, duration: 800})
	vectors := map[string][]byte{target: encodeVector(1, 0, 0), lookalike: encodeVector(0.9, 0.1, 0), unrelated: encodeVector(0, 0, 1)}
	for videoID, vector := range vectors {
		f.insert("Embedding", db.Vars{"subjectKind": "video", "subjectId": videoID, "kind": "visual", "model": "clip-test", "vector": vector})
	}
	matches, err := f.engine.Related(f.ctx, schema.RecordRef{Kind: schema.RecordKindVideo, Id: target}, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{linked, likely, samePerformer, sameStudio, lookalike} {
		if positionOf(matches, id) < 0 {
			t.Errorf("%s missing from %+v", id, matches)
		}
	}
	for _, id := range []string{target, unrelated, blocked} {
		if positionOf(matches, id) >= 0 {
			t.Errorf("%s is related", id)
		}
	}
	if positionOf(matches, linked) != 0 {
		t.Errorf("linked gallery at %d", positionOf(matches, linked))
	}
	if !hasEvidence(matches, samePerformer, "Same performer: Shared Star") || !hasEvidence(matches, lookalike, "Looks similar") {
		t.Errorf("evidence = %+v", matches)
	}
}

// hasEvidence reports whether the match of id carries an evidence detail
// starting with prefix.
func hasEvidence(matches []schema.Match, id, prefix string) bool {
	position := positionOf(matches, id)
	if position < 0 {
		return false
	}
	return slices.ContainsFunc(evidenceOf(matches[position]), func(detail string) bool { return strings.HasPrefix(detail, prefix) })
}

// TestRelatedFromGalleryFindsItsVideo links a gallery back to the video it
// points at.
func TestRelatedFromGalleryFindsItsVideo(t *testing.T) {
	f := newFixture(t)
	videoID := f.video(mediaRow{title: "The video", duration: 600})
	galleryID := f.galleryRow(mediaRow{title: "The set", video: videoID, images: 20})
	matches, err := f.engine.Related(f.ctx, schema.RecordRef{Kind: schema.RecordKindGallery, Id: galleryID}, 5)
	if err != nil {
		t.Fatal(err)
	}
	if positionOf(matches, videoID) != 0 || !hasEvidence(matches, videoID, "Video of this set") {
		t.Errorf("matches = %+v", matches)
	}
}

// TestDecodeVectorNormalizes reads little-endian float32s as a unit vector
// and rejects malformed blobs.
func TestDecodeVectorNormalizes(t *testing.T) {
	vector := decodeVector(encodeVector(3, 4))
	if len(vector) != 2 || cosine(vector, vector) < 0.999 || vector[0] < 0.59 || vector[0] > 0.61 {
		t.Errorf("vector = %v", vector)
	}
	if decodeVector([]byte{1, 2, 3}) != nil || decodeVector(encodeVector(0, 0)) != nil {
		t.Error("malformed vectors decoded")
	}
}
