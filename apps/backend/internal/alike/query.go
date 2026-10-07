package alike

import "strings"

// maxQueries caps how many searches one scene fans out, so a scene with many
// performers does not explode the plugin fan-out.
const maxQueries = 4

// BuildQueries derives the search strings used to retrieve alternate copies of a
// scene. Titles on adult sites are SEO spam and rarely match across sites, so
// queries are driven by the stable identifying signals — performers and studio —
// and fall back to title tokens only when neither is known.
func BuildQueries(target Target) []string {
	queries := make([]string, 0, maxQueries)
	seen := make(map[string]struct{})

	add := func(query string) {
		trimmed := strings.TrimSpace(query)
		if trimmed == "" {
			return
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		queries = append(queries, trimmed)
	}

	performers := cleanNames(target.Performers)

	// All performers together is the most specific query.
	if len(performers) >= 2 {
		add(strings.Join(performers, " "))
	}
	// Studio paired with the lead performer narrows a common performer down.
	if target.Studio != "" && len(performers) >= 1 {
		add(target.Studio + " " + performers[0])
	}
	// The title's meaningful words often carry identifying info (specific acts,
	// setting, series name) even when the full title differs across sites, so it
	// is searched alongside the performer queries, not only as a fallback.
	add(titleQuery(target.Title))
	// Each performer alone catches sites that index only one credited name.
	for _, performer := range performers {
		if len(queries) >= maxQueries {
			break
		}
		add(performer)
	}

	if len(queries) > maxQueries {
		return queries[:maxQueries]
	}
	return queries
}

// cleanNames trims and drops empty performer names while preserving order.
func cleanNames(names []string) []string {
	cleaned := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

// titleQuery is the fallback query for scenes with no performers or studio: the
// first few meaningful title tokens.
func titleQuery(title string) string {
	tokens := titleTokens(title)
	if len(tokens) > 6 {
		tokens = tokens[:6]
	}
	return strings.Join(tokens, " ")
}
