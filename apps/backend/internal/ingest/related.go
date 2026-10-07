package ingest

import (
	"context"
	"fmt"
	"log"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// Where a scene_related edge came from, in display priority order.
const (
	RelatedSourceSite      = "site"
	RelatedSourcePerformer = "performer"
	RelatedSourceTag       = "tag"
	RelatedSourceTitle     = "title"
)

// LinkRelated ingests items as stubs and links each to sceneID as related,
// ranked after any edges the scene already has from the same source. Items
// that resolve to the scene itself or are already linked are skipped. Returns
// the number of new edges.
func (s *Service) LinkRelated(ctx context.Context, sceneID, pluginName, source string, items []plugins.DiscoveredItem) (int, error) {
	nextRank, err := s.database.Int(ctx,
		`SELECT count(*) FROM scene_related WHERE scene = $scene AND source = $source`,
		db.Vars{"scene": sceneID, "source": source})
	if err != nil {
		return 0, fmt.Errorf("count related: %w", err)
	}
	linked := 0
	for _, item := range items {
		if item.MediaType != plugins.MediaTypeScene || item.SourceURL == "" {
			continue
		}
		relatedID, err := s.IngestDiscoveredItem(ctx, pluginName, item)
		if err != nil {
			log.Printf("ingest: related stub %q: %v", item.SourceURL, err)
			continue
		}
		if relatedID == "" || relatedID == sceneID {
			continue
		}
		inserted, err := s.insertRelatedEdge(ctx, sceneID, relatedID, source, nextRank)
		if err != nil {
			return linked, err
		}
		if inserted {
			linked++
			nextRank++
		}
	}
	return linked, nil
}

// insertRelatedEdge adds one edge; false when the pair was already linked.
func (s *Service) insertRelatedEdge(ctx context.Context, sceneID, relatedID, source string, rank int) (bool, error) {
	affected, err := s.database.Exec(ctx,
		`INSERT INTO scene_related (id, scene, related, source, rank)
		 VALUES ($id, $scene, $related, $source, $rank)
		 ON CONFLICT (scene, related) DO NOTHING`,
		db.Vars{
			"id":      db.NewRecordID("scene_related"),
			"scene":   sceneID,
			"related": relatedID,
			"source":  source,
			"rank":    rank,
		})
	if err != nil {
		return false, fmt.Errorf("insert related edge: %w", err)
	}
	return affected > 0, nil
}

// RelatedCount returns how many related edges sceneID has.
func (s *Service) RelatedCount(ctx context.Context, sceneID string) (int, error) {
	return s.database.Int(ctx, `SELECT count(*) FROM scene_related WHERE scene = $scene`, db.Vars{"scene": sceneID})
}

// MarkDetailFetched records that sceneID's detail page has been fetched.
func (s *Service) MarkDetailFetched(ctx context.Context, sceneID string) error {
	_, err := s.database.Exec(ctx, `UPDATE scene SET detail_fetched_at = $now WHERE id = $id`,
		db.Vars{"now": db.Now(), "id": sceneID})
	return err
}

// DetailFetched reports whether sceneID's detail page has been fetched.
func (s *Service) DetailFetched(ctx context.Context, sceneID string) (bool, error) {
	count, err := s.database.Int(ctx,
		`SELECT count(*) FROM scene WHERE id = $id AND detail_fetched_at IS NOT NULL`, db.Vars{"id": sceneID})
	return count > 0, err
}

// GalleryHasImages reports whether galleryID has any content images yet.
func (s *Service) GalleryHasImages(ctx context.Context, galleryID string) (bool, error) {
	count, err := s.database.Int(ctx,
		`SELECT count(*) FROM image WHERE media = $id AND content = 1`, db.Vars{"id": galleryID})
	return count > 0, err
}

// SceneSearchTerms returns what a related search can query for sceneID: its
// credited performer names and tag names (in stored order) and its title.
func (s *Service) SceneSearchTerms(ctx context.Context, sceneID string) (performers, tags []string, title string, err error) {
	row, err := s.GetByID(ctx, sceneID)
	if err != nil || row == nil {
		return nil, nil, "", err
	}
	title, _ = row["title"].(string)
	performers, err = s.namesInOrder(ctx, "performer", row["performers"])
	if err != nil {
		return nil, nil, "", err
	}
	tags, err = s.namesInOrder(ctx, "tag", row["tags"])
	return performers, tags, title, err
}

// namesInOrder resolves a JSON list of ids from table to their names, keeping
// the list's order.
func (s *Service) namesInOrder(ctx context.Context, table string, ids any) ([]string, error) {
	if err := validNameTable(table); err != nil {
		return nil, err
	}
	return s.database.Strings(ctx,
		`SELECT `+table+`.name FROM json_each($ids) AS member
		 JOIN `+table+` ON `+table+`.id = member.value
		 ORDER BY member.key`,
		db.Vars{"ids": ids})
}

// validNameTable limits namesInOrder to the tables it is written for.
func validNameTable(table string) error {
	if table == "performer" || table == "tag" {
		return nil
	}
	return fmt.Errorf("names: unsupported table %q", table)
}
