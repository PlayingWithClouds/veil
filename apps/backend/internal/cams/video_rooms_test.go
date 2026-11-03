package cams

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/schema"
)

// roomIDs lists the ids of rooms.
func roomIDs(rooms []schema.CamModel) []string {
	ids := []string{}
	for _, camModel := range rooms {
		ids = append(ids, camModel.Id)
	}
	return ids
}

// TestVideoRoomsNameTheRoomWithItsSite checks that a title names a room only
// by its whole username together with its site, that a recording names the
// room it came from, and that the rooms come back with their live state.
func TestVideoRoomsNameTheRoomWithItsSite(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	cases := []struct {
		title string
		named bool
	}{
		{"nova ALPHA webcam rip 2026.11.03", true},
		{"Alpha - Nova January 6 2023", true},
		{"nova solo in the shower", false},
		{"novakane ALPHA webcam rip", false},
	}
	for _, testCase := range cases {
		rooms, err := f.service.VideoRooms(f.ctx, schema.Video{Media: schema.Media{Title: testCase.title, Performers: []string{}}})
		if err != nil {
			t.Fatal(err)
		}
		named := len(rooms) == 1 && rooms[0].Id == f.roomID
		if named != testCase.named {
			t.Errorf("%q named rooms %v, want named %v", testCase.title, roomIDs(rooms), testCase.named)
		}
	}

	recordedFrom := f.roomID
	rooms, _ := f.service.VideoRooms(f.ctx, schema.Video{RecordedFrom: &recordedFrom, Media: schema.Media{Title: "untitled", Performers: []string{}}})
	if len(rooms) != 1 || rooms[0].Live == nil {
		t.Errorf("recording's rooms = %+v, want nova checked live", rooms)
	}
}
