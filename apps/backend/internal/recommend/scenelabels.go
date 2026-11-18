package recommend

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Scene labels are the tags the phone's scene model gives stretches of videos
// (ai segments: "Blowjob", "Doggystyle"). How the user treats those stretches
// says what they're into, without them saying a word: the share of a label
// they actually play, the seconds of it they play again, the sections of it
// they seek past, the label on screen when they finish. Each is compared with
// the other labels, joined with the cover picks (covertags.go) and the tag's
// taste affinity into one score, and told back as evidence they can correct.

// SceneLabelTaste is how much the user seems to like one scene label, why,
// and what they answered about it.
type SceneLabelTaste struct {
	Tag schema.Tag
	// Score is in [-1, 1]: above 0 into it, below 0 avoided.
	Score float64
	// Certainty in [0, 1] is how much the score rests on: little means it's
	// a guess worth asking about.
	Certainty float64
	Evidence  []string
	// Confirmed is the user's answer: true for yes, false for not really,
	// nil when they haven't said.
	Confirmed *bool
	// VideoCount is how many videos have the label.
	VideoCount int
}

// labelEvidence is what the event log says about one label.
type labelEvidence struct {
	// presentSeconds is how long the label lasts in the videos played;
	// watchedSeconds how much of it played (replays included),
	// replaySeconds how much of that played more than once.
	presentSeconds float64
	watchedSeconds float64
	replaySeconds  float64
	// passes counts the playbacks that went from before one of its sections
	// to past it; skips those that seeked over it.
	passes int
	skips  int
	// finishes weighs the o events with it on screen (inferred ones at
	// inferredEventShare); finishCount counts them.
	finishes    float64
	finishCount int
}

// labelTotals sums the evidence over every label, the baseline each label is
// compared with.
type labelTotals struct {
	labelEvidence
	// allFinishes is the weight of every o event in a labelled video.
	allFinishes float64
}

// playedSpan is one heatSpan: a stretch of a video that played in a session.
type playedSpan struct {
	span
	videoID string
	session string
}

// finishMoment is an o event at a position of a video.
type finishMoment struct {
	videoID  string
	position float64
	weight   float64
}

// labelHeatQuery reads the window's heatSpans of videos, with their session.
const labelHeatQuery = `SELECT "subjectId", COALESCE("sessionId", substr("at", 1, 10)) AS "session",
	json_extract("data", '` + pathStart + `') AS "start", json_extract("data", '` + pathEnd + `') AS "end"
	FROM "Event"
	WHERE "kind" = 'heatSpan' AND "subjectKind" = 'video' AND "at" >= $since AND "deletedAt" IS NULL
	ORDER BY "at" DESC LIMIT $limit`

// labelFinishQuery reads the window's o events that say where in a video.
const labelFinishQuery = `SELECT "subjectId", json_extract("data", '` + pathPosition + `') AS "position",
	json_extract("data", '` + pathInferred + `') AS "inferred"
	FROM "Event"
	WHERE "kind" = 'o' AND "subjectKind" = 'video' AND "at" >= $since AND "deletedAt" IS NULL
	  AND json_extract("data", '` + pathPosition + `') IS NOT NULL
	ORDER BY "at" DESC LIMIT $limit`

// sceneLabelCountQuery lists every tag of a live video's ai segments with how
// many videos carry it, most common first.
const sceneLabelCountQuery = `SELECT "segment"."tag" AS "tag", COUNT(DISTINCT "segment"."subjectId") AS "count"
	FROM "Segment" AS "segment" JOIN "Video" AS "video" ON "video"."id" = "segment"."subjectId"
	WHERE "segment"."subjectKind" = 'video' AND "segment"."origin" = 'ai' AND "segment"."deletedAt" IS NULL
	  AND "segment"."tag" IS NOT NULL AND "video"."deletedAt" IS NULL
	GROUP BY "segment"."tag" ORDER BY "count" DESC LIMIT $limit`

// sceneLabels scores every scene label of the library against the user's
// playback, cover picks, tag affinities and answers, best first.
func (e *Engine) sceneLabels(ctx context.Context, taste *profile, answers *verdicts, now time.Time) ([]SceneLabelTaste, error) {
	counts, err := e.database.Query(ctx, sceneLabelCountQuery, db.Vars{"limit": sceneLabelLimit})
	if err != nil {
		return nil, fmt.Errorf("read scene labels: %w", err)
	}
	if len(counts) == 0 {
		return []SceneLabelTaste{}, nil
	}
	evidence, totals, err := e.labelPlayback(ctx, now)
	if err != nil {
		return nil, err
	}
	lifts, err := e.loadCoverTagLifts(ctx, now)
	if err != nil {
		return nil, err
	}
	tags, err := e.tagsByID(ctx, counts)
	if err != nil {
		return nil, err
	}
	out := make([]SceneLabelTaste, 0, len(counts))
	for _, row := range counts {
		tag, found := tags[rowString(row, "tag")]
		if !found {
			continue
		}
		parts := labelParts{evidence: evidence[tag.Id], totals: totals, coverLift: lifts[tag.Id], hasCover: hasKey(lifts, tag.Id), affinity: taste.tags[tag.Id]}
		out = append(out, SceneLabelTaste{
			Tag:        tag,
			Score:      parts.score(),
			Certainty:  parts.certainty(),
			Evidence:   parts.sentences(tag.Name),
			Confirmed:  answers.of(schema.RecordKindTag, tag.Id),
			VideoCount: rowInt(row, "count"),
		})
	}
	sortLabels(out)
	return out, nil
}

// hasKey reports whether the map has the key.
func hasKey(values map[string]float64, key string) bool {
	_, found := values[key]
	return found
}

// tagsByID reads the live tags the count rows name.
func (e *Engine) tagsByID(ctx context.Context, counts []db.Row) (map[string]schema.Tag, error) {
	ids := make([]string, 0, len(counts))
	for _, row := range counts {
		ids = append(ids, rowString(row, "tag"))
	}
	tags, err := db.QueryAs[schema.Tag](ctx, e.database,
		`SELECT * FROM "Tag" WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`, db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("read scene label tags: %w", err)
	}
	out := make(map[string]schema.Tag, len(tags))
	for _, tag := range tags {
		out[tag.Id] = tag
	}
	return out, nil
}

// sortLabels puts the highest scores first, then the most common labels.
func sortLabels(labels []SceneLabelTaste) {
	sort.SliceStable(labels, func(left, right int) bool {
		if labels[left].Score != labels[right].Score {
			return labels[left].Score > labels[right].Score
		}
		return labels[left].VideoCount > labels[right].VideoCount
	})
}

// labelPlayback reads the window's playback and finishes and tallies them per
// label of the videos' ai segments.
func (e *Engine) labelPlayback(ctx context.Context, now time.Time) (map[string]*labelEvidence, labelTotals, error) {
	since := db.FormatTime(now.Add(-signalWindow))
	played, err := e.labelSpans(ctx, since)
	if err != nil {
		return nil, labelTotals{}, err
	}
	finishes, err := e.labelFinishes(ctx, since)
	if err != nil {
		return nil, labelTotals{}, err
	}
	videoIDs := map[string]bool{}
	for _, entry := range played {
		videoIDs[entry.videoID] = true
	}
	for _, finish := range finishes {
		videoIDs[finish.videoID] = true
	}
	segments, err := e.loadAISegments(ctx, setKeys(videoIDs))
	if err != nil {
		return nil, labelTotals{}, err
	}
	evidence := tallyLabels(segments, played, finishes)
	return evidence, totalOf(evidence, finishes, segments), nil
}

// labelSpans reads the window's heatSpans; malformed ones are left out.
func (e *Engine) labelSpans(ctx context.Context, since string) ([]playedSpan, error) {
	rows, err := e.database.Query(ctx, labelHeatQuery, db.Vars{"since": since, "limit": eventReadLimit})
	if err != nil {
		return nil, fmt.Errorf("read label playback: %w", err)
	}
	out := make([]playedSpan, 0, len(rows))
	for _, row := range rows {
		played := span{start: rowFloat(row, "start"), end: rowFloat(row, "end")}
		if played.start < 0 || played.end <= played.start || played.end-played.start > maxHeatSpanSeconds {
			continue
		}
		out = append(out, playedSpan{span: played, videoID: rowString(row, "subjectId"), session: rowString(row, "session")})
	}
	return out, nil
}

// labelFinishes reads the window's o events with a position.
func (e *Engine) labelFinishes(ctx context.Context, since string) ([]finishMoment, error) {
	rows, err := e.database.Query(ctx, labelFinishQuery, db.Vars{"since": since, "limit": eventReadLimit})
	if err != nil {
		return nil, fmt.Errorf("read label finishes: %w", err)
	}
	out := make([]finishMoment, 0, len(rows))
	for _, row := range rows {
		weight := 1.0
		if isTruthy(row["inferred"]) {
			weight = inferredEventShare
		}
		out = append(out, finishMoment{videoID: rowString(row, "subjectId"), position: rowFloat(row, "position"), weight: weight})
	}
	return out, nil
}

// loadAISegments reads the ai segments of the videos, per video and tag, each
// tag's stretches merged where they overlap.
func (e *Engine) loadAISegments(ctx context.Context, videoIDs []string) (map[string]map[string][]span, error) {
	out := map[string]map[string][]span{}
	if len(videoIDs) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx,
		`SELECT "subjectId", "tag", "startSeconds", "endSeconds" FROM "Segment"
		 WHERE "subjectKind" = 'video' AND "subjectId" IN (SELECT value FROM json_each($ids))
		   AND "origin" = 'ai' AND "tag" IS NOT NULL AND "endSeconds" IS NOT NULL AND "deletedAt" IS NULL`,
		db.Vars{"ids": videoIDs})
	if err != nil {
		return nil, fmt.Errorf("read label segments: %w", err)
	}
	for _, row := range rows {
		videoID := rowString(row, "subjectId")
		if out[videoID] == nil {
			out[videoID] = map[string][]span{}
		}
		tagID := rowString(row, "tag")
		out[videoID][tagID] = append(out[videoID][tagID], span{start: rowFloat(row, "startSeconds"), end: rowFloat(row, "endSeconds")})
	}
	for _, byTag := range out {
		for tagID, spans := range byTag {
			byTag[tagID] = mergedSpans(spans)
		}
	}
	return out, nil
}

// mergedSpans sorts spans and joins the overlapping ones.
func mergedSpans(spans []span) []span {
	ordered := append([]span(nil), spans...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left].start < ordered[right].start })
	out := []span{}
	for _, stretch := range ordered {
		last := len(out) - 1
		if last >= 0 && stretch.start <= out[last].end {
			out[last].end = math.Max(out[last].end, stretch.end)
			continue
		}
		out = append(out, stretch)
	}
	return out
}

// overlapSeconds is how long two spans share.
func overlapSeconds(first, second span) float64 {
	return math.Max(0, math.Min(first.end, second.end)-math.Max(first.start, second.start))
}

// tallyLabels walks every labelled section of the played videos against
// their playback and the finishes, per label.
func tallyLabels(segments map[string]map[string][]span, played []playedSpan, finishes []finishMoment) map[string]*labelEvidence {
	byVideo := map[string][]playedSpan{}
	for _, entry := range played {
		byVideo[entry.videoID] = append(byVideo[entry.videoID], entry)
	}
	out := map[string]*labelEvidence{}
	for videoID, byTag := range segments {
		for tagID, sections := range byTag {
			tallySections(evidenceFor(out, tagID), sections, byVideo[videoID])
		}
	}
	for _, finish := range finishes {
		for tagID := range labelsOnScreen(segments[finish.videoID], finish.position) {
			evidence := evidenceFor(out, tagID)
			evidence.finishes += finish.weight
			evidence.finishCount++
		}
	}
	return out
}

// evidenceFor is the label's evidence, created on first use.
func evidenceFor(evidence map[string]*labelEvidence, tagID string) *labelEvidence {
	if evidence[tagID] == nil {
		evidence[tagID] = &labelEvidence{}
	}
	return evidence[tagID]
}

// tallySections adds one video's sections of a label against its playback:
// how much of them played, played again, and how often a playback seeked
// over one. Videos that never played add nothing.
func tallySections(evidence *labelEvidence, sections []span, played []playedSpan) {
	if len(played) == 0 {
		return
	}
	spans := make([]span, 0, len(played))
	for _, entry := range played {
		spans = append(spans, entry.span)
	}
	union := mergedSpans(spans)
	sessions := sessionSpans(played)
	for _, section := range sections {
		evidence.presentSeconds += section.end - section.start
		watched := playedWithin(spans, section)
		evidence.watchedSeconds += watched
		evidence.replaySeconds += watched - playedWithin(union, section)
		tallyPasses(evidence, section, sessions)
	}
}

// sessionSpans groups the playback by session, each merged.
func sessionSpans(played []playedSpan) [][]span {
	bySession := map[string][]span{}
	for _, entry := range played {
		bySession[entry.session] = append(bySession[entry.session], entry.span)
	}
	out := make([][]span, 0, len(bySession))
	for _, session := range sortedKeys(bySession) {
		out = append(out, mergedSpans(bySession[session]))
	}
	return out
}

// playedWithin is how many seconds of section the spans cover, each span on
// its own (a replayed stretch counts again).
func playedWithin(spans []span, section span) float64 {
	total := 0.0
	for _, stretch := range spans {
		total += overlapSeconds(stretch, section)
	}
	return total
}

// tallyPasses counts the sessions that played from before section to past
// it, and the ones of them that played under labelSkipPlayedShare of it.
func tallyPasses(evidence *labelEvidence, section span, sessions [][]span) {
	length := section.end - section.start
	for _, merged := range sessions {
		reached := span{start: merged[0].start, end: merged[len(merged)-1].end}
		if reached.start > section.start || reached.end < section.end {
			continue
		}
		evidence.passes++
		if playedWithin(merged, section) < labelSkipPlayedShare*length {
			evidence.skips++
		}
	}
}

// labelsOnScreen are the labels with a section at position, or one that ended
// within labelFinishGraceSeconds before it.
func labelsOnScreen(byTag map[string][]span, position float64) map[string]bool {
	out := map[string]bool{}
	moment := span{start: position - labelFinishGraceSeconds, end: position}
	for tagID, sections := range byTag {
		if touchesAny(moment, sections) {
			out[tagID] = true
		}
	}
	return out
}

// totalOf sums the labels' evidence; every finish in a labelled video counts
// toward allFinishes once, however many labels were on screen.
func totalOf(evidence map[string]*labelEvidence, finishes []finishMoment, segments map[string]map[string][]span) labelTotals {
	totals := labelTotals{}
	for _, entry := range evidence {
		totals.presentSeconds += entry.presentSeconds
		totals.watchedSeconds += entry.watchedSeconds
		totals.replaySeconds += entry.replaySeconds
		totals.passes += entry.passes
		totals.skips += entry.skips
	}
	for _, finish := range finishes {
		if len(segments[finish.videoID]) > 0 {
			totals.allFinishes += finish.weight
		}
	}
	return totals
}

// labelParts is everything one label's score is made of.
type labelParts struct {
	evidence  *labelEvidence
	totals    labelTotals
	coverLift float64
	hasCover  bool
	affinity  float64
}

// watchRatio is how much of the label played for what was there, against
// every label: above 1 it's watched through more than others.
func (parts labelParts) watchRatio() float64 {
	if parts.evidence == nil || parts.evidence.presentSeconds == 0 || parts.totals.presentSeconds == 0 {
		return 1
	}
	return safeRatio(parts.evidence.watchedSeconds/parts.evidence.presentSeconds, parts.totals.watchedSeconds/parts.totals.presentSeconds)
}

// replayRatio is the share of the label's playback that was a replay,
// against every label's.
func (parts labelParts) replayRatio() float64 {
	if parts.evidence == nil || parts.evidence.watchedSeconds == 0 || parts.totals.watchedSeconds == 0 {
		return 1
	}
	return safeRatio(parts.evidence.replaySeconds/parts.evidence.watchedSeconds, parts.totals.replaySeconds/parts.totals.watchedSeconds)
}

// skipRate is the share of the passes over the label that skipped it.
func (parts labelParts) skipRate() float64 {
	if parts.evidence == nil || parts.evidence.passes == 0 {
		return 0
	}
	return float64(parts.evidence.skips) / float64(parts.evidence.passes)
}

// overallSkipRate is the share of every pass that skipped.
func (parts labelParts) overallSkipRate() float64 {
	if parts.totals.passes == 0 {
		return 0
	}
	return float64(parts.totals.skips) / float64(parts.totals.passes)
}

// finishShare is the share of the finishes in labelled videos that came
// with the label on screen.
func (parts labelParts) finishShare() float64 {
	if parts.evidence == nil || parts.totals.allFinishes == 0 {
		return 0
	}
	return parts.evidence.finishes / parts.totals.allFinishes
}

// finishRatio is the label's share of the finishes against its share of the
// playback: above 1 the user finishes to it more than its screen time says.
func (parts labelParts) finishRatio() float64 {
	if parts.evidence == nil || parts.totals.watchedSeconds == 0 || parts.evidence.watchedSeconds == 0 {
		return 1
	}
	return safeRatio(parts.finishShare(), parts.evidence.watchedSeconds/parts.totals.watchedSeconds)
}

// safeRatio is value / baseline, 1 when the baseline is 0.
func safeRatio(value, baseline float64) float64 {
	if baseline <= 0 {
		return 1
	}
	return value / baseline
}

// ratioTerm maps a ratio onto [-1, 1]: 1 is 0, 4× is 1, a quarter is -1.
func ratioTerm(ratio float64) float64 {
	if ratio <= 0 {
		return -1
	}
	return clampFloat(math.Log2(ratio)/2, -1, 1)
}

// confidence is how far amount gets toward full, in [0, 1].
func confidence(amount, full float64) float64 {
	return clampFloat(amount/full, 0, 1)
}

// score combines the parts, each weighed by how much it rests on.
func (parts labelParts) score() float64 {
	total := labelProfileWeight * parts.affinity
	if parts.hasCover {
		total += labelCoverWeight * parts.coverLift
	}
	evidence := parts.evidence
	if evidence == nil {
		return clampFloat(total, -1, 1)
	}
	total += labelWatchWeight * ratioTerm(parts.watchRatio()) * confidence(evidence.watchedSeconds, labelFullWatchSeconds)
	total += labelReplayWeight * ratioTerm(parts.replayRatio()) * confidence(evidence.replaySeconds, labelFullReplaySeconds)
	skipTerm := clampFloat(2*(parts.overallSkipRate()-parts.skipRate()), -1, 1)
	total += labelSkipWeight * skipTerm * confidence(float64(evidence.passes), labelFullPasses)
	total += labelFinishWeight * ratioTerm(parts.finishRatio()) * confidence(parts.finishEvidence(), labelFullFinishes)
	return clampFloat(total, -1, 1)
}

// finishEvidence is how many finishes the label's finish ratio rests on: the
// ones during it, or as many as its share of the playback would have had, so
// a label hardly watched isn't marked down for having none.
func (parts labelParts) finishEvidence() float64 {
	if parts.evidence == nil || parts.totals.watchedSeconds == 0 {
		return 0
	}
	expected := parts.totals.allFinishes * parts.evidence.watchedSeconds / parts.totals.watchedSeconds
	return math.Max(expected, parts.evidence.finishes)
}

// certainty is how much the score rests on, in [0, 1].
func (parts labelParts) certainty() float64 {
	total := 0.0
	if parts.hasCover {
		total += labelCoverCertainty
	}
	evidence := parts.evidence
	if evidence != nil {
		total += evidence.watchedSeconds/labelFullWatchSeconds + float64(evidence.passes)/labelFullPasses + evidence.finishes/labelFullFinishes
	}
	return clampFloat(total, 0, 1)
}

// sentences tells the parts that stand out in words, strongest kinds first.
func (parts labelParts) sentences(name string) []string {
	label := labelInText(name)
	out := []string{}
	out = appendIf(out, parts.finishSentence(label))
	out = appendIf(out, parts.replaySentence(label))
	out = appendIf(out, parts.skipSentence(label))
	out = appendIf(out, parts.watchSentence(label))
	out = appendIf(out, parts.coverSentence(label))
	out = appendIf(out, parts.affinitySentence(name))
	return out
}

// appendIf appends sentence unless it's empty.
func appendIf(sentences []string, sentence string) []string {
	if sentence == "" {
		return sentences
	}
	return append(sentences, sentence)
}

// finishSentence names the finishes during the label.
func (parts labelParts) finishSentence(label string) string {
	if parts.evidence == nil || parts.evidence.finishCount == 0 {
		return ""
	}
	if parts.evidence.finishCount >= 2 && parts.finishShare() >= 0.5 {
		return "Most of your finishes happen during " + label + " scenes"
	}
	if parts.evidence.finishCount == 1 {
		return "One of your finishes happened during a " + label + " scene"
	}
	return fmt.Sprintf("%d of your finishes happened during %s scenes", parts.evidence.finishCount, label)
}

// replaySentence names a replay ratio that stands out.
func (parts labelParts) replaySentence(label string) string {
	if parts.evidence == nil || parts.evidence.replaySeconds < labelFullReplaySeconds/2 {
		return ""
	}
	ratio := parts.replayRatio()
	if ratio < labelEvidenceRatio {
		return ""
	}
	return "You replay " + label + " scenes " + timesText(ratio) + " more than others"
}

// skipSentence names how often the label is skipped, when it often is.
func (parts labelParts) skipSentence(label string) string {
	if parts.evidence == nil || parts.evidence.skips < 2 || parts.skipRate() < 0.3 {
		return ""
	}
	return fmt.Sprintf("You skipped past %s scenes %d of %d times", label, parts.evidence.skips, parts.evidence.passes)
}

// watchSentence names a watch ratio that stands out.
func (parts labelParts) watchSentence(label string) string {
	if parts.evidence == nil || parts.evidence.presentSeconds < labelFullReplaySeconds {
		return ""
	}
	ratio := parts.watchRatio()
	if ratio >= labelEvidenceRatio {
		return "You stay with " + label + " scenes " + timesText(ratio) + " longer than with others"
	}
	if ratio <= 1/labelEvidenceRatio {
		return "You watch less of " + label + " scenes than of others"
	}
	return ""
}

// coverSentence names a cover lift that stands out.
func (parts labelParts) coverSentence(label string) string {
	if !parts.hasCover {
		return ""
	}
	if parts.coverLift >= labelEvidenceLift {
		return fmt.Sprintf("You pick covers showing %s %d%% more often", label, int(math.Round(parts.coverLift*100)))
	}
	if parts.coverLift <= -labelEvidenceLift {
		return "You pass over covers showing " + label
	}
	return ""
}

// affinitySentence names a tag affinity that stands out.
func (parts labelParts) affinitySentence(name string) string {
	if parts.affinity >= 2*labelEvidenceLift {
		return name + " is one of the strongest tags in what you watch"
	}
	if parts.affinity <= -2*labelEvidenceLift {
		return name + " is in much of what you skip or dislike"
	}
	return ""
}

// timesText writes a ratio as "3×", or "1.5×" under 2.
func timesText(ratio float64) string {
	if ratio >= 2 {
		return fmt.Sprintf("%d×", int(math.Round(ratio)))
	}
	return strings.TrimSuffix(fmt.Sprintf("%.1f", ratio), ".0") + "×"
}

// labelInText is a label's name inside a sentence: lower case, unless it's
// written in capitals ("POV").
func labelInText(name string) string {
	for _, letter := range name {
		if unicode.IsLower(letter) {
			return strings.ToLower(name)
		}
	}
	return name
}
