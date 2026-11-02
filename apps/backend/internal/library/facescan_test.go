package library

import (
	"context"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestPurgeDropsFaceScans purges a scanned video and expects its scan, its
// faces and the performer's face learned from it gone, foreign keys intact.
func TestPurgeDropsFaceScans(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	performerID := video.Performers[0]
	scanID, err := library.database.Insert(ctx, "FaceScan", schema.FaceScan{
		SubjectKind: schema.RecordKindVideo, SubjectId: ref.Id, Model: "test", Frames: 1, Faces: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, subject := range []schema.RecordRef{ref, {Kind: schema.RecordKindPerformer, Id: performerID}} {
		row := db.Vars{"subjectKind": subject.Kind, "subjectId": subject.Id, "kind": schema.EmbeddingKindFace, "model": "test", "vector": []byte{0, 0, 128, 63}, "scan": scanID}
		if _, err := library.database.Insert(ctx, "Embedding", row); err != nil {
			t.Fatal(err)
		}
	}
	if err := library.Trash(ctx, schema.RecordKindVideo, ref.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := library.PurgeTrash(ctx, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"FaceScan", "Embedding"} {
		row, err := library.database.QueryRow(ctx, `SELECT count(*) AS "count" FROM `+db.Quote(table), nil)
		if err != nil || row["count"] != int64(0) {
			t.Errorf("%s rows = %v, err %v", table, row, err)
		}
	}
}
