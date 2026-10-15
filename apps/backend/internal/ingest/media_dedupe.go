package ingest

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// keeperOrder ranks the records of a table by how much they are worth keeping
// when several are one thing: what the user did with them first (history,
// rating, saves, memberships), then whether the page was fetched or the images
// are there, then how many streams, then age.
var keeperOrder = map[string]string{
	"scene": `(SELECT count(*) FROM watch_history WHERE media = scene.id)
		+ (SELECT count(*) FROM user_rating WHERE media = scene.id)
		+ (SELECT count(*) FROM watchlist WHERE media = scene.id)
		+ (SELECT count(*) FROM collection_item WHERE media = scene.id)
		+ (SELECT count(*) FROM o_event WHERE media = scene.id) DESC,
		(detail_fetched_at IS NOT NULL) DESC,
		(SELECT count(*) FROM stream WHERE media = scene.id) DESC, created_at, id`,
	"gallery": `(SELECT count(*) FROM user_rating WHERE media = gallery.id)
		+ (SELECT count(*) FROM collection_item WHERE media = gallery.id) DESC,
		(SELECT count(*) FROM image WHERE media = gallery.id AND content) DESC, created_at, id`,
}

// duplicateGroup is a record to keep and the records that are the same thing.
type duplicateGroup struct {
	survivor string
	merged   []string
}

// MergeDuplicateMedia merges the scenes, galleries and collections that several
// sites carry into one record each, keeping the one the user has used most.
// Returns how many records were merged away.
func (s *Service) MergeDuplicateMedia(ctx context.Context) (int, error) {
	total := 0
	for _, merge := range []func(context.Context) (int, error){
		s.mergeDuplicateScenes, s.mergeDuplicateGalleries, s.mergeDuplicateCollections,
	} {
		count, err := merge(ctx)
		total += count
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// PerformerFinder returns a finder over every performer's name and aliases.
func (s *Service) PerformerFinder(ctx context.Context) (*names.Finder, error) {
	rows, err := s.database.Query(ctx, `SELECT id, name, aliases FROM performer`, nil)
	if err != nil {
		return nil, fmt.Errorf("load performer names: %w", err)
	}
	finder := names.NewFinder()
	for _, row := range rows {
		id, _ := row["id"].(string)
		name, _ := row["name"].(string)
		finder.Add(id, name)
		for _, alias := range jsonStringList(row["aliases"]) {
			finder.Add(id, alias)
		}
	}
	return finder, nil
}

// storedCandidate is the matcher's view of a scene or gallery row. A record that
// credits no one is given the known performers its title names, which can only
// keep two of them apart, never join them.
func storedCandidate(row db.Row, finder *names.Finder) alike.Candidate {
	candidate := alike.Candidate{
		Title:      rowText(row, "title"),
		Performers: jsonStringList(row["performers"]),
	}
	if seconds, ok := row["duration_seconds"].(int64); ok {
		candidate.DurationSeconds = int(seconds)
	}
	if len(candidate.Performers) == 0 {
		candidate.Named = finder.Find(candidate.Title)
	}
	return candidate
}

// rowText reads a text column, "" when absent.
func rowText(row db.Row, column string) string {
	text, _ := row[column].(string)
	return text
}

// RecordingSourcePrefix starts the source_url of a scene made from a live cam
// recording. Recordings are never merged as duplicates: two sessions of one
// model share performer and often runtime, but are different footage.
const RecordingSourcePrefix = "veil-recording:"

// mergeDuplicateScenes merges scenes that are the same video (alike.SameSceneStrict).
func (s *Service) mergeDuplicateScenes(ctx context.Context) (int, error) {
	rows, err := s.database.Query(ctx,
		`SELECT id, title, performers, duration_seconds FROM scene WHERE duration_seconds > 0 AND source_url NOT LIKE '`+RecordingSourcePrefix+`%' ORDER BY `+keeperOrder["scene"], nil)
	if err != nil {
		return 0, fmt.Errorf("load scenes to deduplicate: %w", err)
	}
	finder, err := s.PerformerFinder(ctx)
	if err != nil {
		return 0, err
	}
	candidates := make([]alike.Candidate, len(rows))
	for index, row := range rows {
		candidates[index] = storedCandidate(row, finder)
	}
	groups := groupDuplicates(rows, func(first, second int) bool {
		return alike.SameSceneStrict(candidates[first], candidates[second])
	})
	return s.mergeGroups(ctx, "scene", groups)
}

// mergeDuplicateGalleries merges galleries that are the same image set
// (alike.SameGallery).
func (s *Service) mergeDuplicateGalleries(ctx context.Context) (int, error) {
	rows, err := s.database.Query(ctx,
		`SELECT id, title, performers, image_count FROM gallery WHERE image_count > 0 ORDER BY `+keeperOrder["gallery"], nil)
	if err != nil {
		return 0, fmt.Errorf("load galleries to deduplicate: %w", err)
	}
	finder, err := s.PerformerFinder(ctx)
	if err != nil {
		return 0, err
	}
	candidates := make([]alike.Candidate, len(rows))
	imageCounts := make([]int, len(rows))
	for index, row := range rows {
		candidates[index] = storedCandidate(row, finder)
		if count, ok := row["image_count"].(int64); ok {
			imageCounts[index] = int(count)
		}
	}
	groups := groupDuplicates(rows, func(first, second int) bool {
		return alike.SameGallery(candidates[first], candidates[second], imageCounts[first], imageCounts[second])
	})
	return s.mergeGroups(ctx, "gallery", groups)
}

// groupDuplicates walks rows (best to keep first) and puts each one with the
// first earlier kept row that same says it matches, so a chain of near matches
// never joins records that aren't the same.
func groupDuplicates(rows []db.Row, same func(first, second int) bool) []duplicateGroup {
	var groups []duplicateGroup
	var keptRows []int
	for index := range rows {
		matched := -1
		for groupIndex, keptRow := range keptRows {
			if same(keptRow, index) {
				matched = groupIndex
				break
			}
		}
		if matched < 0 {
			keptRows = append(keptRows, index)
			groups = append(groups, duplicateGroup{survivor: rowText(rows[index], "id")})
			continue
		}
		groups[matched].merged = append(groups[matched].merged, rowText(rows[index], "id"))
	}
	return groups
}

// mergeGroups merges every group's records into its survivor.
func (s *Service) mergeGroups(ctx context.Context, table string, groups []duplicateGroup) (int, error) {
	mergedCount := 0
	for _, group := range groups {
		for _, mergedID := range group.merged {
			if err := s.MergeMedia(ctx, table, group.survivor, mergedID); err != nil {
				return mergedCount, err
			}
			mergedCount++
		}
	}
	if mergedCount > 0 {
		log.Printf("ingest: merged %d duplicate %s records", mergedCount, table)
	}
	return mergedCount, nil
}

// mergeDuplicateCollections merges scraper-created collections that carry the
// same name and share a member, after scenes were merged: a series two sites
// list is one collection. User playlists are never touched.
func (s *Service) mergeDuplicateCollections(ctx context.Context) (int, error) {
	rows, err := s.database.Query(ctx,
		`SELECT id, name, (SELECT count(*) FROM collection_item WHERE collection = collection.id) AS members
		 FROM collection WHERE NOT user_created ORDER BY members DESC, created_at, id`, nil)
	if err != nil {
		return 0, fmt.Errorf("load collections to deduplicate: %w", err)
	}
	groups := groupDuplicates(rows, func(first, second int) bool {
		return names.CompactKey(rowText(rows[first], "name")) == names.CompactKey(rowText(rows[second], "name")) &&
			s.shareMember(ctx, rowText(rows[first], "id"), rowText(rows[second], "id"))
	})
	return s.mergeGroups(ctx, "collection", groups)
}

// shareMember reports whether two collections list a common record.
func (s *Service) shareMember(ctx context.Context, firstID, secondID string) bool {
	count, err := s.database.Int(ctx,
		`SELECT count(*) FROM collection_item AS first JOIN collection_item AS second ON second.media = first.media
		 WHERE first.collection = $first AND second.collection = $second`,
		db.Vars{"first": firstID, "second": secondID})
	return err == nil && count > 0
}

// mergeNewIntoTwin folds a record that this ingest just created into an
// existing record of the same thing, and points the result at the survivor, so
// a search that lists a video the library already has from another site adds
// nothing new.
func (s *Service) mergeNewIntoTwin(ctx context.Context, table string, result *IngestResult) {
	if !result.IsNew {
		return
	}
	survivorID, err := s.mergeIfDuplicate(ctx, table, result.CanonicalID)
	if err != nil {
		log.Printf("ingest: merge duplicate %s %s: %v", table, result.CanonicalID, err)
		return
	}
	if survivorID != result.CanonicalID {
		result.CanonicalID = survivorID
		result.IsNew = false
	}
}

// mergeIfDuplicate merges the freshly created scene or gallery recordID into an
// existing record of the same thing and returns the id that now holds it, or
// recordID itself when it is new. Existing records win, so what the user has
// already done with them stays.
func (s *Service) mergeIfDuplicate(ctx context.Context, table, recordID string) (string, error) {
	existingID, err := s.findExistingTwin(ctx, table, recordID)
	if err != nil || existingID == "" {
		return recordID, err
	}
	if err := s.MergeMedia(ctx, table, existingID, recordID); err != nil {
		return recordID, err
	}
	return existingID, nil
}

// findExistingTwin returns the best-kept other record that is the same scene or
// gallery as recordID, or "".
func (s *Service) findExistingTwin(ctx context.Context, table, recordID string) (string, error) {
	finder, err := s.PerformerFinder(ctx)
	if err != nil {
		return "", err
	}
	row, err := s.database.GetString(ctx, recordID)
	if err != nil || row == nil {
		return "", err
	}
	if strings.HasPrefix(rowText(row, "source_url"), RecordingSourcePrefix) {
		return "", nil
	}
	candidate := storedCandidate(row, finder)
	if table == "gallery" {
		return s.findGalleryTwin(ctx, recordID, row, candidate, finder)
	}
	return s.findSceneTwin(ctx, recordID, candidate, finder)
}

// findSceneTwin looks among the scenes with a nearby runtime.
func (s *Service) findSceneTwin(ctx context.Context, recordID string, candidate alike.Candidate, finder *names.Finder) (string, error) {
	if candidate.DurationSeconds <= 0 {
		return "", nil
	}
	rows, err := s.database.Query(ctx,
		`SELECT id, title, performers, duration_seconds FROM scene
		 WHERE id <> $id AND duration_seconds BETWEEN $low AND $high
		   AND source_url NOT LIKE '`+RecordingSourcePrefix+`%' ORDER BY `+keeperOrder["scene"],
		db.Vars{"id": recordID, "low": candidate.DurationSeconds * 90 / 100, "high": candidate.DurationSeconds * 112 / 100})
	if err != nil {
		return "", fmt.Errorf("find scene twins: %w", err)
	}
	return firstMatch(rows, finder, func(other alike.Candidate) bool { return alike.SameSceneStrict(candidate, other) }), nil
}

// findGalleryTwin looks among the galleries with the same image count.
func (s *Service) findGalleryTwin(ctx context.Context, recordID string, row db.Row, candidate alike.Candidate, finder *names.Finder) (string, error) {
	imageCount, _ := row["image_count"].(int64)
	if imageCount <= 0 {
		return "", nil
	}
	rows, err := s.database.Query(ctx,
		`SELECT id, title, performers, image_count FROM gallery
		 WHERE id <> $id AND image_count = $count ORDER BY `+keeperOrder["gallery"],
		db.Vars{"id": recordID, "count": imageCount})
	if err != nil {
		return "", fmt.Errorf("find gallery twins: %w", err)
	}
	return firstMatch(rows, finder, func(other alike.Candidate) bool {
		return alike.SameGallery(candidate, other, int(imageCount), int(imageCount))
	}), nil
}

// firstMatch returns the id of the first row (in the order given) that matches.
func firstMatch(rows []db.Row, finder *names.Finder, matches func(alike.Candidate) bool) string {
	for _, row := range rows {
		if matches(storedCandidate(row, finder)) {
			return rowText(row, "id")
		}
	}
	return ""
}
