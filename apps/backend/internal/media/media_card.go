package media

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// cardSource describes how to read card fields from one entity table.
type cardSource struct {
	table       string
	titleField  string
	posterField string
}

// cardSources covers every entity type a card can represent (collection
// members, watch history, watchlist).
var cardSources = []cardSource{
	{table: "scene", titleField: "title", posterField: "COALESCE(thumbnail_path, poster_path)"},
	{table: "gallery", titleField: "title", posterField: "cover_path"},
	{table: "image", titleField: "title", posterField: "file_path"},
	{table: "performer", titleField: "name", posterField: "image_path"},
	{table: "studio", titleField: "name", posterField: "image_path"},
}

// MediaCards resolves normalized card data for a set of record ids — one query
// per involved table — preserving the requested order.
func (r *Repository) MediaCards(ctx context.Context, ids []string) ([]*model.MediaCard, error) {
	cards := map[string]*model.MediaCard{}

	for _, source := range cardSources {
		tableIDs := recordIDsForTable(ids, source.table)
		if len(tableIDs) == 0 {
			continue
		}
		query := fmt.Sprintf("SELECT id, %s AS title, %s AS poster_path FROM %s WHERE id IN (SELECT value FROM json_each($ids))",
			source.titleField, source.posterField, source.table)
		rows, err := db.QueryAs[sceneCardRecord](ctx, r.database, query, db.Vars{"ids": tableIDs})
		if err != nil {
			return nil, fmt.Errorf("%s cards: %w", source.table, err)
		}
		for _, rec := range rows {
			id := recordIDString(rec.ID)
			title := rec.Title
			if title == "" && rec.PosterPath != nil {
				title = *rec.PosterPath
			}
			cards[id] = &model.MediaCard{
				MediaID:    id,
				MediaType:  source.table,
				Title:      title,
				PosterPath: rec.PosterPath,
			}
		}
	}

	out := make([]*model.MediaCard, 0, len(ids))
	for _, raw := range ids {
		if card, ok := cards[raw]; ok {
			out = append(out, card)
		}
	}
	return out, nil
}

type sceneCardRecord struct {
	ID         *db.RecordID `json:"id"`
	Title      string       `json:"title"`
	PosterPath *string      `json:"poster_path"`
}

// recordIDsForTable parses "table:id" strings belonging to the given table into
// RecordIDs, skipping ids for other tables.
func recordIDsForTable(ids []string, table string) []db.RecordID {
	out := make([]db.RecordID, 0, len(ids))
	for _, raw := range ids {
		if !strings.HasPrefix(raw, table+":") {
			continue
		}
		rid, err := db.ParseRecordID(raw)
		if err != nil {
			continue
		}
		out = append(out, *rid)
	}
	return out
}
