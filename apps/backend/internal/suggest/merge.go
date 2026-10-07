package suggest

// mergedList dedupes suggestions by their text in matchText form: an entity
// replaces a past search or phrase of the same text, and within the same
// class the higher score wins. The survivor keeps the best score of the group.
type mergedList struct {
	byText map[string]*Suggestion
	order  []string
}

// newMergedList returns an empty list.
func newMergedList() *mergedList {
	return &mergedList{byText: map[string]*Suggestion{}}
}

// addAll merges every suggestion in.
func (list *mergedList) addAll(suggestions []Suggestion) {
	for _, suggestion := range suggestions {
		list.add(suggestion)
	}
}

// add merges one suggestion in.
func (list *mergedList) add(suggestion Suggestion) {
	key := matchText(suggestion.Text)
	existing, found := list.byText[key]
	if !found {
		added := suggestion
		list.byText[key] = &added
		list.order = append(list.order, key)
		return
	}
	bestScore := max(existing.Score, suggestion.Score)
	if replaces(suggestion, *existing) {
		*existing = suggestion
	}
	existing.Score = bestScore
}

// replaces reports whether candidate should stand in for current.
func replaces(candidate, current Suggestion) bool {
	candidateIsEntity := candidate.EntityID != ""
	currentIsEntity := current.EntityID != ""
	if candidateIsEntity != currentIsEntity {
		return candidateIsEntity
	}
	return candidate.Score > current.Score
}

// ranked returns the merged suggestions best first, at most limit.
func (list *mergedList) ranked(limit int) []Suggestion {
	out := make([]Suggestion, 0, len(list.order))
	for _, key := range list.order {
		out = append(out, *list.byText[key])
	}
	return bestOf(out, limit)
}
