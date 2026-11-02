package library

import (
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// MediaQuery is a page of videos or galleries matching a filter.
type MediaQuery struct {
	Filter schema.Filter
	Limit  int
	Offset int
}

// mediaSelect is the FROM clause every media query shares: the table as
// "media" and the user's state on each row as "state" (NULL when none).
func mediaSelect(kind schema.RecordKind) string {
	return ` FROM ` + db.Quote(tableFor(kind)) + ` AS "media"
		LEFT JOIN "UserState" AS "state" ON "state"."subjectKind" = '` + string(kind) + `'
			AND "state"."subjectId" = "media"."id" AND "state"."deletedAt" IS NULL`
}

// filterConditions turns a filter into WHERE conditions on "media" and
// "state". Lists of performers, studios, categories and sites match any of
// them; included tags must all be there, excluded tags none. Blocked media,
// and media crediting a blocked performer, studio or tag, never match.
// Fields a kind doesn't have (duration outside videos and audio, VR outside
// videos) don't filter.
func filterConditions(kind schema.RecordKind, filter schema.Filter) ([]string, db.Vars) {
	builder := &conditionBuilder{vars: db.Vars{"mediaKind": string(kind)}}
	builder.add(`"media"."deletedAt" IS NULL`)
	builder.add(notBlocked)
	if filter.Text != nil && strings.TrimSpace(*filter.Text) != "" {
		builder.bind(`"media"."title" LIKE $text`, "text", "%"+strings.TrimSpace(*filter.Text)+"%")
	}
	addCredits(builder, filter)
	addDates(builder, filter)
	addUserConditions(builder, filter)
	if hasDuration(kind) {
		addDurationConditions(builder, filter)
	}
	if kind == schema.RecordKindVideo && filter.Vr != nil {
		builder.add(flagCondition(`"media"."vr" IS NOT NULL`, *filter.Vr))
	}
	for index, condition := range filter.Fields {
		addFieldCondition(builder, index, condition)
	}
	return builder.conditions, builder.vars
}

// notBlocked excludes blocked media and media crediting anything blocked.
const notBlocked = `NOT EXISTS (SELECT 1 FROM "UserState" AS "blocked"
	WHERE "blocked"."blocked" = 1 AND "blocked"."deletedAt" IS NULL AND (
		("blocked"."subjectKind" = $mediaKind AND "blocked"."subjectId" = "media"."id")
		OR ("blocked"."subjectKind" = 'studio' AND "blocked"."subjectId" = "media"."studio")
		OR ("blocked"."subjectKind" = 'performer' AND "blocked"."subjectId" IN (SELECT value FROM json_each("media"."performers")))
		OR ("blocked"."subjectKind" = 'tag' AND "blocked"."subjectId" IN (SELECT value FROM json_each("media"."tags")))))`

// conditionBuilder collects WHERE conditions and their parameters.
type conditionBuilder struct {
	conditions []string
	vars       db.Vars
}

// add appends a condition without parameters.
func (b *conditionBuilder) add(condition string) {
	b.conditions = append(b.conditions, condition)
}

// bind appends a condition and binds its parameter.
func (b *conditionBuilder) bind(condition, name string, value any) {
	b.conditions = append(b.conditions, condition)
	b.vars[name] = value
}

// addCredits filters by performers, studios, tags, categories and sites.
func addCredits(builder *conditionBuilder, filter schema.Filter) {
	if len(filter.PerformerIds) > 0 {
		builder.bind(`EXISTS (SELECT 1 FROM json_each("media"."performers") WHERE value IN (SELECT value FROM json_each($performerIds)))`, "performerIds", filter.PerformerIds)
	}
	if len(filter.StudioIds) > 0 {
		builder.bind(`"media"."studio" IN (SELECT value FROM json_each($studioIds))`, "studioIds", filter.StudioIds)
	}
	if len(filter.IncludeTagIds) > 0 {
		builder.bind(`NOT EXISTS (SELECT 1 FROM json_each($includeTagIds) AS "wanted"
			WHERE "wanted".value NOT IN (SELECT value FROM json_each("media"."tags")))`, "includeTagIds", filter.IncludeTagIds)
	}
	if len(filter.ExcludeTagIds) > 0 {
		builder.bind(`NOT EXISTS (SELECT 1 FROM json_each("media"."tags") WHERE value IN (SELECT value FROM json_each($excludeTagIds)))`, "excludeTagIds", filter.ExcludeTagIds)
	}
	if len(filter.CategoryIds) > 0 {
		builder.bind(`EXISTS (SELECT 1 FROM "Category" AS "category", json_each("category"."matchTags") AS "match"
			WHERE "category"."id" IN (SELECT value FROM json_each($categoryIds)) AND "category"."deletedAt" IS NULL
				AND "match".value IN (SELECT value FROM json_each("media"."tags")))`, "categoryIds", filter.CategoryIds)
	}
	if len(filter.Sources) > 0 {
		builder.bind(`EXISTS (SELECT 1 FROM "Source" AS "source" WHERE "source"."subjectKind" = $mediaKind
			AND "source"."subjectId" = "media"."id" AND "source"."deletedAt" IS NULL
			AND "source"."plugin" IN (SELECT value FROM json_each($sources)))`, "sources", filter.Sources)
	}
}

// addDates filters by release date and status.
func addDates(builder *conditionBuilder, filter schema.Filter) {
	if filter.DateFrom != nil {
		builder.bind(`"media"."date" >= $dateFrom`, "dateFrom", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		builder.bind(`"media"."date" <= $dateTo`, "dateTo", *filter.DateTo)
	}
	if len(filter.Status) > 0 {
		builder.bind(`"media"."status" IN (SELECT value FROM json_each($status))`, "status", filter.Status)
	}
}

// addUserConditions filters by the user's state: rating, favourite, watched,
// and whether a file is stored.
func addUserConditions(builder *conditionBuilder, filter schema.Filter) {
	if filter.MinRating != nil {
		builder.bind(`"state"."rating" >= $minRating`, "minRating", *filter.MinRating)
	}
	if filter.Favorite != nil {
		builder.add(flagCondition(`COALESCE("state"."favorite", 0) = 1`, *filter.Favorite))
	}
	if filter.Unwatched != nil {
		builder.add(flagCondition(`COALESCE("state"."viewCount", 0) = 0`, *filter.Unwatched))
	}
	if filter.InLibrary != nil {
		builder.add(flagCondition(`EXISTS (SELECT 1 FROM "MediaFile" AS "file" WHERE "file"."subjectKind" = $mediaKind
			AND "file"."subjectId" = "media"."id" AND "file"."deletedAt" IS NULL AND "file"."missing" = 0)`, *filter.InLibrary))
	}
}

// hasDuration reports whether media of kind has a running time.
func hasDuration(kind schema.RecordKind) bool {
	return kind == schema.RecordKindVideo || kind == schema.RecordKindAudio
}

// addDurationConditions filters videos and audio by running time.
func addDurationConditions(builder *conditionBuilder, filter schema.Filter) {
	if filter.MinDurationSeconds != nil {
		builder.bind(`"media"."durationSeconds" >= $minDuration`, "minDuration", *filter.MinDurationSeconds)
	}
	if filter.MaxDurationSeconds != nil {
		builder.bind(`"media"."durationSeconds" <= $maxDuration`, "maxDuration", *filter.MaxDurationSeconds)
	}
}

// fieldOperators maps a custom field condition's operator to SQL.
var fieldOperators = map[string]string{"eq": "=", "ne": "!=", "lt": "<", "gt": ">"}

// addFieldCondition filters by one custom field value (UserState.fields).
func addFieldCondition(builder *conditionBuilder, index int, condition schema.FieldCondition) {
	path := fmt.Sprintf("fieldPath%d", index)
	value := fmt.Sprintf("fieldValue%d", index)
	builder.vars[path] = `$."` + strings.ReplaceAll(condition.FieldDefId, `"`, ``) + `"`
	extracted := `json_extract("state"."fields", $` + path + `)`
	switch condition.Operator {
	case "set":
		builder.add(extracted + ` IS NOT NULL`)
	case "unset":
		builder.add(extracted + ` IS NULL`)
	case "contains":
		builder.bind(`EXISTS (SELECT 1 FROM json_each("state"."fields", $`+path+`) WHERE value = $`+value+`)`, value, condition.Value)
	default:
		operator, known := fieldOperators[condition.Operator]
		if !known {
			operator = "="
		}
		builder.bind(extracted+` `+operator+` $`+value, value, condition.Value)
	}
}

// flagCondition is condition, or its negation when wanted is false.
func flagCondition(condition string, wanted bool) string {
	if wanted {
		return condition
	}
	return "NOT (" + condition + ")"
}

// sortColumns maps a sort field to the expression it orders by.
var sortColumns = map[schema.SortField]string{
	schema.SortFieldRelevance:    `"media"."createdAt"`,
	schema.SortFieldDate:         `"media"."date"`,
	schema.SortFieldAddedAt:      `"media"."createdAt"`,
	schema.SortFieldTitle:        `"media"."title" COLLATE NOCASE`,
	schema.SortFieldRating:       `"state"."rating"`,
	schema.SortFieldViewCount:    `COALESCE("state"."viewCount", 0)`,
	schema.SortFieldLastViewedAt: `"state"."lastViewedAt"`,
}

// orderBy renders the ORDER BY of a filter's sort; newest added first by
// default. Missing values sort last either way.
func orderBy(kind schema.RecordKind, sort *schema.Sort) string {
	if sort == nil {
		return `"media"."createdAt" DESC, "media".rowid DESC`
	}
	if sort.Field == schema.SortFieldRandom {
		return `RANDOM()`
	}
	column, known := sortColumns[sort.Field]
	if sort.Field == schema.SortFieldDuration && hasDuration(kind) {
		column, known = `"media"."durationSeconds"`, true
	}
	if !known {
		column = `"media"."createdAt"`
	}
	direction := "ASC"
	if sort.Descending {
		direction = "DESC"
	}
	return column + ` IS NULL, ` + column + ` ` + direction + `, "media".rowid DESC`
}
