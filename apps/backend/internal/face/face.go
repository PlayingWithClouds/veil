// Package face names the performer in a picture. It sends the picture of a
// scene, gallery or image to every active face:identify plugin, merges and ranks
// what they answer, and credits the performer the user picks.
//
// Named candidates may only come from databases of professional performers or
// cam models; this package takes whatever plugin declares face:identify, so
// that rule is kept at the plugin level (see CLAUDE.md). Similar videos are
// passed through as links and never become candidates. Matching forum posts
// (reverse image search) are passed through too, with the performers their
// accepted answers name counted into a list of their own.
package face

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	// maxVideos bounds how many similar videos are returned, best first.
	maxVideos = 12
	// maxPosts bounds how many matching forum posts are returned, best first.
	maxPosts = 30
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

// VideoMatch is a tube-site video whose face looks like the picture's. It
// names no person and is never offered for crediting.
type VideoMatch struct {
	Title        string
	Similarity   float64
	URL          string
	ThumbnailURL string
	// Sources are the plugins that found the video.
	Sources []string
	// ScenePlugin is the active scene plugin that serves the page, so the app
	// can open it as a scene; empty when no plugin does.
	ScenePlugin string
}

// PostMatch is a post on a porn identification forum whose picture looks
// like the searched one. Names are what its accepted answer names.
type PostMatch struct {
	Title        string
	URL          string
	ThumbnailURL string
	Solved       bool
	Names        []string
	// Sources are the plugins that found the post.
	Sources []string
}

// PostPerformer is a performer the accepted answers of matching posts name.
type PostPerformer struct {
	Name string
	// Posts counts the matching posts naming the performer.
	Posts int
	// ImageURL is the thumbnail of the best post naming the performer.
	ImageURL string
	// PostURL is the best post naming the performer.
	PostURL string
	// PerformerID is the library's performer of that name, empty when there is none.
	PerformerID string
}

// Identification is the outcome of one identify run.
type Identification struct {
	Candidates []Candidate
	// Videos are similar videos from tube sites, best first.
	Videos []VideoMatch
	// Posts are matching forum posts, best first; PostPerformers the
	// performers their answers name, most often named first.
	Posts          []PostMatch
	PostPerformers []PostPerformer
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
	// scenePluginFor names the active scene plugin serving a page URL, or "".
	scenePluginFor func(pageURL string) string
	// loadPicture returns the bytes behind a stored image URL.
	loadPicture func(ctx context.Context, pictureURL string) ([]byte, error)
	// threadPluginFor is the active post:find plugin serving a thread URL, or nil.
	threadPluginFor func(threadURL string) *plugins.Plugin
	findPost        func(ctx context.Context, plugin *plugins.Plugin, threadURL string) (*plugins.ForumPost, error)
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
		scenePluginFor: func(pageURL string) string {
			for _, plugin := range registry.WithCapability(plugins.FindCapability(plugins.MediaTypeScene)) {
				if plugin.ServesURL(pageURL) {
					return plugin.Meta.Name
				}
			}
			return ""
		},
		identify:    runner.IdentifyFace,
		loadPicture: loadPicture,
		threadPluginFor: func(threadURL string) *plugins.Plugin {
			for _, plugin := range registry.WithCapability(plugins.CapabilityPostFind) {
				if plugin.ServesURL(threadURL) {
					return plugin
				}
			}
			return nil
		},
		findPost: runner.FindPost,
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

// IdentifyURL runs the face services on the picture at an http(s) address,
// loaded through the image cache, and returns the merged candidates, best first.
func (s *Service) IdentifyURL(ctx context.Context, pictureURL string) (*Identification, error) {
	parsed, err := url.Parse(pictureURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("%q is not a web address", pictureURL)
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
	videos     []plugins.FaceVideoMatch
	posts      []plugins.FacePostMatch
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
				answers[index].videos = result.Videos
				answers[index].posts = result.Posts
			}
		}()
	}
	group.Wait()
	return answers
}

// rank merges the answers into one candidate list and marks the performers
// the library already has.
func (s *Service) rank(ctx context.Context, answers []answer) *Identification {
	posts := mergePosts(answers)
	identification := &Identification{
		Candidates:     mergeAnswers(answers),
		Videos:         mergeVideos(answers),
		Posts:          posts,
		PostPerformers: postPerformers(posts),
	}
	for _, reply := range answers {
		if reply.err != nil {
			log.Printf("face:identify via plugin %s: %v", reply.plugin, reply.err)
			identification.Failures = append(identification.Failures, fmt.Sprintf("%s: %v", reply.plugin, reply.err))
			continue
		}
		identification.Sources = append(identification.Sources, reply.plugin)
	}
	s.markScenePlugins(identification.Videos)
	for index := range identification.Candidates {
		candidate := &identification.Candidates[index]
		candidate.PerformerID = s.libraryPerformer(ctx, candidate.Name)
	}
	for index := range identification.PostPerformers {
		performer := &identification.PostPerformers[index]
		performer.PerformerID = s.libraryPerformer(ctx, performer.Name)
	}
	return identification
}

// libraryPerformer is the id of the library's performer called name, empty
// when there is none or the lookup failed.
func (s *Service) libraryPerformer(ctx context.Context, name string) string {
	performerID, err := s.library.FindPerformer(ctx, name)
	if err != nil {
		log.Printf("face:identify: look up %q: %v", name, err)
		return ""
	}
	return performerID
}

// markScenePlugins records, per video, the scene plugin that can open its page.
func (s *Service) markScenePlugins(videos []VideoMatch) {
	if s.scenePluginFor == nil {
		return
	}
	for index := range videos {
		videos[index].ScenePlugin = s.scenePluginFor(videos[index].URL)
	}
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

// mergeVideos joins the videos of all services by page URL (best similarity
// wins, sources are united). Best first, at most maxVideos.
func mergeVideos(answers []answer) []VideoMatch {
	byURL := map[string]*VideoMatch{}
	var order []string
	for _, reply := range answers {
		for _, found := range reply.videos {
			if found.URL == "" {
				continue
			}
			merged, seen := byURL[found.URL]
			if !seen {
				merged = &VideoMatch{URL: found.URL}
				byURL[found.URL] = merged
				order = append(order, found.URL)
			}
			if found.Similarity > merged.Similarity || !seen {
				merged.Similarity = found.Similarity
				merged.Title = found.Title
			}
			if merged.ThumbnailURL == "" {
				merged.ThumbnailURL = found.ThumbnailURL
			}
			if !contains(merged.Sources, reply.plugin) {
				merged.Sources = append(merged.Sources, reply.plugin)
			}
		}
	}
	videos := make([]VideoMatch, 0, len(order))
	for _, pageURL := range order {
		videos = append(videos, *byURL[pageURL])
	}
	sort.SliceStable(videos, func(left, right int) bool {
		return videos[left].Similarity > videos[right].Similarity
	})
	if len(videos) > maxVideos {
		videos = videos[:maxVideos]
	}
	return videos
}

// mergePosts joins the posts of all services by page URL, keeping each
// service's order (the sites rank, they give no similarity) and uniting
// sources. At most maxPosts.
func mergePosts(answers []answer) []PostMatch {
	byURL := map[string]*PostMatch{}
	var order []string
	for _, reply := range answers {
		for _, found := range reply.posts {
			if found.URL == "" {
				continue
			}
			merged, seen := byURL[found.URL]
			if !seen {
				merged = &PostMatch{URL: found.URL, Title: found.Title, ThumbnailURL: found.ThumbnailURL}
				byURL[found.URL] = merged
				order = append(order, found.URL)
			}
			if found.Solved && len(merged.Names) == 0 {
				merged.Solved = true
				merged.Names = found.Names
			}
			if !contains(merged.Sources, reply.plugin) {
				merged.Sources = append(merged.Sources, reply.plugin)
			}
		}
	}
	if len(order) > maxPosts {
		order = order[:maxPosts]
	}
	posts := make([]PostMatch, 0, len(order))
	for _, pageURL := range order {
		posts = append(posts, *byURL[pageURL])
	}
	return posts
}

// postPerformers counts the performers the posts' answers name, by name_key.
// Most often named first, ties in the order of their best post.
func postPerformers(posts []PostMatch) []PostPerformer {
	byKey := map[string]*PostPerformer{}
	var order []string
	for _, post := range posts {
		for _, name := range post.Names {
			key := names.Key(name)
			if key == "" {
				continue
			}
			counted, seen := byKey[key]
			if !seen {
				counted = &PostPerformer{Name: name, ImageURL: post.ThumbnailURL, PostURL: post.URL}
				byKey[key] = counted
				order = append(order, key)
			}
			counted.Posts++
		}
	}
	performers := make([]PostPerformer, 0, len(order))
	for _, key := range order {
		performers = append(performers, *byKey[key])
	}
	sort.SliceStable(performers, func(left, right int) bool {
		return performers[left].Posts > performers[right].Posts
	})
	if len(performers) > maxCandidates {
		performers = performers[:maxCandidates]
	}
	return performers
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
