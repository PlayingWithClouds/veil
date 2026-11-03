package recommend

import (
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// camRoomRow inserts a cam room with the tags and gender.
func (f *fixture) camRoomRow(username, gender string, tags ...string) string {
	return f.insert("CamModel", db.Vars{
		"plugin": "alpha", "username": username, "displayName": username, "tags": tags,
		"gender": gender, "autoRecord": false,
	})
}

// TestCamWatchWeight checks that a room left at once counts against it and a
// longer watch for it, growing slower than its length up to the cap.
func TestCamWatchWeight(t *testing.T) {
	if camWatchWeight(4) != weightCamQuickLeave {
		t.Errorf("4 s = %v, want the quick leave weight", camWatchWeight(4))
	}
	minute, hour := camWatchWeight(60), camWatchWeight(3600)
	if minute <= 0 || hour <= minute || hour > weightCamWatchCap {
		t.Errorf("a minute = %v, an hour = %v", minute, hour)
	}
	if camWatchWeight(100*3600) != weightCamWatchCap {
		t.Errorf("100 h = %v, want the cap", camWatchWeight(100*3600))
	}
}

// TestCamWatchesFeedEveryTaste checks that watching a room makes the room,
// its gender, its site and its tags liked: a tag the library has lands on
// the library tag (so videos tagged with it rank up), one it hasn't stays a
// cam tag. A room left at once is disliked, and a blocked room says nothing.
func TestCamWatchesFeedEveryTaste(t *testing.T) {
	f := newFixture(t)
	libraryTag := f.named("Tag", "Big Boobs")
	watched := f.camRoomRow("regular", "female", "bigboobs", "lovense")
	left := f.camRoomRow("left", "male", "chat")
	blocked := f.camRoomRow("blocked", "trans", "feet")
	f.state(schema.RecordKindCamModel, blocked, db.Vars{"blocked": true})
	f.event(schema.EventKindProgress, schema.RecordKindCamModel, watched, time.Hour, map[string]any{"watched": 1200})
	f.event(schema.EventKindProgress, schema.RecordKindCamModel, watched, 50*time.Minute, map[string]any{"watched": 30})
	f.event(schema.EventKindProgress, schema.RecordKindCamModel, left, time.Hour, map[string]any{"watched": 4})
	f.event(schema.EventKindProgress, schema.RecordKindCamModel, blocked, time.Hour, map[string]any{"watched": 3000})

	taste, err := f.engine.CamTaste(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if taste.Rooms[watched] != 1 || taste.Rooms[left] >= 0 || taste.Rooms[blocked] != 0 {
		t.Errorf("rooms = %v", taste.Rooms)
	}
	if taste.Genders["female"] != 1 || taste.Genders["trans"] != 0 {
		t.Errorf("genders = %v", taste.Genders)
	}
	if taste.Tags["bigboobs"] <= 0 || taste.Tags["lovense"] <= 0 || taste.Tags["feet"] != 0 || taste.Tags["chat"] >= 0 {
		t.Errorf("tags = %v", taste.Tags)
	}
	if taste.Sites["alpha"] <= 0 {
		t.Errorf("sites = %v", taste.Sites)
	}
	profile, _, _, err := f.engine.profileAt(f.ctx, f.now)
	if err != nil {
		t.Fatal(err)
	}
	if profile.tags[libraryTag] <= 0 {
		t.Errorf("the video taste of %s = %v, want liked", libraryTag, profile.tags[libraryTag])
	}
	if len(profile.seeds) != 0 {
		t.Errorf("seeds = %v, want no room among them", profile.seeds)
	}
}

// TestFollowedCamRoomIsLiked checks that following a room counts for it
// without any watch.
func TestFollowedCamRoomIsLiked(t *testing.T) {
	f := newFixture(t)
	followed := f.camRoomRow("followed", "couple", "outdoor")
	f.insert("Follow", db.Vars{
		"name": "followed", "filter": map[string]any{}, "targetKind": schema.RecordKindCamModel, "targetId": followed,
		"intervalHours": 1, "enabled": true, "notify": true,
	})

	taste, err := f.engine.CamTaste(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if taste.Rooms[followed] != 1 || taste.Tags["outdoor"] != 1 {
		t.Errorf("rooms = %v, tags = %v", taste.Rooms, taste.Tags)
	}
}
