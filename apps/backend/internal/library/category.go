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

// tableCategory is the table categories live in.
const tableCategory = "Category"

// CategoryEdit is a new category or a change to one. Nil fields stay as
// they are (or take their default on create); an empty string clears the
// parent or cover.
type CategoryEdit struct {
	Name        *string
	Slug        *string
	ParentID    *string
	MatchTagIDs *[]string
	CoverURI    *string
	Position    *int
}

// CreateCategory adds a category of our own taxonomy. Media is in it when
// it carries any of its match tags. The slug defaults to the folded name.
func (l *Library) CreateCategory(ctx context.Context, edit CategoryEdit) (*schema.Category, error) {
	if edit.Name == nil || strings.TrimSpace(*edit.Name) == "" {
		return nil, fmt.Errorf("a category needs a name")
	}
	fields, err := l.categoryFields(ctx, "", edit)
	if err != nil {
		return nil, err
	}
	if _, hasSlug := fields["slug"]; !hasSlug {
		fields["slug"] = slugOf(*edit.Name)
	}
	fillListDefaults(fields, "matchTags")
	if _, hasPosition := fields["position"]; !hasPosition {
		fields["position"] = 0
	}
	if err := l.requireFreeSlug(ctx, "", fields["slug"].(string)); err != nil {
		return nil, err
	}
	id, err := l.database.Insert(ctx, tableCategory, fields)
	if err != nil {
		return nil, err
	}
	return l.Category(ctx, id)
}

// EditCategory changes a category.
func (l *Library) EditCategory(ctx context.Context, id string, edit CategoryEdit) (*schema.Category, error) {
	fields, err := l.categoryFields(ctx, id, edit)
	if err != nil {
		return nil, err
	}
	if slug, hasSlug := fields["slug"].(string); hasSlug {
		if err := l.requireFreeSlug(ctx, id, slug); err != nil {
			return nil, err
		}
	}
	if _, err := l.recorder.Update(ctx, schema.RecordKindCategory, id, fields, provenance.User()); err != nil {
		return nil, err
	}
	return l.Category(ctx, id)
}

// categoryFields turns an edit into columns, refusing a parent that would
// make a cycle.
func (l *Library) categoryFields(ctx context.Context, id string, edit CategoryEdit) (db.Vars, error) {
	fields := db.Vars{}
	if edit.Name != nil {
		fields["name"] = strings.TrimSpace(*edit.Name)
	}
	if edit.Slug != nil {
		fields["slug"] = slugOf(*edit.Slug)
	}
	if edit.MatchTagIDs != nil {
		fields["matchTags"] = *edit.MatchTagIDs
	}
	if edit.Position != nil {
		fields["position"] = *edit.Position
	}
	if edit.CoverURI != nil {
		fields["cover"] = assetOf(*edit.CoverURI)
	}
	if edit.ParentID == nil {
		return fields, nil
	}
	setOptionalText(fields, "parent", edit.ParentID)
	return fields, l.requireNoCycle(ctx, id, strings.TrimSpace(*edit.ParentID))
}

// requireNoCycle fails when parentID is the category itself or one of its
// descendants.
func (l *Library) requireNoCycle(ctx context.Context, id, parentID string) error {
	for current := parentID; current != "" && id != ""; {
		if current == id {
			return fmt.Errorf("a category can't be inside itself")
		}
		parent, err := l.database.Strings(ctx, `SELECT "parent" FROM "Category" WHERE "id" = $id`, db.Vars{"id": current})
		if err != nil {
			return err
		}
		current = firstString(parent)
	}
	return nil
}

// requireFreeSlug fails when another category has slug.
func (l *Library) requireFreeSlug(ctx context.Context, id, slug string) error {
	if slug == "" {
		return fmt.Errorf("a category needs a slug")
	}
	taken, err := l.database.Int(ctx, `SELECT COUNT(*) FROM "Category" WHERE "slug" = $slug AND "id" != $id`, db.Vars{"slug": slug, "id": id})
	if err != nil {
		return err
	}
	if taken > 0 {
		return fmt.Errorf("another category already has the slug %q", slug)
	}
	return nil
}

// slugOf folds a name into a URL slug: "Big Tits" → "big-tits".
func slugOf(name string) string {
	return strings.ReplaceAll(names.Key(name), " ", "-")
}

// Categories lists every category, by position then name. The tree is
// built from parent.
func (l *Library) Categories(ctx context.Context) ([]schema.Category, error) {
	return db.QueryAs[schema.Category](ctx, l.database,
		`SELECT * FROM "Category" WHERE "deletedAt" IS NULL ORDER BY "position", "name" COLLATE NOCASE`, nil)
}

// Category loads one category; nil when it doesn't exist or is trashed.
func (l *Library) Category(ctx context.Context, id string) (*schema.Category, error) {
	return db.Get[schema.Category](ctx, l.database, tableCategory, id)
}

// CategoryBySlug loads the category with slug.
func (l *Library) CategoryBySlug(ctx context.Context, slug string) (*schema.Category, error) {
	return db.QueryOneAs[schema.Category](ctx, l.database,
		`SELECT * FROM "Category" WHERE "slug" = $slug AND "deletedAt" IS NULL`, db.Vars{"slug": slug})
}

// CategoriesOfTags lists the categories media with these tags belongs to.
func (l *Library) CategoriesOfTags(ctx context.Context, tagIDs []string) ([]schema.Category, error) {
	return db.QueryAs[schema.Category](ctx, l.database,
		`SELECT * FROM "Category" WHERE "deletedAt" IS NULL
		 AND EXISTS (SELECT 1 FROM json_each("matchTags") WHERE value IN (SELECT value FROM json_each($tags)))
		 ORDER BY "position", "name" COLLATE NOCASE`,
		db.Vars{"tags": tagIDs})
}

// ReorderCategories gives categories the positions of their order in ids.
func (l *Library) ReorderCategories(ctx context.Context, ids []string) error {
	for position, id := range ids {
		if _, err := l.recorder.Update(ctx, schema.RecordKindCategory, id, db.Vars{"position": position}, provenance.User()); err != nil {
			return err
		}
	}
	return nil
}
