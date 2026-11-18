// Package dedupe finds the same video stored twice from different sites by
// its pictures, and merges the copies that are provably one scene.
//
// Titles are a weak signal across sites (reworded, translated, the cast
// added), so a sweep hashes every video's cover and the stills the site shows
// across it (difference hashes, alike.DecodePosterHash) and compares videos of
// near-equal runtime. Strict matches (alike.SameSceneByImages, or the strict
// title matcher) are merged automatically; merges are journaled, so they can
// be undone. Looser pairs stay hidden pairs (library.DetectAlike).
package dedupe

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/schema"
)

// picturesKey prefixes the cache entry holding a video's picture hashes:
// "<cover hex>|<frame hex>,<frame hex>".
const picturesKey = "pictures:"

// passInterval is how often the sweep runs.
const passInterval = 10 * time.Minute

// hashBatch is how many videos one pass hashes.
const hashBatch = 60

// framesHashed is how many of a video's stills are hashed, spread over it.
const framesHashed = 4

// placeholderCount is how many videos may share one cover hash before the
// cover counts as a site's placeholder and stops proving anything.
const placeholderCount = 3

// pairBatch is how many waiting pairs one pass checks for a strict match.
const pairBatch = 200

// ImageLoader returns an image's bytes from a remote URL or the blob store
// (imgcache.Handler).
type ImageLoader interface {
	Load(ctx context.Context, raw string) ([]byte, error)
}

// Sweeper hashes pictures and merges provable duplicates.
type Sweeper struct {
	library *library.Library
	cache   *cache.Client
	images  ImageLoader
}

// New returns a sweeper over the library.
func New(lib *library.Library, cacheClient *cache.Client, images ImageLoader) *Sweeper {
	return &Sweeper{library: lib, cache: cacheClient, images: images}
}

// Run sweeps at start and then every passInterval until ctx ends.
func (s *Sweeper) Run(ctx context.Context) {
	for {
		if err := s.Pass(ctx); err != nil {
			log.Printf("dedupe: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(passInterval):
		}
	}
}

// Pass hashes a batch of unhashed videos, merges each into a stored copy it
// provably is, then merges the waiting alike pairs that are provably one.
func (s *Sweeper) Pass(ctx context.Context) error {
	hashed, err := s.hashBatch(ctx)
	if err != nil {
		return err
	}
	all, err := s.allPictures(ctx)
	if err != nil {
		return err
	}
	covers := coverCounts(all)
	for _, video := range hashed {
		if err := s.mergeCopies(ctx, video, all, covers); err != nil {
			log.Printf("dedupe: video %s: %v", video.Id, err)
		}
	}
	return s.mergeProvablePairs(ctx, all, covers)
}

// hashBatch hashes the pictures of up to hashBatch videos that have none yet.
func (s *Sweeper) hashBatch(ctx context.Context) ([]schema.Video, error) {
	videos, err := s.library.VideosMissingPictures(ctx, picturesKey, hashBatch)
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		pictures := s.hashPictures(ctx, video)
		if err := s.cache.Set(ctx, picturesKey+video.Id, encodePictures(pictures), 0); err != nil {
			return nil, err
		}
	}
	return videos, nil
}

// hashPictures hashes a video's cover and a few of its stills; pictures that
// can't be loaded or decoded stay 0 (unknown).
func (s *Sweeper) hashPictures(ctx context.Context, video schema.Video) alike.Pictures {
	pictures := alike.Pictures{}
	if video.Cover != nil {
		pictures.Cover = s.hash(ctx, video.Cover.Uri)
	}
	for _, frame := range spread(video.PreviewFrames, framesHashed) {
		hash := s.hash(ctx, frame.Uri)
		if hash != 0 {
			pictures.Frames = append(pictures.Frames, hash)
		}
	}
	return pictures
}

// hash loads and hashes one image; 0 when that fails.
func (s *Sweeper) hash(ctx context.Context, uri string) uint64 {
	if strings.TrimSpace(uri) == "" {
		return 0
	}
	data, err := s.images.Load(ctx, uri)
	if err != nil {
		return 0
	}
	hash, err := alike.DecodePosterHash(data)
	if err != nil {
		return 0
	}
	return hash
}

// spread picks up to count assets evenly across the list.
func spread(assets []schema.Asset, count int) []schema.Asset {
	if len(assets) <= count {
		return assets
	}
	picked := make([]schema.Asset, 0, count)
	for index := range count {
		picked = append(picked, assets[index*len(assets)/count])
	}
	return picked
}

// allPictures reads every hashed video's pictures by video id.
func (s *Sweeper) allPictures(ctx context.Context) (map[string]alike.Pictures, error) {
	entries, err := s.cache.Entries(ctx, picturesKey)
	if err != nil {
		return nil, err
	}
	out := make(map[string]alike.Pictures, len(entries))
	for key, value := range entries {
		out[strings.TrimPrefix(key, picturesKey)] = decodePictures(value)
	}
	return out, nil
}

// coverCounts counts the videos per cover hash, to spot placeholders.
func coverCounts(all map[string]alike.Pictures) map[uint64]int {
	counts := map[uint64]int{}
	for _, pictures := range all {
		if pictures.Cover != 0 {
			counts[pictures.Cover]++
		}
	}
	return counts
}

// coversUsable reports whether neither cover is a placeholder.
func coversUsable(covers map[uint64]int, first, second alike.Pictures) bool {
	return covers[first.Cover] <= placeholderCount && covers[second.Cover] <= placeholderCount
}

// mergeCopies merges the video into the first stored video of similar
// runtime it is provably the same scene as (the older record survives).
func (s *Sweeper) mergeCopies(ctx context.Context, video schema.Video, all map[string]alike.Pictures, covers map[uint64]int) error {
	others, err := s.library.SimilarRuntime(ctx, &video)
	if err != nil {
		return err
	}
	target, err := s.library.AlikeCandidate(ctx, &video)
	if err != nil {
		return err
	}
	for _, other := range others {
		candidate, err := s.library.AlikeCandidate(ctx, &other)
		if err != nil {
			return err
		}
		if !provablySame(target, candidate, all[video.Id], all[other.Id], covers) {
			continue
		}
		keep, merge := older(video, other)
		log.Printf("dedupe: merging video %s into %s (same scene)", merge, keep)
		return s.library.Merge(ctx, schema.RecordKindVideo, keep, merge)
	}
	return nil
}

// mergeProvablePairs merges the waiting alike pairs whose videos are provably
// one: by pictures, the strict title matcher or a shared file fingerprint.
func (s *Sweeper) mergeProvablePairs(ctx context.Context, all map[string]alike.Pictures, covers map[uint64]int) error {
	pairs, err := s.library.Alikes(ctx, schema.RecordKindVideo, pairBatch, 0)
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		same, err := s.pairProvablySame(ctx, pair, all, covers)
		if err != nil {
			return err
		}
		if !same {
			continue
		}
		log.Printf("dedupe: merging pair %s (video %s into %s)", pair.Id, pair.SecondId, pair.FirstId)
		if err := s.library.Merge(ctx, schema.RecordKindVideo, pair.FirstId, pair.SecondId); err != nil {
			return fmt.Errorf("merge pair %s: %w", pair.Id, err)
		}
	}
	return nil
}

// pairProvablySame checks one pair by fingerprint and strict title match
// (library.ProvablySame), then by pictures.
func (s *Sweeper) pairProvablySame(ctx context.Context, pair schema.Alike, all map[string]alike.Pictures, covers map[uint64]int) (bool, error) {
	same, err := s.library.ProvablySame(ctx, schema.RecordKindVideo, pair.FirstId, pair.SecondId)
	if err != nil || same {
		return same, err
	}
	first, err := s.library.Video(ctx, pair.FirstId)
	if err != nil || first == nil {
		return false, err
	}
	second, err := s.library.Video(ctx, pair.SecondId)
	if err != nil || second == nil {
		return false, err
	}
	firstCandidate, err := s.library.AlikeCandidate(ctx, first)
	if err != nil {
		return false, err
	}
	secondCandidate, err := s.library.AlikeCandidate(ctx, second)
	if err != nil {
		return false, err
	}
	return provablySame(firstCandidate, secondCandidate, all[first.Id], all[second.Id], covers), nil
}

// provablySame is the merge rule: the strict title matcher, or the pictures.
func provablySame(first, second alike.Candidate, firstPictures, secondPictures alike.Pictures, covers map[uint64]int) bool {
	if alike.SameSceneStrict(first, second) {
		return true
	}
	return alike.SameSceneByImages(first, second, firstPictures, secondPictures, coversUsable(covers, firstPictures, secondPictures))
}

// older returns the survivor (the record stored first) and the one merged into it.
func older(first, second schema.Video) (string, string) {
	if second.CreatedAt < first.CreatedAt {
		return second.Id, first.Id
	}
	return first.Id, second.Id
}

// encodePictures writes pictures as "<cover hex>|<frame hex>,…".
func encodePictures(pictures alike.Pictures) string {
	frames := make([]string, 0, len(pictures.Frames))
	for _, frame := range pictures.Frames {
		frames = append(frames, strconv.FormatUint(frame, 16))
	}
	return strconv.FormatUint(pictures.Cover, 16) + "|" + strings.Join(frames, ",")
}

// decodePictures reads what encodePictures wrote; unreadable parts are 0.
func decodePictures(value string) alike.Pictures {
	coverText, framesText, _ := strings.Cut(value, "|")
	pictures := alike.Pictures{}
	pictures.Cover, _ = strconv.ParseUint(coverText, 16, 64)
	for _, text := range strings.Split(framesText, ",") {
		frame, err := strconv.ParseUint(text, 16, 64)
		if err == nil && frame != 0 {
			pictures.Frames = append(pictures.Frames, frame)
		}
	}
	return pictures
}
