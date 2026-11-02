package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/user"
)

// TestMergePerformers merges a performer into another and expects the
// credits, the user's rating and the merged name to land on the survivor.
func TestMergePerformers(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	ada, bea := video.Performers[0], video.Performers[1]
	rating := 7.0
	if _, err := user.New(library.DB(), "").Rate(ctx, schema.RecordKindPerformer, bea, &rating); err != nil {
		t.Fatal(err)
	}
	if err := library.Merge(ctx, schema.RecordKindPerformer, ada, bea); err != nil {
		t.Fatal(err)
	}
	video, _ = library.Video(ctx, ref.Id)
	if len(video.Performers) != 1 || video.Performers[0] != ada {
		t.Errorf("performers = %v, want only the survivor", video.Performers)
	}
	survivor, _ := library.Performer(ctx, ada)
	if len(survivor.Aliases) != 1 || survivor.Aliases[0] != "Bea" {
		t.Errorf("aliases = %v", survivor.Aliases)
	}
	state, _ := user.New(library.DB(), "").State(ctx, schema.RecordKindPerformer, ada)
	if state == nil || state.Rating == nil || *state.Rating != 7 {
		t.Errorf("state = %+v, want the merged rating", state)
	}
	if gone, _ := library.Performer(ctx, bea); gone != nil {
		t.Error("the merged performer still exists")
	}
}

// TestUndoMerge merges a performer into another, renames the survivor, then
// undoes the merge: the merged performer, their credit and rating come back,
// the survivor loses the merged alias but keeps the later rename.
func TestUndoMerge(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	ada, bea := video.Performers[0], video.Performers[1]
	rating := 7.0
	states := user.New(library.DB(), "")
	if _, err := states.Rate(ctx, schema.RecordKindPerformer, bea, &rating); err != nil {
		t.Fatal(err)
	}
	if err := library.Merge(ctx, schema.RecordKindPerformer, ada, bea); err != nil {
		t.Fatal(err)
	}
	if _, err := library.DB().Exec(ctx, `UPDATE "Performer" SET "details" = 'later' WHERE "id" = $id`, db.Vars{"id": ada}); err != nil {
		t.Fatal(err)
	}
	merges, err := library.Merges(ctx, schema.RecordKindPerformer, ada)
	if err != nil || len(merges) != 1 {
		t.Fatalf("merges = %v, %v", merges, err)
	}
	if _, err := library.UndoMerge(ctx, merges[0].Id); err != nil {
		t.Fatal(err)
	}

	video, _ = library.Video(ctx, ref.Id)
	if len(video.Performers) != 2 || video.Performers[1] != bea {
		t.Errorf("performers = %v, want both again", video.Performers)
	}
	back, _ := library.Performer(ctx, bea)
	if back == nil || back.Name != "Bea" {
		t.Fatalf("merged performer = %+v, want Bea back", back)
	}
	survivor, _ := library.Performer(ctx, ada)
	if len(survivor.Aliases) != 0 {
		t.Errorf("survivor aliases = %v, want the merged name gone", survivor.Aliases)
	}
	if survivor.Details == nil || *survivor.Details != "later" {
		t.Errorf("survivor details = %v, want the later edit kept", survivor.Details)
	}
	state, _ := states.State(ctx, schema.RecordKindPerformer, bea)
	if state == nil || state.Rating == nil || *state.Rating != 7 {
		t.Errorf("merged state = %+v, want the rating back", state)
	}
	if _, err := library.UndoMerge(ctx, merges[0].Id); err == nil {
		t.Error("undoing twice succeeded")
	}
}

// TestAlikeCopiesHideAndMerge stores the same scene from two sites, expects
// the newer copy hidden as a pair, refuses to merge it without proof, and
// merges it once both share a fingerprint.
func TestAlikeCopiesHideAndMerge(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	first, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	copyPage := foundScene()
	copyPage[0].Scene.ExternalID = "other-1"
	copyPage[0].Scene.Title = "Pool day HD"
	copyPage[0].Scene.Performers = []plugins.ScenePerformer{{Name: "Ada Lane"}, {Name: "Bea"}}
	second, err := library.IngestFound(ctx, "other", copyPage)
	if err != nil {
		t.Fatal(err)
	}
	assertVideos(t, library, "with a hidden copy", schema.Filter{}, first.Id)
	pairs, err := library.Alikes(ctx, schema.RecordKindVideo, 10, 0)
	if err != nil || len(pairs) != 1 || pairs[0].SecondId != second.Id {
		t.Fatalf("pairs = %+v, err %v", pairs, err)
	}
	for _, id := range []string{first.Id, second.Id} {
		if _, err := library.DB().Insert(ctx, "Fingerprint", db.Vars{"subjectKind": "video", "subjectId": id, "algorithm": "oshash", "value": "abc123"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := library.MergeAlike(ctx, pairs[0].Id); err != nil {
		t.Fatal(err)
	}
	sources, _ := library.Sources(ctx, schema.RecordKindVideo, first.Id)
	if len(sources) != 2 {
		t.Errorf("sources = %d, want both sites on the survivor", len(sources))
	}
	assertVideos(t, library, "after merging", schema.Filter{}, first.Id)
}

// TestUnprovenAlikeStaysApart refuses to merge a pair with no fingerprint
// the strict matcher can't confirm.
func TestUnprovenAlikeStaysApart(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	original := foundScene()
	original[0].Scene.Title = "Pool party sunset lake"
	if _, err := library.IngestFound(ctx, "site", original); err != nil {
		t.Fatal(err)
	}
	copyPage := foundScene()
	copyPage[0].Scene.ExternalID = "other-1"
	copyPage[0].Scene.Title = "Pool party sunset lake ride bikes"
	copyPage[0].Scene.Performers = nil
	if _, err := library.IngestFound(ctx, "other", copyPage); err != nil {
		t.Fatal(err)
	}
	pairs, _ := library.Alikes(ctx, schema.RecordKindVideo, 10, 0)
	if len(pairs) != 1 {
		t.Fatalf("pairs = %+v", pairs)
	}
	if err := library.MergeAlike(ctx, pairs[0].Id); err == nil {
		t.Error("an unproven pair was merged")
	}
	if err := library.DismissAlike(ctx, pairs[0].Id); err != nil {
		t.Fatal(err)
	}
	videos, _ := library.Videos(ctx, MediaQuery{})
	if len(videos) != 2 {
		t.Errorf("videos = %d after dismissing, want both", len(videos))
	}
}
