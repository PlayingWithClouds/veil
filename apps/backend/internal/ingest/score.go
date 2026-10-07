package ingest

import (
	"math"
	"strings"
	"unicode"
)

// MatchScore represents how similar an observation is to a canonical record.
type MatchScore struct {
	Total      float64 // 0.0–1.0 overall similarity
	TitleScore float64
	YearScore  float64
	ExternalID bool // true = definitive match (same plugin + external_id)
}

// ScoreThresholdAuto is the minimum score for an automatic merge.
const ScoreThresholdAuto = 0.85

// ScoreThresholdManual is the minimum score to surface for manual review.
const ScoreThresholdManual = 0.60

// ScoreMatch computes a similarity score between two media representations.
// obs and canonical are plain maps with the same snake_case field keys.
func ScoreMatch(obs, canonical map[string]any) MatchScore {
	var s MatchScore

	// Definitive match: same plugin produced the same external_id.
	if obsPlugin, _ := obs["plugin"].(string); obsPlugin != "" {
		if canPlugin, _ := canonical["plugin_name"].(string); obsPlugin == canPlugin {
			if obsExt, _ := obs["external_id"].(string); obsExt != "" {
				if canExt, _ := canonical["external_id"].(string); obsExt == canExt {
					s.ExternalID = true
					s.Total = 1.0
					s.TitleScore = 1.0
					s.YearScore = 1.0
					return s
				}
			}
		}
	}

	// Title similarity. Movies use `title`, series use `name`.
	obsTitle := titleOf(obs)
	canTitle := titleOf(canonical)
	if obsTitle != "" && canTitle != "" {
		s.TitleScore = jaroWinkler(normalize(obsTitle), normalize(canTitle))
	}

	// Year proximity.
	obsYear := intField(obs, "year")
	canYear := intField(canonical, "year")
	if obsYear > 0 && canYear > 0 {
		diff := math.Abs(float64(obsYear - canYear))
		switch {
		case diff == 0:
			s.YearScore = 1.0
		case diff == 1:
			s.YearScore = 0.7
		case diff == 2:
			s.YearScore = 0.3
		default:
			s.YearScore = 0.0
		}
	}

	// Weighted total: title is most important (0.7), year secondary (0.3).
	// If year is unknown on either side, weight falls entirely on title.
	if obsYear == 0 || canYear == 0 {
		s.Total = s.TitleScore
	} else {
		s.Total = s.TitleScore*0.7 + s.YearScore*0.3
	}

	return s
}

// titleOf returns the display title for matching: movies use `title`, series `name`.
func titleOf(m map[string]any) string {
	if t, _ := m["title"].(string); t != "" {
		return t
	}
	n, _ := m["name"].(string)
	return n
}

func intField(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// normalize strips punctuation/case for comparison.
func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// jaroWinkler returns a similarity score in [0,1] between two strings.
func jaroWinkler(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	jaro := jaroSimilarity(a, b)
	if jaro < 0.7 {
		return jaro
	}

	// Winkler prefix bonus (up to 4 chars).
	prefix := 0
	for i := 0; i < min(4, min(len(a), len(b))); i++ {
		if a[i] == b[i] {
			prefix++
		} else {
			break
		}
	}
	return jaro + float64(prefix)*0.1*(1-jaro)
}

func jaroSimilarity(a, b string) float64 {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	matchDist := max(la, lb)/2 - 1
	if matchDist < 0 {
		matchDist = 0
	}

	aMatched := make([]bool, la)
	bMatched := make([]bool, lb)
	matches := 0
	for i := range ra {
		lo := max(0, i-matchDist)
		hi := min(lb-1, i+matchDist)
		for j := lo; j <= hi; j++ {
			if bMatched[j] || ra[i] != rb[j] {
				continue
			}
			aMatched[i] = true
			bMatched[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}

	transpositions := 0
	k := 0
	for i := range ra {
		if !aMatched[i] {
			continue
		}
		for !bMatched[k] {
			k++
		}
		if ra[i] != rb[k] {
			transpositions++
		}
		k++
	}

	m := float64(matches)
	return (m/float64(la) + m/float64(lb) + (m-float64(transpositions)/2)/m) / 3
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
