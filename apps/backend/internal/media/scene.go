package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// EntityFilters holds the shared browse filters for the scene-model entities.
type EntityFilters struct {
	Search      *string
	Sort        *string
	StudioID    *string
	PerformerID *string
	TagID       *string
	Category    *string
	// Advanced scene filters. Nil means the dimension is unconstrained.
	MinRating   *float64
	MinDuration *int
	MaxDuration *int
	DateFrom    *string
	DateTo      *string
	// DisabledPlugins hides scenes that a now-disabled plugin discovered but never
	// fully scraped (stubs with no streams). Empty means no such filtering.
	DisabledPlugins []string
	// Sources keeps only records some of these plugins observed. Empty = any.
	Sources []string
}

// sourceCondition matches records observed by one of the $sources plugins.
func sourceCondition(table string) string {
	return fmt.Sprintf(`EXISTS (SELECT 1 FROM observation
		WHERE observation.target = %s.id AND observation.plugin IN (SELECT value FROM json_each($sources)))`, table)
}

// hiddenStubSceneIDs returns the scenes a disabled plugin discovered but never
// fully scraped (stubs with no streams). Fully scraped scenes are always kept —
// disabling a plugin only hides its unfetched stubs, not content already in the
// library. Best-effort: on error it returns nil so listings never break.
func (r *Repository) hiddenStubSceneIDs(ctx context.Context, disabled []string) []db.RecordID {
	if len(disabled) == 0 {
		return nil
	}
	targets, err := r.database.Strings(ctx,
		`SELECT DISTINCT target FROM observation
			WHERE plugin IN (SELECT value FROM json_each($disabled))
			AND target NOT IN (SELECT media FROM stream)`,
		db.Vars{"disabled": disabled})
	if err != nil {
		return nil
	}
	return parseRecordIDs(targets)
}

// excludedSceneIDs merges the disabled-plugin stubs with the blocklist into a
// single set of scene ids to omit from a listing.
func (r *Repository) excludedSceneIDs(ctx context.Context, filters EntityFilters) []db.RecordID {
	hidden := r.hiddenStubSceneIDs(ctx, filters.DisabledPlugins)
	hidden = append(hidden, r.BlockedSceneRIDs(ctx)...)
	return dedupeRecordIDs(hidden)
}

// ExcludedSceneIDs returns the scenes every listing hides: blocked ones and
// unfetched stubs of the given inactive plugins.
func (r *Repository) ExcludedSceneIDs(ctx context.Context, disabledPlugins []string) []db.RecordID {
	return r.excludedSceneIDs(ctx, EntityFilters{DisabledPlugins: disabledPlugins})
}

// ScenesByIDs loads the given scenes in the given order, skipping unknown ids.
func (r *Repository) ScenesByIDs(ctx context.Context, ids []string) ([]*model.Scene, error) {
	if len(ids) == 0 {
		return []*model.Scene{}, nil
	}
	rows, err := r.database.Query(ctx,
		`SELECT scene.* FROM json_each($ids) AS link JOIN scene ON scene.id = link.value ORDER BY link.key`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("scenes by ids: %w", err)
	}
	out := make([]*model.Scene, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.sceneFromMap(ctx, row))
	}
	return out, nil
}

// parseRecordIDs parses "table:id" strings, skipping malformed ones.
func parseRecordIDs(values []string) []db.RecordID {
	out := make([]db.RecordID, 0, len(values))
	for _, value := range values {
		parsed, err := db.ParseRecordID(value)
		if err != nil {
			continue
		}
		out = append(out, *parsed)
	}
	return out
}

func dedupeRecordIDs(ids []db.RecordID) []db.RecordID {
	if len(ids) == 0 {
		return ids
	}
	seen := make(map[string]bool, len(ids))
	out := make([]db.RecordID, 0, len(ids))
	for i := range ids {
		key := ids[i].String()
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ids[i])
	}
	return out
}

// RandomScenes returns a random selection of scenes, applying the blocklist and hiding disabled-plugin stubs.
func (r *Repository) RandomScenes(ctx context.Context, limit int, filters EntityFilters) ([]*model.Scene, error) {
	if limit <= 0 {
		limit = 5
	}
	vars := db.Vars{"limit": limit}
	where := ""
	if hidden := r.excludedSceneIDs(ctx, filters); len(hidden) > 0 {
		vars["hidden"] = hidden
		where = "WHERE " + excludedIDCondition
	}
	query := fmt.Sprintf("SELECT * FROM scene %s ORDER BY random() LIMIT $limit", where)
	rows, err := r.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("random scenes: %w", err)
	}
	out := make([]*model.Scene, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.sceneFromMap(ctx, row))
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// map[string]any field accessors for decoded rows.

func mString(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func mStringPtr(m map[string]any, key string) *string {
	s, ok := m[key].(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

func mIntPtr(m map[string]any, key string) *int {
	switch v := m[key].(type) {
	case int:
		return &v
	case int64:
		n := int(v)
		return &n
	case float64:
		n := int(v)
		return &n
	case uint64:
		n := int(v)
		return &n
	}
	return nil
}

func mInt(m map[string]any, key string) int {
	if p := mIntPtr(m, key); p != nil {
		return *p
	}
	return 0
}

func mFloatPtr(m map[string]any, key string) *float64 {
	switch v := m[key].(type) {
	case float64:
		return &v
	case float32:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	}
	return nil
}

func mBool(m map[string]any, key string) bool {
	return db.AsBool(m[key])
}

func mStringSlice(m map[string]any, key string) []string {
	out := []string{}
	switch xs := m[key].(type) {
	case []string:
		return xs
	case []any:
		for _, x := range xs {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

// asRecordID reads a link column value ("table:id" string) as a record id.
func asRecordID(v any) *db.RecordID {
	switch id := v.(type) {
	case string:
		parsed, err := db.ParseRecordID(id)
		if err != nil {
			return nil
		}
		return parsed
	case db.RecordID:
		return &id
	case *db.RecordID:
		return id
	}
	return nil
}

func mID(m map[string]any) string {
	return recordIDString(asRecordID(m["id"]))
}

func mRIDList(m map[string]any, key string) []db.RecordID {
	out := []db.RecordID{}
	xs, ok := m[key].([]any)
	if !ok {
		return out
	}
	for _, x := range xs {
		if rid := asRecordID(x); rid != nil {
			out = append(out, *rid)
		}
	}
	return out
}

// sceneCountWhere counts scenes matching a WHERE clause that binds $id.
func (r *Repository) sceneCountWhere(ctx context.Context, clause string, rid *db.RecordID) int {
	if rid == nil {
		return 0
	}
	count, err := r.database.Int(ctx,
		fmt.Sprintf("SELECT count(*) FROM scene WHERE %s", clause),
		db.Vars{"id": *rid})
	if err != nil {
		return 0
	}
	return count
}

// ---------------------------------------------------------------------------
// Tag

func tagFromMap(m map[string]any) *model.Tag {
	return &model.Tag{
		ID:          mID(m),
		Name:        mString(m, "name"),
		Aliases:     mStringSlice(m, "aliases"),
		Description: mStringPtr(m, "description"),
		Category:    mStringPtr(m, "category"),
		CreatedAt:   mString(m, "created_at"),
		UpdatedAt:   mString(m, "updated_at"),
	}
}

func (r *Repository) ListTags(ctx context.Context, filters EntityFilters, limit, offset *int) ([]*model.Tag, error) {
	vars := db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)}
	where := ""
	if filters.Search != nil && *filters.Search != "" {
		vars["q"] = *filters.Search
		where = "WHERE " + searchCondition("name")
	} else if filters.Category != nil && *filters.Category != "" {
		vars["cat"] = *filters.Category
		where = "WHERE category = $cat"
	}
	query := fmt.Sprintf("SELECT * FROM tag %s ORDER BY name ASC LIMIT $limit OFFSET $offset", where)
	rows, err := r.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	out := make([]*model.Tag, 0, len(rows))
	for _, row := range rows {
		tag := tagFromMap(row)
		tag.SceneCount = r.sceneCountWhere(ctx, jsonContains("scene.tags", "$id"), asRecordID(row["id"]))
		out = append(out, tag)
	}
	return out, nil
}

func (r *Repository) GetTag(ctx context.Context, id string) (*model.Tag, error) {
	m, err := r.getByID(ctx, id)
	if err != nil || m == nil {
		return nil, err
	}
	tag := tagFromMap(m)
	tag.SceneCount = r.sceneCountWhere(ctx, jsonContains("scene.tags", "$id"), asRecordID(m["id"]))
	return tag, nil
}

// ---------------------------------------------------------------------------
// Studio

func (r *Repository) studioFromMap(ctx context.Context, m map[string]any, parent *model.Studio) *model.Studio {
	return &model.Studio{
		ID:        mID(m),
		Name:      mString(m, "name"),
		Aliases:   mStringSlice(m, "aliases"),
		URL:       mStringPtr(m, "url"),
		Parent:    parent,
		ImagePath: mStringPtr(m, "image_path"),
		Details:   mStringPtr(m, "details"),
		Tags:      r.tagsByIDs(ctx, mRIDList(m, "tags")),
		CreatedAt: mString(m, "created_at"),
		UpdatedAt: mString(m, "updated_at"),
	}
}

func (r *Repository) ListStudios(ctx context.Context, filters EntityFilters, limit, offset *int) ([]*model.Studio, error) {
	vars := db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)}
	var conditions []string
	if filters.Search != nil && *filters.Search != "" {
		vars["q"] = *filters.Search
		conditions = append(conditions, searchCondition("name"))
	}
	if rid := parseFilterID(filters.TagID); rid != nil {
		vars["tag"] = *rid
		conditions = append(conditions, jsonContains("studio.tags", "$tag"))
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + joinAnd(conditions)
	}
	query := fmt.Sprintf("SELECT * FROM studio %s ORDER BY name ASC LIMIT $limit OFFSET $offset", where)
	rows, err := r.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("list studios: %w", err)
	}
	out := make([]*model.Studio, 0, len(rows))
	for _, row := range rows {
		studio := r.studioFromMap(ctx, row, nil)
		studio.SceneCount = r.sceneCountWhere(ctx, "studio = $id", asRecordID(row["id"]))
		out = append(out, studio)
	}
	return out, nil
}

func (r *Repository) GetStudio(ctx context.Context, id string) (*model.Studio, error) {
	m, err := r.getByID(ctx, id)
	if err != nil || m == nil {
		return nil, err
	}
	var parent *model.Studio
	if pid := asRecordID(m["parent"]); pid != nil {
		parent = r.studioLite(ctx, pid)
	}
	studio := r.studioFromMap(ctx, m, parent)
	studio.SceneCount = r.sceneCountWhere(ctx, "studio = $id", asRecordID(m["id"]))
	return studio, nil
}

// SetStudioTags replaces a studio's tags with exactly the given tag ids.
func (r *Repository) SetStudioTags(ctx context.Context, studioID string, tagIDs []string) (*model.Studio, error) {
	if err := r.setTagsField(ctx, studioID, tagIDs); err != nil {
		return nil, err
	}
	return r.GetStudio(ctx, studioID)
}

// setTagsField writes a tags record-link array onto any tag-carrying record.
func (r *Repository) setTagsField(ctx context.Context, recordID string, tagIDs []string) error {
	rid, err := db.ParseRecordID(recordID)
	if err != nil {
		return fmt.Errorf("invalid record id: %w", err)
	}
	tags := make([]db.RecordID, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		tagRID, parseErr := db.ParseRecordID(tagID)
		if parseErr != nil {
			return fmt.Errorf("invalid tag id %q: %w", tagID, parseErr)
		}
		tags = append(tags, *tagRID)
	}
	if err := r.database.Merge(ctx, *rid, db.Vars{"tags": tags}); err != nil {
		return fmt.Errorf("set tags: %w", err)
	}
	return nil
}

// studioLite loads a studio without its parent chain or scene count (used for
// nested references to avoid unbounded recursion / N+1 counting).
func (r *Repository) studioLite(ctx context.Context, rid *db.RecordID) *model.Studio {
	if rid == nil {
		return nil
	}
	m, err := r.getByID(ctx, rid.String())
	if err != nil || m == nil {
		return nil
	}
	return r.studioFromMap(ctx, m, nil)
}

// ---------------------------------------------------------------------------
// Performer

func (r *Repository) performerFromMap(ctx context.Context, m map[string]any, withCount bool) *model.Performer {
	p := &model.Performer{
		ID:           mID(m),
		Name:         mString(m, "name"),
		Aliases:      mStringSlice(m, "aliases"),
		Details:      mStringPtr(m, "details"),
		Gender:       mStringPtr(m, "gender"),
		Birthdate:    mStringPtr(m, "birthdate"),
		DeathDate:    mStringPtr(m, "death_date"),
		Country:      mStringPtr(m, "country"),
		Ethnicity:    mStringPtr(m, "ethnicity"),
		EyeColor:     mStringPtr(m, "eye_color"),
		HairColor:    mStringPtr(m, "hair_color"),
		HeightCm:     mIntPtr(m, "height_cm"),
		WeightKg:     mIntPtr(m, "weight_kg"),
		Measurements: mStringPtr(m, "measurements"),
		FakeTits:     mStringPtr(m, "fake_tits"),
		Tattoos:      mStringPtr(m, "tattoos"),
		Piercings:    mStringPtr(m, "piercings"),
		CareerLength: mStringPtr(m, "career_length"),
		URL:          mStringPtr(m, "url"),
		Twitter:      mStringPtr(m, "twitter"),
		Instagram:    mStringPtr(m, "instagram"),
		ImagePath:    mStringPtr(m, "image_path"),
		Favorite:     mBool(m, "favorite"),
		Rating:       mFloatPtr(m, "rating"),
		Tags:         r.tagsByIDs(ctx, mRIDList(m, "tags")),
		CreatedAt:    mString(m, "created_at"),
		UpdatedAt:    mString(m, "updated_at"),
	}
	if withCount {
		p.SceneCount = r.sceneCountWhere(ctx, jsonContains("scene.performers", "$id"), asRecordID(m["id"]))
	}
	return p
}

func (r *Repository) ListPerformers(ctx context.Context, filters EntityFilters, limit, offset *int) ([]*model.Performer, error) {
	vars := db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)}
	var conditions []string
	if filters.Search != nil && *filters.Search != "" {
		vars["q"] = *filters.Search
		conditions = append(conditions, searchCondition("name"))
	}
	if rid := parseFilterID(filters.TagID); rid != nil {
		vars["tag"] = *rid
		conditions = append(conditions, jsonContains("performer.tags", "$tag"))
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + joinAnd(conditions)
	}
	query := fmt.Sprintf("SELECT * FROM performer %s ORDER BY name ASC LIMIT $limit OFFSET $offset", where)
	rows, err := r.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("list performers: %w", err)
	}
	out := make([]*model.Performer, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.performerFromMap(ctx, row, true))
	}
	return out, nil
}

func (r *Repository) GetPerformer(ctx context.Context, id string) (*model.Performer, error) {
	m, err := r.getByID(ctx, id)
	if err != nil || m == nil {
		return nil, err
	}
	return r.performerFromMap(ctx, m, true), nil
}

func (r *Repository) SetPerformerFavorite(ctx context.Context, id string, favorite bool) (*model.Performer, error) {
	rid, err := db.ParseRecordID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid performer id: %w", err)
	}
	if err := r.database.Merge(ctx, *rid, db.Vars{"favorite": favorite}); err != nil {
		return nil, fmt.Errorf("set favorite: %w", err)
	}
	return r.GetPerformer(ctx, id)
}

// ---------------------------------------------------------------------------
// Scene

func (r *Repository) sceneFromMap(ctx context.Context, m map[string]any) *model.Scene {
	return &model.Scene{
		ID:              mID(m),
		ExternalID:      mString(m, "external_id"),
		SourceURL:       mString(m, "source_url"),
		Title:           mString(m, "title"),
		Details:         mStringPtr(m, "details"),
		Date:            mStringPtr(m, "date"),
		DurationSeconds: mIntPtr(m, "duration_seconds"),
		Rating:          mFloatPtr(m, "rating"),
		ViewCount:       mInt(m, "view_count"),
		Organized:       mBool(m, "organized"),
		PosterPath:      mStringPtr(m, "poster_path"),
		PreviewVideo:    mStringPtr(m, "preview_video"),
		PreviewImages:   mStringSlice(m, "preview_images"),
		Studio:          r.studioLite(ctx, asRecordID(m["studio"])),
		Performers:      r.performersByIDs(ctx, mRIDList(m, "performers")),
		Tags:            r.tagsByIDs(ctx, mRIDList(m, "tags")),
		TagMatch:        tagMatchFromMap(m),
		CreatedAt:       mString(m, "created_at"),
		UpdatedAt:       mString(m, "updated_at"),
	}
}

// tagMatchFromMap converts the query-projected direct_match flag into the
// GraphQL tagMatch value. Nil when the listing had no tag filter.
func tagMatchFromMap(m map[string]any) *string {
	v, ok := m["direct_match"]
	if !ok {
		return nil
	}
	match := "inherited"
	if db.AsBool(v) {
		match = "direct"
	}
	return &match
}

// inheritedTagCondition matches rows of table whose own tags, studio tags, or
// credited performers' tags contain $tag. Columns are table-qualified because
// json_each exposes its own id/key/value columns inside the subqueries.
func inheritedTagCondition(table string) string {
	return "(" + jsonContains(table+".tags", "$tag") +
		" OR EXISTS (SELECT 1 FROM studio, json_each(studio.tags) AS studio_tag" +
		" WHERE studio.id = " + table + ".studio AND studio_tag.value = $tag)" +
		" OR EXISTS (SELECT 1 FROM json_each(" + table + ".performers) AS credit" +
		" JOIN performer ON performer.id = credit.value, json_each(performer.tags) AS performer_tag" +
		" WHERE performer_tag.value = $tag))"
}

// inheritedTagProjection selects every column of table plus the direct_match
// flag (own tags contain $tag) that feeds tagMatch and ranks direct matches first.
func inheritedTagProjection(table string) string {
	return table + ".*, " + jsonContains(table+".tags", "$tag") + " AS direct_match"
}

// jsonContains matches when the JSON array column contains the bound value.
func jsonContains(column, parameter string) string {
	return "EXISTS (SELECT 1 FROM json_each(" + column + ") WHERE value = " + parameter + ")"
}

// searchCondition is a case-insensitive substring match against $q. instr
// rather than LIKE so % and _ in user input stay literal.
func searchCondition(column string) string {
	return "instr(lower(" + column + "), lower($q)) > 0"
}

// excludedIDCondition drops rows whose id is in the bound $hidden list.
const excludedIDCondition = "id NOT IN (SELECT value FROM json_each($hidden))"

func (r *Repository) ListScenes(ctx context.Context, filters EntityFilters, limit, offset *int) ([]*model.Scene, error) {
	vars := db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)}
	var conditions []string
	if filters.Search != nil && *filters.Search != "" {
		vars["q"] = *filters.Search
		conditions = append(conditions, searchCondition("title"))
	}
	if rid := parseFilterID(filters.StudioID); rid != nil {
		vars["studio"] = *rid
		conditions = append(conditions, "studio = $studio")
	}
	if rid := parseFilterID(filters.PerformerID); rid != nil {
		vars["performer"] = *rid
		conditions = append(conditions, jsonContains("scene.performers", "$performer"))
	}
	tagFiltered := false
	if rid := parseFilterID(filters.TagID); rid != nil {
		vars["tag"] = *rid
		tagFiltered = true
		conditions = append(conditions, inheritedTagCondition("scene"))
	}
	if filters.MinRating != nil {
		vars["minRating"] = *filters.MinRating
		conditions = append(conditions, "rating >= $minRating")
	}
	if filters.MinDuration != nil {
		vars["minDuration"] = *filters.MinDuration
		conditions = append(conditions, "duration_seconds >= $minDuration")
	}
	if filters.MaxDuration != nil {
		vars["maxDuration"] = *filters.MaxDuration
		conditions = append(conditions, "duration_seconds <= $maxDuration")
	}
	if filters.DateFrom != nil && *filters.DateFrom != "" {
		vars["dateFrom"] = *filters.DateFrom
		conditions = append(conditions, "date >= $dateFrom")
	}
	if filters.DateTo != nil && *filters.DateTo != "" {
		vars["dateTo"] = *filters.DateTo
		conditions = append(conditions, "date <= $dateTo")
	}
	if len(filters.Sources) > 0 {
		vars["sources"] = filters.Sources
		conditions = append(conditions, sourceCondition("scene"))
	}
	if hidden := r.excludedSceneIDs(ctx, filters); len(hidden) > 0 {
		vars["hidden"] = hidden
		conditions = append(conditions, excludedIDCondition)
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + joinAnd(conditions)
	}
	projection, order := "*", sceneSort(filters.Sort)
	if tagFiltered {
		projection = inheritedTagProjection("scene")
		order = "direct_match DESC, " + order
	}
	query := fmt.Sprintf("SELECT %s FROM scene %s ORDER BY %s LIMIT $limit OFFSET $offset", projection, where, order)
	rows, err := r.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("list scenes: %w", err)
	}
	out := make([]*model.Scene, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.sceneFromMap(ctx, row))
	}
	return out, nil
}

func (r *Repository) GetScene(ctx context.Context, id string) (*model.Scene, error) {
	m, err := r.getByID(ctx, id)
	if err != nil || m == nil {
		return nil, err
	}
	return r.sceneFromMap(ctx, m), nil
}

func sceneSort(sort *string) string {
	if sort == nil {
		return "created_at DESC"
	}
	switch *sort {
	case "title":
		return "title ASC"
	case "rating":
		return "rating DESC"
	case "date":
		return "date DESC"
	case "views":
		return "view_count DESC"
	default:
		return "created_at DESC"
	}
}

// ---------------------------------------------------------------------------
// Relation loaders + shared helpers.

func (r *Repository) tagsByIDs(ctx context.Context, ids []db.RecordID) []*model.Tag {
	out := []*model.Tag{}
	if len(ids) == 0 {
		return out
	}
	rows, err := r.database.Query(ctx,
		"SELECT tag.* FROM json_each($ids) AS link JOIN tag ON tag.id = link.value ORDER BY link.key",
		db.Vars{"ids": ids})
	if err != nil {
		return out
	}
	for _, row := range rows {
		out = append(out, tagFromMap(row))
	}
	return out
}

func (r *Repository) performersByIDs(ctx context.Context, ids []db.RecordID) []*model.Performer {
	out := []*model.Performer{}
	if len(ids) == 0 {
		return out
	}
	rows, err := r.database.Query(ctx,
		"SELECT performer.* FROM json_each($ids) AS link JOIN performer ON performer.id = link.value ORDER BY link.key",
		db.Vars{"ids": ids})
	if err != nil {
		return out
	}
	for _, row := range rows {
		out = append(out, r.performerFromMap(ctx, row, false))
	}
	return out
}

// getByID loads a single record by "table:id".
func (r *Repository) getByID(ctx context.Context, id string) (map[string]any, error) {
	rid, err := db.ParseRecordID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q: %w", id, err)
	}
	return r.database.Get(ctx, *rid)
}

func parseFilterID(id *string) *db.RecordID {
	if id == nil || *id == "" {
		return nil
	}
	rid, err := db.ParseRecordID(*id)
	if err != nil {
		return nil
	}
	return rid
}

func joinAnd(conditions []string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}

// RelatedScenes returns scenes linked to sceneID by a visit: the site's own
// related list first, then fallback-search hits by performer, tag and title,
// each in rank order. Blocked scenes and disabled-plugin stubs are skipped.
func (r *Repository) RelatedScenes(ctx context.Context, sceneID string, limit int, disabledPlugins []string) ([]*model.Scene, error) {
	if limit <= 0 {
		limit = 24
	}
	rows, err := r.database.Query(ctx,
		`SELECT scene.* FROM scene_related
		 JOIN scene ON scene.id = scene_related.related
		 WHERE scene_related.scene = $scene
		   AND scene.id NOT IN (SELECT value FROM json_each($hidden))
		 ORDER BY CASE scene_related.source
		            WHEN 'site' THEN 0 WHEN 'performer' THEN 1 WHEN 'tag' THEN 2 ELSE 3 END,
		          scene_related.rank
		 LIMIT $limit`,
		db.Vars{"scene": sceneID, "hidden": r.excludedSceneIDs(ctx, EntityFilters{DisabledPlugins: disabledPlugins}), "limit": limit})
	if err != nil {
		return nil, fmt.Errorf("related scenes: %w", err)
	}
	out := make([]*model.Scene, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.sceneFromMap(ctx, row))
	}
	return out, nil
}
