package files

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestScanAddsMatchesAndMarksMissing scans a root twice: the first scan
// creates a video, an audio and a folder gallery with fingerprints; the
// second matches a copied file to the same video and flags a removed one.
func TestScanAddsMatchesAndMarksMissing(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	root, folder := env.libraryRoot(t, nil)
	video := patterned(200*1024, 1)
	writeFile(t, filepath.Join(folder, "Pool_day.mp4"), video)
	writeFile(t, filepath.Join(folder, "song.mp3"), patterned(4096, 2))
	writeFile(t, filepath.Join(folder, "set", "one.jpg"), patterned(512, 3))
	writeFile(t, filepath.Join(folder, "set", "two.jpg"), patterned(512, 4))
	writeFile(t, filepath.Join(folder, "notes.txt"), []byte("not media"))
	writeFile(t, filepath.Join(folder, ".hidden", "x.mp4"), patterned(64, 5))

	env.scan(t, root.Id)
	if len(env.added) != 4 {
		t.Fatalf("added %d files, want 4: %+v", len(env.added), env.added)
	}
	videoFile := addedAt(t, env.added, "media/Pool_day.mp4")
	stored, _ := db.Get[schema.Video](ctx, env.database, "Video", videoFile.SubjectId)
	if stored == nil || stored.Title != "Pool day" || stored.Status != schema.MediaStatusAvailable {
		t.Errorf("video = %+v", stored)
	}
	fingerprints, _ := env.service.Fingerprints(ctx, schema.RecordRef{Kind: schema.RecordKindVideo, Id: videoFile.SubjectId})
	if len(fingerprints) != 2 {
		t.Errorf("fingerprints = %+v, want oshash and md5", fingerprints)
	}
	image := addedAt(t, env.added, "media/set/one.jpg")
	gallery, _ := db.QueryOneAs[schema.Gallery](ctx, env.database,
		`SELECT "Gallery".* FROM "Gallery" JOIN "Image" ON "Image"."gallery" = "Gallery"."id" WHERE "Image"."id" = $id`,
		db.Vars{"id": image.SubjectId})
	if gallery == nil || gallery.Title != "set" || gallery.ImageCount != 2 || gallery.Cover == nil {
		t.Errorf("gallery = %+v", gallery)
	}

	writeFile(t, filepath.Join(folder, "copy.mp4"), video)
	if err := os.Remove(filepath.Join(folder, "song.mp3")); err != nil {
		t.Fatal(err)
	}
	env.scan(t, root.Id)
	copied := addedAt(t, env.added, "media/copy.mp4")
	if copied.SubjectId != videoFile.SubjectId {
		t.Errorf("copy belongs to %s, want the matching video %s", copied.SubjectId, videoFile.SubjectId)
	}
	song, _ := env.service.Files(ctx, schema.RecordKindAudio, addedAt(t, env.added, "media/song.mp3").SubjectId)
	if len(song) != 1 || !song[0].Missing {
		t.Errorf("song files = %+v, want one missing", song)
	}
}

// addedAt finds the added file at a location path.
func addedAt(t *testing.T, added []schema.MediaFile, path string) schema.MediaFile {
	t.Helper()
	for _, file := range added {
		if file.Path == path {
			return file
		}
	}
	t.Fatalf("no file added at %s", path)
	return schema.MediaFile{}
}

// TestEnqueueDueScans queues a root's scan only once its interval passed.
func TestEnqueueDueScans(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	root, _ := env.libraryRoot(t, nil)
	if queued, err := env.service.EnqueueDueScans(ctx, time.Now()); err != nil || queued != 1 {
		t.Fatalf("never scanned: queued %d, err %v", queued, err)
	}
	env.runJob(t, schema.JobKindLibraryScan, env.service.ScanTask())
	if queued, _ := env.service.EnqueueDueScans(ctx, time.Now()); queued != 0 {
		t.Errorf("just scanned: queued %d", queued)
	}
	if queued, _ := env.service.EnqueueDueScans(ctx, time.Now().Add(61*time.Minute)); queued != 1 {
		t.Errorf("interval passed: queued %d", queued)
	}
	manual := 0
	if _, err := env.service.UpdateRoot(ctx, root.Id, RootEdit{ScanIntervalMinutes: &manual}); err != nil {
		t.Fatal(err)
	}
	if queued, _ := env.service.EnqueueDueScans(ctx, time.Now().Add(24*time.Hour)); queued != 0 {
		t.Errorf("manual root: queued %d", queued)
	}
}

// TestOrganize plans moves from the template and applies them, leaving
// files already in place alone.
func TestOrganize(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	template := "{studio}/{date} {title}.{ext}"
	root, folder := env.libraryRoot(t, &template)
	writeFile(t, filepath.Join(folder, "incoming", "clip.MP4"), patterned(2048, 1))
	writeFile(t, filepath.Join(folder, "song.mp3"), patterned(2048, 2))
	env.scan(t, root.Id)
	clip := addedAt(t, env.added, "media/incoming/clip.MP4")
	studioID, err := env.database.Insert(ctx, "Studio", db.Vars{"name": "Sun: Studio", "aliases": []string{}, "tags": []string{}, "nameKey": "sun studio"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.database.Exec(ctx, `UPDATE "Video" SET "title" = 'Pool day?', "date" = '2024-05-01', "studio" = $studio WHERE "id" = $id`,
		db.Vars{"studio": studioID, "id": clip.SubjectId}); err != nil {
		t.Fatal(err)
	}

	moves, err := env.service.PlanOrganize(ctx, root.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 1 || moves[0].From != "incoming/clip.MP4" || moves[0].To != "Sun_ Studio/2024-05-01 Pool day_.mp4" {
		t.Fatalf("moves = %+v", moves)
	}
	if moved, err := env.service.ApplyOrganize(ctx, root.Id, moves); err != nil || moved != 1 {
		t.Fatalf("moved %d, err %v", moved, err)
	}
	if !fileExists(filepath.Join(folder, "Sun_ Studio", "2024-05-01 Pool day_.mp4")) {
		t.Error("file not at its new place")
	}
	moved, _ := env.service.File(ctx, clip.Id)
	if moved.Path != "media/Sun_ Studio/2024-05-01 Pool day_.mp4" {
		t.Errorf("row path = %s", moved.Path)
	}
	if _, err := env.service.ApplyOrganize(ctx, root.Id, moves); err == nil {
		t.Error("a stale plan was applied again")
	}
}

// TestTrashFileAndPurge moves a removed file into the root's trash folder
// and deletes it once the retention passed.
func TestTrashFileAndPurge(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	root, folder := env.libraryRoot(t, nil)
	writeFile(t, filepath.Join(folder, "a", "clip.mp4"), patterned(1024, 1))
	env.scan(t, root.Id)
	clip := addedAt(t, env.added, "media/a/clip.mp4")

	if err := env.service.TrashFile(ctx, clip.Id); err != nil {
		t.Fatal(err)
	}
	trashed := filepath.Join(folder, trashFolder, "a", "clip.mp4")
	if !fileExists(trashed) || fileExists(filepath.Join(folder, "a", "clip.mp4")) {
		t.Fatal("file not moved to the trash folder")
	}
	if files, _ := env.service.Files(ctx, clip.SubjectKind, clip.SubjectId); len(files) != 0 {
		t.Errorf("files = %+v, want none", files)
	}
	env.scan(t, root.Id)
	if len(env.added) != 1 {
		t.Errorf("the scan picked up the trash folder: %+v", env.added)
	}
	if purged, _ := env.service.PurgeTrash(ctx, time.Now().Add(-time.Hour)); purged != 0 || !fileExists(trashed) {
		t.Errorf("purged %d before the retention passed", purged)
	}
	if purged, err := env.service.PurgeTrash(ctx, time.Now().Add(time.Hour)); err != nil || purged != 1 || fileExists(trashed) {
		t.Errorf("purged %d, err %v", purged, err)
	}
	if row, _ := env.database.QueryRow(ctx, `SELECT "id" FROM "MediaFile" WHERE "id" = $id`, db.Vars{"id": clip.Id}); row != nil {
		t.Error("trashed row not purged")
	}
}

// TestAddRootRefusesOutsidePaths keeps roots inside their location.
func TestAddRootRefusesOutsidePaths(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	location, err := env.devices.AddLocation(ctx, schema.StorageKindLibrary, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.service.AddRoot(ctx, location.Id, "../elsewhere", nil, nil); err == nil {
		t.Error("root outside its location accepted")
	}
	bad := "{studio}/{rating}"
	if _, err := env.service.AddRoot(ctx, location.Id, ".", nil, &bad); err == nil {
		t.Error("unknown placeholder accepted")
	}
}
