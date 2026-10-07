package ingest

import (
	"context"
	"log"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// ingestCollection upserts a scraper-created collection by source_url, links
// its tags, and syncs ordered members. Member URLs without a matching record
// are returned on the result so the orchestrator can queue them for scraping.
func (s *Service) ingestCollection(ctx context.Context, pluginName string, collection *plugins.Collection) (*IngestResult, error) {
	obs := s.baseObs(collection, pluginName, collection.SourceURL)

	result, err := s.upsert(ctx, "collection", "source_url", collection.SourceURL, obs, pluginName)
	if err != nil {
		return nil, err
	}

	s.linkTags(ctx, "collection", result.CanonicalID, collection.Tags, pluginName)
	s.setCreatedByPlugin(ctx, result.CanonicalID, pluginName)
	result.UnresolvedMemberURLs = s.syncCollectionMembers(ctx, result.CanonicalID, collection.MemberURLs)
	return result, nil
}

// syncCollectionMembers resolves ordered member source URLs to records and
// upserts collection_item rows. Returns the URLs that matched no record.
func (s *Service) syncCollectionMembers(ctx context.Context, collectionID string, memberURLs []string) []string {
	collectionRID, err := db.ParseRecordID(collectionID)
	if err != nil {
		log.Printf("ingest: invalid collection id %q: %v", collectionID, err)
		return nil
	}

	var unresolved []string
	for position, url := range memberURLs {
		if url == "" {
			continue
		}
		memberRID, found := s.contentRecordBySourceURL(ctx, url)
		if !found {
			unresolved = append(unresolved, url)
			continue
		}
		s.upsertCollectionItem(ctx, *collectionRID, memberRID, position)
	}
	return unresolved
}

// contentRecordBySourceURL finds the scene/gallery/image record with this
// source URL.
func (s *Service) contentRecordBySourceURL(ctx context.Context, url string) (db.RecordID, bool) {
	for _, table := range []string{"scene", "gallery", "image"} {
		if recordID, found := s.findIDBy(ctx, table, "source_url", url); found {
			return recordID, true
		}
	}
	return db.RecordID{}, false
}

// upsertCollectionItem creates the membership row or moves it to a new position.
func (s *Service) upsertCollectionItem(ctx context.Context, collectionRID, memberRID db.RecordID, position int) {
	_, err := s.database.Exec(ctx,
		`INSERT INTO collection_item (id, collection, media, position) VALUES ($id, $collection, $media, $position)
		ON CONFLICT (collection, media) DO UPDATE SET position = excluded.position`,
		db.Vars{
			"id":         db.NewRecordID("collection_item"),
			"collection": collectionRID,
			"media":      memberRID,
			"position":   position,
		})
	if err != nil {
		log.Printf("ingest: upsert collection_item: %v", err)
	}
}

// setCreatedByPlugin stamps the creating plugin on a scraper-created record.
func (s *Service) setCreatedByPlugin(ctx context.Context, recordID, pluginName string) {
	if err := s.setField(ctx, recordID, "created_by", s.pluginRecordID(ctx, pluginName)); err != nil {
		log.Printf("ingest: set created_by on %s: %v", recordID, err)
	}
}
