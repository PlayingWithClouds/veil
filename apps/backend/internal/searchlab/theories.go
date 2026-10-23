package searchlab

import (
	"sort"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// Theory is one way of turning the plugins' lists for a query into a single
// ranking. Add new ideas here and they are scored next to the others.
type Theory struct {
	Name        string
	Description string
	Rank        func(query Query, lists []List) []plugins.DiscoveredItem
}

// Theories lists every theory the eval scores, baselines first.
var Theories = []Theory{
	{"plugin-order", "each plugin's list after the other, in config order (roughly arrival order)", rankPluginOrder},
	{"interleave", "round robin over the plugins, each in its own order", rankInterleave},
	{"rrf", "reciprocal rank fusion of the site orders, nothing else", rankFusion},
	{"text", "text match of the query against the title, fusion breaks ties", rankText},
	{"searchrank", "the production live-search scorer (no library, taste or history)", rankSearchrank},
	{"searchrank+listfit", "production scorer, each site's rank share scaled by how well its list fits the query", rankSearchrankListFit},
	{"formula", "the production formula re-built here: text + site rank + quality (no cross-site folding)", formulaTheory(true, true)},
	{"formula-noquality", "formula without the view count / rating term", formulaTheory(true, false)},
	{"formula-nositerank", "formula without the site rank term", formulaTheory(false, true)},
	{"text+agreement", "text, plus a bonus per other site that carries the same video", rankTextAgreement},
	{"text+sitetrust", "text, plus each site's mean grade on the OTHER queries (leave one out, so no peeking)", nil},
	{"text+seeds", "text, plus the seed confidence (whole phrase, or part of it repeated by one uploader)", rankTextSeeds},
	{"text+seeds+expansion", "text+seeds, plus the related lists of a niche query's confident hits", rankExpansion},
}

// Theory weights.
const (
	// agreementBonus is added per other site carrying the same video.
	agreementBonus = 0.3
	// siteTrustWeight scales a site's leave-one-out mean grade (0..1).
	siteTrustWeight = 0.5
	// seedWeight scales a result's seed confidence.
	seedWeight = 0.5
	// expansionWeight scales a related result's seed confidence, damped by
	// its place in the related list.
	expansionWeight = 0.7
	expansionDecay  = 0.1
)

// fusionConstant is the usual reciprocal rank fusion constant.
const fusionConstant = 60.0

// List fit: a list whose titles match the query poorly (a site answering with
// its popular videos) gets a lower rank share, down to listFitFloor.
const (
	listFitFloor   = 0.3
	listFitCeiling = 1.5
)

// rankPluginOrder concatenates the lists.
func rankPluginOrder(query Query, lists []List) []plugins.DiscoveredItem {
	lists = keywordLists(lists)
	out := []plugins.DiscoveredItem{}
	for _, list := range lists {
		out = append(out, list.Items...)
	}
	return out
}

// rankInterleave takes the first of every list, then the second, and so on.
func rankInterleave(query Query, lists []List) []plugins.DiscoveredItem {
	lists = keywordLists(lists)
	out := []plugins.DiscoveredItem{}
	for place := 0; ; place++ {
		added := false
		for _, list := range lists {
			if place < len(list.Items) {
				out = append(out, list.Items[place])
				added = true
			}
		}
		if !added {
			return out
		}
	}
}

// scoredItem is a result and a theory's score for it.
type scoredItem struct {
	item  plugins.DiscoveredItem
	score float64
}

// fusionScores returns every result's reciprocal rank fusion score. Results
// sit on one site each, so fusion only orders places across sites.
func fusionScores(lists []List) []scoredItem {
	out := []scoredItem{}
	for _, list := range lists {
		for index, item := range list.Items {
			out = append(out, scoredItem{item: item, score: 1 / (fusionConstant + float64(index+1))})
		}
	}
	return out
}

// rankFusion orders by reciprocal rank fusion.
func rankFusion(query Query, lists []List) []plugins.DiscoveredItem {
	lists = keywordLists(lists)
	return sortedItems(fusionScores(lists))
}

// rankText orders by text match, fusion breaking ties.
func rankText(query Query, lists []List) []plugins.DiscoveredItem {
	lists = keywordLists(lists)
	words := searchrank.QueryWords(query.Text)
	scored := fusionScores(lists)
	for index := range scored {
		scored[index].score += searchrank.TextMatch(words, searchrank.TextFields{Title: scored[index].item.Title})
	}
	return sortedItems(scored)
}

// rankSearchrank feeds every list through the production ranker.
func rankSearchrank(query Query, lists []List) []plugins.DiscoveredItem {
	lists = keywordLists(lists)
	return runRanker(searchrank.New(query.Text, 0), lists)
}

// rankSearchrankListFit is the production ranker with each site's rank share
// scaled by its list fit relative to the other sites for this query.
func rankSearchrankListFit(query Query, lists []List) []plugins.DiscoveredItem {
	lists = keywordLists(lists)
	ranker := searchrank.New(query.Text, 0)
	ranker.SetLearned(searchrank.Learned{SourceShares: listFitShares(query, lists)})
	return runRanker(ranker, lists)
}

// runRanker adds the lists to a ranker in site order and returns its ranking.
func runRanker(ranker *searchrank.Ranker, lists []List) []plugins.DiscoveredItem {
	for _, list := range lists {
		for _, item := range list.Items {
			if item.MediaType == "" {
				item.MediaType = plugins.MediaTypeScene
			}
			ranker.Add(searchrank.Hit{Source: list.Plugin, Item: item})
		}
	}
	out := []plugins.DiscoveredItem{}
	for _, ranked := range ranker.Ranked() {
		out = append(out, ranked.Hit.Item)
	}
	return out
}

// ListFit is how well a site's list answers the query: the mean text match of
// its titles, in [0, 1]. A site that ignores the query scores near 0.
func ListFit(query Query, list List) float64 {
	if len(list.Items) == 0 {
		return 0
	}
	words := searchrank.QueryWords(query.Text)
	total := 0.0
	for _, item := range list.Items {
		total += searchrank.TextMatch(words, searchrank.TextFields{Title: item.Title})
	}
	return total / float64(len(list.Items))
}

// listFitShares maps each plugin to its list fit divided by the mean fit of
// the query's lists, bounded to [listFitFloor, listFitCeiling].
func listFitShares(query Query, lists []List) map[string]float64 {
	fits := map[string]float64{}
	total := 0.0
	for _, list := range lists {
		fits[list.Plugin] = ListFit(query, list)
		total += fits[list.Plugin]
	}
	shares := map[string]float64{}
	if total == 0 {
		return shares
	}
	mean := total / float64(len(lists))
	for plugin, fit := range fits {
		shares[plugin] = min(listFitCeiling, max(listFitFloor, fit/mean))
	}
	return shares
}

// sortedItems orders scored results best first, keeping the input order on ties.
func sortedItems(scored []scoredItem) []plugins.DiscoveredItem {
	sort.SliceStable(scored, func(left, right int) bool { return scored[left].score > scored[right].score })
	out := make([]plugins.DiscoveredItem, 0, len(scored))
	for _, entry := range scored {
		out = append(out, entry.item)
	}
	return out
}

// formulaTheory re-builds the production per-result formula, with the site
// rank and quality terms switchable, to see what each term is worth.
func formulaTheory(siteRank, quality bool) func(query Query, lists []List) []plugins.DiscoveredItem {
	return func(query Query, lists []List) []plugins.DiscoveredItem {
		words := searchrank.QueryWords(query.Text)
		scored := []scoredItem{}
		for _, list := range keywordLists(lists) {
			for index, item := range list.Items {
				score := searchrank.WeightText * searchrank.TextMatch(words, searchrank.TextFields{Title: item.Title})
				if siteRank {
					score += searchrank.WeightSiteRank * (searchrank.RankFusionConstant + 1) / (searchrank.RankFusionConstant + float64(index+1))
				}
				if quality {
					score += searchrank.WeightQuality * (searchrank.Quality(item.Rating, item.ViewCount) - searchrank.NeutralQuality)
				}
				scored = append(scored, scoredItem{item: item, score: score})
			}
		}
		return sortedItems(scored)
	}
}

// rankTextAgreement adds a bonus for every other site that lists the same
// video (alike.SameScene: runtime and title or cast).
func rankTextAgreement(query Query, lists []List) []plugins.DiscoveredItem {
	lists = keywordLists(lists)
	words := searchrank.QueryWords(query.Text)
	scored := []scoredItem{}
	for listIndex, list := range lists {
		for place, item := range list.Items {
			score := searchrank.TextMatch(words, searchrank.TextFields{Title: item.Title}) + fusionTiebreak(place)
			score += agreementBonus * float64(otherSitesCarrying(item, lists, listIndex))
			scored = append(scored, scoredItem{item: item, score: score})
		}
	}
	return sortedItems(scored)
}

// otherSitesCarrying counts the lists other than skip that hold the same video.
func otherSitesCarrying(item plugins.DiscoveredItem, lists []List, skip int) int {
	count := 0
	for index, list := range lists {
		if index == skip || list.Plugin == lists[skip].Plugin {
			continue
		}
		for _, other := range list.Items {
			if alike.SameScene(candidateOf(item), candidateOf(other)) {
				count++
				break
			}
		}
	}
	return count
}

// candidateOf describes a result for alike's matchers.
func candidateOf(item plugins.DiscoveredItem) alike.Candidate {
	return alike.Candidate{Title: item.Title, DurationSeconds: item.DurationSeconds}
}

// siteTrustTheory scores text plus each site's mean grade on the other
// queries. It reads labels, so it is built per evaluation (see Evaluate) and
// leaves the query being ranked out of the trust estimate.
func siteTrustTheory(config *Config, snapshot *Snapshot, labels *Labels) func(query Query, lists []List) []plugins.DiscoveredItem {
	return func(query Query, lists []List) []plugins.DiscoveredItem {
		trust := siteTrust(config, snapshot, labels, query.ID)
		words := searchrank.QueryWords(query.Text)
		scored := []scoredItem{}
		for _, list := range keywordLists(lists) {
			for place, item := range list.Items {
				score := searchrank.TextMatch(words, searchrank.TextFields{Title: item.Title}) + fusionTiebreak(place)
				score += siteTrustWeight * trust[list.Plugin]
				scored = append(scored, scoredItem{item: item, score: score})
			}
		}
		return sortedItems(scored)
	}
}

// siteTrust is every site's mean grade (0..1) over the keyword results of all
// queries except leaveOut.
func siteTrust(config *Config, snapshot *Snapshot, labels *Labels, leaveOut string) map[string]float64 {
	totals, counts := map[string]float64{}, map[string]int{}
	for _, query := range config.Queries {
		if query.ID == leaveOut {
			continue
		}
		for _, list := range keywordLists(snapshot.ListsFor(query.ID)) {
			for _, item := range list.Items {
				grade, _ := labels.Grade(query.ID, ItemKey(item))
				totals[list.Plugin] += float64(grade) / GradePerfect
				counts[list.Plugin]++
			}
		}
	}
	trust := map[string]float64{}
	for plugin, total := range totals {
		trust[plugin] = total / float64(counts[plugin])
	}
	return trust
}

// rankTextSeeds adds the seed confidence to the text match.
func rankTextSeeds(query Query, lists []List) []plugins.DiscoveredItem {
	return sortedItems(seededScores(query, lists))
}

// seededScores scores keyword results by text match plus seed confidence.
func seededScores(query Query, lists []List) []scoredItem {
	confidences := SeedConfidences(query, lists)
	words := searchrank.QueryWords(query.Text)
	scored := []scoredItem{}
	for _, list := range keywordLists(lists) {
		for place, item := range list.Items {
			score := searchrank.TextMatch(words, searchrank.TextFields{Title: item.Title}) + fusionTiebreak(place)
			score += seedWeight * confidences[ItemKey(item)]
			scored = append(scored, scoredItem{item: item, score: score})
		}
	}
	return scored
}

// rankExpansion is text+seeds plus the related lists of the confident hits:
// a related result scores its seed's confidence, damped by its place, on top
// of its own text match. A result already in the keyword lists keeps the
// better of its two scores.
func rankExpansion(query Query, lists []List) []plugins.DiscoveredItem {
	scored := seededScores(query, lists)
	index := map[string]int{}
	for position, entry := range scored {
		index[ItemKey(entry.item)] = position
	}
	words := searchrank.QueryWords(query.Text)
	for _, list := range lists {
		if list.Kind != KindRelated {
			continue
		}
		for place, item := range list.Items {
			score := searchrank.TextMatch(words, searchrank.TextFields{Title: item.Title})
			score += expansionWeight * list.SeedConfidence / (1 + expansionDecay*float64(place))
			if position, known := index[ItemKey(item)]; known {
				scored[position].score = max(scored[position].score, score)
				continue
			}
			index[ItemKey(item)] = len(scored)
			scored = append(scored, scoredItem{item: item, score: score})
		}
	}
	return sortedItems(scored)
}

// fusionTiebreak is a tiny site-rank term that only orders equal scores.
func fusionTiebreak(place int) float64 {
	return 0.001 / (1 + float64(place))
}
