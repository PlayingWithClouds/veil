package library

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// MediaEdit is the user's change to a video or gallery. Nil fields stay as
// they are; an empty string clears an optional field.
type MediaEdit struct {
	Title        *string
	Details      *string
	Date         *string
	StudioID     *string
	PerformerIDs *[]string
	TagIDs       *[]string
}

// NamedEdit is the user's change to a performer, studio or tag. Fields the
// kind doesn't have are ignored.
type NamedEdit struct {
	Name    *string
	Aliases *[]string
	Details *string
	// Performers only.
	Links *[]string
	// Studios only.
	URL      *string
	ParentID *string
}

// EditMedia applies the user's change to a video or gallery. Every field it
// sets is locked against later scrapes.
func (l *Library) EditMedia(ctx context.Context, kind schema.RecordKind, id string, edit MediaEdit) error {
	fields := db.Vars{}
	if edit.Title != nil {
		title := strings.TrimSpace(*edit.Title)
		if title == "" {
			return fmt.Errorf("a title can't be empty")
		}
		fields["title"] = title
	}
	setOptionalText(fields, "details", edit.Details)
	setOptionalText(fields, "date", edit.Date)
	setOptionalText(fields, "studio", edit.StudioID)
	if edit.PerformerIDs != nil {
		fields["performers"] = *edit.PerformerIDs
	}
	if edit.TagIDs != nil {
		fields["tags"] = *edit.TagIDs
	}
	_, err := l.recorder.Update(ctx, kind, id, fields, provenance.User())
	return err
}

// EditNamed applies the user's change to a performer, studio or tag. A new
// name refolds the name key; a tag can't take a name another tag has.
func (l *Library) EditNamed(ctx context.Context, kind schema.RecordKind, id string, edit NamedEdit) error {
	fields := db.Vars{}
	if err := l.setName(ctx, kind, id, fields, edit.Name); err != nil {
		return err
	}
	if edit.Aliases != nil {
		fields["aliases"] = *edit.Aliases
	}
	if kind == schema.RecordKindTag {
		_, err := l.recorder.Update(ctx, kind, id, fields, provenance.User())
		return err
	}
	setOptionalText(fields, "details", edit.Details)
	if kind == schema.RecordKindPerformer && edit.Links != nil {
		fields["links"] = *edit.Links
	}
	if kind == schema.RecordKindStudio {
		setOptionalText(fields, "url", edit.URL)
		setOptionalText(fields, "parent", edit.ParentID)
	}
	_, err := l.recorder.Update(ctx, kind, id, fields, provenance.User())
	return err
}

// setName sets name and nameKey from a new name, when there is one.
func (l *Library) setName(ctx context.Context, kind schema.RecordKind, id string, fields db.Vars, name *string) error {
	if name == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*name)
	key := names.Key(trimmed)
	if key == "" {
		return fmt.Errorf("a name can't be empty")
	}
	if kind == schema.RecordKindTag {
		existing, err := l.findByNameKey(ctx, tableTag, key)
		if err != nil {
			return err
		}
		if existing != "" && existing != id {
			return fmt.Errorf("another tag is already named %q", trimmed)
		}
	}
	fields["name"] = trimmed
	fields["nameKey"] = key
	return nil
}

// setOptionalText sets an optional text column from an edit: nil leaves it,
// blank clears it.
func setOptionalText(fields db.Vars, column string, value *string) {
	if value == nil {
		return
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		fields[column] = nil
		return
	}
	fields[column] = trimmed
}
