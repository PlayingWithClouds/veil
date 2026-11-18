package automations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// ActionFullLength searches the sites for the full-length version of the
// user's favourite and collected clips.
const ActionFullLength = "fullLength"

// FullLengthHuntID is the fixed id of the built-in full-length hunt.
const FullLengthHuntID = "7a2c4e61-3b9d-4f8a-a1c5-6e0d2b8f9c37"

// fullLengthHuntCron runs the full-length hunt daily at 05:00.
const fullLengthHuntCron = "0 5 * * *"

// fullLengthBatch is how many clips one run searches for.
const fullLengthBatch = 30

// fullLengthClipSeconds is the runtime below which a kept video counts as a
// clip whose full version is worth looking for.
const fullLengthClipSeconds = 720

// fullLengthRetry is how long a clip whose full version wasn't found rests
// before it's searched for again.
const fullLengthRetry = 7 * 24 * time.Hour

// fullLengthKey prefixes the cache entry recording a clip was searched for:
// its value is the full version's id, or "" while none was found.
const fullLengthKey = "fulllength:"

// FullLengthKeeper keeps a found full version where the user keeps the clip.
// Main supplies it from the user service.
type FullLengthKeeper interface {
	SetFavorite(ctx context.Context, kind schema.RecordKind, id string, favorite bool) (*schema.UserState, error)
	AddItems(ctx context.Context, collectionID string, items []schema.RecordRef, position *int) error
}

// FullLengthAction is the fullLength action over keeper. With no input it
// takes the user's favourite or collected clips (videos shorter than
// fullLengthClipSeconds or titled trailer, teaser or preview) not searched
// for lately; for each it searches every site for the full version
// (alike.FullLengthOf), and when found favourites it like the clip, adds it to
// the clip's collections right after it and notifies the user. A dry run
// only counts. Output: the full versions found.
func (s *Service) FullLengthAction(keeper FullLengthKeeper) Action {
	return func(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error) {
		clips, err := s.keptClips(ctx, input)
		if err != nil {
			return Result{}, err
		}
		if run.DryRun || s.searcher == nil {
			return Result{Output: []schema.RecordRef{}, Touched: []schema.RecordRef{}, Report: fmt.Sprintf("would search for %d clips", len(clips))}, nil
		}
		result := Result{Output: []schema.RecordRef{}, Touched: []schema.RecordRef{}}
		for _, clip := range clips {
			found, err := s.fullVersionOf(ctx, clip)
			if err != nil {
				return result, err
			}
			if err := s.rememberSearched(ctx, clip.Id, found); err != nil {
				return result, err
			}
			if found == nil {
				continue
			}
			ref := schema.RecordRef{Kind: schema.RecordKindVideo, Id: found.Id}
			if err := s.keepLikeClip(ctx, keeper, clip.Id, ref); err != nil {
				return result, err
			}
			s.notifyFullLength(ctx, clip, ref)
			result.Output = append(result.Output, ref)
			result.Touched = append(result.Touched, ref)
		}
		result.Report = fmt.Sprintf("found the full version of %d of %d clips", len(result.Output), len(clips))
		return result, nil
	}
}

// keptClipsQuery selects favourite or collected clips not searched for
// lately ($now in unix seconds, as CacheEntry stores expiry).
const keptClipsQuery = `SELECT "video".* FROM "Video" AS "video"
	WHERE "video"."deletedAt" IS NULL AND "video"."status" != 'wanted'
	  AND ("video"."durationSeconds" < $clipSeconds OR lower("video"."title") LIKE '%trailer%'
	       OR lower("video"."title") LIKE '%teaser%' OR lower("video"."title") LIKE '%preview%')
	  AND (EXISTS (SELECT 1 FROM "UserState" AS "state" WHERE "state"."subjectKind" = 'video'
	         AND "state"."subjectId" = "video"."id" AND "state"."favorite" = 1 AND "state"."deletedAt" IS NULL)
	    OR EXISTS (SELECT 1 FROM "CollectionItem" AS "item" JOIN "Collection" AS "collection" ON "collection"."id" = "item"."collection"
	         WHERE "item"."subjectKind" = 'video' AND "item"."subjectId" = "video"."id"
	           AND "item"."deletedAt" IS NULL AND "collection"."deletedAt" IS NULL AND "collection"."kind" != 'queue'))
	  AND NOT EXISTS (SELECT 1 FROM "CacheEntry" WHERE "key" = $prefix || "video"."id"
	         AND ("expiresAt" IS NULL OR "expiresAt" > $now))`

// keptClips loads the clips among refs, or the next batch of kept clips
// when refs is empty.
func (s *Service) keptClips(ctx context.Context, refs []schema.RecordRef) ([]schema.Video, error) {
	query := keptClipsQuery + ` ORDER BY "video"."createdAt" DESC LIMIT $limit`
	vars := db.Vars{"clipSeconds": fullLengthClipSeconds, "prefix": fullLengthKey, "now": s.now().Unix(), "limit": fullLengthBatch}
	if len(refs) > 0 {
		query = keptClipsQuery + ` AND "video"."id" IN (SELECT value FROM json_each($ids)) LIMIT $limit`
		vars["ids"] = idsOfKind(refs, schema.RecordKindVideo)
	}
	return db.QueryAs[schema.Video](ctx, s.database, query, vars)
}

// fullVersionOf searches the sites for the clip's full version.
func (s *Service) fullVersionOf(ctx context.Context, clip schema.Video) (*schema.Video, error) {
	target, err := s.candidate(ctx, clip)
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(searchableTitle(clip.Title) + " " + strings.Join(target.Performers, " "))
	for _, ref := range s.searcher.Search(ctx, plugins.MediaTypeScene, query, huntSearchLimit) {
		if ref.Kind != schema.RecordKindVideo || ref.Id == clip.Id {
			continue
		}
		video, err := s.library.Video(ctx, ref.Id)
		if err != nil || video == nil {
			return nil, err
		}
		candidate, err := s.candidate(ctx, *video)
		if err != nil {
			return nil, err
		}
		if alike.FullLengthOf(target, candidate) {
			return video, nil
		}
	}
	return nil, nil
}

// searchableTitle drops the words that mark a clip from its title, so the
// search finds the full scene rather than more trailers.
func searchableTitle(title string) string {
	words := strings.Fields(title)
	kept := make([]string, 0, len(words))
	for _, word := range words {
		switch strings.ToLower(strings.Trim(word, "()[]-:|")) {
		case "trailer", "teaser", "preview", "clip", "excerpt":
			continue
		}
		kept = append(kept, word)
	}
	return strings.Join(kept, " ")
}

// rememberSearched records the search: for good once found, else so the clip
// rests for fullLengthRetry.
func (s *Service) rememberSearched(ctx context.Context, clipID string, found *schema.Video) error {
	value := ""
	var expires any = s.now().Add(fullLengthRetry).Unix()
	if found != nil {
		value = found.Id
		expires = nil
	}
	_, err := s.database.Exec(ctx,
		`INSERT INTO "CacheEntry" ("key", "value", "expiresAt") VALUES ($key, $value, $expires)
		 ON CONFLICT ("key") DO UPDATE SET "value" = excluded."value", "expiresAt" = excluded."expiresAt"`,
		db.Vars{"key": fullLengthKey + clipID, "value": value, "expires": expires})
	return err
}

// keepLikeClip favourites the full version when the clip is a favourite and
// adds it right after the clip in each collection holding the clip.
func (s *Service) keepLikeClip(ctx context.Context, keeper FullLengthKeeper, clipID string, full schema.RecordRef) error {
	favourite, err := s.database.Int(ctx,
		`SELECT COUNT(*) FROM "UserState" WHERE "subjectKind" = 'video' AND "subjectId" = $id AND "favorite" = 1 AND "deletedAt" IS NULL`,
		db.Vars{"id": clipID})
	if err != nil {
		return err
	}
	if favourite > 0 {
		if _, err := keeper.SetFavorite(ctx, schema.RecordKindVideo, full.Id, true); err != nil {
			return err
		}
	}
	rows, err := s.database.Query(ctx,
		`SELECT "collection", "position" FROM "CollectionItem"
		 WHERE "subjectKind" = 'video' AND "subjectId" = $id AND "deletedAt" IS NULL`,
		db.Vars{"id": clipID})
	if err != nil {
		return err
	}
	for _, row := range rows {
		collectionID, _ := row["collection"].(string)
		position := rowInt(row["position"]) + 1
		if err := keeper.AddItems(ctx, collectionID, []schema.RecordRef{full}, &position); err != nil {
			return err
		}
	}
	return nil
}

// rowInt reads an integer column that SQLite returns as int64.
func rowInt(value any) int {
	number, _ := value.(int64)
	return int(number)
}

// notifyFullLength tells the user a clip's full version was found.
func (s *Service) notifyFullLength(ctx context.Context, clip schema.Video, full schema.RecordRef) {
	if s.notifier == nil {
		return
	}
	notification := notify.NewNotification{Kind: "fullLengthFound", Title: "Full version found: " + clip.Title, Subject: &full}
	if _, err := s.notifier.Notify(ctx, notification); err != nil {
		log.Printf("automations: notify full length of %s: %v", clip.Id, err)
	}
}
