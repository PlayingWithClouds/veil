package searchlab

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// Cutoffs the metrics are taken at: a first screen and a full page.
const (
	shortCutoff = 10
	pageCutoff  = 24
)

// QueryScore is one theory's result on one query.
type QueryScore struct {
	QueryID string  `json:"query_id"`
	NDCG10  float64 `json:"ndcg10"`
	NDCG24  float64 `json:"ndcg24"`
	// Precision10 is the share of the top 10 graded RelevantGrade or better.
	Precision10 float64 `json:"precision10"`
	// Unlabelled10 counts top-10 results without a grade (scored as wrong).
	Unlabelled10 int `json:"unlabelled10"`
}

// TheoryScore is one theory across every query.
type TheoryScore struct {
	Theory      string       `json:"theory"`
	Queries     []QueryScore `json:"queries"`
	MeanNDCG10  float64      `json:"mean_ndcg10"`
	MeanNDCG24  float64      `json:"mean_ndcg24"`
	MeanPrec10  float64      `json:"mean_precision10"`
	Unlabelled  int          `json:"unlabelled10"`
	Description string       `json:"description"`
}

// ListScore is what the labels say about one site's list for one query, next
// to what the list-fit heuristic predicted without labels.
type ListScore struct {
	QueryID     string  `json:"query_id"`
	Plugin      string  `json:"plugin"`
	Results     int     `json:"results"`
	Precision10 float64 `json:"precision10"`
	MeanGrade   float64 `json:"mean_grade"`
	ListFit     float64 `json:"list_fit"`
	Error       string  `json:"error,omitempty"`
}

// Report is the whole evaluation.
type Report struct {
	Theories []TheoryScore `json:"theories"`
	Lists    []ListScore   `json:"lists"`
	// ListFitCorrelation is the Spearman rank correlation between list fit and
	// labelled precision over all lists: does list fit spot bad search?
	ListFitCorrelation float64 `json:"list_fit_correlation"`
	Labelled           int     `json:"labelled"`
	Total              int     `json:"total"`
}

// Evaluate scores every theory and every site list against the labels.
func Evaluate(config *Config, snapshot *Snapshot, labels *Labels) *Report {
	report := &Report{}
	for _, theory := range Theories {
		report.Theories = append(report.Theories, scoreTheory(theory, config, snapshot, labels))
	}
	for _, query := range config.Queries {
		for _, list := range snapshot.ListsFor(query.ID) {
			report.Lists = append(report.Lists, scoreList(query, list, labels))
			report.Total += len(list.Items)
			report.Labelled += labelledCount(query.ID, list.Items, labels)
		}
	}
	report.ListFitCorrelation = listFitCorrelation(report.Lists)
	return report
}

// scoreTheory runs one theory on every query.
func scoreTheory(theory Theory, config *Config, snapshot *Snapshot, labels *Labels) TheoryScore {
	score := TheoryScore{Theory: theory.Name, Description: theory.Description}
	for _, query := range config.Queries {
		lists := snapshot.ListsFor(query.ID)
		ranking := theory.Rank(query, lists)
		queryScore := scoreRanking(query.ID, ranking, idealGrades(query.ID, lists, labels), labels)
		score.Queries = append(score.Queries, queryScore)
		score.MeanNDCG10 += queryScore.NDCG10
		score.MeanNDCG24 += queryScore.NDCG24
		score.MeanPrec10 += queryScore.Precision10
		score.Unlabelled += queryScore.Unlabelled10
	}
	if count := float64(len(config.Queries)); count > 0 {
		score.MeanNDCG10 /= count
		score.MeanNDCG24 /= count
		score.MeanPrec10 /= count
	}
	return score
}

// scoreRanking computes the metrics of one ranking.
func scoreRanking(queryID string, ranking []plugins.DiscoveredItem, ideal []int, labels *Labels) QueryScore {
	grades := make([]int, len(ranking))
	score := QueryScore{QueryID: queryID}
	for index, item := range ranking {
		grade, found := labels.Grade(queryID, ItemKey(item))
		grades[index] = grade
		if index < shortCutoff && !found {
			score.Unlabelled10++
		}
	}
	score.NDCG10 = ndcg(grades, ideal, shortCutoff)
	score.NDCG24 = ndcg(grades, ideal, pageCutoff)
	score.Precision10 = precision(grades, shortCutoff)
	return score
}

// idealGrades are all labelled grades of a query, best first: the ranking a
// perfect theory would produce.
func idealGrades(queryID string, lists []List, labels *Labels) []int {
	grades := []int{}
	for _, list := range lists {
		for _, item := range list.Items {
			if grade, found := labels.Grade(queryID, ItemKey(item)); found {
				grades = append(grades, grade)
			}
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(grades)))
	return grades
}

// ndcg is the normalized discounted cumulative gain at cutoff, gain 2^grade−1.
func ndcg(grades, ideal []int, cutoff int) float64 {
	best := dcg(ideal, cutoff)
	if best == 0 {
		return 0
	}
	return dcg(grades, cutoff) / best
}

// dcg is the discounted cumulative gain of the first cutoff grades.
func dcg(grades []int, cutoff int) float64 {
	total := 0.0
	for index := 0; index < len(grades) && index < cutoff; index++ {
		total += (math.Pow(2, float64(grades[index])) - 1) / math.Log2(float64(index+2))
	}
	return total
}

// precision is the share of the first cutoff places graded relevant; missing
// places count as misses.
func precision(grades []int, cutoff int) float64 {
	hits := 0
	for index := 0; index < len(grades) && index < cutoff; index++ {
		if grades[index] >= RelevantGrade {
			hits++
		}
	}
	return float64(hits) / float64(cutoff)
}

// scoreList grades one site's own list.
func scoreList(query Query, list List, labels *Labels) ListScore {
	score := ListScore{QueryID: query.ID, Plugin: list.Plugin, Results: len(list.Items), ListFit: ListFit(query, list), Error: list.Error}
	grades := []int{}
	gradeTotal := 0
	for _, item := range list.Items {
		grade, _ := labels.Grade(query.ID, ItemKey(item))
		grades = append(grades, grade)
		gradeTotal += grade
	}
	score.Precision10 = precision(grades, shortCutoff)
	if len(grades) > 0 {
		score.MeanGrade = float64(gradeTotal) / float64(len(grades))
	}
	return score
}

// labelledCount counts the items with a grade.
func labelledCount(queryID string, items []plugins.DiscoveredItem, labels *Labels) int {
	count := 0
	for _, item := range items {
		if _, found := labels.Grade(queryID, ItemKey(item)); found {
			count++
		}
	}
	return count
}

// listFitCorrelation is the Spearman correlation of list fit and labelled
// precision over the lists that returned anything.
func listFitCorrelation(lists []ListScore) float64 {
	fits, precisions := []float64{}, []float64{}
	for _, list := range lists {
		if list.Results == 0 {
			continue
		}
		fits = append(fits, list.ListFit)
		precisions = append(precisions, list.Precision10)
	}
	return pearson(ranks(fits), ranks(precisions))
}

// ranks replaces values by their ranks, ties sharing the mean rank.
func ranks(values []float64) []float64 {
	order := make([]int, len(values))
	for index := range order {
		order[index] = index
	}
	sort.SliceStable(order, func(left, right int) bool { return values[order[left]] < values[order[right]] })
	out := make([]float64, len(values))
	for start := 0; start < len(order); {
		end := start
		for end+1 < len(order) && values[order[end+1]] == values[order[start]] {
			end++
		}
		for index := start; index <= end; index++ {
			out[order[index]] = float64(start+end)/2 + 1
		}
		start = end + 1
	}
	return out
}

// pearson is the Pearson correlation of two series; 0 when undefined.
func pearson(first, second []float64) float64 {
	count := float64(len(first))
	if count < 2 {
		return 0
	}
	firstMean, secondMean := mean(first), mean(second)
	covariance, firstVariance, secondVariance := 0.0, 0.0, 0.0
	for index := range first {
		firstDelta, secondDelta := first[index]-firstMean, second[index]-secondMean
		covariance += firstDelta * secondDelta
		firstVariance += firstDelta * firstDelta
		secondVariance += secondDelta * secondDelta
	}
	if firstVariance == 0 || secondVariance == 0 {
		return 0
	}
	return covariance / math.Sqrt(firstVariance*secondVariance)
}

// mean is the arithmetic mean.
func mean(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

// Print writes the report as readable tables.
func (report *Report) Print(writer io.Writer, config *Config) {
	fmt.Fprintf(writer, "labelled %d of %d results\n\n", report.Labelled, report.Total)
	fmt.Fprintf(writer, "%-20s %8s %8s %8s %6s", "theory", "nDCG@10", "nDCG@24", "P@10", "unlab")
	for _, query := range config.Queries {
		fmt.Fprintf(writer, " %10s", truncate(query.ID, 10))
	}
	fmt.Fprintln(writer, "   (per query: nDCG@10)")
	for _, theory := range report.Theories {
		fmt.Fprintf(writer, "%-20s %8.3f %8.3f %8.3f %6d", theory.Theory, theory.MeanNDCG10, theory.MeanNDCG24, theory.MeanPrec10, theory.Unlabelled)
		for _, query := range theory.Queries {
			fmt.Fprintf(writer, " %10.3f", query.NDCG10)
		}
		fmt.Fprintln(writer)
	}
	fmt.Fprintf(writer, "\n%-12s %-14s %7s %6s %10s %8s\n", "query", "plugin", "results", "P@10", "mean grade", "list fit")
	for _, list := range report.Lists {
		line := fmt.Sprintf("%-12s %-14s %7d %6.2f %10.2f %8.2f", truncate(list.QueryID, 12), list.Plugin, list.Results, list.Precision10, list.MeanGrade, list.ListFit)
		if list.Error != "" {
			line += "  error: " + truncate(list.Error, 60)
		}
		fmt.Fprintln(writer, line)
	}
	fmt.Fprintf(writer, "\nlist fit vs labelled P@10, Spearman: %.2f (1 = list fit spots bad search perfectly)\n", report.ListFitCorrelation)
	for _, theory := range report.Theories {
		fmt.Fprintf(writer, "  %-20s %s\n", theory.Theory, theory.Description)
	}
}

// truncate shortens text to at most length runes.
func truncate(text string, length int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= length {
		return string(runes)
	}
	return string(runes[:length-1]) + "…"
}
