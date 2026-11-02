package faces

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// testModel is the model name the tests scan with.
const testModel = "test-faces/1"

// openService returns the service on a fresh database.
func openService(t *testing.T) (*Service, *db.DB) {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return New(database), database
}

// insert stores a row and returns its id.
func insert(t *testing.T, database *db.DB, table string, row any) string {
	t.Helper()
	id, err := database.Insert(context.Background(), table, row)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// video stores a video crediting performers.
func video(t *testing.T, database *db.DB, performers ...string) string {
	t.Helper()
	return insert(t, database, "Video", db.Vars{"title": "Scene", "status": "available", "performers": performers, "tags": []string{}})
}

// frames are faces of people (identity axes), one frame per entry; each
// entry lists the axes seen in that frame.
func frames(entries ...[]int) []Face {
	var faces []Face
	for index, axes := range entries {
		for _, axis := range axes {
			faces = append(faces, Face{
				Vector: noisy(axis, 7-index%3, 0.1+0.05*float32(index)), AtSeconds: at(float64(10 * index)),
				Region: map[string]any{"x": 0.1, "y": 0.1, "width": 0.2, "height": 0.3},
			})
		}
	}
	return faces
}

// scanVideo records a video scan of faces.
func scanVideo(t *testing.T, service *Service, videoID string, faces []Face) {
	t.Helper()
	subject := schema.RecordRef{Kind: schema.RecordKindVideo, Id: videoID}
	if _, err := service.RecordScan(context.Background(), Scan{Subject: subject, Model: testModel, Frames: 8, Faces: faces}); err != nil {
		t.Fatal(err)
	}
}

// TestGuessesFromPicturesAndLearnedFaces learns one performer's face from
// her picture and credited video and another's from the video crediting
// only her, then guesses both in an uncredited video; a credit taken back
// stops counting.
func TestGuessesFromPicturesAndLearnedFaces(t *testing.T) {
	ctx := context.Background()
	service, database := openService(t)
	ada := insert(t, database, "Performer", schema.Performer{Name: "Ada", Image: &schema.Asset{Uri: "blob://ada"}})
	bea := insert(t, database, "Performer", schema.Performer{Name: "Bea"})
	picture := "blob://ada"
	pictureScan := Scan{
		Subject: schema.RecordRef{Kind: schema.RecordKindPerformer, Id: ada}, Model: testModel,
		Picture: &picture, Frames: 1, Faces: []Face{{Vector: identity(0)}},
	}
	if _, err := service.RecordScan(ctx, pictureScan); err != nil {
		t.Fatal(err)
	}
	pictures, err := service.Pictures(ctx, ada, testModel)
	if err != nil || len(pictures) != 1 || !pictures[0].Scanned {
		t.Fatalf("pictures = %+v, err %v", pictures, err)
	}

	scanVideo(t, service, video(t, database, ada, bea), frames([]int{0, 1}, []int{0}, []int{0}))
	beaOnly := video(t, database, bea)
	scanVideo(t, service, beaOnly, frames([]int{1}, []int{1}, []int{1}, []int{1}, []int{2}))
	uncredited := video(t, database)
	scanVideo(t, service, uncredited, frames([]int{0, 1}, []int{0}, []int{1}, []int{2}))

	found, err := service.VideoFaces(ctx, uncredited, "")
	if err != nil {
		t.Fatal(err)
	}
	if found.Scan == nil || found.Scan.Faces != 5 || len(found.People) != 3 {
		t.Fatalf("scan = %+v, people = %d", found.Scan, len(found.People))
	}
	if guessed := guessedPerformers(found); guessed[ada] == 0 || guessed[bea] == 0 || len(guessed) != 2 {
		t.Errorf("guesses = %+v", found.Guesses)
	}

	if err := database.Update(ctx, "Video", beaOnly, db.Vars{"performers": []string{}}); err != nil {
		t.Fatal(err)
	}
	found, err = service.VideoFaces(ctx, uncredited, testModel)
	if err != nil {
		t.Fatal(err)
	}
	if guessed := guessedPerformers(found); guessed[bea] != 0 || guessed[ada] == 0 {
		t.Errorf("after the credit was taken back, guesses = %+v", found.Guesses)
	}
}

// TestRescanReplacesTheEarlierScan keeps one scan per video and model, with
// only the new scan's faces.
func TestRescanReplacesTheEarlierScan(t *testing.T) {
	ctx := context.Background()
	service, database := openService(t)
	videoID := video(t, database)
	scanVideo(t, service, videoID, frames([]int{0}, []int{0}, []int{1}))
	scanVideo(t, service, videoID, frames([]int{0}))
	scans, err := service.Scans(ctx, schema.RecordRef{Kind: schema.RecordKindVideo, Id: videoID})
	if err != nil || len(scans) != 1 || scans[0].Faces != 1 {
		t.Fatalf("scans = %+v, err %v", scans, err)
	}
	row, err := database.QueryRow(ctx, `SELECT count(*) AS "count" FROM "Embedding" WHERE "subjectId" = $id`, db.Vars{"id": videoID})
	if err != nil || row["count"] != int64(1) {
		t.Errorf("embeddings = %v, err %v", row, err)
	}
	if _, err := service.RecordScan(ctx, Scan{Subject: schema.RecordRef{Kind: schema.RecordKindVideo, Id: "missing"}, Model: testModel}); err == nil {
		t.Error("scanned a video that doesn't exist")
	}
}

// TestPerformerMatches finds the uncredited video showing a performer's
// learned face, and neither the video crediting her nor one without her.
func TestPerformerMatches(t *testing.T) {
	ctx := context.Background()
	service, database := openService(t)
	ada := insert(t, database, "Performer", schema.Performer{Name: "Ada"})
	credited := video(t, database, ada)
	scanVideo(t, service, credited, frames([]int{0}, []int{0}, []int{0}, []int{0}))
	showsHer := video(t, database)
	scanVideo(t, service, showsHer, frames([]int{0, 1}, []int{0}, []int{1}))
	without := video(t, database)
	scanVideo(t, service, without, frames([]int{2}, []int{2}))

	matches, err := service.PerformerMatches(ctx, ada, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].VideoID != showsHer {
		t.Errorf("matches = %+v, want only %s", matches, showsHer)
	}
}

// guessedPerformers counts the guesses per performer.
func guessedPerformers(found *VideoFaces) map[string]int {
	guessed := map[string]int{}
	for _, guess := range found.Guesses {
		guessed[guess.PerformerID]++
	}
	return guessed
}
