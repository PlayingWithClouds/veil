package suggest

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// entity is a tag, performer or studio a suggestion can point at.
type entity struct {
	kind       Kind
	id         string
	name       string
	imageURL   string
	sceneCount int
	// matchNames are the name and aliases in matchText form.
	matchNames []string
}

// entityKey is one lookup string of an entity: its name, an alias, or the
// tail of either from a later word on.
type entityKey struct {
	key   string
	runes []rune
	// entity indexes index.entities.
	entity int
	// whole is set for a full name or alias, unset for a word tail.
	whole bool
}

// phrase is a word n-gram mined from scene titles.
type phrase struct {
	text  string
	count int
}

// index is the in-memory snapshot suggestions are matched against.
type index struct {
	builtAt  time.Time
	entities []entity
	// keys is sorted by key, so a prefix is one binary-searched range.
	keys []entityKey
	// entityByID maps an entity id to its position in entities.
	entityByID map[string]int
	// phrases is sorted by text.
	phrases []phrase
}

// entitySources reads every tag, performer and studio with its scene count.
var entitySources = []struct {
	kind  Kind
	query string
}{
	{KindTag, `SELECT tag.id, tag.name, tag.aliases, NULL AS image_path, coalesce(counts.scenes, 0) AS scenes
		FROM tag LEFT JOIN (
			SELECT entry.value AS target, count(*) AS scenes FROM scene, json_each(scene.tags) AS entry GROUP BY entry.value
		) AS counts ON counts.target = tag.id`},
	{KindPerformer, `SELECT performer.id, performer.name, performer.aliases, performer.image_path, coalesce(counts.scenes, 0) AS scenes
		FROM performer LEFT JOIN (
			SELECT entry.value AS target, count(*) AS scenes FROM scene, json_each(scene.performers) AS entry GROUP BY entry.value
		) AS counts ON counts.target = performer.id`},
	{KindStudio, `SELECT studio.id, studio.name, studio.aliases, studio.image_path, coalesce(counts.scenes, 0) AS scenes
		FROM studio LEFT JOIN (
			SELECT scene.studio AS target, count(*) AS scenes FROM scene WHERE scene.studio IS NOT NULL GROUP BY scene.studio
		) AS counts ON counts.target = studio.id`},
}

// buildIndex reads entities and scene titles and indexes them.
func buildIndex(ctx context.Context, database *db.DB, now time.Time) (*index, error) {
	built := &index{builtAt: now, entityByID: map[string]int{}}
	for _, source := range entitySources {
		rows, err := database.Query(ctx, source.query, nil)
		if err != nil {
			return nil, fmt.Errorf("suggestion index %s: %w", strings.ToLower(string(source.kind)), err)
		}
		for _, row := range rows {
			built.addEntity(entityFromRow(source.kind, row))
		}
	}
	sort.Slice(built.keys, func(left, right int) bool {
		return built.keys[left].key < built.keys[right].key
	})
	titles, err := database.Strings(ctx,
		`SELECT title FROM scene WHERE title <> '' ORDER BY created_at DESC LIMIT $limit`,
		db.Vars{"limit": phraseSceneLimit})
	if err != nil {
		return nil, fmt.Errorf("suggestion index titles: %w", err)
	}
	built.phrases = minePhrases(titles)
	return built, nil
}

// entityFromRow converts one entity row.
func entityFromRow(kind Kind, row db.Row) entity {
	name := rowString(row, "name")
	matchNames := []string{matchText(name)}
	for _, alias := range rowStringList(row, "aliases") {
		matchNames = append(matchNames, matchText(alias))
	}
	return entity{
		kind:       kind,
		id:         rowString(row, "id"),
		name:       name,
		imageURL:   rowString(row, "image_path"),
		sceneCount: db.AsInt(row["scenes"]),
		matchNames: matchNames,
	}
}

// addEntity stores an entity and its lookup keys; nameless entities are skipped.
func (built *index) addEntity(added entity) {
	if added.id == "" || added.matchNames[0] == "" {
		return
	}
	position := len(built.entities)
	built.entities = append(built.entities, added)
	built.entityByID[added.id] = position
	seen := map[string]bool{}
	for _, name := range added.matchNames {
		built.addKey(position, name, true, seen)
		for _, tail := range wordSuffixes(name) {
			built.addKey(position, tail, false, seen)
		}
	}
}

// addKey appends one lookup key unless the entity already has it.
func (built *index) addKey(position int, key string, whole bool, seen map[string]bool) {
	if key == "" || seen[key] {
		return
	}
	seen[key] = true
	built.keys = append(built.keys, entityKey{key: key, runes: []rune(key), entity: position, whole: whole})
}

// keysWithPrefix returns the keys starting with prefix.
func (built *index) keysWithPrefix(prefix string) []entityKey {
	start := sort.Search(len(built.keys), func(position int) bool {
		return built.keys[position].key >= prefix
	})
	end := start
	for end < len(built.keys) && strings.HasPrefix(built.keys[end].key, prefix) {
		end++
	}
	return built.keys[start:end]
}

// phrasesWithPrefix returns the phrases starting with prefix.
func (built *index) phrasesWithPrefix(prefix string) []phrase {
	start := sort.Search(len(built.phrases), func(position int) bool {
		return built.phrases[position].text >= prefix
	})
	end := start
	for end < len(built.phrases) && strings.HasPrefix(built.phrases[end].text, prefix) {
		end++
	}
	return built.phrases[start:end]
}

// entityFor looks an entity up by id; nil when it isn't indexed.
func (built *index) entityFor(id string) *entity {
	position, found := built.entityByID[id]
	if !found {
		return nil
	}
	return &built.entities[position]
}

// minePhrases counts the word n-grams of the titles (each once per title) and
// keeps the recurring ones, sorted by text.
func minePhrases(titles []string) []phrase {
	counts := map[string]int{}
	titlePhrases := []string{}
	for _, title := range titles {
		titlePhrases = titleNGrams(words(title), titlePhrases[:0])
		for _, text := range uniqueStrings(titlePhrases) {
			counts[text]++
		}
	}
	out := make([]phrase, 0, len(counts)/4)
	for text, count := range counts {
		if count >= minimumPhraseCount {
			out = append(out, phrase{text: text, count: count})
		}
	}
	sort.Slice(out, func(left, right int) bool {
		return out[left].text < out[right].text
	})
	return out
}

// titleNGrams appends the title's n-grams of 1..maxPhraseWords words that
// neither start nor end with a stop word or a number.
func titleNGrams(titleWords []string, out []string) []string {
	for start := range titleWords {
		if !isPhraseEdge(titleWords[start]) {
			continue
		}
		for end := start + 1; end <= min(start+maxPhraseWords, len(titleWords)); end++ {
			text := strings.Join(titleWords[start:end], " ")
			if isPhraseEdge(titleWords[end-1]) && len(text) >= minimumPhraseLength {
				out = append(out, text)
			}
		}
	}
	return out
}

// isPhraseEdge reports whether a word may start or end a phrase.
func isPhraseEdge(word string) bool {
	return !stopWords[word] && strings.TrimLeft(word, "0123456789") != ""
}

// uniqueStrings removes repeats in place, keeping first occurrences. Titles
// yield a few dozen n-grams, so the quadratic scan beats a map.
func uniqueStrings(values []string) []string {
	out := values[:0]
	for _, value := range values {
		if !containsString(out, value) {
			out = append(out, value)
		}
	}
	return out
}

// containsString reports whether values holds value.
func containsString(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

// rowString reads a text column, "" when absent.
func rowString(row db.Row, key string) string {
	switch value := row[key].(type) {
	case string:
		return value
	case []byte:
		return string(value)
	}
	return ""
}

// rowStringList reads a JSON array column of strings, decoded or raw.
func rowStringList(row db.Row, key string) []string {
	values, isList := row[key].([]any)
	if !isList {
		return decodeStringList(rowString(row, key))
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if text, isText := value.(string); isText {
			out = append(out, text)
		}
	}
	return out
}

// decodeStringList parses a JSON string array, nil when it isn't one.
func decodeStringList(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}
