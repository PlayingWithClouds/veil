package recommend

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// signal is one weighted user interaction with a video or gallery.
type signal struct {
	item   schema.RecordRef
	weight float64
	at     time.Time
	// disliked marks an explicit thumbs-down or low rating, which also
	// excludes the item.
	disliked bool
	// weak marks a signal that says nothing about the item's content (shown
	// but not opened), so the runtime preference leaves it out.
	weak bool
	// explicit marks a rating or favourite, which never decays below
	// explicitDecayFloor.
	explicit bool
	// sessionEnd marks a watch nothing else followed; stopSeconds is where
	// its playback was left.
	sessionEnd  bool
	stopSeconds float64
}

// entitySignal is a verdict on a performer, studio or tag itself: a like,
// favourite or rating.
type entitySignal struct {
	kind   schema.RecordKind
	id     string
	weight float64
	at     time.Time
}

// signals is everything the profile is built from.
type signals struct {
	items    []signal
	entities []entitySignal
	// watches are the video watches of the window, for the watched penalty.
	watches []watch
	// ignored counts the showings no open followed, per item.
	ignored map[schema.RecordRef]ignoredImpression
}

// ignoredImpression counts the showings of one item that no open followed.
type ignoredImpression struct {
	count     int
	lastShown time.Time
}

// fixedWeightQueries are events that weigh the same every time: (kind, id,
// at) rows since $since, newest first.
var fixedWeightQueries = []struct {
	eventKind schema.EventKind
	weight    float64
}{
	{schema.EventKindO, weightO},
	{schema.EventKindChoice, weightChoice},
	{schema.EventKindSkip, weightSkip},
}

// fixedWeightQuery reads one event kind's media subjects since $since.
const fixedWeightQuery = `SELECT "subjectKind", "subjectId", "at" FROM "Event"
	WHERE "kind" = $eventKind AND "at" >= $since AND "deletedAt" IS NULL
	  AND "subjectKind" IN (SELECT value FROM json_each($mediaKinds))
	ORDER BY "at" DESC LIMIT $limit`

// collectSignals reads every interaction inside the signal window.
func (e *Engine) collectSignals(ctx context.Context, now time.Time) (*signals, error) {
	since := db.FormatTime(now.Add(-signalWindow))
	collected := &signals{}
	for _, source := range fixedWeightQueries {
		batch, err := e.fixedWeightSignals(ctx, source.eventKind, source.weight, since)
		if err != nil {
			return nil, err
		}
		collected.items = append(collected.items, batch...)
	}
	steps := []func(context.Context, *signals, string, time.Time) error{
		e.likeSignals,
		e.stateSignals,
		e.saveSignals,
		e.galleryViewSignals,
		e.watchSignals,
		e.ignoredImpressionSignals,
	}
	for _, step := range steps {
		if err := step(ctx, collected, since, now); err != nil {
			return nil, err
		}
	}
	return collected, nil
}

// fixedWeightSignals gives every event of one kind the same weight.
func (e *Engine) fixedWeightSignals(ctx context.Context, eventKind schema.EventKind, weight float64, since string) ([]signal, error) {
	rows, err := e.database.Query(ctx, fixedWeightQuery,
		db.Vars{"eventKind": eventKind, "since": since, "mediaKinds": mediaKinds, "limit": eventReadLimit})
	if err != nil {
		return nil, fmt.Errorf("read %s signals: %w", eventKind, err)
	}
	out := make([]signal, 0, len(rows))
	for _, row := range rows {
		out = append(out, signal{item: rowRef(row, "subjectKind", "subjectId"), weight: weight, at: rowTime(row, "at")})
	}
	return out, nil
}

// likeSignals turns like events into likes and dislikes of media, and into
// direct verdicts on performers, studios and tags.
func (e *Engine) likeSignals(ctx context.Context, collected *signals, since string, now time.Time) error {
	rows, err := e.database.Query(ctx,
		`SELECT "subjectKind", "subjectId", "at", json_extract("data", '`+pathValue+`') AS "value" FROM "Event"
		 WHERE "kind" = 'like' AND "at" >= $since AND "deletedAt" IS NULL AND "subjectKind" IS NOT NULL
		 ORDER BY "at" DESC LIMIT $limit`,
		db.Vars{"since": since, "limit": eventReadLimit})
	if err != nil {
		return fmt.Errorf("read like signals: %w", err)
	}
	for _, row := range rows {
		weight := weightLike
		if rowFloat(row, "value") < 0 {
			weight = weightDislike
		}
		addVerdict(collected, rowRef(row, "subjectKind", "subjectId"), weight, rowTime(row, "at"), false)
	}
	return nil
}

// addVerdict adds a like, dislike, rating or favourite: on media as an item
// signal (negative ones exclude the item), on anything else as an entity
// signal.
func addVerdict(collected *signals, ref schema.RecordRef, weight float64, at time.Time, explicit bool) {
	if isMediaKind(ref.Kind) {
		disliked := weight <= weightDislike
		collected.items = append(collected.items, signal{item: ref, weight: weight, at: at, disliked: disliked, explicit: explicit})
		return
	}
	if isEntityKind(ref.Kind) {
		collected.entities = append(collected.entities, entitySignal{kind: ref.Kind, id: ref.Id, weight: weight, at: at})
	}
}

// isMediaKind reports whether the feed recommends records of kind.
func isMediaKind(kind schema.RecordKind) bool {
	return kind == schema.RecordKindVideo || kind == schema.RecordKindGallery
}

// isEntityKind reports whether kind is a taste dimension of its own.
func isEntityKind(kind schema.RecordKind) bool {
	return kind == schema.RecordKindPerformer || kind == schema.RecordKindStudio || kind == schema.RecordKindTag
}

// stateSignals turns ratings and favourites (UserState) into verdicts. They
// are read regardless of the window and dated by the state's last change.
func (e *Engine) stateSignals(ctx context.Context, collected *signals, since string, now time.Time) error {
	rows, err := e.database.Query(ctx,
		`SELECT "subjectKind", "subjectId", "rating", "favorite", "updatedAt" FROM "UserState"
		 WHERE "deletedAt" IS NULL AND ("rating" IS NOT NULL OR "favorite" = 1)
		   AND "subjectKind" IN ('video', 'gallery', 'performer', 'studio', 'tag')
		 ORDER BY "updatedAt" DESC LIMIT $limit`,
		db.Vars{"limit": eventReadLimit})
	if err != nil {
		return fmt.Errorf("read rating signals: %w", err)
	}
	for _, row := range rows {
		ref := rowRef(row, "subjectKind", "subjectId")
		at := rowTime(row, "updatedAt")
		if db.AsBool(row["favorite"]) {
			addVerdict(collected, ref, weightFavorite, at, true)
		}
		if row["rating"] != nil {
			addRating(collected, ref, rowFloat(row, "rating"), at)
		}
	}
	return nil
}

// addRating adds a rating as a verdict growing linearly from the neutral
// rating; a rating under ratingDislikeBelow is a dislike.
func addRating(collected *signals, ref schema.RecordRef, rating float64, at time.Time) {
	if rating < ratingDislikeBelow {
		addVerdict(collected, ref, weightDislike, at, true)
		return
	}
	weight := weightLike * (rating - ratingNeutral) / ratingNeutral
	if weight == 0 {
		return
	}
	addVerdict(collected, ref, weight, at, true)
}

// saveSignals counts media put in the user's own collections and watchlist.
func (e *Engine) saveSignals(ctx context.Context, collected *signals, since string, now time.Time) error {
	rows, err := e.database.Query(ctx,
		`SELECT "item"."subjectKind", "item"."subjectId", "item"."createdAt" FROM "CollectionItem" AS "item"
		 JOIN "Collection" AS "collection" ON "collection"."id" = "item"."collection"
		 WHERE "collection"."kind" IN ('user', 'watchlist') AND "collection"."deletedAt" IS NULL
		   AND "item"."deletedAt" IS NULL AND "item"."createdAt" >= $since
		   AND "item"."subjectKind" IN (SELECT value FROM json_each($mediaKinds))
		 ORDER BY "item"."createdAt" DESC LIMIT $limit`,
		db.Vars{"since": since, "mediaKinds": mediaKinds, "limit": eventReadLimit})
	if err != nil {
		return fmt.Errorf("read save signals: %w", err)
	}
	for _, row := range rows {
		collected.items = append(collected.items, signal{item: rowRef(row, "subjectKind", "subjectId"), weight: weightSave, at: rowTime(row, "createdAt")})
	}
	return nil
}

// galleryViewSignals counts gallery views, one per gallery and day.
func (e *Engine) galleryViewSignals(ctx context.Context, collected *signals, since string, now time.Time) error {
	rows, err := e.database.Query(ctx,
		`SELECT "subjectId", MAX("at") AS "at" FROM "Event"
		 WHERE "kind" = 'view' AND "subjectKind" = 'gallery' AND "at" >= $since AND "deletedAt" IS NULL
		 GROUP BY "subjectId", substr("at", 1, 10) ORDER BY "at" DESC LIMIT $limit`,
		db.Vars{"since": since, "limit": eventReadLimit})
	if err != nil {
		return fmt.Errorf("read gallery views: %w", err)
	}
	for _, row := range rows {
		ref := schema.RecordRef{Kind: schema.RecordKindGallery, Id: rowString(row, "subjectId")}
		collected.items = append(collected.items, signal{item: ref, weight: weightGalleryView, at: rowTime(row, "at")})
	}
	return nil
}

// ignoredImpressionQuery counts, per item, the showings on the recommendation
// surfaces since $since that the user never engaged with afterwards.
const ignoredImpressionQuery = `SELECT "shown"."subjectKind", "shown"."subjectId", COUNT(*) AS "ignored", MAX("shown"."at") AS "lastShown"
	FROM "Event" AS "shown"
	WHERE "shown"."kind" = 'impression' AND "shown"."at" >= $since AND "shown"."deletedAt" IS NULL
	  AND "shown"."subjectKind" IN (SELECT value FROM json_each($mediaKinds))
	  AND json_extract("shown"."data", '` + pathSurface + `') IN (SELECT value FROM json_each($surfaces))
	  AND NOT EXISTS (SELECT 1 FROM "Event" AS "engaged"
	    WHERE "engaged"."subjectKind" = "shown"."subjectKind" AND "engaged"."subjectId" = "shown"."subjectId"
	      AND "engaged"."kind" IN ('choice', 'view', 'progress', 'heatSpan') AND "engaged"."at" >= "shown"."at"
	      AND "engaged"."deletedAt" IS NULL)
	GROUP BY "shown"."subjectKind", "shown"."subjectId"
	ORDER BY "lastShown" DESC LIMIT $limit`

// ignoredImpressionSignals turns "shown but not opened" counts into weak
// negative signals, capped per item, and keeps the counts for fatigue.
func (e *Engine) ignoredImpressionSignals(ctx context.Context, collected *signals, since string, now time.Time) error {
	rows, err := e.database.Query(ctx, ignoredImpressionQuery, db.Vars{
		"since":      db.FormatTime(now.Add(-ignoredImpressionWindow)),
		"mediaKinds": mediaKinds,
		"surfaces":   recommendationSurfaces,
		"limit":      eventReadLimit,
	})
	if err != nil {
		return fmt.Errorf("read ignored impressions: %w", err)
	}
	collected.ignored = make(map[schema.RecordRef]ignoredImpression, len(rows))
	for _, row := range rows {
		ref := rowRef(row, "subjectKind", "subjectId")
		impression := ignoredImpression{count: rowInt(row, "ignored"), lastShown: rowTime(row, "lastShown")}
		collected.ignored[ref] = impression
		weight := weightIgnoredImpression * float64(min(impression.count, ignoredImpressionCap))
		collected.items = append(collected.items, signal{item: ref, weight: weight, at: impression.lastShown, weak: true})
	}
	return nil
}

// signalItems lists the distinct items the signals are about.
func signalItems(collected *signals) []schema.RecordRef {
	seen := map[schema.RecordRef]bool{}
	out := []schema.RecordRef{}
	for _, entry := range collected.items {
		if !seen[entry.item] {
			seen[entry.item] = true
			out = append(out, entry.item)
		}
	}
	return out
}
