package media

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// blocklistRecord mirrors the `blocklist_entry` table.
type blocklistRecord struct {
	ID        *db.RecordID `json:"id,omitempty"`
	Kind      string       `json:"kind"`
	Target    *db.RecordID `json:"target,omitempty"`
	Label     *string      `json:"label,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

func (b blocklistRecord) toModel() *model.BlocklistEntry {
	return &model.BlocklistEntry{
		ID:        recordIDString(b.ID),
		Kind:      b.Kind,
		TargetID:  recordIDString(b.Target),
		Label:     b.Label,
		CreatedAt: b.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ListBlocklist returns the blocked entities, newest first.
func (r *Repository) ListBlocklist(ctx context.Context) ([]*model.BlocklistEntry, error) {
	rows, err := db.QueryAs[blocklistRecord](ctx, r.database,
		`SELECT * FROM blocklist_entry ORDER BY created_at DESC`, nil)
	if err != nil {
		return nil, fmt.Errorf("list blocklist: %w", err)
	}
	out := make([]*model.BlocklistEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toModel())
	}
	return out, nil
}

// AddBlock blocks a tag/performer/studio (idempotent: an existing block on the
// same target is replaced).
func (r *Repository) AddBlock(ctx context.Context, kind, targetID string, label *string) (*model.BlocklistEntry, error) {
	target, err := db.ParseRecordID(targetID)
	if err != nil {
		return nil, fmt.Errorf("invalid target id: %w", err)
	}
	fields := map[string]any{"kind": kind, "target": *target}
	if label != nil && *label != "" {
		fields["label"] = *label
	}
	var id db.RecordID
	err = r.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM blocklist_entry WHERE target = $target`, db.Vars{"target": *target}); err != nil {
			return fmt.Errorf("clear existing block: %w", err)
		}
		insertedID, insertErr := tx.Insert(ctx, "blocklist_entry", fields)
		id = insertedID
		return insertErr
	})
	if err != nil {
		return nil, fmt.Errorf("add block: %w", err)
	}
	row, err := db.QueryOneAs[blocklistRecord](ctx, r.database,
		`SELECT * FROM blocklist_entry WHERE id = $id`, db.Vars{"id": id})
	if err != nil {
		return nil, fmt.Errorf("add block: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("add block: no record returned")
	}
	return row.toModel(), nil
}

// RemoveBlock unblocks a target.
func (r *Repository) RemoveBlock(ctx context.Context, targetID string) (bool, error) {
	target, err := db.ParseRecordID(targetID)
	if err != nil {
		return false, fmt.Errorf("invalid target id: %w", err)
	}
	if _, err := r.database.Exec(ctx,
		`DELETE FROM blocklist_entry WHERE target = $target`,
		db.Vars{"target": *target}); err != nil {
		return false, fmt.Errorf("remove block: %w", err)
	}
	return true, nil
}

// notBlockedCondition drops rows of an entity table (tag, performer or studio)
// that are on the blocklist, for the index listings.
func notBlockedCondition(table string) string {
	return "id NOT IN (SELECT target FROM blocklist_entry WHERE kind = '" + table + "')"
}

// blockedTargets returns the blocked target record ids grouped by kind.
func (r *Repository) blockedTargets(ctx context.Context) (tags, performers, studios []db.RecordID, err error) {
	rows, err := db.QueryAs[blocklistRecord](ctx, r.database,
		`SELECT kind, target FROM blocklist_entry`, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("blocked targets: %w", err)
	}
	for _, row := range rows {
		if row.Target == nil {
			continue
		}
		switch row.Kind {
		case "tag":
			tags = append(tags, *row.Target)
		case "performer":
			performers = append(performers, *row.Target)
		case "studio":
			studios = append(studios, *row.Target)
		}
	}
	return tags, performers, studios, nil
}

// BlockedSceneRIDs returns scene ids that should be hidden because they carry a
// blocked tag/performer/studio. Best-effort: on error returns nil.
func (r *Repository) BlockedSceneRIDs(ctx context.Context) []db.RecordID {
	tags, performers, studios, err := r.blockedTargets(ctx)
	if err != nil {
		return nil
	}
	if len(tags) == 0 && len(performers) == 0 && len(studios) == 0 {
		return nil
	}
	scenes, err := r.database.Strings(ctx,
		`SELECT id FROM scene
			WHERE EXISTS (SELECT 1 FROM json_each(scene.tags) WHERE value IN (SELECT value FROM json_each($tags)))
			OR EXISTS (SELECT 1 FROM json_each(scene.performers) WHERE value IN (SELECT value FROM json_each($performers)))
			OR scene.studio IN (SELECT value FROM json_each($studios))`,
		db.Vars{"tags": tags, "performers": performers, "studios": studios})
	if err != nil {
		return nil
	}
	return parseRecordIDs(scenes)
}

// BlockedTargetIDSet returns the blocked entity ids as a lookup set of
// "table:id" strings, for dropping blocked entities from search results.
func (r *Repository) BlockedTargetIDSet(ctx context.Context) map[string]bool {
	tags, performers, studios, err := r.blockedTargets(ctx)
	if err != nil {
		return nil
	}
	set := make(map[string]bool, len(tags)+len(performers)+len(studios))
	for _, group := range [][]db.RecordID{tags, performers, studios} {
		for i := range group {
			set[group[i].String()] = true
		}
	}
	return set
}
