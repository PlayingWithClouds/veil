package recommend

import (
	"slices"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// heat logs a played stretch of a video in a session, ago before now.
func (f *fixture) heat(videoID, session string, start, end float64, ago time.Duration) {
	f.insert("Event", db.Vars{
		"at": db.FormatTime(f.now.Add(-ago)), "kind": schema.EventKindHeatSpan, "subjectKind": schema.RecordKindVideo,
		"subjectId": videoID, "sessionId": session, "data": map[string]any{"start": start, "end": end},
	})
}

// sceneLabelNamed is the taste's entry of the label, or fails the test.
func sceneLabelNamed(t *testing.T, taste Taste, name string) SceneLabelTaste {
	t.Helper()
	for _, label := range taste.SceneLabels {
		if label.Tag.Name == name {
			return label
		}
	}
	t.Fatalf("no scene label %q in %+v", name, taste.SceneLabels)
	return SceneLabelTaste{}
}

// TestSceneLabelsLearnFromPlayback scores a label the user plays and finishes
// to above one they seek past, and says why.
func TestSceneLabelsLearnFromPlayback(t *testing.T) {
	f := newFixture(t)
	doggy := f.tag("Doggystyle")
	kissing := f.tag("Kissing")
	video := f.video(mediaRow{title: "Scanned", duration: 300})
	f.aiSegment(video, doggy, 0, 100)
	f.aiSegment(video, kissing, 100, 200)
	f.aiSegment(video, doggy, 200, 300)
	sessions := []string{db.NewID(), db.NewID()}
	for index, session := range sessions {
		ago := time.Duration(index+1) * time.Hour
		f.heat(video, session, 0, 100, ago)
		f.heat(video, session, 200, 300, ago)
		f.event(schema.EventKindO, schema.RecordKindVideo, video, ago, map[string]any{"position": 290.0, "inferred": index == 1})
	}
	taste, err := f.engine.Taste(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	liked, skipped := sceneLabelNamed(t, taste, "Doggystyle"), sceneLabelNamed(t, taste, "Kissing")
	if liked.Score <= skipped.Score || skipped.Score >= 0 {
		t.Errorf("scores: doggystyle %v, kissing %v", liked.Score, skipped.Score)
	}
	if !slices.Contains(liked.Evidence, "Most of your finishes happen during doggystyle scenes") {
		t.Errorf("doggystyle evidence = %v", liked.Evidence)
	}
	if !slices.Contains(skipped.Evidence, "You skipped past kissing scenes 2 of 2 times") {
		t.Errorf("kissing evidence = %v", skipped.Evidence)
	}
	if liked.Certainty <= 0 || liked.VideoCount != 1 || liked.Confirmed != nil {
		t.Errorf("doggystyle = %+v", liked)
	}
}

// TestSceneLabelAnswersMakeLikedScenes takes a yes on a label as a liked
// scene, a later 0 as no answer and a -1 as not really.
func TestSceneLabelAnswersMakeLikedScenes(t *testing.T) {
	f := newFixture(t)
	doggy := f.tag("Doggystyle")
	f.aiSegment(f.video(mediaRow{title: "Scanned", duration: 300}), doggy, 0, 100)
	steps := []struct {
		value     int
		confirmed *bool
		liked     bool
	}{
		{1, answerOf(1), true},
		{0, nil, false},
		{-1, answerOf(-1), false},
	}
	for index, step := range steps {
		f.event(schema.EventKindLike, schema.RecordKindTag, doggy, time.Duration(len(steps)-index)*time.Minute, map[string]any{"value": step.value})
		taste, err := f.engine.Taste(f.ctx)
		if err != nil {
			t.Fatal(err)
		}
		label := sceneLabelNamed(t, taste, "Doggystyle")
		if (label.Confirmed == nil) != (step.confirmed == nil) || (label.Confirmed != nil && *label.Confirmed != *step.confirmed) {
			t.Errorf("after %d: confirmed = %v", step.value, label.Confirmed)
		}
		liked, err := f.engine.LikedSceneTags(f.ctx)
		if err != nil {
			t.Fatal(err)
		}
		if (len(liked) == 1) != step.liked {
			t.Errorf("after %d: liked scenes = %+v", step.value, liked)
		}
	}
}

// TestSiteAnswersCountAsVerdicts likes a site through a like without a
// subject naming it.
func TestSiteAnswersCountAsVerdicts(t *testing.T) {
	f := newFixture(t)
	f.insert("Event", db.Vars{"at": db.FormatTime(f.now.Add(-time.Hour)), "kind": schema.EventKindLike, "data": map[string]any{"site": "tube", "value": 1}})
	taste, err := f.engine.Taste(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(taste.Sites) != 1 || taste.Sites[0].Id != "tube" || taste.Sites[0].Weight <= 0 || taste.Sites[0].Confirmed == nil || !*taste.Sites[0].Confirmed {
		t.Errorf("sites = %+v", taste.Sites)
	}
}

// TestLabelSentencesNameWhatStandsOut words a replay ratio, a cover lift and
// leaves out what doesn't stand out.
func TestLabelSentencesNameWhatStandsOut(t *testing.T) {
	parts := labelParts{
		evidence:  &labelEvidence{presentSeconds: 300, watchedSeconds: 300, replaySeconds: 200},
		totals:    labelTotals{labelEvidence: labelEvidence{presentSeconds: 1000, watchedSeconds: 1000, replaySeconds: 250}},
		coverLift: 0.4,
		hasCover:  true,
	}
	sentences := parts.sentences("POV")
	expected := []string{"You replay POV scenes 3× more than others", "You pick covers showing POV 40% more often"}
	if !slices.Equal(sentences, expected) {
		t.Errorf("sentences = %v", sentences)
	}
	if parts.score() <= 0 {
		t.Errorf("score = %v", parts.score())
	}
}
