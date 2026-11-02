package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Collection tables.
const (
	collectionTable = "Collection"
	itemTable       = "CollectionItem"
)

// builtInCollections are the collections every library has, by kind.
var builtInCollections = map[schema.CollectionKind]string{
	schema.CollectionKindWatchlist: "Watchlist",
	schema.CollectionKindQueue:     "Queue",
}

// EnsureBuiltInCollections creates the watchlist and the play queue when
// they don't exist yet.
func (s *Service) EnsureBuiltInCollections(ctx context.Context) error {
	for kind, name := range builtInCollections {
		if _, err := s.BuiltIn(ctx, kind); err == nil {
			continue
		}
		if _, err := s.database.Insert(ctx, collectionTable, db.Vars{"name": name, "kind": kind}); err != nil {
			return err
		}
	}
	return nil
}

// BuiltIn returns the watchlist or the queue.
func (s *Service) BuiltIn(ctx context.Context, kind schema.CollectionKind) (*schema.Collection, error) {
	collection, err := db.QueryOneAs[schema.Collection](ctx, s.database,
		`SELECT * FROM "Collection" WHERE "kind" = $kind AND "deletedAt" IS NULL ORDER BY "createdAt" LIMIT 1`, db.Vars{"kind": kind})
	if err != nil {
		return nil, err
	}
	if collection == nil {
		return nil, fmt.Errorf("no %s collection", kind)
	}
	return collection, nil
}

// Collections lists every collection: the watchlist and queue first, then
// the user's by name.
func (s *Service) Collections(ctx context.Context) ([]schema.Collection, error) {
	return db.QueryAs[schema.Collection](ctx, s.database,
		`SELECT * FROM "Collection" WHERE "deletedAt" IS NULL
		 ORDER BY "kind" = 'user', "kind", "name" COLLATE NOCASE`, nil)
}

// Collection loads one collection; nil when it doesn't exist.
func (s *Service) Collection(ctx context.Context, id string) (*schema.Collection, error) {
	return db.Get[schema.Collection](ctx, s.database, collectionTable, id)
}

// CreateCollection adds a collection of the user's.
func (s *Service) CreateCollection(ctx context.Context, name, details string) (*schema.Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("a collection needs a name")
	}
	fields := db.Vars{"name": name, "kind": schema.CollectionKindUser}
	if strings.TrimSpace(details) != "" {
		fields["details"] = details
	}
	id, err := s.database.Insert(ctx, collectionTable, fields)
	if err != nil {
		return nil, err
	}
	return s.Collection(ctx, id)
}

// EditCollection renames a collection or changes its details or cover. Nil
// fields stay; empty details or cover clear them.
func (s *Service) EditCollection(ctx context.Context, id string, name, details, coverURI *string) (*schema.Collection, error) {
	fields := db.Vars{}
	if name != nil {
		if strings.TrimSpace(*name) == "" {
			return nil, fmt.Errorf("a collection needs a name")
		}
		fields["name"] = strings.TrimSpace(*name)
	}
	setOptional(fields, "details", details)
	if coverURI != nil {
		fields["cover"] = assetOrNil(*coverURI)
	}
	if _, err := provenance.New(s.database).Update(ctx, schema.RecordKindCollection, id, fields, provenance.User()); err != nil {
		return nil, err
	}
	return s.Collection(ctx, id)
}

// setOptional sets an optional text column: nil leaves it, blank clears it.
func setOptional(fields db.Vars, column string, value *string) {
	if value == nil {
		return
	}
	if strings.TrimSpace(*value) == "" {
		fields[column] = nil
		return
	}
	fields[column] = *value
}

// assetOrNil wraps a URI as an asset; nil for a blank one.
func assetOrNil(uri string) *schema.Asset {
	if strings.TrimSpace(uri) == "" {
		return nil
	}
	return &schema.Asset{Uri: strings.TrimSpace(uri)}
}

// RequireUserCollection fails for the built-in collections, which can't be
// trashed.
func (s *Service) RequireUserCollection(ctx context.Context, id string) error {
	collection, err := s.Collection(ctx, id)
	if err != nil {
		return err
	}
	if collection == nil {
		return fmt.Errorf("no collection %s", id)
	}
	if collection.Kind != schema.CollectionKindUser {
		return fmt.Errorf("the %s can't be deleted", collection.Name)
	}
	return nil
}

// Items lists a collection's items in order.
func (s *Service) Items(ctx context.Context, collectionID string, limit, offset int) ([]schema.CollectionItem, error) {
	return db.QueryAs[schema.CollectionItem](ctx, s.database,
		`SELECT * FROM "CollectionItem" WHERE "collection" = $collection AND "deletedAt" IS NULL
		 ORDER BY "position", rowid LIMIT $limit OFFSET $offset`,
		db.Vars{"collection": collectionID, "limit": limit, "offset": offset})
}

// ItemCount counts a collection's items.
func (s *Service) ItemCount(ctx context.Context, collectionID string) (int, error) {
	return s.database.Int(ctx, `SELECT COUNT(*) FROM "CollectionItem" WHERE "collection" = $collection AND "deletedAt" IS NULL`,
		db.Vars{"collection": collectionID})
}

// AddItems puts records into a collection at position (the end when nil,
// 0 = first, as "play next" does for the queue). A record already in it
// moves there instead of appearing twice.
func (s *Service) AddItems(ctx context.Context, collectionID string, refs []schema.RecordRef, position *int) error {
	return s.database.Tx(ctx, func(tx *db.DB) error {
		ids, err := itemSubjects(ctx, tx, collectionID)
		if err != nil {
			return err
		}
		ids = withoutRefs(ids, refs)
		at := len(ids)
		if position != nil && *position >= 0 && *position < at {
			at = *position
		}
		ordered := append(append(append([]schema.RecordRef{}, ids[:at]...), refs...), ids[at:]...)
		return rewriteItems(ctx, tx, collectionID, ordered)
	})
}

// RemoveItems takes records out of a collection.
func (s *Service) RemoveItems(ctx context.Context, collectionID string, refs []schema.RecordRef) error {
	return s.database.Tx(ctx, func(tx *db.DB) error {
		ids, err := itemSubjects(ctx, tx, collectionID)
		if err != nil {
			return err
		}
		return rewriteItems(ctx, tx, collectionID, withoutRefs(ids, refs))
	})
}

// MoveItem puts a record of the collection at a new position.
func (s *Service) MoveItem(ctx context.Context, collectionID string, ref schema.RecordRef, position int) error {
	return s.AddItems(ctx, collectionID, []schema.RecordRef{ref}, &position)
}

// ClearItems empties a collection (e.g. the queue).
func (s *Service) ClearItems(ctx context.Context, collectionID string) error {
	_, err := s.database.Exec(ctx, `DELETE FROM "CollectionItem" WHERE "collection" = $collection`, db.Vars{"collection": collectionID})
	return err
}

// CollectionsWith lists the collections a record is in.
func (s *Service) CollectionsWith(ctx context.Context, ref schema.RecordRef) ([]schema.Collection, error) {
	return db.QueryAs[schema.Collection](ctx, s.database,
		`SELECT * FROM "Collection" WHERE "deletedAt" IS NULL AND "id" IN (
		   SELECT "collection" FROM "CollectionItem" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL)
		 ORDER BY "kind" = 'user', "kind", "name" COLLATE NOCASE`,
		db.Vars{"kind": ref.Kind, "id": ref.Id})
}

// itemSubjects lists the records of a collection in order.
func itemSubjects(ctx context.Context, tx *db.DB, collectionID string) ([]schema.RecordRef, error) {
	items, err := db.QueryAs[schema.CollectionItem](ctx, tx,
		`SELECT * FROM "CollectionItem" WHERE "collection" = $collection AND "deletedAt" IS NULL ORDER BY "position", rowid`,
		db.Vars{"collection": collectionID})
	if err != nil {
		return nil, err
	}
	refs := make([]schema.RecordRef, 0, len(items))
	for _, item := range items {
		refs = append(refs, schema.RecordRef{Kind: item.SubjectKind, Id: item.SubjectId})
	}
	return refs, nil
}

// withoutRefs drops the given records from a list.
func withoutRefs(refs, dropped []schema.RecordRef) []schema.RecordRef {
	drop := map[schema.RecordRef]bool{}
	for _, ref := range dropped {
		drop[ref] = true
	}
	kept := []schema.RecordRef{}
	for _, ref := range refs {
		if !drop[ref] {
			kept = append(kept, ref)
		}
	}
	return kept
}

// rewriteItems replaces a collection's items with refs in order. Rows are
// rewritten rather than renumbered: collections are short, and positions
// stay dense.
func rewriteItems(ctx context.Context, tx *db.DB, collectionID string, refs []schema.RecordRef) error {
	if _, err := tx.Exec(ctx, `DELETE FROM "CollectionItem" WHERE "collection" = $collection`, db.Vars{"collection": collectionID}); err != nil {
		return err
	}
	for position, ref := range refs {
		if err := requireRecord(ctx, tx, ref.Kind, ref.Id); err != nil {
			return err
		}
		row := db.Vars{"collection": collectionID, "subjectKind": ref.Kind, "subjectId": ref.Id, "position": position}
		if _, err := tx.Insert(ctx, itemTable, row); err != nil {
			return err
		}
	}
	return nil
}
