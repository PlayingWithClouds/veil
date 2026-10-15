// Package face names the performer in a picture. It sends the picture of a
// scene, gallery or image to every active face:identify plugin, merges and ranks
// what they answer, and credits the performer the user picks.
//
// Only plugins backed by databases of professional performers or cam models
// may offer face:identify; this package takes whatever plugin declares it, so
// that rule is kept at the plugin level (see CLAUDE.md).
package face

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const (
	// pluginTimeout is how long one recognition service may take.
	pluginTimeout = 45 * time.Second
	// maxPictureBytes is the biggest picture sent; the services accept 6 MB.
	maxPictureBytes = 5 << 20
	// maxCandidates bounds how many candidates are returned, best first.
	maxCandidates = 20
)

// Candidate is a performer one or more services think the picture shows.
type Candidate struct {
	Name string
	// Similarity runs 0..1: the best any service gave.
	Similarity  float64
	ImageURL    string
	ProfileURLs []string
	// Sources are the plugins that named the candidate.
	Sources []string
	// PerformerID is the library's performer of that name, empty when there is none.
	PerformerID string
}

// Identification is the outcome of one identify run.
type Identification struct {
	Candidates []Candidate
	// Sources are the plugins that answered; Failures describe those that did not.
	Sources  []string
	Failures []string
	// PictureURL is the stored picture that was sent; empty for a supplied one.
	PictureURL string
}

// Library is the performer store: lookups and crediting (ingest.Service).
type Library interface {
	FindPerformer(ctx context.Context, name string) (string, error)
	CreditPerformer(ctx context.Context, mediaID, name string) (string, error)
}

// Service identifies faces. Safe for concurrent use.
type Service struct {
	database *db.DB
	library  Library
	// services lists the active plugins that offer face:identify.
	services func() []*plugins.Plugin
	identify func(ctx context.Context, plugin *plugins.Plugin, args plugins.FaceIdentifyArgs) (*plugins.FaceIdentifyResult, error)
	// loadPicture returns the bytes behind a stored image URL.
	loadPicture func(ctx context.Context, pictureURL string) ([]byte, error)
}

// New returns a service asking the registry's face:identify plugins through
// the runner, reading pictures through loadPicture (the image cache).
func New(database *db.DB, library Library, registry *plugins.Registry, runner *plugins.Runner, loadPicture func(ctx context.Context, pictureURL string) ([]byte, error)) *Service {
	return &Service{
		database: database,
		library:  library,
		services: func() []*plugins.Plugin {
			return registry.WithCapability(plugins.CapabilityFaceIdentify)
		},
		identify:    runner.IdentifyFace,
		loadPicture: loadPicture,
	}
}

// Identify runs the face services on the picture of mediaID (a scene's
// thumbnail, an image, or a gallery's image imageID or first image) and
// returns the merged candidates, best first.
func (s *Service) Identify(ctx context.Context, mediaID, imageID string) (*Identification, error) {
	pictureURL, err := s.pictureURL(ctx, mediaID, imageID)
	if err != nil {
		return nil, err
	}
	picture, err := s.loadPicture(ctx, pictureURL)
	if err != nil {
		return nil, fmt.Errorf("load picture: %w", err)
	}
	identification, err := s.IdentifyPicture(ctx, picture)
	if err != nil {
		return nil, err
	}
	identification.PictureURL = pictureURL
	return identification, nil
}

// IdentifyPicture runs the face services on a picture the caller supplies,
// such as a video frame the user picked, and returns the merged candidates, best first.
func (s *Service) IdentifyPicture(ctx context.Context, picture []byte) (*Identification, error) {
	services := s.services()
	if len(services) == 0 {
		return nil, fmt.Errorf("no face recognition plugin is installed")
	}
	if len(picture) == 0 || len(picture) > maxPictureBytes {
		return nil, fmt.Errorf("picture is empty or larger than %d MB", maxPictureBytes>>20)
	}

	answers := s.askServices(ctx, services, plugins.FaceIdentifyArgs{
		ImageBase64: base64.StdEncoding.EncodeToString(picture),
		MimeType:    http.DetectContentType(picture),
	})
	return s.rank(ctx, answers), nil
}

// CreditPerformer credits the chosen name to the media: a scene or gallery;
// an image is credited on its gallery when it has one. Returns the performer's id.
func (s *Service) CreditPerformer(ctx context.Context, mediaID, name string) (string, error) {
	target, err := s.creditTarget(ctx, mediaID)
	if err != nil {
		return "", err
	}
	return s.library.CreditPerformer(ctx, target, name)
}

// creditTarget maps an image to its gallery; scenes and galleries stand for themselves.
func (s *Service) creditTarget(ctx context.Context, mediaID string) (string, error) {
	parsed, err := db.ParseRecordID(mediaID)
	if err != nil {
		return "", err
	}
	if parsed.Table != "image" {
		return mediaID, nil
	}
	owner, err := s.database.QueryRow(ctx, `SELECT media FROM image WHERE id = $id`, db.Vars{"id": mediaID})
	if err != nil {
		return "", err
	}
	galleryID, _ := owner["media"].(string)
	gallery, err := db.ParseRecordID(galleryID)
	if err != nil || gallery.Table != "gallery" {
		return mediaID, nil
	}
	return galleryID, nil
}

// answer is what one service replied.
type answer struct {
	plugin     string
	candidates []plugins.FaceCandidate
	err        error
}

// askServices queries every service concurrently, each with its own time limit.
func (s *Service) askServices(ctx context.Context, services []*plugins.Plugin, args plugins.FaceIdentifyArgs) []answer {
	answers := make([]answer, len(services))
	var group sync.WaitGroup
	for index, service := range services {
		group.Add(1)
		go func() {
			defer group.Done()
			serviceContext, cancel := context.WithTimeout(ctx, pluginTimeout)
			defer cancel()
			result, err := s.identify(serviceContext, service, args)
			answers[index] = answer{plugin: service.Meta.Name, err: err}
			if err == nil {
				answers[index].candidates = result.Candidates
			}
		}()
	}
	group.Wait()
	return answers
}

// rank merges the answers into one candidate list and marks the performers
// the library already has.
func (s *Service) rank(ctx context.Context, answers []answer) *Identification {
	identification := &Identification{Candidates: mergeAnswers(answers)}
	for _, reply := range answers {
		if reply.err != nil {
			log.Printf("face:identify via plugin %s: %v", reply.plugin, reply.err)
			identification.Failures = append(identification.Failures, fmt.Sprintf("%s: %v", reply.plugin, reply.err))
			continue
		}
		identification.Sources = append(identification.Sources, reply.plugin)
	}
	for index := range identification.Candidates {
		candidate := &identification.Candidates[index]
		performerID, err := s.library.FindPerformer(ctx, candidate.Name)
		if err != nil {
			log.Printf("face:identify: look up %q: %v", candidate.Name, err)
			continue
		}
		candidate.PerformerID = performerID
	}
	return identification
}

// mergeAnswers joins candidates of the same name_key across services: the best
// similarity wins, profile links and sources are united. Best first, at most maxCandidates.
func mergeAnswers(answers []answer) []Candidate {
	byKey := map[string]*Candidate{}
	var order []string
	for _, reply := range answers {
		for _, found := range reply.candidates {
			key := names.Key(found.Name)
			if key == "" {
				continue
			}
			merged, seen := byKey[key]
			if !seen {
				merged = &Candidate{Name: found.Name}
				byKey[key] = merged
				order = append(order, key)
			}
			mergeInto(merged, found, reply.plugin)
		}
	}
	candidates := make([]Candidate, 0, len(order))
	for _, key := range order {
		candidates = append(candidates, *byKey[key])
	}
	sort.SliceStable(candidates, func(left, right int) bool {
		return candidates[left].Similarity > candidates[right].Similarity
	})
	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}
	return candidates
}

// mergeInto folds one service's candidate into the merged one.
func mergeInto(merged *Candidate, found plugins.FaceCandidate, plugin string) {
	if found.Similarity > merged.Similarity {
		merged.Similarity = found.Similarity
		merged.Name = found.Name
	}
	if merged.ImageURL == "" {
		merged.ImageURL = found.ImageURL
	}
	for _, profile := range found.ProfileURLs {
		if !contains(merged.ProfileURLs, profile) {
			merged.ProfileURLs = append(merged.ProfileURLs, profile)
		}
	}
	if !contains(merged.Sources, plugin) {
		merged.Sources = append(merged.Sources, plugin)
	}
}

// contains reports whether values holds value.
func contains(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

// pictureURL picks the stored image to send: an image's own file, a gallery's
// chosen or first image (else its cover), a scene's thumbnail or poster.
func (s *Service) pictureURL(ctx context.Context, mediaID, imageID string) (string, error) {
	parsed, err := db.ParseRecordID(mediaID)
	if err != nil {
		return "", err
	}
	switch parsed.Table {
	case "scene":
		return s.firstText(ctx, `SELECT COALESCE(thumbnail_path, poster_path) AS picture FROM scene WHERE id = $media`, mediaID, imageID)
	case "image":
		return s.firstText(ctx, `SELECT file_path AS picture FROM image WHERE id = $media`, mediaID, imageID)
	case "gallery":
		return s.galleryPictureURL(ctx, mediaID, imageID)
	}
	return "", fmt.Errorf("no picture to identify on a %s", parsed.Table)
}

// galleryPictureURL is the picture of imageID in the gallery, else its first content image, else its cover.
func (s *Service) galleryPictureURL(ctx context.Context, galleryID, imageID string) (string, error) {
	if imageID != "" {
		return s.firstText(ctx, `SELECT file_path AS picture FROM image WHERE id = $image AND media = $media`, galleryID, imageID)
	}
	pictureURL, err := s.firstText(ctx, `SELECT file_path AS picture FROM image WHERE media = $media AND content = 1 ORDER BY position, id LIMIT 1`, galleryID, imageID)
	if err == nil {
		return pictureURL, nil
	}
	return s.firstText(ctx, `SELECT cover_path AS picture FROM gallery WHERE id = $media`, galleryID, imageID)
}

// firstText runs a query yielding one "picture" column and returns it, or an
// error when the record has none.
func (s *Service) firstText(ctx context.Context, query, mediaID, imageID string) (string, error) {
	row, err := s.database.QueryRow(ctx, query, db.Vars{"media": mediaID, "image": imageID})
	if err != nil {
		return "", err
	}
	pictureURL, _ := row["picture"].(string)
	if pictureURL == "" {
		return "", fmt.Errorf("%s has no picture to identify", mediaID)
	}
	return pictureURL, nil
}
