package user

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestBuiltInCollectionsAndOrder ensures the watchlist and queue, then adds,
// moves and removes queue items.
func TestBuiltInCollectionsAndOrder(t *testing.T) {
	ctx := context.Background()
	service, first := openTestService(t)
	second, err := service.database.Insert(ctx, "Video", db.Vars{"title": "Two", "status": "available", "performers": []string{}, "tags": []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.EnsureBuiltInCollections(ctx); err != nil {
		t.Fatal(err)
	}
	if err := service.EnsureBuiltInCollections(ctx); err != nil {
		t.Fatal(err)
	}
	collections, _ := service.Collections(ctx)
	if len(collections) != 2 {
		t.Fatalf("collections = %+v", collections)
	}
	queue, err := service.BuiltIn(ctx, schema.CollectionKindQueue)
	if err != nil {
		t.Fatal(err)
	}
	refOf := func(id string) schema.RecordRef { return schema.RecordRef{Kind: schema.RecordKindVideo, Id: id} }
	if err := service.AddItems(ctx, queue.Id, []schema.RecordRef{refOf(first)}, nil); err != nil {
		t.Fatal(err)
	}
	front := 0
	if err := service.AddItems(ctx, queue.Id, []schema.RecordRef{refOf(second)}, &front); err != nil {
		t.Fatal(err)
	}
	items, _ := service.Items(ctx, queue.Id, 10, 0)
	if len(items) != 2 || items[0].SubjectId != second {
		t.Fatalf("queue = %+v, want the play-next video first", items)
	}
	if err := service.AddItems(ctx, queue.Id, []schema.RecordRef{refOf(second)}, nil); err != nil {
		t.Fatal(err)
	}
	items, _ = service.Items(ctx, queue.Id, 10, 0)
	if len(items) != 2 || items[1].SubjectId != second {
		t.Fatalf("queue = %+v, want the re-added video moved to the end", items)
	}
	if err := service.RemoveItems(ctx, queue.Id, []schema.RecordRef{refOf(first)}); err != nil {
		t.Fatal(err)
	}
	if count, _ := service.ItemCount(ctx, queue.Id); count != 1 {
		t.Errorf("count = %d after removing one", count)
	}
	if err := service.RequireUserCollection(ctx, queue.Id); err == nil {
		t.Error("the queue may be deleted")
	}
}

// TestCustomFieldValues sets values of a choice and a number field and
// rejects ones that don't fit.
func TestCustomFieldValues(t *testing.T) {
	ctx := context.Background()
	service, videoID := openTestService(t)
	mood, err := service.SaveFieldDef(ctx, "", schema.FieldDef{Name: "Mood", Kinds: []schema.RecordKind{schema.RecordKindVideo}, FieldType: schema.FieldTypeChoice, Options: []string{"soft", "rough"}})
	if err != nil {
		t.Fatal(err)
	}
	score, err := service.SaveFieldDef(ctx, "", schema.FieldDef{Name: "Score", Kinds: []schema.RecordKind{schema.RecordKindVideo}, FieldType: schema.FieldTypeNumber})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetFieldValue(ctx, schema.RecordKindVideo, videoID, mood.Id, "soft"); err != nil {
		t.Fatal(err)
	}
	state, err := service.SetFieldValue(ctx, schema.RecordKindVideo, videoID, score.Id, 7.5)
	if err != nil {
		t.Fatal(err)
	}
	if state.Fields[mood.Id] != "soft" || state.Fields[score.Id] != 7.5 {
		t.Errorf("fields = %+v", state.Fields)
	}
	if _, err := service.SetFieldValue(ctx, schema.RecordKindVideo, videoID, mood.Id, "angry"); err == nil {
		t.Error("a value outside the options was accepted")
	}
	state, _ = service.SetFieldValue(ctx, schema.RecordKindVideo, videoID, mood.Id, nil)
	if _, set := state.Fields[mood.Id]; set {
		t.Errorf("fields = %+v after clearing", state.Fields)
	}
}
