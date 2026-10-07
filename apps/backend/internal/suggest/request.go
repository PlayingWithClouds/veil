package suggest

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/recommend"
)

// request is the state of one Suggest call: one index, profile and blocklist
// for every candidate.
type request struct {
	service *Service
	index   *index
	now     time.Time
	profile *recommend.Profile
	blocked map[string]bool
	limit   int
	// blockedNames caches the matchText names and aliases of blocked entities.
	blockedNames map[string]bool
}

// entityKindLabels names the entity kinds in the detail line.
var entityKindLabels = map[Kind]string{
	KindTag:       "Tag",
	KindPerformer: "Performer",
	KindStudio:    "Studio",
}

// typedQuery ranks past searches, entities and title phrases against the
// typed text (in matchText form).
func (r *request) typedQuery(ctx context.Context, typed string) ([]Suggestion, error) {
	searches, err := r.service.pastSearches(ctx, historyScanLimit)
	if err != nil {
		return nil, err
	}
	merged := newMergedList()
	merged.addAll(r.historyMatches(searches, typed))
	merged.addAll(r.entityMatches(typed))
	merged.addAll(r.phraseMatches(typed))
	return merged.ranked(r.limit), nil
}

// emptyQuery lists the recent searches, newest first, then the entities the
// user likes most (the most popular ones on a cold start).
func (r *request) emptyQuery(ctx context.Context) ([]Suggestion, error) {
	searches, err := r.service.pastSearches(ctx, emptyRecentLimit+len(r.blocked))
	if err != nil {
		return nil, err
	}
	out := []Suggestion{}
	seen := map[string]bool{}
	for _, search := range searches {
		if len(out) >= min(emptyRecentLimit, r.limit) {
			break
		}
		if r.isBlockedName(search.matchQuery) || seen[search.matchQuery] {
			continue
		}
		seen[search.matchQuery] = true
		out = append(out, Suggestion{Kind: KindRecent, Text: search.query, Score: search.popularity(r.now)})
	}
	picks := r.tastePicks()
	if len(picks) == 0 {
		picks = r.popularPicks()
	}
	for _, pick := range picks {
		if len(out) >= r.limit {
			break
		}
		key := matchText(pick.Text)
		if !seen[key] {
			seen[key] = true
			out = append(out, pick)
		}
	}
	return out, nil
}

// historyMatches scores the past searches the typed text prefixes or
// word-prefixes, keeping the best typedRecentLimit.
func (r *request) historyMatches(searches []pastSearch, typed string) []Suggestion {
	out := []Suggestion{}
	for _, search := range searches {
		quality := textMatch(typed, search.matchQuery)
		if quality == matchNone || r.isBlockedName(search.matchQuery) {
			continue
		}
		score := suggestionScore(qualityValue(quality, 0), search.popularity(r.now), 0, KindRecent)
		out = append(out, Suggestion{Kind: KindRecent, Text: search.query, Score: score})
	}
	return bestOf(out, typedRecentLimit)
}

// entityMatches scores the tags, performers and studios whose name or alias
// matches the typed text, adding typo-tolerant matches while the list is short.
func (r *request) entityMatches(typed string) []Suggestion {
	matches := map[int]entityMatch{}
	for _, key := range r.index.keysWithPrefix(typed) {
		matches[key.entity] = betterMatch(matches[key.entity], entityMatch{quality: keyQuality(key, typed)})
	}
	if len(matches) < r.limit {
		r.addFuzzyMatches(typed, matches)
	}
	out := make([]Suggestion, 0, len(matches))
	for position, match := range matches {
		matched := &r.index.entities[position]
		if r.blocked[matched.id] {
			continue
		}
		out = append(out, r.entitySuggestion(matched, qualityValue(match.quality, match.edits)))
	}
	return out
}

// entityMatch is the best way one entity matched.
type entityMatch struct {
	quality matchQuality
	// edits is the typo count of a fuzzy match.
	edits int
}

// betterMatch returns the stronger of two matches.
func betterMatch(current, candidate entityMatch) entityMatch {
	if candidate.quality > current.quality {
		return candidate
	}
	if candidate.quality == current.quality && candidate.edits < current.edits {
		return candidate
	}
	return current
}

// keyQuality rates a key that starts with the typed text.
func keyQuality(key entityKey, typed string) matchQuality {
	if !key.whole {
		return matchWordPrefix
	}
	if key.key == typed {
		return matchExact
	}
	return matchPrefix
}

// addFuzzyMatches scans every key for a typo-tolerant prefix match, skipping
// entities that already matched without typos.
func (r *request) addFuzzyMatches(typed string, matches map[int]entityMatch) {
	typedRunes := []rune(typed)
	maxEdits := allowedEdits(len(typedRunes))
	if maxEdits == 0 {
		return
	}
	scratch := make([]int, 2*(len(typedRunes)+maxEdits+1))
	for _, key := range r.index.keys {
		existing, found := matches[key.entity]
		if (found && existing.quality > matchFuzzy) || len(key.runes) < len(typedRunes)-maxEdits {
			continue
		}
		edits := fuzzyPrefixDistance(typedRunes, key.runes, maxEdits, scratch)
		if edits <= maxEdits {
			matches[key.entity] = betterMatch(existing, entityMatch{quality: matchFuzzy, edits: edits})
		}
	}
}

// phraseMatches scores the title phrases that extend the typed text, keeping
// the best queryCompletionLimit.
func (r *request) phraseMatches(typed string) []Suggestion {
	top := []Suggestion{}
	for _, candidate := range r.index.phrasesWithPrefix(typed) {
		if candidate.text == typed || r.isBlockedName(candidate.text) {
			continue
		}
		score := suggestionScore(qualityPrefix, logPopularity(candidate.count, phraseReferenceCount), 0, KindQuery)
		top = insertTop(top, Suggestion{Kind: KindQuery, Text: candidate.text, Score: score}, queryCompletionLimit)
	}
	return top
}

// tastePicks turns the profile's top tags, performers and studios into
// suggestions, strongest affinity first.
func (r *request) tastePicks() []Suggestion {
	out := []Suggestion{}
	dimensions := []map[string]float64{r.profile.Tags, r.profile.Performers, r.profile.Studios}
	for _, affinities := range dimensions {
		for _, id := range recommend.TopEntities(affinities, tastePicksPerKind) {
			picked := r.index.entityFor(id)
			if picked == nil || r.blocked[id] {
				continue
			}
			out = append(out, r.entitySuggestion(picked, affinities[id]))
		}
	}
	sortSuggestions(out)
	return out
}

// popularPicks suggests the entities with the most scenes, for a cold start.
func (r *request) popularPicks() []Suggestion {
	out := []Suggestion{}
	for position := range r.index.entities {
		candidate := &r.index.entities[position]
		if candidate.sceneCount == 0 || r.blocked[candidate.id] {
			continue
		}
		out = insertTop(out, r.entitySuggestion(candidate, qualityExact), r.limit)
	}
	return out
}

// entitySuggestion scores an entity suggestion given its match quality.
func (r *request) entitySuggestion(matched *entity, quality float64) Suggestion {
	popularity := logPopularity(matched.sceneCount, entityReferenceScenes)
	return Suggestion{
		Kind:     matched.kind,
		Text:     matched.name,
		EntityID: matched.id,
		ImageURL: matched.imageURL,
		Detail:   entityDetail(matched),
		Score:    suggestionScore(quality, popularity, r.affinity(matched), matched.kind),
	}
}

// affinity is the taste profile's affinity for an entity, in [-1, 1].
func (r *request) affinity(matched *entity) float64 {
	switch matched.kind {
	case KindTag:
		return r.profile.Tags[matched.id]
	case KindPerformer:
		return r.profile.Performers[matched.id]
	case KindStudio:
		return r.profile.Studios[matched.id]
	}
	return 0
}

// isBlockedName reports whether text (matchText form) is the name or an
// alias of a blocked entity, so searches for it aren't suggested either.
func (r *request) isBlockedName(text string) bool {
	if r.blockedNames == nil {
		r.blockedNames = r.blockedNameSet()
	}
	return r.blockedNames[text]
}

// blockedNameSet collects the matchText names and aliases of blocked entities.
func (r *request) blockedNameSet() map[string]bool {
	out := map[string]bool{}
	for id := range r.blocked {
		blockedEntity := r.index.entityFor(id)
		if blockedEntity == nil {
			continue
		}
		for _, name := range blockedEntity.matchNames {
			out[name] = true
		}
	}
	return out
}

// entityDetail is the grey line of an entity suggestion, e.g. "Tag · 128 videos".
func entityDetail(described *entity) string {
	label := entityKindLabels[described.kind]
	switch described.sceneCount {
	case 0:
		return label
	case 1:
		return label + " · 1 video"
	}
	return fmt.Sprintf("%s · %d videos", label, described.sceneCount)
}

// qualityValue maps a match quality to its score factor; fuzzy matches lose
// strength with every edit.
func qualityValue(quality matchQuality, edits int) float64 {
	switch quality {
	case matchExact:
		return qualityExact
	case matchPrefix:
		return qualityPrefix
	case matchWordPrefix:
		return qualityWordPrefix
	case matchFuzzy:
		return qualityFuzzy / float64(max(1, edits))
	}
	return 0
}

// suggestionScore combines match quality, popularity in [0, 1] and taste
// affinity in [-1, 1] with the kind's weight.
func suggestionScore(quality, popularity, affinity float64, kind Kind) float64 {
	return quality * (1 + popularityWeight*popularity) * (1 + affinityWeight*affinity) * kindWeights[kind]
}

// bestOf sorts the suggestions best first and keeps at most limit.
func bestOf(suggestions []Suggestion, limit int) []Suggestion {
	sortSuggestions(suggestions)
	if len(suggestions) > limit {
		return suggestions[:limit]
	}
	return suggestions
}

// insertTop inserts a suggestion into a best-first list capped at limit.
func insertTop(top []Suggestion, added Suggestion, limit int) []Suggestion {
	if len(top) >= limit && !ranksBefore(added, top[len(top)-1]) {
		return top
	}
	position := sort.Search(len(top), func(index int) bool {
		return ranksBefore(added, top[index])
	})
	top = append(top, Suggestion{})
	copy(top[position+1:], top[position:])
	top[position] = added
	if len(top) > limit {
		top = top[:limit]
	}
	return top
}

// sortSuggestions orders suggestions best first.
func sortSuggestions(suggestions []Suggestion) {
	sort.SliceStable(suggestions, func(left, right int) bool {
		return ranksBefore(suggestions[left], suggestions[right])
	})
}

// ranksBefore orders by score, then by text so equal scores stay stable.
func ranksBefore(left, right Suggestion) bool {
	if left.Score != right.Score {
		return left.Score > right.Score
	}
	return left.Text < right.Text
}
