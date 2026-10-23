// Package mediakeep keeps listing media of plugins whose site signs its media
// URLs with an expiry (meta.expiring_media). A stub's poster is copied into the
// blob store while its URL still works, and the preview clip URL is refreshed
// whenever the scene is listed again.
package mediakeep

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/storage"
)

// queueSize bounds the pending copies; a full queue drops the stub, which is
// picked up again the next time a listing shows it.
const queueSize = 512

// fetchTimeout bounds one poster download.
const fetchTimeout = 20 * time.Second

// pluginLookup finds an installed plugin by name; plugins.Registry.Get in production.
type pluginLookup func(name string) (*plugins.Plugin, bool)

// stub is one listed scene waiting for its media to be kept.
type stub struct {
	sceneID string
	item    plugins.DiscoveredItem
}

// Keeper copies the posters of expiring-media stubs into the blob store in the
// background.
type Keeper struct {
	database *db.DB
	store    *storage.Client
	lookup   pluginLookup
	queue    chan stub
}

// New creates a keeper; Run starts its worker.
func New(database *db.DB, store *storage.Client, registry *plugins.Registry) *Keeper {
	return &Keeper{database: database, store: store, lookup: registry.Get, queue: make(chan stub, queueSize)}
}

// Run copies queued posters until ctx ends.
func (k *Keeper) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case pending := <-k.queue:
			if err := k.keep(ctx, pending); err != nil {
				log.Printf("mediakeep %s: %v", pending.sceneID, err)
			}
		}
	}
}

// StubIngested queues a listed scene of an expiring-media plugin. It never
// blocks the ingest that calls it.
func (k *Keeper) StubIngested(pluginName, sceneID string, item plugins.DiscoveredItem) {
	if !strings.HasPrefix(sceneID, "scene:") || !k.expiringMedia(pluginName) {
		return
	}
	select {
	case k.queue <- stub{sceneID: sceneID, item: item}:
	default:
	}
}

// expiringMedia reports whether the plugin declares expiring media URLs.
func (k *Keeper) expiringMedia(pluginName string) bool {
	plugin, found := k.lookup(pluginName)
	return found && plugin.Meta.ExpiringMedia
}

// keep refreshes the scene's preview clip and stores its poster.
func (k *Keeper) keep(ctx context.Context, pending stub) error {
	scene, err := k.database.Query(ctx, `SELECT poster_path, preview_video FROM scene WHERE id = $id`,
		db.Vars{"id": pending.sceneID})
	if err != nil || len(scene) == 0 {
		return err
	}
	currentPoster, _ := scene[0]["poster_path"].(string)
	currentPreview, _ := scene[0]["preview_video"].(string)
	if err := k.refreshPreview(ctx, pending, currentPreview); err != nil {
		return err
	}
	if !replaceable(currentPoster, pending.item.PosterPath) {
		return nil
	}
	blobURL, err := k.storePoster(ctx, pending.item)
	if err != nil {
		return err
	}
	_, err = k.database.Exec(ctx, `UPDATE scene SET poster_path = $poster WHERE id = $id`,
		db.Vars{"id": pending.sceneID, "poster": blobURL})
	return err
}

// refreshPreview swaps the scene's preview clip for the listing's fresh URL
// when the current one comes from the same site.
func (k *Keeper) refreshPreview(ctx context.Context, pending stub, currentPreview string) error {
	fresh := pending.item.PreviewVideo
	if fresh == "" || fresh == currentPreview || !sameSite(currentPreview, fresh) {
		return nil
	}
	_, err := k.database.Exec(ctx, `UPDATE scene SET preview_video = $preview WHERE id = $id`,
		db.Vars{"id": pending.sceneID, "preview": fresh})
	return err
}

// replaceable reports whether the stored poster should be replaced by a copy
// of the listing's: it is not kept in the blob store yet and is empty or comes
// from the same site (a scene several sites carry keeps another site's poster).
func replaceable(currentPoster, listedPoster string) bool {
	if !isRemote(listedPoster) {
		return false
	}
	if _, isBlob := storage.KeyFromURL(currentPoster); isBlob {
		return false
	}
	return currentPoster == "" || sameSite(currentPoster, listedPoster)
}

// storePoster downloads the listing's poster into posters/ (the key a scene
// visit uses too) and returns its blob URL.
func (k *Keeper) storePoster(ctx context.Context, item plugins.DiscoveredItem) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	body, contentType, err := download(ctx, item.PosterPath, item.SourceURL)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("posters/%s%s", item.ExternalID, posterExtension(item.PosterPath))
	if err := k.store.Put(ctx, key, bytes.NewReader(body), int64(len(body)), contentType); err != nil {
		return "", err
	}
	return k.store.DirectURL(key), nil
}

// download fetches an image with the page as referer.
func download(ctx context.Context, imageURL, referer string) ([]byte, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	if referer != "" {
		request.Header.Set("Referer", referer)
	}
	response, err := outbound.Client.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("poster fetch %s: HTTP %d", imageURL, response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := response.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	return body, contentType, nil
}

// posterExtension is the URL path's extension, .jpg when there is none.
func posterExtension(posterURL string) string {
	parsed, err := url.Parse(posterURL)
	if err != nil {
		return ".jpg"
	}
	extension := path.Ext(parsed.Path)
	if extension == "" || len(extension) > 5 {
		return ".jpg"
	}
	return extension
}

// isRemote reports whether value is an http(s) URL.
func isRemote(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

// sameSite reports whether two remote URLs share their registrable domain
// (the last two host labels), so b1. and b3. of one CDN count as one site.
func sameSite(left, right string) bool {
	leftDomain := siteDomain(left)
	return leftDomain != "" && leftDomain == siteDomain(right)
}

// siteDomain returns the last two labels of the URL's host, empty when the
// value is no remote URL.
func siteDomain(value string) string {
	if !isRemote(value) {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	labels := strings.Split(parsed.Hostname(), ".")
	if len(labels) < 2 {
		return parsed.Hostname()
	}
	return strings.Join(labels[len(labels)-2:], ".")
}
