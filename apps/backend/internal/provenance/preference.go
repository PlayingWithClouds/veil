package provenance

import (
	"slices"

	"github.com/playingwithclouds/veil/internal/schema"
)

// fieldPreference ranks the sites per field of a kind, best first: a site
// higher in the list takes a field over from one lower down or unlisted,
// while among unlisted sites the first to fill a field keeps it. Profile
// databases know a performer's facts, bio and portrait better than the tube
// sites that also describe them. The user's edits stay locked regardless.
var fieldPreference = map[schema.RecordKind]map[string][]string{
	schema.RecordKindPerformer: {
		"attributes": {"freeones", "aylo"},
		"details":    {"freeones", "aylo"},
		"image":      {"freeones", "aylo"},
	},
}

// outranks reports whether the plugin author ranks above owner for one
// field of a kind.
func outranks(kind schema.RecordKind, column, author, owner string) bool {
	ranking := fieldPreference[kind][column]
	return rankIn(ranking, author) < rankIn(ranking, owner)
}

// rankIn is a plugin's place in a ranking; unlisted plugins share the place
// after the last.
func rankIn(ranking []string, plugin string) int {
	index := slices.Index(ranking, plugin)
	if index < 0 {
		return len(ranking)
	}
	return index
}
