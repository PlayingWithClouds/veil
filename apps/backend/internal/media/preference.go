package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
)

// RecordPreference stores a comparative choice: `chosen` was picked over
// `rejected` when the two were shown side by side.
func (r *Repository) RecordPreference(ctx context.Context, chosenID, rejectedID string) error {
	chosen, err := db.ParseRecordID(chosenID)
	if err != nil {
		return fmt.Errorf("invalid chosen id: %w", err)
	}
	rejected, err := db.ParseRecordID(rejectedID)
	if err != nil {
		return fmt.Errorf("invalid rejected id: %w", err)
	}
	_, err = r.database.Insert(ctx, "preference_event", map[string]any{"chosen": *chosen, "rejected": *rejected})
	if err != nil {
		return fmt.Errorf("record preference: %w", err)
	}
	return nil
}
