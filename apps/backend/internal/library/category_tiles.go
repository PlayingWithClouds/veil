package library

import (
	"context"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// CategoryTile is a category as a grid of them shows it for one media kind:
// how much of that kind is in it (subcategories included) and a picture.
type CategoryTile struct {
	Category schema.Category
	Count    int
	// The category's own cover, else the newest cover among its media.
	Preview *schema.Asset
}

// categoryCount is how many media one category holds.
type categoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// categoryPreview is the newest cover among a category's media.
type categoryPreview struct {
	Category string        `json:"category"`
	Cover    *schema.Asset `json:"cover"`
}

// pairsTable reads $pairs, a JSON list of [category, tag] pairs, as a table.
const pairsTable = `WITH "pairs"("category", "tag") AS (
	SELECT json_extract(value, '$[0]'), json_extract(value, '$[1]') FROM json_each($pairs))`

// CategoryTiles lists every category with its count and picture for the
// media of kind, in the order of Categories. Blocked and trashed media don't
// count. Two queries for all categories, rather than two per category.
func (l *Library) CategoryTiles(ctx context.Context, kind schema.MediaKind) ([]CategoryTile, error) {
	recordKind, err := recordKindOf(kind)
	if err != nil {
		return nil, err
	}
	categories, err := l.Categories(ctx)
	if err != nil {
		return nil, err
	}
	pairs, err := l.categoryTagPairs(ctx, categories)
	if err != nil {
		return nil, err
	}
	conditions, vars := filterConditions(recordKind, schema.Filter{})
	vars["pairs"] = pairs
	from := mediaSelect(recordKind) + `, json_each("media"."tags") AS "carried"
		JOIN "pairs" ON "pairs"."tag" = "carried".value
		WHERE ` + strings.Join(conditions, " AND ")
	counts, err := db.QueryAs[categoryCount](ctx, l.database, pairsTable+`
		SELECT "pairs"."category" AS "category", COUNT(DISTINCT "media"."id") AS "count"`+from+`
		GROUP BY "pairs"."category"`, vars)
	if err != nil {
		return nil, err
	}
	// SQLite takes the bare column (the cover) from the row with the MAX.
	previews, err := db.QueryAs[categoryPreview](ctx, l.database, pairsTable+`
		SELECT "pairs"."category" AS "category", "media"."cover" AS "cover", MAX("media"."createdAt")`+from+`
		AND "media"."cover" IS NOT NULL GROUP BY "pairs"."category"`, vars)
	if err != nil {
		return nil, err
	}
	return categoryTiles(categories, counts, previews), nil
}

// categoryTiles joins the categories with their counts and previews.
func categoryTiles(categories []schema.Category, counts []categoryCount, previews []categoryPreview) []CategoryTile {
	countOf := map[string]int{}
	for _, row := range counts {
		countOf[row.Category] = row.Count
	}
	previewOf := map[string]*schema.Asset{}
	for _, row := range previews {
		previewOf[row.Category] = row.Cover
	}
	tiles := make([]CategoryTile, 0, len(categories))
	for _, category := range categories {
		preview := category.Cover
		if preview == nil {
			preview = previewOf[category.Id]
		}
		tiles = append(tiles, CategoryTile{Category: category, Count: countOf[category.Id], Preview: preview})
	}
	return tiles
}

// categoryTagPairs lists [category, tag] for every tag that puts media in a
// category: its match tags and the tags its terms name, its subcategories'
// included.
func (l *Library) categoryTagPairs(ctx context.Context, categories []schema.Category) ([][2]string, error) {
	tagsOfTerm, err := l.tagsOfTerms(ctx, categories)
	if err != nil {
		return nil, err
	}
	children := map[string][]string{}
	byID := map[string]schema.Category{}
	for _, category := range categories {
		byID[category.Id] = category
		if category.Parent != nil {
			children[*category.Parent] = append(children[*category.Parent], category.Id)
		}
	}
	pairs := [][2]string{}
	for _, category := range categories {
		seen := map[string]bool{}
		for _, member := range subtreeOf(category.Id, children) {
			for _, tag := range tagsMatching(byID[member], tagsOfTerm) {
				if seen[tag] {
					continue
				}
				seen[tag] = true
				pairs = append(pairs, [2]string{category.Id, tag})
			}
		}
	}
	return pairs, nil
}

// subtreeOf lists a category and every category below it.
func subtreeOf(id string, children map[string][]string) []string {
	out := []string{}
	pending := []string{id}
	for len(pending) > 0 {
		current := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		out = append(out, current)
		pending = append(pending, children[current]...)
	}
	return out
}

// tagsMatching lists the tags one category matches by itself.
func tagsMatching(category schema.Category, tagsOfTerm map[string][]string) []string {
	tags := append([]string{}, category.MatchTags...)
	for _, term := range category.MatchTerms {
		tags = append(tags, tagsOfTerm[term]...)
	}
	return tags
}

// tagsOfTerms maps each term of the categories to the live tags whose
// folded name it is.
func (l *Library) tagsOfTerms(ctx context.Context, categories []schema.Category) (map[string][]string, error) {
	terms := []string{}
	for _, category := range categories {
		terms = append(terms, category.MatchTerms...)
	}
	rows, err := l.database.Query(ctx, `SELECT "id", "nameKey" FROM "Tag"
		WHERE "deletedAt" IS NULL AND "nameKey" IN (SELECT value FROM json_each($terms))`, db.Vars{"terms": terms})
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, row := range rows {
		key, _ := row["nameKey"].(string)
		id, _ := row["id"].(string)
		out[key] = append(out[key], id)
	}
	return out, nil
}
