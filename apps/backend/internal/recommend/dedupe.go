package recommend

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// duplicateQuery reads what identifies a scene across sites for the ids in $ids.
const duplicateQuery = `SELECT id, title, performers, duration_seconds
	FROM scene WHERE id IN (SELECT value FROM json_each($ids))`

// dropDuplicates keeps one card per video that several sites carry. scored is
// best first, so the copy that ranks highest (the best source for this user)
// stays and the others are dropped. Matching is internal/alike's SameScene on
// stored data only; scenes without a known runtime never match. A scene that
// credits no one (a search stub) is given the known performers its title names,
// which can only keep two videos apart, never join them.
func (e *Engine) dropDuplicates(ctx context.Context, scored []scoredItem) ([]scoredItem, error) {
	identities, err := e.loadIdentities(ctx, scoredSceneIDs(scored))
	if err != nil {
		return nil, err
	}
	kept := make([]scoredItem, 0, len(scored))
	keptIdentities := make([]alike.Candidate, 0, len(scored))
	for _, item := range scored {
		identity, known := identities[item.SceneID]
		if known && matchesAny(identity, keptIdentities) {
			continue
		}
		kept = append(kept, item)
		if known {
			keptIdentities = append(keptIdentities, identity)
		}
	}
	return kept, nil
}

// matchesAny reports whether identity is the same video as one of others.
func matchesAny(identity alike.Candidate, others []alike.Candidate) bool {
	for _, other := range others {
		if alike.SameScene(identity, other) {
			return true
		}
	}
	return false
}

// loadIdentities reads the matcher view of the scenes, keyed by scene id.
// Performers are compared by record id, which is already canonical across sites.
func (e *Engine) loadIdentities(ctx context.Context, ids []string) (map[string]alike.Candidate, error) {
	out := map[string]alike.Candidate{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx, duplicateQuery, db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("load duplicate identities: %w", err)
	}
	finder, err := e.performerFinder(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		identity := alike.Candidate{
			Title:           rowString(row, "title"),
			Performers:      rowStrings(row, "performers"),
			DurationSeconds: rowInt(row, "duration_seconds"),
		}
		if len(identity.Performers) == 0 {
			identity.Named = finder.Find(identity.Title)
		}
		out[rowString(row, "id")] = identity
	}
	return out, nil
}

// performerFinder finds the performers a title names, by their names and aliases.
func (e *Engine) performerFinder(ctx context.Context) (*names.Finder, error) {
	rows, err := e.database.Query(ctx, `SELECT id, name, aliases FROM performer`, nil)
	if err != nil {
		return nil, fmt.Errorf("load performer names: %w", err)
	}
	finder := names.NewFinder()
	for _, row := range rows {
		id := rowString(row, "id")
		finder.Add(id, rowString(row, "name"))
		for _, alias := range rowStrings(row, "aliases") {
			finder.Add(id, alias)
		}
	}
	return finder, nil
}
