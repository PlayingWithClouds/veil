package media

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// collectionMemberTables lists the tables a collection may contain. Collections
// cannot contain collections (no nesting).
var collectionMemberTables = map[string]bool{
	"scene":     true,
	"gallery":   true,
	"image":     true,
	"performer": true,
	"studio":    true,
}

// collectionRecord mirrors the `collection` table.
type collectionRecord struct {
	ID          *db.RecordID  `json:"id,omitempty"`
	UserCreated bool          `json:"user_created"`
	Name        string        `json:"name"`
	Details     *string       `json:"details,omitempty"`
	CoverPath   *string       `json:"cover_path,omitempty"`
	SourceURL   *string       `json:"source_url,omitempty"`
	ExternalID  *string       `json:"external_id,omitempty"`
	Tags        []db.RecordID `json:"tags"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (r *Repository) collectionToModel(ctx context.Context, c collectionRecord) *model.Collection {
	origin := "scraped"
	if c.UserCreated {
		origin = "user"
	}
	itemCount, _ := r.collectionItemCount(ctx, *c.ID)
	coverPath := c.CoverPath
	if coverPath == nil || *coverPath == "" {
		coverPath = r.collectionCover(ctx, *c.ID)
	}
	return &model.Collection{
		ID:         recordIDString(c.ID),
		Name:       c.Name,
		Details:    c.Details,
		CoverPath:  coverPath,
		ItemCount:  itemCount,
		Origin:     origin,
		SourceURL:  c.SourceURL,
		ExternalID: c.ExternalID,
		Tags:       r.tagsByIDs(ctx, c.Tags),
		CreatedAt:  c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  c.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// CreateCollection creates a named user collection (playlist).
func (r *Repository) CreateCollection(ctx context.Context, name string) (*model.Collection, error) {
	id, err := r.database.Insert(ctx, "collection", map[string]any{"name": name, "user_created": true})
	if err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}
	return r.requireCollection(ctx, id, "create collection")
}

// RenameCollection changes a collection's name.
func (r *Repository) RenameCollection(ctx context.Context, collectionID, name string) (*model.Collection, error) {
	id, err := db.ParseRecordID(collectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection id: %w", err)
	}
	if err := r.database.Merge(ctx, *id, map[string]any{"name": name}); err != nil {
		return nil, fmt.Errorf("rename collection: %w", err)
	}
	return r.requireCollection(ctx, *id, "rename collection")
}

// requireCollection loads a collection that must exist after a write.
func (r *Repository) requireCollection(ctx context.Context, id db.RecordID, operation string) (*model.Collection, error) {
	collection, err := r.GetCollection(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	if collection == nil {
		return nil, fmt.Errorf("%s: collection %s not found", operation, id)
	}
	return collection, nil
}

// DeleteCollection removes a collection and all its membership rows.
func (r *Repository) DeleteCollection(ctx context.Context, collectionID string) (bool, error) {
	id, err := db.ParseRecordID(collectionID)
	if err != nil {
		return false, fmt.Errorf("invalid collection id: %w", err)
	}
	err = r.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM collection_item WHERE collection = $id`, db.Vars{"id": *id}); err != nil {
			return fmt.Errorf("delete collection items: %w", err)
		}
		if err := tx.Delete(ctx, *id); err != nil {
			return fmt.Errorf("delete collection: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddToCollection appends any allowed entity to a collection (idempotent),
// placing it last. Asset images (content = false) are rejected.
func (r *Repository) AddToCollection(ctx context.Context, collectionID, mediaID string) (bool, error) {
	collection, media, err := parseCollectionMedia(collectionID, mediaID)
	if err != nil {
		return false, err
	}
	if media.Table == "image" {
		if err := r.requireContentImage(ctx, *media); err != nil {
			return false, err
		}
	}
	if exists, err := r.collectionContains(ctx, *collection, *media); err != nil {
		return false, err
	} else if exists {
		return true, nil
	}
	position, err := r.collectionItemCount(ctx, *collection)
	if err != nil {
		return false, err
	}
	_, err = r.database.Insert(ctx, "collection_item", map[string]any{
		"collection": *collection,
		"media":      *media,
		"position":   position,
	})
	if err != nil {
		return false, fmt.Errorf("add to collection: %w", err)
	}
	return true, nil
}

// RemoveFromCollection removes a member from a collection.
func (r *Repository) RemoveFromCollection(ctx context.Context, collectionID, mediaID string) (bool, error) {
	collection, media, err := parseCollectionMedia(collectionID, mediaID)
	if err != nil {
		return false, err
	}
	_, err = r.database.Exec(ctx,
		`DELETE FROM collection_item WHERE collection = $collection AND media = $media`,
		db.Vars{"collection": *collection, "media": *media})
	if err != nil {
		return false, fmt.Errorf("remove from collection: %w", err)
	}
	return true, nil
}

// ReorderCollection sets each member's position to its index in mediaIDs.
func (r *Repository) ReorderCollection(ctx context.Context, collectionID string, mediaIDs []string) (bool, error) {
	collection, err := db.ParseRecordID(collectionID)
	if err != nil {
		return false, fmt.Errorf("invalid collection id: %w", err)
	}
	mediaRIDs := make([]db.RecordID, 0, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		media, err := db.ParseRecordID(mediaID)
		if err != nil {
			return false, fmt.Errorf("invalid media id: %w", err)
		}
		mediaRIDs = append(mediaRIDs, *media)
	}
	err = r.database.Tx(ctx, func(tx *db.DB) error {
		for position, media := range mediaRIDs {
			_, err := tx.Exec(ctx,
				`UPDATE collection_item SET position = $position WHERE collection = $collection AND media = $media`,
				db.Vars{"position": position, "collection": *collection, "media": media})
			if err != nil {
				return fmt.Errorf("reorder collection: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetCollectionTags replaces a collection's tags with exactly the given tag ids.
func (r *Repository) SetCollectionTags(ctx context.Context, collectionID string, tagIDs []string) (*model.Collection, error) {
	if err := r.setTagsField(ctx, collectionID, tagIDs); err != nil {
		return nil, err
	}
	return r.GetCollection(ctx, collectionID)
}

// ListCollections returns every collection, newest first. origin filters to
// "user" (user-created playlists) or "scraped" (scraper-created).
func (r *Repository) ListCollections(ctx context.Context, search, tagID, origin *string, limit, offset *int) ([]*model.Collection, error) {
	vars := db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)}
	var conditions []string
	if search != nil && *search != "" {
		vars["q"] = *search
		conditions = append(conditions, searchCondition("name"))
	}
	if rid := parseFilterID(tagID); rid != nil {
		vars["tag"] = *rid
		conditions = append(conditions, jsonContains("collection.tags", "$tag"))
	}
	if origin != nil && *origin == "user" {
		conditions = append(conditions, "user_created = 1")
	}
	if origin != nil && *origin == "scraped" {
		conditions = append(conditions, "user_created = 0")
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + joinAnd(conditions)
	}

	query := fmt.Sprintf(
		"SELECT * FROM collection %s ORDER BY created_at DESC LIMIT $limit OFFSET $offset", where)
	rows, err := db.QueryAs[collectionRecord](ctx, r.database, query, vars)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	out := make([]*model.Collection, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.collectionToModel(ctx, row))
	}
	return out, nil
}

// GetCollection returns one collection with its item count and cover.
func (r *Repository) GetCollection(ctx context.Context, collectionID string) (*model.Collection, error) {
	id, err := db.ParseRecordID(collectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection id: %w", err)
	}
	row, err := db.QueryOneAs[collectionRecord](ctx, r.database,
		`SELECT * FROM collection WHERE id = $id`, db.Vars{"id": *id})
	if err != nil {
		return nil, fmt.Errorf("get collection: %w", err)
	}
	if row == nil {
		return nil, nil
	}
	return r.collectionToModel(ctx, *row), nil
}

// CollectionMembers returns a collection's members in manual sort order, shaped
// as cards for mixed-type rendering.
func (r *Repository) CollectionMembers(ctx context.Context, collectionID string) ([]*model.CollectionMember, error) {
	items, err := r.collectionItems(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, recordIDString(item.Media))
	}
	cards, err := r.MediaCards(ctx, ids)
	if err != nil {
		return nil, err
	}
	cardsByID := make(map[string]*model.MediaCard, len(cards))
	for _, card := range cards {
		cardsByID[card.MediaID] = card
	}

	out := make([]*model.CollectionMember, 0, len(items))
	for _, item := range items {
		id := recordIDString(item.Media)
		card, ok := cardsByID[id]
		if !ok {
			continue
		}
		out = append(out, &model.CollectionMember{
			MediaID:    id,
			MediaType:  card.MediaType,
			Position:   item.Position,
			Title:      card.Title,
			PosterPath: card.PosterPath,
		})
	}
	return out, nil
}

// CollectionIdsForMedia returns the ids of the user-created collections
// containing the given record.
func (r *Repository) CollectionIdsForMedia(ctx context.Context, mediaID string) ([]string, error) {
	media, err := db.ParseRecordID(mediaID)
	if err != nil {
		return nil, fmt.Errorf("invalid media id: %w", err)
	}
	ids, err := r.database.Strings(ctx,
		`SELECT collection_item.collection FROM collection_item
			JOIN collection ON collection.id = collection_item.collection
			WHERE collection_item.media = $media AND collection.user_created = 1`,
		db.Vars{"media": *media})
	if err != nil {
		return nil, fmt.Errorf("collection ids for media: %w", err)
	}
	return ids, nil
}

// --- helpers ---------------------------------------------------------------

type collectionItemRow struct {
	Media    *db.RecordID `json:"media"`
	Position int          `json:"position"`
}

func (r *Repository) collectionItems(ctx context.Context, collectionID string) ([]collectionItemRow, error) {
	id, err := db.ParseRecordID(collectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection id: %w", err)
	}
	rows, err := db.QueryAs[collectionItemRow](ctx, r.database,
		`SELECT media, position FROM collection_item WHERE collection = $id ORDER BY position ASC`,
		db.Vars{"id": *id})
	if err != nil {
		return nil, fmt.Errorf("collection items: %w", err)
	}
	return rows, nil
}

func (r *Repository) collectionItemCount(ctx context.Context, collection db.RecordID) (int, error) {
	return r.database.Int(ctx,
		`SELECT count(*) FROM collection_item WHERE collection = $id`,
		db.Vars{"id": collection})
}

// collectionCover returns the first member's card poster, or nil.
func (r *Repository) collectionCover(ctx context.Context, collection db.RecordID) *string {
	firstMedia, err := r.database.Strings(ctx,
		`SELECT media FROM collection_item WHERE collection = $id ORDER BY position ASC LIMIT 1`,
		db.Vars{"id": collection})
	if err != nil || len(firstMedia) == 0 {
		return nil
	}
	cards, err := r.MediaCards(ctx, firstMedia)
	if err != nil || len(cards) == 0 {
		return nil
	}
	return cards[0].PosterPath
}

func (r *Repository) collectionContains(ctx context.Context, collection, media db.RecordID) (bool, error) {
	count, err := r.database.Int(ctx,
		`SELECT count(*) FROM collection_item WHERE collection = $collection AND media = $media`,
		db.Vars{"collection": collection, "media": media})
	if err != nil {
		return false, fmt.Errorf("collection contains: %w", err)
	}
	return count > 0, nil
}

// requireContentImage rejects asset images: only content images are collectible.
func (r *Repository) requireContentImage(ctx context.Context, media db.RecordID) error {
	m, err := r.getByID(ctx, media.String())
	if err != nil || m == nil {
		return fmt.Errorf("image %s not found", media.String())
	}
	if !mBool(m, "content") {
		return fmt.Errorf("image %s is an asset, not content", media.String())
	}
	return nil
}

func parseCollectionMedia(collectionID, mediaID string) (*db.RecordID, *db.RecordID, error) {
	collection, err := db.ParseRecordID(collectionID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid collection id: %w", err)
	}
	media, err := db.ParseRecordID(mediaID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid media id: %w", err)
	}
	if !collectionMemberTables[media.Table] {
		return nil, nil, fmt.Errorf("%s records cannot be collection members", media.Table)
	}
	return collection, media, nil
}
