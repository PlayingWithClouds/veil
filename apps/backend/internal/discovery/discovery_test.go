package discovery

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// resultListFunc answers a list call with its whole result at once.
type resultListFunc func(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, args plugins.ListArgs) (*plugins.ItemsResult, error)

// newTestService creates a discovery service on a temporary database whose
// list calls are answered by list, each result delivered as one batch.
func newTestService(t *testing.T, list resultListFunc) *Service {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return newStreamingTestService(t, database, func(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, args plugins.ListArgs, onItems func([]plugins.DiscoveredItem)) (*plugins.ItemsResult, error) {
		result, err := list(ctx, plugin, entity, args)
		if err == nil {
			onItems(result.Items)
		}
		return result, err
	})
}

// newStreamingTestService creates a discovery service whose list calls are
// answered by list, batches and all.
func newStreamingTestService(t *testing.T, database *db.DB, list listFunc) *Service {
	t.Helper()
	return &Service{ingestSvc: ingest.New(database), list: list}
}

// newTestPlugin creates a plugin that lists the given entity types.
func newTestPlugin(name string, entities ...plugins.MediaType) *plugins.Plugin {
	var capabilities []plugins.Capability
	for _, entity := range entities {
		capabilities = append(capabilities, plugins.ListCapability(entity))
	}
	return &plugins.Plugin{Meta: plugins.PluginMeta{Name: name, Capabilities: capabilities}}
}

// testItems returns count stubs of entity, unique per plugin and entity.
func testItems(pluginName string, entity plugins.MediaType, count int) []plugins.DiscoveredItem {
	items := make([]plugins.DiscoveredItem, 0, count)
	for index := range count {
		id := pluginName + "-" + string(entity) + "-" + string(rune('a'+index))
		items = append(items, plugins.DiscoveredItem{
			Title:      id,
			MediaType:  entity,
			SourceURL:  "https://" + pluginName + ".test/" + id,
			ExternalID: id,
		})
	}
	return items
}

// TestRunHandsOnFastListWhileSlowListRuns checks that a plugin's scenes are
// delivered while its gallery list is still running.
func TestRunHandsOnFastListWhileSlowListRuns(t *testing.T) {
	sceneDelivered := make(chan struct{})
	service := newTestService(t, func(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, args plugins.ListArgs) (*plugins.ItemsResult, error) {
		if entity == plugins.MediaTypeScene {
			return &plugins.ItemsResult{Items: testItems(plugin.Meta.Name, entity, 2)}, nil
		}
		select {
		case <-sceneDelivered:
		case <-time.After(5 * time.Second):
			return nil, errors.New("scenes were not delivered while the gallery list ran")
		}
		return &plugins.ItemsResult{Items: testItems(plugin.Meta.Name, entity, 1)}, nil
	})

	var delivered []Hit
	err := service.Run(context.Background(), Options{Query: "q", Limit: 5},
		[]*plugins.Plugin{newTestPlugin("site", plugins.MediaTypeScene, plugins.MediaTypeGallery)},
		func(hit Hit) {
			if hit.Item.MediaType == plugins.MediaTypeScene && len(delivered) == 0 {
				close(sceneDelivered)
			}
			delivered = append(delivered, hit)
		})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(delivered) != 3 {
		t.Fatalf("delivered %d hits, want 3", len(delivered))
	}
	if delivered[0].Item.MediaType != plugins.MediaTypeScene {
		t.Fatalf("first hit is a %s, want the scene that finished first", delivered[0].Item.MediaType)
	}
}

// TestRunFailsPluginOnlyWhenNothingCameBack checks that one failing entity type
// doesn't fail a plugin whose other lists returned results.
func TestRunFailsPluginOnlyWhenNothingCameBack(t *testing.T) {
	service := newTestService(t, func(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, args plugins.ListArgs) (*plugins.ItemsResult, error) {
		switch plugin.Meta.Name {
		case "partial":
			if entity == plugins.MediaTypeScene {
				return nil, errors.New("scene list broke")
			}
			return &plugins.ItemsResult{Items: testItems("partial", entity, 1)}, nil
		case "broken":
			return nil, errors.New("site is down")
		default:
			return &plugins.ItemsResult{}, nil
		}
	})

	hitCount := 0
	type report struct {
		count  int
		failed bool
	}
	reports := map[string][]report{}
	options := Options{Query: "q", Limit: 5, OnPluginDone: func(plugin string, itemCount int, failed bool) {
		reports[plugin] = append(reports[plugin], report{itemCount, failed})
	}}
	err := service.Run(context.Background(), options,
		[]*plugins.Plugin{
			newTestPlugin("partial", plugins.MediaTypeScene, plugins.MediaTypeGallery),
			newTestPlugin("broken", plugins.MediaTypeScene, plugins.MediaTypeGallery),
			newTestPlugin("empty", plugins.MediaTypeScene, plugins.MediaTypeGallery),
		},
		func(Hit) { hitCount++ })

	if hitCount != 1 {
		t.Fatalf("got %d hits, want the partial plugin's one gallery", hitCount)
	}
	if err == nil {
		t.Fatal("run returned no error, want the broken plugin's failure")
	}
	if err.Error() != "broken: site is down" {
		t.Fatalf("error = %q, want only the broken plugin's first error", err)
	}
	want := map[string]report{"partial": {1, false}, "broken": {0, true}, "empty": {0, false}}
	for plugin, expected := range want {
		got := reports[plugin]
		if len(got) != 1 || got[0] != expected {
			t.Errorf("%s reported %+v, want one %+v after both its lists", plugin, got, expected)
		}
	}
}

func TestListJobsKeepOnlyTheWantedEntity(t *testing.T) {
	both := &plugins.Plugin{Meta: plugins.PluginMeta{Name: "both", Capabilities: []plugins.Capability{plugins.CapabilitySceneList, plugins.CapabilityGalleryList}}}
	if jobs := listJobs([]*plugins.Plugin{both}, ""); len(jobs) != 2 {
		t.Errorf("unrestricted jobs = %d, want 2", len(jobs))
	}
	jobs := listJobs([]*plugins.Plugin{both}, plugins.MediaTypeScene)
	if len(jobs) != 1 || jobs[0].entity != plugins.MediaTypeScene {
		t.Errorf("scene-only jobs = %+v", jobs)
	}
}

// TestRunHandsOnEachBatchWhileThePluginCrawls checks that a plugin's first
// page reaches onHit before its second page is listed.
func TestRunHandsOnEachBatchWhileThePluginCrawls(t *testing.T) {
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	firstPageShown := make(chan struct{})
	pages := testItems("crawler", plugins.MediaTypeScene, 4)
	service := newStreamingTestService(t, database, func(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, args plugins.ListArgs, onItems func([]plugins.DiscoveredItem)) (*plugins.ItemsResult, error) {
		onItems(pages[:2])
		select {
		case <-firstPageShown:
		case <-time.After(time.Second):
			return nil, errors.New("first page was not handed on while crawling")
		}
		onItems(pages[2:])
		return &plugins.ItemsResult{Items: pages}, nil
	})

	var hits []string
	err = service.Run(context.Background(), Options{Query: "q"}, []*plugins.Plugin{newTestPlugin("crawler", plugins.MediaTypeScene)}, func(hit Hit) {
		hits = append(hits, hit.Item.ExternalID)
		if len(hits) == 2 {
			close(firstPageShown)
		}
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(hits) != 4 {
		t.Errorf("hits = %v, want 4", hits)
	}
}
