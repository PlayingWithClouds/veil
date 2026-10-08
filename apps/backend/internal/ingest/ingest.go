// Package ingest implements the observation-based ingestion pipeline.
//
// Flow:
//  1. Plugin returns a ScrapeResult (scene / performer / studio / tag / gallery).
//  2. Resolve the canonical record by its stable identity (source_url for
//     scenes/galleries/performers, name for studios/tags), merging into an
//     existing record or creating a new one.
//  3. Resolve name-based relations (studio, performers, tags) into record links
//     and, for performers, a performs_in graph edge.
//  4. Store the raw payload as an observation and write one changes row per
//     field the observation filled in.
package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// Service handles writing observations, matching, and merging.
type Service struct {
	database *db.DB
}

func New(database *db.DB) *Service {
	return &Service{database: database}
}

// IngestResult describes what happened during ingestion of one observation.
type IngestResult struct {
	ObservationID string
	CanonicalID   string
	IsNew         bool    // true = new canonical record created
	Score         float64 // 0 if IsNew
	Mode          string  // "auto" | "new"
	// UnresolvedMemberURLs lists collection member source URLs that had no
	// matching record yet; the orchestrator enqueues scrape jobs for them.
	UnresolvedMemberURLs []string
}

// canonicalColumns whitelists the scalar columns each canonical table accepts.
// Relation fields (studio/performers/tags) and asset fields (images/downloads)
// are handled separately, not written from this map.
var canonicalColumns = map[string][]string{
	"scene": {
		"external_id", "source_url", "title", "details", "date",
		"duration_seconds", "rating", "view_count", "poster_path", "preview_video", "preview_images",
	},
	"gallery": {
		"external_id", "source_url", "title", "details", "date", "cover_path",
	},
	"performer": {
		"external_id", "source_url", "name", "aliases", "details", "gender",
		"birthdate", "death_date", "country", "ethnicity", "eye_color", "hair_color",
		"height_cm", "weight_kg", "measurements", "fake_tits", "tattoos", "piercings",
		"career_length", "url", "twitter", "instagram", "image_path",
	},
	"studio": {"external_id", "source_url", "name", "aliases", "url", "image_path", "details"},
	"tag":    {"external_id", "name", "aliases", "description", "category"},
	"image": {
		"external_id", "source_url", "title", "details", "date", "type",
		"file_path", "width", "height", "aspect_ratio", "rating", "content",
	},
	"collection": {"external_id", "source_url", "name", "details", "cover_path"},
}

// IngestScrapeResults processes a stream of ScrapeResult items in order. Each
// item is resolved independently; the first item's result is returned as the root.
func (s *Service) IngestScrapeResults(ctx context.Context, pluginName, sourceURL string, results []*plugins.ScrapeResult) (*IngestResult, error) {
	var firstResult *IngestResult
	for _, result := range results {
		var ingestResult *IngestResult
		var err error

		switch result.Type {
		case plugins.MediaTypeScene:
			if result.Scene == nil {
				continue
			}
			ingestResult, err = s.ingestScene(ctx, pluginName, sourceURL, result.Scene)

		case plugins.MediaTypePerformer:
			if result.Performer == nil {
				continue
			}
			ingestResult, err = s.ingestPerformer(ctx, pluginName, result.Performer)

		case plugins.MediaTypeStudio:
			if result.Studio == nil {
				continue
			}
			ingestResult, err = s.ingestStudio(ctx, pluginName, result.Studio)

		case plugins.MediaTypeTag:
			if result.Tag == nil {
				continue
			}
			ingestResult, err = s.ingestTag(ctx, pluginName, result.Tag)

		case plugins.MediaTypeGallery:
			if result.Gallery == nil {
				continue
			}
			ingestResult, err = s.ingestGallery(ctx, pluginName, sourceURL, result.Gallery)

		case plugins.MediaTypeImage:
			if result.Image == nil {
				continue
			}
			ingestResult, err = s.ingestImage(ctx, pluginName, result.Image)

		case plugins.MediaTypeCollection:
			if result.Collection == nil {
				continue
			}
			ingestResult, err = s.ingestCollection(ctx, pluginName, result.Collection)

		default:
			log.Printf("ingest: unknown media type %q — skipping", result.Type)
			continue
		}

		if err != nil {
			log.Printf("ingest: %s: %v", result.Type, err)
			continue
		}
		if firstResult == nil {
			firstResult = ingestResult
		} else if ingestResult != nil && len(ingestResult.UnresolvedMemberURLs) > 0 {
			firstResult.UnresolvedMemberURLs = append(firstResult.UnresolvedMemberURLs, ingestResult.UnresolvedMemberURLs...)
		}
	}

	if firstResult == nil {
		return nil, fmt.Errorf("no results ingested from %d items", len(results))
	}
	return firstResult, nil
}

// ingestScene upserts a scene by source_url, resolves its studio/performers/tags
// relations, records the observation, and syncs its stream sources.
func (s *Service) ingestScene(ctx context.Context, pluginName, sourceURL string, scene *plugins.Scene) (*IngestResult, error) {
	// The related list becomes stubs + edges below, not observation data.
	withoutRelated := *scene
	withoutRelated.Related = nil
	obs := s.baseObs(&withoutRelated, pluginName, sceneURL(scene, sourceURL))

	result, err := s.upsert(ctx, "scene", "source_url", sceneURL(scene, sourceURL), obs, pluginName)
	if err != nil {
		return nil, err
	}

	s.linkStudio(ctx, "scene", result.CanonicalID, scene.Studio, pluginName)
	s.linkPerformers(ctx, "scene", result.CanonicalID, scene.Performers, pluginName)
	s.linkTags(ctx, "scene", result.CanonicalID, scene.Tags, pluginName)
	s.syncStreamsLogged(ctx, result.CanonicalID, pluginName, scene.Downloads)
	s.syncSceneMarkersLogged(ctx, result.CanonicalID, pluginName, scene.Markers)
	if len(scene.Related) > 0 {
		if _, err := s.LinkRelated(ctx, result.CanonicalID, pluginName, RelatedSourceSite, scene.Related); err != nil {
			log.Printf("ingest: related for %s: %v", result.CanonicalID, err)
		}
	}
	return result, nil
}

// ingestGallery mirrors ingestScene for image sets.
func (s *Service) ingestGallery(ctx context.Context, pluginName, sourceURL string, gallery *plugins.Gallery) (*IngestResult, error) {
	url := gallery.SourceURL
	if url == "" {
		url = sourceURL
	}
	obs := s.baseObs(gallery, pluginName, url)

	result, err := s.upsert(ctx, "gallery", "source_url", url, obs, pluginName)
	if err != nil {
		return nil, err
	}

	s.linkStudio(ctx, "gallery", result.CanonicalID, gallery.Studio, pluginName)
	s.linkPerformers(ctx, "gallery", result.CanonicalID, gallery.Performers, pluginName)
	s.linkTags(ctx, "gallery", result.CanonicalID, gallery.Tags, pluginName)
	s.syncGalleryImagesLogged(ctx, result.CanonicalID, pluginName, gallery.Images)
	return result, nil
}

// ingestPerformer upserts a performer by source_url, then resolves its tags.
func (s *Service) ingestPerformer(ctx context.Context, pluginName string, performer *plugins.Performer) (*IngestResult, error) {
	identity, field := performer.SourceURL, "source_url"
	if identity == "" {
		identity, field = performer.Name, "name"
	}
	obs := s.baseObs(performer, pluginName, performer.SourceURL)

	result, err := s.upsert(ctx, "performer", field, identity, obs, pluginName)
	if err != nil {
		return nil, err
	}
	s.linkTags(ctx, "performer", result.CanonicalID, performer.Tags, pluginName)
	return result, nil
}

// ingestStudio upserts a studio by source_url when present, else by name. Tags
// and the parent studio are resolved to record links afterwards.
func (s *Service) ingestStudio(ctx context.Context, pluginName string, studio *plugins.Studio) (*IngestResult, error) {
	identity, field := studio.SourceURL, "source_url"
	if identity == "" {
		identity, field = studio.Name, "name"
	}
	obs := s.baseObs(studio, pluginName, studio.SourceURL)

	result, err := s.upsert(ctx, "studio", field, identity, obs, pluginName)
	if err != nil {
		return nil, err
	}
	s.linkTags(ctx, "studio", result.CanonicalID, studio.Tags, pluginName)
	s.linkParentStudio(ctx, result.CanonicalID, studio.Parent)
	return result, nil
}

// linkParentStudio resolves a parent studio name and links it, building the
// network → studio hierarchy.
func (s *Service) linkParentStudio(ctx context.Context, studioID, parentName string) {
	parentName = strings.TrimSpace(parentName)
	if parentName == "" {
		return
	}
	parentID, err := s.findOrCreateNamed(ctx, "studio", parentName)
	if err != nil {
		log.Printf("ingest: resolve parent studio %q: %v", parentName, err)
		return
	}
	if parentID.String() == studioID {
		return
	}
	if err := s.setField(ctx, studioID, "parent", parentID); err != nil {
		log.Printf("ingest: set parent on %s: %v", studioID, err)
	}
}

// ingestTag upserts a tag by name.
func (s *Service) ingestTag(ctx context.Context, pluginName string, tag *plugins.Tag) (*IngestResult, error) {
	obs := s.baseObs(tag, pluginName, "")
	return s.upsert(ctx, "tag", "name", tag.Name, obs, pluginName)
}

// baseObs flattens an entity to a map and stamps the common observation fields.
func (s *Service) baseObs(entity any, pluginName, sourceURL string) map[string]any {
	obs := toMap(entity)
	obs["plugin"] = pluginName
	obs["source_url"] = sourceURL
	obs["observed_at"] = time.Now().UTC()
	return obs
}

func sceneURL(scene *plugins.Scene, fallback string) string {
	if scene.SourceURL != "" {
		return scene.SourceURL
	}
	return fallback
}

// upsert resolves a canonical record by (identityField = identityValue): merging
// into an existing record or creating a new one. It then writes the observation
// (with target set) and a merge record, returning the ingest result.
func (s *Service) upsert(ctx context.Context, table, identityField, identityValue string, obs map[string]any, pluginName string) (*IngestResult, error) {
	canonicalID, isNew, changed, err := s.resolveByIdentity(ctx, table, identityField, identityValue, obs)
	if err != nil {
		return nil, err
	}

	canonicalRID, parseErr := db.ParseRecordID(canonicalID)
	if parseErr != nil {
		return nil, fmt.Errorf("invalid canonical id %q: %w", canonicalID, parseErr)
	}
	obs["target"] = *canonicalRID

	obsIDStr, err := s.writeObservation(ctx, obs)
	if err != nil {
		return nil, err
	}

	mode := "auto"
	score := 1.0
	if isNew {
		mode, score = "new", 0
	}
	s.writeChanges(ctx, changeSet{
		canonicalID:   canonicalID,
		observationID: obsIDStr,
		pluginName:    pluginName,
		score:         score,
		mode:          mode,
		fields:        changed,
		values:        obs,
	})

	return &IngestResult{
		ObservationID: obsIDStr,
		CanonicalID:   canonicalID,
		IsNew:         isNew,
		Score:         score,
		Mode:          mode,
	}, nil
}

// tablesWithExternalID are the canonical tables whose external_id identifies a
// record across URL variants (sites link one video under several paths).
var tablesWithExternalID = map[string]bool{"scene": true, "gallery": true, "performer": true, "studio": true}

// findByIdentity returns the id of the record matching field = value, falling
// back to the observation's plugin-prefixed external_id, or "" when neither
// matches.
func (s *Service) findByIdentity(ctx context.Context, table, field, value string, obs map[string]any) (string, error) {
	if value != "" {
		matches, err := s.database.Strings(ctx,
			fmt.Sprintf("SELECT id FROM %s WHERE %s = $value LIMIT 1", table, field),
			db.Vars{"value": value})
		if err != nil {
			return "", fmt.Errorf("lookup %s by %s: %w", table, field, err)
		}
		if len(matches) > 0 {
			return matches[0], nil
		}
	}
	externalID, _ := obs["external_id"].(string)
	if externalID == "" || field == "external_id" || !tablesWithExternalID[table] {
		return "", nil
	}
	matches, err := s.database.Strings(ctx,
		fmt.Sprintf("SELECT id FROM %s WHERE external_id = $externalID LIMIT 1", table),
		db.Vars{"externalID": externalID})
	if err != nil {
		return "", fmt.Errorf("lookup %s by external_id: %w", table, err)
	}
	if len(matches) > 0 {
		return matches[0], nil
	}
	return "", nil
}

// resolveByIdentity finds an existing record by its identity field or creates a
// new one. Returns (canonicalID, isNew, changedFields, err).
func (s *Service) resolveByIdentity(ctx context.Context, table, field, value string, obs map[string]any) (string, bool, []string, error) {
	existingID, err := s.findByIdentity(ctx, table, field, value, obs)
	if err != nil {
		return "", false, nil, err
	}
	if existingID != "" {
		changed, mergeErr := s.applyMerge(ctx, existingID, obs)
		if mergeErr != nil {
			return "", false, nil, fmt.Errorf("apply merge: %w", mergeErr)
		}
		return existingID, false, changed, nil
	}

	canonicalID, createErr := s.createCanonical(ctx, table, obs)
	if createErr != nil {
		return "", false, nil, fmt.Errorf("create canonical: %w", createErr)
	}
	return canonicalID, true, whitelistKeys(table, obs), nil
}

// createCanonical inserts a new canonical record from an observation, keeping
// only the whitelisted scalar columns for the table.
func (s *Service) createCanonical(ctx context.Context, table string, obs map[string]any) (string, error) {
	created, err := s.database.Insert(ctx, table, pickColumns(table, obs))
	if err != nil {
		return "", fmt.Errorf("create %s: %w", table, err)
	}
	return created.String(), nil
}

// applyMerge fills empty canonical columns from the observation. Returns the list
// of field names that changed.
func (s *Service) applyMerge(ctx context.Context, canonicalID string, obs map[string]any) ([]string, error) {
	rid, err := db.ParseRecordID(canonicalID)
	if err != nil {
		return nil, fmt.Errorf("invalid canonical id %q: %w", canonicalID, err)
	}

	current, err := s.database.Get(ctx, *rid)
	if err != nil {
		return nil, fmt.Errorf("fetch canonical %s: %w", canonicalID, err)
	}
	if current == nil {
		return nil, fmt.Errorf("fetch canonical %s: not found", canonicalID)
	}

	table := TableFromID(canonicalID)
	updates := map[string]any{}
	var changed []string
	for _, k := range canonicalColumns[table] {
		v, ok := obs[k]
		if !ok || isEmpty(v) || !isEmpty(current[k]) {
			continue
		}
		updates[k] = v
		changed = append(changed, k)
	}

	if len(changed) > 0 {
		if err = s.database.Merge(ctx, *rid, updates); err != nil {
			return nil, fmt.Errorf("merge canonical %s: %w", canonicalID, err)
		}
	}
	return changed, nil
}

// ---------------------------------------------------------------------------
// Relation resolution: name/reference strings → record links + graph edges.

// linkStudio resolves a studio reference to a record and sets it on the parent.
// Matching prefers stable identity (source_url, external_id) over the name so
// the same studio scraped from different pages doesn't duplicate.
func (s *Service) linkStudio(ctx context.Context, table, parentID string, ref *plugins.StudioRef, pluginName string) {
	if ref == nil || strings.TrimSpace(ref.Name) == "" {
		return
	}
	studioID, err := s.findOrCreateStudio(ctx, ref)
	if err != nil {
		log.Printf("ingest: resolve studio %q: %v", ref.Name, err)
		return
	}
	if err := s.setField(ctx, parentID, "studio", studioID); err != nil {
		log.Printf("ingest: set studio on %s: %v", parentID, err)
	}
}

// findOrCreateStudio matches a StudioRef by source_url, then external_id, then
// name, creating the studio if nothing matches. Identity hints and the logo
// from the ref are backfilled onto the record so future scrapes match directly.
func (s *Service) findOrCreateStudio(ctx context.Context, ref *plugins.StudioRef) (db.RecordID, error) {
	for field, value := range map[string]string{"source_url": ref.SourceURL, "external_id": ref.ExternalID} {
		if value == "" {
			continue
		}
		if studioID, found := s.findIDBy(ctx, "studio", field, value); found {
			s.backfillStudioIdentity(ctx, studioID, ref)
			return studioID, nil
		}
	}

	studioID, err := s.findOrCreateNamed(ctx, "studio", strings.TrimSpace(ref.Name))
	if err != nil {
		return db.RecordID{}, err
	}
	s.backfillStudioIdentity(ctx, studioID, ref)
	return studioID, nil
}

// backfillStudioIdentity fills empty source_url/external_id/image_path columns from the ref.
func (s *Service) backfillStudioIdentity(ctx context.Context, studioID db.RecordID, ref *plugins.StudioRef) {
	fields := map[string]string{"source_url": ref.SourceURL, "external_id": ref.ExternalID, "image_path": ref.ImagePath}
	for field, value := range fields {
		if value == "" {
			continue
		}
		_, err := s.database.Exec(ctx,
			fmt.Sprintf("UPDATE studio SET %s = $value WHERE id = $id AND %s IS NULL", field, field),
			db.Vars{"id": studioID, "value": value})
		if err != nil {
			log.Printf("ingest: backfill studio %s %s: %v", studioID, field, err)
		}
	}
}

// linkPerformers resolves performer names to records, sets the denormalized
// array on the parent, and writes a performs_in credit for each performer.
func (s *Service) linkPerformers(ctx context.Context, table, parentID string, performers []plugins.ScenePerformer, pluginName string) {
	if len(performers) == 0 {
		return
	}
	ids := make([]db.RecordID, 0, len(performers))
	for _, p := range performers {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		performerID, err := s.findOrCreateNamed(ctx, "performer", name)
		if err != nil {
			log.Printf("ingest: resolve performer %q: %v", name, err)
			continue
		}
		ids = append(ids, performerID)
		if err := s.writeCredit(ctx, performerID, parentID, p); err != nil {
			log.Printf("ingest: performs_in credit: %v", err)
		}
	}
	if len(ids) > 0 {
		if err := s.setField(ctx, parentID, "performers", ids); err != nil {
			log.Printf("ingest: set performers on %s: %v", parentID, err)
		}
	}
}

// linkTags resolves tag names to records and sets the array on the parent.
func (s *Service) linkTags(ctx context.Context, table, parentID string, tags []string, pluginName string) {
	ids := s.resolveNamedRecords(ctx, "tag", withoutNoiseTags(tags))
	if len(ids) == 0 {
		return
	}
	if err := s.setField(ctx, parentID, "tags", ids); err != nil {
		log.Printf("ingest: set tags on %s: %v", parentID, err)
	}
}

// setField merges a single field onto a canonical record.
func (s *Service) setField(ctx context.Context, recordID, field string, value any) error {
	rid, err := db.ParseRecordID(recordID)
	if err != nil {
		return fmt.Errorf("invalid record id %q: %w", recordID, err)
	}
	return s.database.Merge(ctx, *rid, map[string]any{field: value})
}

// writeCredit records a performer's appearance in a media record once; an
// existing credit for the same (performer, media) pair is left untouched.
func (s *Service) writeCredit(ctx context.Context, performerID db.RecordID, mediaID string, performer plugins.ScenePerformer) error {
	vars := db.Vars{
		"id":            db.NewRecordID("performs_in"),
		"performer":     performerID,
		"media":         mediaID,
		"credited_as":   nil,
		"billing_order": nil,
	}
	if performer.As != "" {
		vars["credited_as"] = performer.As
	}
	if performer.Order != 0 {
		vars["billing_order"] = performer.Order
	}
	_, err := s.database.Exec(ctx,
		`INSERT INTO performs_in (id, performer, media, credited_as, billing_order)
		VALUES ($id, $performer, $media, $credited_as, $billing_order)
		ON CONFLICT (performer, media) DO NOTHING`, vars)
	return err
}

// ---------------------------------------------------------------------------
// Observation + change history writers.

// observationColumns are the observation fields stored in their own columns;
// every other payload field lives in the data JSON column.
var observationColumns = map[string]bool{
	"id": true, "target": true, "plugin": true, "source_url": true,
	"confidence": true, "status": true, "job": true, "observed_at": true,
}

// writeObservation inserts into the observation table, returning its record ID.
func (s *Service) writeObservation(ctx context.Context, obs map[string]any) (string, error) {
	row := map[string]any{}
	data := map[string]any{}
	for key, value := range obs {
		if observationColumns[key] {
			row[key] = value
			continue
		}
		data[key] = value
	}
	row["data"] = data
	created, err := s.database.Insert(ctx, "observation", row)
	if err != nil {
		return "", fmt.Errorf("write observation: %w", err)
	}
	return created.String(), nil
}

// flattenObservation merges an observation row's data payload back into the
// row so callers see the flat record the plugin emitted. Fixed columns win.
func flattenObservation(row db.Row) map[string]any {
	flat := map[string]any{}
	if data, isObject := row["data"].(map[string]any); isObject {
		for key, value := range data {
			flat[key] = value
		}
	}
	for key, value := range row {
		if key != "data" {
			flat[key] = value
		}
	}
	return flat
}

// latestObservation returns the newest flattened observation for a target, or
// nil when there is none.
func (s *Service) latestObservation(ctx context.Context, targetID string) (map[string]any, error) {
	row, err := s.database.QueryRow(ctx,
		`SELECT * FROM observation WHERE target = $target ORDER BY observed_at DESC, rowid DESC LIMIT 1`,
		db.Vars{"target": targetID})
	if err != nil || row == nil {
		return nil, err
	}
	return flattenObservation(row), nil
}

// changeSet describes the fields one observation wrote onto a canonical record.
type changeSet struct {
	canonicalID   string
	observationID string
	pluginName    string
	score         float64
	mode          string // "new" | "auto" | "manual"
	changedBy     *string
	fields        []string
	values        map[string]any // observation payload the field values come from
}

// writeChanges persists one changes row per filled field. Merges only fill
// empty columns, so every row is an "added" with no original value. Errors are
// logged but not returned because a failed history row shouldn't roll back
// the actual merge.
func (s *Service) writeChanges(ctx context.Context, changes changeSet) {
	for _, field := range changes.fields {
		row := map[string]any{
			"canonical":   changes.canonicalID,
			"observation": changes.observationID,
			"plugin":      changes.pluginName,
			"key":         field,
			"action":      "added",
			"value":       changes.values[field],
			"score":       changes.score,
			"mode":        changes.mode,
		}
		if changes.changedBy != nil {
			row["changed_by"] = *changes.changedBy
		}
		if _, err := s.database.Insert(ctx, "changes", row); err != nil {
			log.Printf("ingest: write change %s.%s: %v", changes.canonicalID, field, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Named-record resolution (studio / performer / tag lookup tables).

// resolveNamedRecords maps names to record ids in a name-keyed table, creating
// any that don't exist yet. Duplicates are collapsed.
func (s *Service) resolveNamedRecords(ctx context.Context, table string, names []string) []db.RecordID {
	seen := map[string]bool{}
	seenIDs := map[db.RecordID]bool{}
	out := make([]db.RecordID, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		id, err := s.findOrCreateNamed(ctx, table, name)
		if err != nil {
			log.Printf("ingest: resolve %s %q: %v", table, name, err)
			continue
		}
		// Differently spelled names can resolve to one record ("Anal", "anal").
		if seenIDs[id] {
			continue
		}
		seenIDs[id] = true
		out = append(out, id)
	}
	return out
}

// findOrCreateNamed returns the id of the row with this name, creating it if
// absent. On a create race (UNIQUE name) it falls back to a second lookup.
func (s *Service) findOrCreateNamed(ctx context.Context, table, name string) (db.RecordID, error) {
	if id, ok, err := s.findNamed(ctx, table, name); err != nil {
		return db.RecordID{}, err
	} else if ok {
		return id, nil
	}

	created, createErr := s.database.Insert(ctx, table, map[string]any{"name": name})
	if createErr == nil {
		return created, nil
	}
	if id, ok, findErr := s.findNamed(ctx, table, name); findErr == nil && ok {
		return id, nil
	}
	return db.RecordID{}, createErr
}

func (s *Service) findNamed(ctx context.Context, table, name string) (db.RecordID, bool, error) {
	// Tags merge synonyms on ingest: a name matching an existing tag's alias
	// resolves to that canonical tag instead of creating a duplicate. Sites
	// disagree on casing ("Anal" / "anal"), so tags also match case-insensitively,
	// preferring an exact spelling when both exist.
	query := fmt.Sprintf("SELECT id FROM %s WHERE name = $name LIMIT 1", table)
	if table == "tag" {
		query = `SELECT id FROM tag
			WHERE name = $name COLLATE NOCASE
			   OR EXISTS (SELECT 1 FROM json_each(tag.aliases) WHERE value = $name COLLATE NOCASE)
			ORDER BY name = $name DESC LIMIT 1`
	}
	matches, err := s.database.Strings(ctx, query, db.Vars{"name": name})
	if err != nil {
		return db.RecordID{}, false, err
	}
	if len(matches) == 0 {
		return db.RecordID{}, false, nil
	}
	parsed, err := db.ParseRecordID(matches[0])
	if err != nil {
		return db.RecordID{}, false, err
	}
	return *parsed, true, nil
}

// findIDBy returns the id of the first row in table whose field equals value.
func (s *Service) findIDBy(ctx context.Context, table, field, value string) (db.RecordID, bool) {
	matches, err := s.database.Strings(ctx,
		fmt.Sprintf("SELECT id FROM %s WHERE %s = $value LIMIT 1", table, field),
		db.Vars{"value": value})
	if err != nil || len(matches) == 0 {
		return db.RecordID{}, false
	}
	parsed, err := db.ParseRecordID(matches[0])
	if err != nil {
		return db.RecordID{}, false
	}
	return *parsed, true
}

// pluginRecordID returns the plugin registry row for pluginName, or the
// deterministic "plugin:<name>" id when the plugin has no registry row.
func (s *Service) pluginRecordID(ctx context.Context, pluginName string) db.RecordID {
	if pluginID, found, err := s.findNamed(ctx, "plugin", pluginName); err == nil && found {
		return pluginID
	}
	return db.RecordID{Table: "plugin", ID: pluginName}
}

// ---------------------------------------------------------------------------
// Stub ingestion + manual merge.

// IngestDiscoveredItem inserts a minimal stub canonical record for a
// DiscoveredItem so it is immediately navigable before a full scrape completes.
func (s *Service) IngestDiscoveredItem(ctx context.Context, pluginName string, item plugins.DiscoveredItem) (string, error) {
	var results []*plugins.ScrapeResult
	switch item.MediaType {
	case plugins.MediaTypeScene:
		results = []*plugins.ScrapeResult{{Type: item.MediaType, Scene: &plugins.Scene{
			Type: item.MediaType, ExternalID: item.ExternalID, SourceURL: item.SourceURL,
			Title: item.Title, Date: item.Date, Duration: item.DurationSeconds,
			Rating: item.Rating, ViewCount: item.ViewCount,
			PosterPath: item.PosterPath, PreviewImages: item.PreviewImages, PreviewVideo: item.PreviewVideo,
			Studio: item.Studio,
			// Deterministic sources from the listing attach a stream to the stub
			// right away, so browsed scenes are playable without a full scrape.
			Downloads: item.Downloads,
		}}}
	case plugins.MediaTypeGallery:
		results = []*plugins.ScrapeResult{{Type: item.MediaType, Gallery: &plugins.Gallery{
			Type: item.MediaType, ExternalID: item.ExternalID, SourceURL: item.SourceURL,
			Title: item.Title, Date: item.Date, CoverPath: item.PosterPath,
		}}}
	case plugins.MediaTypePerformer:
		results = []*plugins.ScrapeResult{{Type: item.MediaType, Performer: &plugins.Performer{
			Type: item.MediaType, ExternalID: item.ExternalID, SourceURL: item.SourceURL,
			Name: item.Title, ImagePath: item.PosterPath,
		}}}
	case plugins.MediaTypeStudio:
		results = []*plugins.ScrapeResult{{Type: item.MediaType, Studio: &plugins.Studio{
			Type: item.MediaType, ExternalID: item.ExternalID, SourceURL: item.SourceURL,
			Name: item.Title, ImagePath: item.PosterPath,
		}}}
	default:
		return "", fmt.Errorf("unsupported media type for stub ingestion: %s", item.MediaType)
	}

	ingestResult, err := s.IngestScrapeResults(ctx, pluginName, item.SourceURL, results)
	if err != nil {
		return "", err
	}
	return ingestResult.CanonicalID, nil
}

// ManualMerge merges a specific observation into a specific canonical record,
// recording mergedBy as the user who confirmed the match.
func (s *Service) ManualMerge(ctx context.Context, canonicalID, observationID, mergedBy string) (*IngestResult, error) {
	obsRID, err := db.ParseRecordID(observationID)
	if err != nil {
		return nil, fmt.Errorf("invalid observation id: %w", err)
	}

	row, err := s.database.Get(ctx, *obsRID)
	if err != nil || row == nil {
		return nil, fmt.Errorf("observation %s not found", observationID)
	}
	obs := flattenObservation(row)

	plugin, _ := obs["plugin"].(string)
	changed, mergeErr := s.applyMerge(ctx, canonicalID, obs)
	if mergeErr != nil {
		return nil, mergeErr
	}
	s.writeChanges(ctx, changeSet{
		canonicalID:   canonicalID,
		observationID: observationID,
		pluginName:    plugin,
		score:         1.0,
		mode:          "manual",
		changedBy:     &mergedBy,
		fields:        changed,
		values:        obs,
	})

	return &IngestResult{
		ObservationID: observationID,
		CanonicalID:   canonicalID,
		Score:         1.0,
		Mode:          "manual",
	}, nil
}

// ---------------------------------------------------------------------------
// helpers

func toMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

// pickColumns copies only the whitelisted scalar columns for a table.
func pickColumns(table string, obs map[string]any) map[string]any {
	rec := make(map[string]any)
	for _, k := range canonicalColumns[table] {
		if v, ok := obs[k]; ok && !isEmpty(v) {
			rec[k] = v
		}
	}
	return rec
}

// whitelistKeys returns the whitelisted columns present and non-empty in obs.
func whitelistKeys(table string, obs map[string]any) []string {
	var keys []string
	for _, k := range canonicalColumns[table] {
		if v, ok := obs[k]; ok && !isEmpty(v) {
			keys = append(keys, k)
		}
	}
	return keys
}

func recordIDString(m map[string]any) string {
	id, ok := m["id"]
	if !ok {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	case db.RecordID:
		return v.String()
	case *db.RecordID:
		if v != nil {
			return v.String()
		}
		return ""
	}
	return fmt.Sprintf("%v", id)
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []any:
		return len(val) == 0
	case []string:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	case float64:
		return val == 0
	case int:
		return val == 0
	case int64:
		return val == 0
	}
	return false
}
