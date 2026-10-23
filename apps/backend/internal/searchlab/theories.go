package searchlab

import (
	"sort"

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
}

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
	out := []plugins.DiscoveredItem{}
	for _, list := range lists {
		out = append(out, list.Items...)
	}
	return out
}

// rankInterleave takes the first of every list, then the second, and so on.
func rankInterleave(query Query, lists []List) []plugins.DiscoveredItem {
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
	return sortedItems(fusionScores(lists))
}

// rankText orders by text match, fusion breaking ties.
func rankText(query Query, lists []List) []plugins.DiscoveredItem {
	words := searchrank.QueryWords(query.Text)
	scored := fusionScores(lists)
	for index := range scored {
		scored[index].score += searchrank.TextMatch(words, searchrank.TextFields{Title: scored[index].item.Title})
	}
	return sortedItems(scored)
}

// rankSearchrank feeds every list through the production ranker.
func rankSearchrank(query Query, lists []List) []plugins.DiscoveredItem {
	return runRanker(searchrank.New(query.Text, 0), lists)
}

// rankSearchrankListFit is the production ranker with each site's rank share
// scaled by its list fit relative to the other sites for this query.
func rankSearchrankListFit(query Query, lists []List) []plugins.DiscoveredItem {
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
