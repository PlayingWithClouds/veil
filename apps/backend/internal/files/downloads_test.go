package files

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/storage"
)

// videoWithStream inserts a video with a source and an MP4 stream at url.
func (env *testEnv) videoWithStream(t *testing.T, url string) schema.RecordRef {
	t.Helper()
	ctx := context.Background()
	video := env.newVideo(t, "Pool day")
	sourceID, err := env.database.Insert(ctx, tableSource, db.Vars{
		"subjectKind": video.Kind, "subjectId": video.Id, "plugin": "site", "externalId": "v1", "url": url,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.database.Insert(ctx, tableStream, db.Vars{
		"subjectKind": video.Kind, "subjectId": video.Id, "source": sourceID, "url": url, "mimeType": "video/mp4",
		"headers": map[string]string{}, "height": 720,
	}); err != nil {
		t.Fatal(err)
	}
	return video
}

// TestDownload downloads a video's stream into the blob store, records it as
// a file of the video, keeps it pinned and deletes it again.
func TestDownload(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	content := patterned(300*1024, 9)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.ServeContent(writer, request, "clip.mp4", time.Now(), bytes.NewReader(content))
	}))
	defer server.Close()
	video := env.videoWithStream(t, server.URL+"/clip.mp4")

	jobID, err := env.service.Download(ctx, video, nil)
	if err != nil {
		t.Fatal(err)
	}
	if again, _ := env.service.Download(ctx, video, nil); again != jobID {
		t.Errorf("second download queued %s, want %s", again, jobID)
	}
	finished := env.runJob(t, jobs.KindDownload, env.service.DownloadTask())
	downloadURL := payloadString(finished, "download_url")
	key, isBlob := storage.KeyFromURL(downloadURL)
	if !isBlob || !env.store.Exists(ctx, key) {
		t.Fatalf("download_url = %q", downloadURL)
	}
	pinned, _ := env.queue.DownloadURLs(ctx)
	if !slices.Contains(pinned, downloadURL) {
		t.Errorf("pinned = %v", pinned)
	}
	files, _ := env.service.Files(ctx, video.Kind, video.Id)
	if len(files) != 1 || files[0].SizeBytes != int64(len(content)) || env.service.FileURL(ctx, files[0]) != downloadURL {
		t.Fatalf("files = %+v", files)
	}
	if again, _ := env.service.Download(ctx, video, nil); again != jobID {
		t.Errorf("downloading again queued %s, want the finished %s", again, jobID)
	}
	downloads, _ := env.service.Downloads(ctx)
	if len(downloads) != 1 || downloads[0].File == nil || downloads[0].File.Id != files[0].Id {
		t.Errorf("downloads = %+v", downloads)
	}

	if err := env.service.DeleteDownload(ctx, jobID); err != nil {
		t.Fatal(err)
	}
	if env.store.Exists(ctx, key) {
		t.Error("blob kept")
	}
	if files, _ := env.service.Files(ctx, video.Kind, video.Id); len(files) != 0 {
		t.Errorf("files = %+v after delete", files)
	}
	if downloads, _ := env.service.Downloads(ctx); len(downloads) != 0 {
		t.Errorf("downloads = %+v after delete", downloads)
	}
}

// TestDownloadWaitsForPlayback holds a download back while a stream plays
// when downloads while streaming are off, without using up an attempt.
func TestDownloadWaitsForPlayback(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	if _, err := env.settings.Update(ctx, map[string]any{"allowDownloadsWhileStreaming": false}); err != nil {
		t.Fatal(err)
	}
	env.streamCache.MarkStreamActivity(ctx)
	video := env.videoWithStream(t, "http://127.0.0.1:1/clip.mp4")
	jobID, err := env.service.Download(ctx, video, nil)
	if err != nil {
		t.Fatal(err)
	}
	job, _ := env.queue.Claim(ctx, []jobs.Kind{jobs.KindDownload})
	if err := env.service.DownloadTask()(ctx, job, func(float64) {}); err != nil {
		t.Fatal(err)
	}
	held, _ := env.queue.Get(ctx, jobID)
	if held.Status != jobs.StatusQueued || held.Attempts != 0 || held.RunAt == nil || !held.RunAt.After(time.Now()) {
		t.Errorf("held = %+v", held)
	}
}

// TestPacerHoldsToTheLimit waits until received bytes fit the rate.
func TestPacerHoldsToTheLimit(t *testing.T) {
	limiter := newPacer(100)
	started := time.Now()
	limiter.wait(context.Background(), 10*1024)
	if elapsed := time.Since(started); elapsed < 80*time.Millisecond {
		t.Errorf("10 KiB at 100 KiB/s took %v", elapsed)
	}
}
