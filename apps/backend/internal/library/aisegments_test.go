package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestReplaceAISegmentsKeepsOthers replaces a video's ai segments twice and
// expects the user's segment to stay and tags to be found by name.
func TestReplaceAISegmentsKeepsOthers(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	start := 5.0
	if _, err := library.AddSegment(ctx, schema.RecordKindVideo, ref.Id, SegmentEdit{StartSeconds: &start}); err != nil {
		t.Fatal(err)
	}
	first := []AISegment{{StartSeconds: 10, EndSeconds: 40, Tag: "Blowjob"}, {StartSeconds: 50, EndSeconds: 90, Tag: "Cowgirl"}}
	if _, err := library.ReplaceAISegments(ctx, schema.RecordKindVideo, ref.Id, first); err != nil {
		t.Fatal(err)
	}
	segments, err := library.ReplaceAISegments(ctx, schema.RecordKindVideo, ref.Id, []AISegment{{StartSeconds: 20, EndSeconds: 60, Tag: "blowjob"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 2 || segments[0].Origin != schema.SegmentOriginUser || segments[1].Origin != schema.SegmentOriginAi {
		t.Fatalf("segments = %+v", segments)
	}
	tags, err := library.Tags(ctx, NameFilter{Text: "blowjob"})
	if err != nil || len(tags) != 1 || segments[1].Tag == nil || *segments[1].Tag != tags[0].Id {
		t.Errorf("tags = %+v, segment tag %v, err %v", tags, segments[1].Tag, err)
	}
	if _, err := library.ReplaceAISegments(ctx, schema.RecordKindVideo, ref.Id, []AISegment{{StartSeconds: 20, EndSeconds: 10, Tag: "Anal"}}); err == nil {
		t.Error("a backwards segment was accepted")
	}
}

// TestSceneMatchesRanksByMatchingSeconds finds videos by their ai segment
// tags, the most covered first, overlaps counted once.
func TestSceneMatchesRanksByMatchingSeconds(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	short, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	second := foundScene()
	second[0].Scene.ExternalID = "v2"
	second[0].Scene.SourceURL = "https://site.test/v2"
	second[0].Scene.Title = "Another evening entirely"
	second[0].Scene.Duration = 1900
	second[0].Scene.PosterPath = "https://cdn.test/v2.jpg"
	second[0].Scene.Performers = []plugins.ScenePerformer{{Name: "Cleo"}}
	long, err := library.IngestFound(ctx, "site", second)
	if err != nil {
		t.Fatal(err)
	}
	mustReplace(t, library, short.Id, []AISegment{{StartSeconds: 0, EndSeconds: 30, Tag: "Cowgirl"}, {StartSeconds: 100, EndSeconds: 400, Tag: "Anal"}})
	mustReplace(t, library, long.Id, []AISegment{{StartSeconds: 0, EndSeconds: 60, Tag: "Cowgirl"}, {StartSeconds: 30, EndSeconds: 90, Tag: "Blowjob"}})
	matches, err := library.SceneMatches(ctx, []string{"cowgirl", "Blowjob", "Unknown"}, nil, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 || matches[0].VideoID != long.Id || matches[0].MatchingSeconds != 90 || len(matches[0].Segments) != 2 {
		t.Fatalf("matches = %+v", matches)
	}
	ordered, err := library.SceneMatches(ctx, []string{"Cowgirl"}, []string{short.Id, long.Id}, 10, 0)
	if err != nil || len(ordered) != 2 || ordered[0].VideoID != short.Id {
		t.Errorf("ordered = %+v, err %v", ordered, err)
	}
}

// TestSceneLabelFilterNeedsEveryLabel keeps the videos whose ai segments carry
// every asked label, the most matching seconds first, and counts the labels.
func TestSceneLabelFilterNeedsEveryLabel(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	first, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	second := foundScene()
	second[0].Scene.ExternalID = "v2"
	second[0].Scene.SourceURL = "https://site.test/v2"
	second[0].Scene.Title = "Another evening entirely"
	second[0].Scene.PosterPath = "https://cdn.test/v2.jpg"
	second[0].Scene.Performers = []plugins.ScenePerformer{{Name: "Cleo"}}
	other, err := library.IngestFound(ctx, "site", second)
	if err != nil {
		t.Fatal(err)
	}
	mustReplace(t, library, first.Id, []AISegment{{StartSeconds: 0, EndSeconds: 30, Tag: "Blowjob"}, {StartSeconds: 0, EndSeconds: 30, Tag: "POV"}})
	mustReplace(t, library, other.Id, []AISegment{{StartSeconds: 0, EndSeconds: 300, Tag: "Blowjob"}, {StartSeconds: 0, EndSeconds: 10, Tag: "POV"}})
	counts, err := library.SceneLabelCounts(ctx, 10)
	if err != nil || len(counts) != 2 || counts[0].Count != 2 {
		t.Fatalf("counts = %+v, err %v", counts, err)
	}
	blowjob, pov := tagIDNamed(t, counts, "Blowjob"), tagIDNamed(t, counts, "POV")
	both, err := library.Videos(ctx, MediaQuery{Filter: schema.Filter{SceneTagIds: []string{blowjob, pov}}, Limit: 10})
	if err != nil || len(both) != 2 || both[0].Id != other.Id {
		t.Fatalf("both labels = %+v, err %v", both, err)
	}
	mustReplace(t, library, other.Id, []AISegment{{StartSeconds: 0, EndSeconds: 300, Tag: "Blowjob"}})
	narrowed, err := library.Videos(ctx, MediaQuery{Filter: schema.Filter{SceneTagIds: []string{blowjob, pov}}, Limit: 10})
	if err != nil || len(narrowed) != 1 || narrowed[0].Id != first.Id {
		t.Errorf("after losing POV = %+v, err %v", narrowed, err)
	}
}

// tagIDNamed is the id of the counted tag with name, or fails the test.
func tagIDNamed(t *testing.T, counts []TagCount, name string) string {
	t.Helper()
	for _, count := range counts {
		if count.Name == name {
			return count.Id
		}
	}
	t.Fatalf("no tag %q in %+v", name, counts)
	return ""
}

// mustReplace stores a video's ai segments or fails the test.
func mustReplace(t *testing.T, library *Library, videoID string, segments []AISegment) {
	t.Helper()
	if _, err := library.ReplaceAISegments(context.Background(), schema.RecordKindVideo, videoID, segments); err != nil {
		t.Fatal(err)
	}
}
