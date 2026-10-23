package recommend

import (
	"sort"
	"time"
)

// Source and reason of the "Watch again" shelf. It is a row only: its items
// never enter the ranked feed as candidates.
const (
	SourceRewatch = "rewatch"
	ReasonRewatch = "rewatch"
)

// lastSessionEnds is, per scene with a session-ending watch signal, when it
// was last watched.
func lastSessionEnds(signals []signal) map[string]time.Time {
	out := map[string]time.Time{}
	for _, entry := range signals {
		if entry.sessionEnd && entry.at.After(out[entry.sceneID]) {
			out[entry.sceneID] = entry.at
		}
	}
	return out
}

// rewatchLift is how far a session ender's watched penalty has given way, in
// [0, 1]: 0 until rewatchFreeAfter, then linear over rewatchRampDuration.
func rewatchLift(sinceWatched time.Duration) float64 {
	lift := float64(sinceWatched-rewatchFreeAfter) / float64(rewatchRampDuration)
	return max(0, min(1, lift))
}

// rewatchableScenes lists the session enders last watched at least
// rewatchRowMinAge ago and still wanted (positive decayed weight), strongest
// first, at most rewatchRowLimit.
func rewatchableScenes(sessionEnds map[string]time.Time, sceneWeights map[string]float64, features map[string]*sceneFeatures, now time.Time) []Seed {
	out := []Seed{}
	for sceneID, endedAt := range sessionEnds {
		scene, found := features[sceneID]
		if !found || sceneWeights[sceneID] <= 0 || now.Sub(endedAt) < rewatchRowMinAge {
			continue
		}
		out = append(out, Seed{SceneID: sceneID, Credit: scene.credit, Weight: sceneWeights[sceneID]})
	}
	sort.Slice(out, func(left, right int) bool {
		if out[left].Weight != out[right].Weight {
			return out[left].Weight > out[right].Weight
		}
		return out[left].SceneID < out[right].SceneID
	})
	if len(out) > rewatchRowLimit {
		out = out[:rewatchRowLimit]
	}
	return out
}

// rewatchItems turns the profile's rewatchable scenes into shelf items,
// leaving out scenes that must never be recommended (blocked, disliked,
// inactive plugin).
func rewatchItems(state *snapshot) []Item {
	items := []Item{}
	for _, seed := range state.profile.Rewatchable {
		if state.excluded[seed.SceneID] || state.profile.Disliked[seed.SceneID] {
			continue
		}
		items = append(items, Item{
			SceneID: seed.SceneID,
			Source:  SourceRewatch,
			Reason:  Reason{Kind: ReasonRewatch, Text: "Watch again"},
			Score:   seed.Weight,
		})
	}
	return items
}

// withRewatchRow inserts the "Watch again" shelf at rewatchRowPosition when it
// has enough scenes, then applies the row and item caps.
func withRewatchRow(rows []Row, items []Item, rowLimit, perRow int) []Row {
	if len(items) < minRowItems {
		return rows
	}
	shelf := Row{Key: ReasonRewatch, Title: "Watch again", Reason: Reason{Kind: ReasonRewatch, Text: "Watch again"}, Items: items}
	position := min(rewatchRowPosition, len(rows))
	withShelf := make([]Row, 0, len(rows)+1)
	withShelf = append(withShelf, rows[:position]...)
	withShelf = append(withShelf, shelf)
	withShelf = append(withShelf, rows[position:]...)
	return trimRows(withShelf, rowLimit, perRow)
}
