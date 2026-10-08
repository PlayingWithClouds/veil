package recommend

import (
	"context"
	"fmt"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
)

// TasteEntry is one entity of the taste summary.
type TasteEntry struct {
	// ID is the record id; the plugin name for sites.
	ID        string
	Name      string
	ImagePath string
	// Affinity is in [-1, 1], normalized within its dimension.
	Affinity float64
}

// Taste is the user's profile for display: per dimension the strongest liked
// entries first, then the strongest disliked ones.
type Taste struct {
	SignalCount int
	Tags        []TasteEntry
	Performers  []TasteEntry
	Studios     []TasteEntry
	Sites       []TasteEntry
}

// Taste summarizes the current profile with at most limit liked and limit
// disliked entries per dimension, names and photos resolved.
func (e *Engine) Taste(ctx context.Context, limit int) (*Taste, error) {
	profile, _, err := e.profileAt(ctx, e.now())
	if err != nil {
		return nil, err
	}
	taste := &Taste{SignalCount: profile.SignalCount}
	for _, dimension := range []struct {
		affinities map[string]float64
		target     *[]TasteEntry
	}{
		{profile.Tags, &taste.Tags},
		{profile.Performers, &taste.Performers},
		{profile.Studios, &taste.Studios},
	} {
		entries, err := e.tasteEntries(ctx, dimension.affinities, limit)
		if err != nil {
			return nil, err
		}
		*dimension.target = entries
	}
	taste.Sites = siteEntries(profile.Sites, limit)
	return taste, nil
}

// tasteEntries picks the strongest liked and disliked ids and resolves their
// names and photos; ids whose record is gone are skipped.
func (e *Engine) tasteEntries(ctx context.Context, affinities map[string]float64, limit int) ([]TasteEntry, error) {
	picked := strongestFirst(affinities, limit)
	if len(picked) == 0 {
		return []TasteEntry{}, nil
	}
	ids := make([]string, 0, len(picked))
	for _, entry := range picked {
		ids = append(ids, entry.ID)
	}
	rows, err := e.database.Query(ctx,
		`SELECT id, name, image_path FROM performer WHERE id IN (SELECT value FROM json_each($ids))
		 UNION ALL SELECT id, name, image_path FROM studio WHERE id IN (SELECT value FROM json_each($ids))
		 UNION ALL SELECT id, name, NULL FROM tag WHERE id IN (SELECT value FROM json_each($ids))`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("taste entities: %w", err)
	}
	known := map[string]db.Row{}
	for _, row := range rows {
		known[rowString(row, "id")] = row
	}
	out := make([]TasteEntry, 0, len(picked))
	for _, entry := range picked {
		row, found := known[entry.ID]
		if !found {
			continue
		}
		entry.Name = rowString(row, "name")
		entry.ImagePath = rowString(row, "image_path")
		out = append(out, entry)
	}
	return out, nil
}

// siteEntries lists the strongest liked and disliked sites; a site's id and
// name are both its plugin name.
func siteEntries(affinities map[string]float64, limit int) []TasteEntry {
	picked := strongestFirst(affinities, limit)
	for index := range picked {
		picked[index].Name = picked[index].ID
	}
	return picked
}

// strongestFirst returns up to limit positive entries, highest first, followed
// by up to limit negative ones, most negative first. Ties break by id.
func strongestFirst(affinities map[string]float64, limit int) []TasteEntry {
	var liked, disliked []TasteEntry
	for id, affinity := range affinities {
		if affinity > 0 {
			liked = append(liked, TasteEntry{ID: id, Affinity: affinity})
		}
		if affinity < 0 {
			disliked = append(disliked, TasteEntry{ID: id, Affinity: affinity})
		}
	}
	sortByMagnitude(liked)
	sortByMagnitude(disliked)
	return append(truncateEntries(liked, limit), truncateEntries(disliked, limit)...)
}

// sortByMagnitude orders entries by absolute affinity, strongest first.
func sortByMagnitude(entries []TasteEntry) {
	sort.Slice(entries, func(left, right int) bool {
		leftMagnitude, rightMagnitude := absolute(entries[left].Affinity), absolute(entries[right].Affinity)
		if leftMagnitude != rightMagnitude {
			return leftMagnitude > rightMagnitude
		}
		return entries[left].ID < entries[right].ID
	})
}

// truncateEntries keeps the first limit entries.
func truncateEntries(entries []TasteEntry, limit int) []TasteEntry {
	if len(entries) > limit {
		return entries[:limit]
	}
	return entries
}

// absolute is |value|.
func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
