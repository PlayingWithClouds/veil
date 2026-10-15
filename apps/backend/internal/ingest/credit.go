package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// FindPerformer returns the id of the performer called name (matched like
// ingest does: spelling-insensitively, through aliases), or "" when the library
// has none.
func (s *Service) FindPerformer(ctx context.Context, name string) (string, error) {
	performerID, found, err := s.findNamed(ctx, "performer", strings.TrimSpace(name))
	if err != nil || !found {
		return "", err
	}
	return performerID.String(), nil
}

// CreditPerformer credits the performer called name to a scene, gallery or
// image, creating the performer record when the library has none yet, and
// returns the performer's id. A performer already credited stays credited once.
func (s *Service) CreditPerformer(ctx context.Context, mediaID, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("performer name is empty")
	}
	media, err := db.ParseRecordID(mediaID)
	if err != nil {
		return "", fmt.Errorf("invalid media id %q: %w", mediaID, err)
	}
	if media.Table != "scene" && media.Table != "gallery" && media.Table != "image" {
		return "", fmt.Errorf("performers cannot be credited to a %s", media.Table)
	}
	row, err := s.database.Get(ctx, *media)
	if err != nil {
		return "", err
	}
	if row == nil {
		return "", fmt.Errorf("%s not found", mediaID)
	}

	performerID, err := s.findOrCreateNamed(ctx, "performer", name)
	if err != nil {
		return "", fmt.Errorf("resolve performer %q: %w", name, err)
	}
	if err := s.writeCredit(ctx, performerID, mediaID, plugins.ScenePerformer{Name: name}); err != nil {
		return "", fmt.Errorf("credit performer: %w", err)
	}
	if err := s.addToPerformers(ctx, mediaID, row, performerID); err != nil {
		return "", err
	}
	return performerID.String(), nil
}

// addToPerformers appends performerID to the record's denormalized
// performers list unless it is already there.
func (s *Service) addToPerformers(ctx context.Context, mediaID string, row db.Row, performerID db.RecordID) error {
	performers := jsonStringList(row["performers"])
	for _, existing := range performers {
		if existing == performerID.String() {
			return nil
		}
	}
	return s.setField(ctx, mediaID, "performers", append(performers, performerID.String()))
}
