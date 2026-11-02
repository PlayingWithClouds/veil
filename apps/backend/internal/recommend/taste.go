package recommend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TasteEntry is one entity of the taste summary.
type TasteEntry struct {
	// Id is the record id; the plugin name for sites.
	Id    string
	Name  string
	Image *schema.Asset
	// Weight is the affinity in [-1, 1], normalized within its dimension.
	Weight float64
}

// Taste is the user's profile for display: per dimension the strongest liked
// entries first, then the strongest disliked ones (at most tasteEntryLimit
// of each).
type Taste struct {
	// SignalCount is how many signals the profile was built from; 0 is a
	// cold start.
	SignalCount int
	Tags        []TasteEntry
	Performers  []TasteEntry
	Studios     []TasteEntry
	Sites       []TasteEntry
}

// Taste summarizes the current profile, names and pictures resolved.
func (e *Engine) Taste(ctx context.Context) (Taste, error) {
	taste, _, _, err := e.profileAt(ctx, e.now())
	if err != nil {
		return Taste{}, err
	}
	summary := Taste{SignalCount: taste.signalCount, Sites: siteEntries(taste.sites)}
	dimensions := []struct {
		table      string
		affinities map[string]float64
		target     *[]TasteEntry
	}{
		{"Tag", taste.tags, &summary.Tags},
		{"Performer", taste.performers, &summary.Performers},
		{"Studio", taste.studios, &summary.Studios},
	}
	for _, dimension := range dimensions {
		entries, err := e.tasteEntries(ctx, dimension.table, dimension.affinities)
		if err != nil {
			return Taste{}, err
		}
		*dimension.target = entries
	}
	return summary, nil
}

// pictureColumn is the picture column of each taste table; tags have none.
var pictureColumn = map[string]string{"Performer": `"image"`, "Studio": `"image"`, "Tag": `NULL`}

// tasteEntries picks the strongest liked and disliked ids of one table and
// resolves their names and pictures; ids whose record is gone are skipped.
func (e *Engine) tasteEntries(ctx context.Context, table string, affinities map[string]float64) ([]TasteEntry, error) {
	picked := strongestFirst(affinities)
	if len(picked) == 0 {
		return []TasteEntry{}, nil
	}
	ids := make([]string, 0, len(picked))
	for _, entry := range picked {
		ids = append(ids, entry.Id)
	}
	rows, err := e.database.Query(ctx,
		`SELECT "id", "name", `+pictureColumn[table]+` AS "image" FROM `+db.Quote(table)+`
		 WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("taste %s: %w", table, err)
	}
	known := map[string]db.Row{}
	for _, row := range rows {
		known[rowString(row, "id")] = row
	}
	out := make([]TasteEntry, 0, len(picked))
	for _, entry := range picked {
		row, found := known[entry.Id]
		if !found {
			continue
		}
		entry.Name = rowString(row, "name")
		entry.Image = rowAsset(row, "image")
		out = append(out, entry)
	}
	return out, nil
}

// rowAsset decodes an asset column; nil when NULL or malformed.
func rowAsset(row db.Row, column string) *schema.Asset {
	text := rowString(row, column)
	if text == "" {
		return nil
	}
	var asset schema.Asset
	if json.Unmarshal([]byte(text), &asset) != nil {
		return nil
	}
	return &asset
}

// siteEntries lists the strongest liked and disliked sites; a site's id and
// name are both its plugin name.
func siteEntries(affinities map[string]float64) []TasteEntry {
	picked := strongestFirst(affinities)
	for index := range picked {
		picked[index].Name = picked[index].Id
	}
	return picked
}

// strongestFirst returns up to tasteEntryLimit positive entries, highest
// first, followed by up to tasteEntryLimit negative ones, most negative
// first. Ties break by id.
func strongestFirst(affinities map[string]float64) []TasteEntry {
	liked, disliked := []TasteEntry{}, []TasteEntry{}
	for id, affinity := range affinities {
		if affinity > 0 {
			liked = append(liked, TasteEntry{Id: id, Weight: affinity})
		}
		if affinity < 0 {
			disliked = append(disliked, TasteEntry{Id: id, Weight: affinity})
		}
	}
	sortByMagnitude(liked)
	sortByMagnitude(disliked)
	return append(liked[:min(tasteEntryLimit, len(liked))], disliked[:min(tasteEntryLimit, len(disliked))]...)
}

// sortByMagnitude orders entries by absolute affinity, strongest first.
func sortByMagnitude(entries []TasteEntry) {
	sort.Slice(entries, func(left, right int) bool {
		leftMagnitude, rightMagnitude := absolute(entries[left].Weight), absolute(entries[right].Weight)
		if leftMagnitude != rightMagnitude {
			return leftMagnitude > rightMagnitude
		}
		return entries[left].Id < entries[right].Id
	})
}

// absolute is |value|.
func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
