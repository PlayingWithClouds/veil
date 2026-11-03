package recommend

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// camWatchQuery reads, per cam room and session, the longest the room
// played (see events.go: progress events carry how long it has played).
const camWatchQuery = `SELECT "subjectId", COALESCE("sessionId", substr("at", 1, 10)) AS "session",
	MAX(CAST(json_extract("data", '` + pathWatched + `') AS REAL)) AS "watched", MAX("at") AS "at"
	FROM "Event"
	WHERE "kind" = 'progress' AND "subjectKind" = $camModel AND "at" >= $since AND "deletedAt" IS NULL
	GROUP BY "subjectId", "session" ORDER BY "at" DESC LIMIT $limit`

// camChoiceQuery reads the cam room cards picked.
const camChoiceQuery = `SELECT "subjectId", "at" FROM "Event"
	WHERE "kind" = 'choice' AND "subjectKind" = $camModel AND "at" >= $since AND "deletedAt" IS NULL
	ORDER BY "at" DESC LIMIT $limit`

// camKeptQuery reads the rooms the user follows or auto-records, dated by
// when they started to.
const camKeptQuery = `SELECT "targetId" AS "id", "createdAt" AS "at" FROM "Follow"
	WHERE "targetKind" = $camModel AND "enabled" = 1 AND "deletedAt" IS NULL
	UNION ALL
	SELECT "id", "updatedAt" AS "at" FROM "CamModel" WHERE "autoRecord" = 1 AND "deletedAt" IS NULL`

// camBlockedQuery reads the rooms the user blocked.
const camBlockedQuery = `SELECT "subjectId" FROM "UserState"
	WHERE "subjectKind" = $camModel AND "blocked" = 1 AND "deletedAt" IS NULL`

// camSignals reads what the user did with cam rooms: how long they watched
// each (a room left at once counts against it), which cards they picked, and
// which rooms they follow or auto-record. Blocked rooms say nothing.
func (e *Engine) camSignals(ctx context.Context, collected *signals, since string, now time.Time) error {
	vars := db.Vars{"camModel": schema.RecordKindCamModel, "since": since, "limit": eventReadLimit}
	blocked, err := e.database.Strings(ctx, camBlockedQuery, vars)
	if err != nil {
		return fmt.Errorf("read blocked cam rooms: %w", err)
	}
	blockedSet := toSet(blocked)
	watches, err := e.database.Query(ctx, camWatchQuery, vars)
	if err != nil {
		return fmt.Errorf("read cam watches: %w", err)
	}
	for _, row := range watches {
		addCamSignal(collected, blockedSet, row, "subjectId", camWatchWeight(rowFloat(row, "watched")), false)
	}
	choices, err := e.database.Query(ctx, camChoiceQuery, vars)
	if err != nil {
		return fmt.Errorf("read cam choices: %w", err)
	}
	for _, row := range choices {
		addCamSignal(collected, blockedSet, row, "subjectId", weightChoice, false)
	}
	kept, err := e.database.Query(ctx, camKeptQuery, vars)
	if err != nil {
		return fmt.Errorf("read kept cam rooms: %w", err)
	}
	for _, row := range kept {
		addCamSignal(collected, blockedSet, row, "id", weightFavorite, true)
	}
	return nil
}

// addCamSignal adds one signal on the room a row names, unless it is blocked.
func addCamSignal(collected *signals, blocked map[string]bool, row db.Row, idColumn string, weight float64, explicit bool) {
	id := rowString(row, idColumn)
	if id == "" || blocked[id] {
		return
	}
	ref := schema.RecordRef{Kind: schema.RecordKindCamModel, Id: id}
	collected.cams = append(collected.cams, signal{item: ref, weight: weight, at: rowTime(row, "at"), explicit: explicit})
}

// camWatchWeight weighs a session of watching a room: a room left within
// camQuickLeaveSeconds counts against it, a longer watch for it, growing with
// the log of its minutes.
func camWatchWeight(watchedSeconds float64) float64 {
	if watchedSeconds < camQuickLeaveSeconds {
		return weightCamQuickLeave
	}
	return math.Min(weightCamWatch*math.Log1p(watchedSeconds/60), weightCamWatchCap)
}

// camRoom is what the taste reads of a cam room.
type camRoom struct {
	id        string
	name      string
	site      string
	gender    string
	performer string
	// tags are spelled as the site spells them.
	tags []string
}

// camInputs describes the signalled cam rooms.
type camInputs struct {
	rooms map[string]*camRoom
	// tagIDs maps names.CompactKey of every library tag name to its id, so
	// a room's "bigboobs" lands on the library's "Big Boobs".
	tagIDs map[string]string
}

// loadCamInputs reads the signalled rooms and, when there are any, the
// library's tag names.
func (e *Engine) loadCamInputs(ctx context.Context, camSignals []signal) (*camInputs, error) {
	inputs := &camInputs{rooms: map[string]*camRoom{}, tagIDs: map[string]string{}}
	ids := []string{}
	for _, entry := range camSignals {
		ids = append(ids, entry.item.Id)
	}
	if len(ids) == 0 {
		return inputs, nil
	}
	rooms, err := e.database.Query(ctx,
		`SELECT "id", "displayName", "plugin", "gender", "performer", "tags" FROM "CamModel"
		 WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("load cam rooms: %w", err)
	}
	for _, row := range rooms {
		room := &camRoom{
			id: rowString(row, "id"), name: rowString(row, "displayName"), site: rowString(row, "plugin"),
			gender: rowString(row, "gender"), performer: rowString(row, "performer"), tags: rowStrings(row, "tags"),
		}
		inputs.rooms[room.id] = room
	}
	tags, err := e.database.Query(ctx, `SELECT "id", "name" FROM "Tag" WHERE "deletedAt" IS NULL`, nil)
	if err != nil {
		return nil, fmt.Errorf("load tag names: %w", err)
	}
	for _, row := range tags {
		inputs.tagIDs[names.CompactKey(rowString(row, "name"))] = rowString(row, "id")
	}
	return inputs, nil
}

// addCamSignals adds the decayed weight of each signalled room to the room
// itself and its gender, and camTasteFactor of it to its performer, site and
// tags, the way addItem spreads a video's.
func (taste *profile) addCamSignals(camSignals []signal, inputs *camInputs, now time.Time) {
	if inputs == nil {
		return
	}
	for ref, weight := range decayedItemWeights(camSignals, now, false) {
		room, found := inputs.rooms[ref.Id]
		if !found {
			continue
		}
		taste.camRooms[room.id] += weight
		if room.gender != "" {
			taste.camGenders[room.gender] += weight
		}
		taste.addCamRoom(room, camTasteFactor*weight, inputs.tagIDs)
	}
}

// addCamRoom spreads a room's weight onto its performer, site and tags; a
// tag no library tag is spelled like goes to camTags.
func (taste *profile) addCamRoom(room *camRoom, weight float64, tagIDs map[string]string) {
	if room.performer != "" {
		taste.performers[room.performer] += weight
	}
	if room.site != "" {
		taste.sites[room.site] += weight
	}
	if len(room.tags) == 0 {
		return
	}
	tagShare := weight / math.Sqrt(float64(len(room.tags)))
	for _, tag := range room.tags {
		key := names.CompactKey(tag)
		if key == "" {
			continue
		}
		tagID, known := tagIDs[key]
		if !known {
			taste.camTags[key] += tagShare
			continue
		}
		taste.tags[tagID] += tagShare * taste.rarity.weight(tagID)
	}
}
