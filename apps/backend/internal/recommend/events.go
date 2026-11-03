package recommend

// The engine learns only from the event log (internal/user, table "Event")
// and UserState. This is the event data clients must send for it, per kind;
// the subject is the record the event is about. Fields not listed are kept
// but ignored. Send a sessionId per playback session where possible: watches
// without one are grouped by day.
//
//   - view: {"surface": "home"} (optional). A video or gallery was opened.
//   - progress: {"position": seconds}. Resume point while a video plays.
//   - finish: {"position": seconds} (optional). The video played to its end.
//   - o: {"position": seconds} (optional). Where in the video it happened;
//     the position marks a best moment.
//   - search: {"text": "…"}. No subject.
//   - impression: {"surface": "home", "position": n}. The item was shown on a
//     surface at 0-based position n. Only surfaces in recommendationSurfaces
//     teach the engine; others are kept for analysis.
//   - choice: {"surface": "home", "position": n}. The item was picked from a
//     surface (a click on its card).
//   - heatSpan: {"start": seconds, "end": seconds}. A stretch of the video
//     that actually played, sent when playback of a contiguous stretch ends
//     (seek, pause, stop). Their sum is how long a watch played; their overlap
//     is replay heat.
//   - skip: {"surface": "shorts", "position": seconds} (position optional).
//     The item (video, clip or segment) was swiped away, after position
//     seconds of playing.
//   - like: {"value": 1} or {"value": -1}. Thumbs up or down on a video,
//     gallery, performer, studio or tag. A thumbs-down item is never
//     recommended.
//
// Cam rooms (subject kind camModel) send their own data:
//
//   - impression and choice: {"surface": "cams", "position": n}, as above.
//     A choice counts as a small like of the room.
//   - view: {"surface": "cams"} (optional), with a sessionId. The room was
//     opened.
//   - progress: {"watched": seconds}, with the view's sessionId. How long the
//     room has played in this session, sent every half minute while it plays
//     and once when it is left. The longest of a session counts.

// Event data fields the engine reads, as json_extract paths.
const (
	pathSurface  = "$.surface"
	pathPosition = "$.position"
	pathStart    = "$.start"
	pathEnd      = "$.end"
	pathValue    = "$.value"
	pathText     = "$.text"
	pathWatched  = "$.watched"
)

// Surfaces.
const (
	// surfaceHome is where Recommendations is shown.
	surfaceHome = "home"
	// surfaceGalleries is the galleries' recommended feed, on the galleries
	// page and under a gallery.
	surfaceGalleries = "galleries"
)

// recommendationSurfaces are the surfaces whose showings were the engine's
// choice: only they teach the source priors and count as "shown but not
// opened". A card passed by on a performer page or in search results says
// nothing against what the engine would have picked.
var recommendationSurfaces = []string{surfaceHome, surfaceGalleries}

// mediaKinds are the record kinds the feed recommends and learns from.
var mediaKinds = []string{"video", "gallery"}
