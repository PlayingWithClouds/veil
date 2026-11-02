package library

import (
	"context"
	"fmt"
	"math"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// alikeCandidateLimit bounds how many stored videos a new one is compared
// with: those whose runtime is close to its own.
const alikeCandidateLimit = 200

// alikeDurationWindow is the runtime difference (a share of the runtime)
// within which stored videos are compared; SameScene itself is stricter.
const alikeDurationWindow = 0.05

// DetectAlike compares a video with the stored videos of similar runtime and
// records each that's probably the same scene (alike.SameScene) as an Alike
// pair: the newer record of the pair is hidden from lists until the user
// dismisses the pair or merges it. Nothing is merged here.
func (l *Library) DetectAlike(ctx context.Context, videoID string) error {
	video, err := l.Video(ctx, videoID)
	if err != nil || video == nil || video.DurationSeconds == nil {
		return err
	}
	target, err := l.alikeCandidate(ctx, video)
	if err != nil {
		return err
	}
	others, err := l.similarRuntime(ctx, video)
	if err != nil {
		return err
	}
	for _, other := range others {
		if err := l.recordIfAlike(ctx, video, target, other); err != nil {
			return err
		}
	}
	return nil
}

// similarRuntime lists live videos whose runtime is within the window of
// the video's.
func (l *Library) similarRuntime(ctx context.Context, video *schema.Video) ([]schema.Video, error) {
	duration := *video.DurationSeconds
	window := math.Max(duration*alikeDurationWindow, 2)
	return db.QueryAs[schema.Video](ctx, l.database,
		`SELECT * FROM "Video" WHERE "deletedAt" IS NULL AND "id" != $id
		 AND "durationSeconds" BETWEEN $low AND $high ORDER BY "createdAt" LIMIT $limit`,
		db.Vars{"id": video.Id, "low": duration - window, "high": duration + window, "limit": alikeCandidateLimit})
}

// recordIfAlike stores the pair when the two videos look like one scene.
func (l *Library) recordIfAlike(ctx context.Context, video *schema.Video, target alike.Candidate, other schema.Video) error {
	candidate, err := l.alikeCandidate(ctx, &other)
	if err != nil || !alike.SameScene(target, candidate) {
		return err
	}
	first, second := other.Id, video.Id
	if video.CreatedAt < other.CreatedAt {
		first, second = video.Id, other.Id
	}
	score := alike.Score(alike.Target{Performers: target.Performers, Title: target.Title, DurationSeconds: target.DurationSeconds, Date: target.Date}, candidate)
	_, err = l.database.Exec(ctx,
		`INSERT INTO "Alike" ("id", "subjectKind", "firstId", "secondId", "score", "reason", "dismissed")
		 VALUES ($id, $kind, $first, $second, $score, $reason, 0)
		 ON CONFLICT ("subjectKind", "firstId", "secondId") DO NOTHING`,
		db.Vars{"id": db.NewID(), "kind": schema.RecordKindVideo, "first": first, "second": second, "score": score, "reason": "same runtime and title or cast"})
	return err
}

// alikeCandidate describes a video for the alike matchers.
func (l *Library) alikeCandidate(ctx context.Context, video *schema.Video) (alike.Candidate, error) {
	performers, err := l.PerformersByID(ctx, video.Performers)
	if err != nil {
		return alike.Candidate{}, err
	}
	candidate := alike.Candidate{Title: video.Title}
	for _, performer := range performers {
		candidate.Performers = append(candidate.Performers, performer.Name)
	}
	if video.DurationSeconds != nil {
		candidate.DurationSeconds = int(math.Round(*video.DurationSeconds))
	}
	if video.Date != nil {
		candidate.Date = *video.Date
	}
	return candidate, nil
}

// Alikes lists the pairs waiting for review, most likely first.
func (l *Library) Alikes(ctx context.Context, kind schema.RecordKind, limit, offset int) ([]schema.Alike, error) {
	vars := db.Vars{"kind": kind}
	vars["limit"], vars["offset"] = pageBounds(limit, offset)
	table := db.Quote(tableFor(kind))
	return db.QueryAs[schema.Alike](ctx, l.database,
		`SELECT * FROM "Alike" WHERE "subjectKind" = $kind AND "dismissed" = 0 AND "deletedAt" IS NULL
		 AND "firstId" IN (SELECT "id" FROM `+table+` WHERE "deletedAt" IS NULL)
		 AND "secondId" IN (SELECT "id" FROM `+table+` WHERE "deletedAt" IS NULL)
		 ORDER BY "score" DESC, "createdAt" LIMIT $limit OFFSET $offset`, vars)
}

// AlikePair loads one pair.
func (l *Library) AlikePair(ctx context.Context, id string) (*schema.Alike, error) {
	return db.Get[schema.Alike](ctx, l.database, "Alike", id)
}

// DismissAlike records that the two are different: both show in lists
// again and the pair isn't suggested again.
func (l *Library) DismissAlike(ctx context.Context, id string) error {
	return l.database.Update(ctx, "Alike", id, db.Vars{"dismissed": true})
}

// MergeAlike merges a pair's second record into its first, but only when
// they are provably the same: they share a file fingerprint, or the strict
// matcher (alike.SameSceneStrict, alike.SameGallery) agrees. A loose match
// stays a hidden pair.
func (l *Library) MergeAlike(ctx context.Context, id string) error {
	pair, err := l.AlikePair(ctx, id)
	if err != nil {
		return err
	}
	if pair == nil {
		return fmt.Errorf("no pair %s", id)
	}
	same, err := l.ProvablySame(ctx, pair.SubjectKind, pair.FirstId, pair.SecondId)
	if err != nil {
		return err
	}
	if !same {
		return fmt.Errorf("the two aren't provably the same: they need a matching fingerprint first")
	}
	return l.Merge(ctx, pair.SubjectKind, pair.FirstId, pair.SecondId)
}

// ProvablySame reports whether two media records are the same for merging:
// a shared fingerprint, or the strict matcher for their kind.
func (l *Library) ProvablySame(ctx context.Context, kind schema.RecordKind, firstID, secondID string) (bool, error) {
	shared, err := l.shareFingerprint(ctx, kind, firstID, secondID)
	if err != nil || shared {
		return shared, err
	}
	switch kind {
	case schema.RecordKindVideo:
		return l.strictlySameVideo(ctx, firstID, secondID)
	case schema.RecordKindGallery:
		return l.strictlySameGallery(ctx, firstID, secondID)
	}
	return false, nil
}

// shareFingerprint reports whether two records have a fingerprint of the
// same algorithm and value.
func (l *Library) shareFingerprint(ctx context.Context, kind schema.RecordKind, firstID, secondID string) (bool, error) {
	count, err := l.database.Int(ctx,
		`SELECT COUNT(*) FROM "Fingerprint" AS "first" JOIN "Fingerprint" AS "second"
		   ON "first"."algorithm" = "second"."algorithm" AND "first"."value" = "second"."value"
		 WHERE "first"."subjectKind" = $kind AND "first"."subjectId" = $first AND "first"."deletedAt" IS NULL
		   AND "second"."subjectKind" = $kind AND "second"."subjectId" = $second AND "second"."deletedAt" IS NULL`,
		db.Vars{"kind": kind, "first": firstID, "second": secondID})
	return count > 0, err
}

// strictlySameVideo runs the strict scene matcher on two videos.
func (l *Library) strictlySameVideo(ctx context.Context, firstID, secondID string) (bool, error) {
	first, err := l.Video(ctx, firstID)
	if err != nil || first == nil {
		return false, err
	}
	second, err := l.Video(ctx, secondID)
	if err != nil || second == nil {
		return false, err
	}
	firstCandidate, err := l.alikeCandidate(ctx, first)
	if err != nil {
		return false, err
	}
	secondCandidate, err := l.alikeCandidate(ctx, second)
	if err != nil {
		return false, err
	}
	return alike.SameSceneStrict(firstCandidate, secondCandidate), nil
}

// strictlySameGallery runs the gallery matcher on two galleries.
func (l *Library) strictlySameGallery(ctx context.Context, firstID, secondID string) (bool, error) {
	first, err := l.Gallery(ctx, firstID)
	if err != nil || first == nil {
		return false, err
	}
	second, err := l.Gallery(ctx, secondID)
	if err != nil || second == nil {
		return false, err
	}
	return alike.SameGallery(alike.Candidate{Title: first.Title}, alike.Candidate{Title: second.Title}, int(first.ImageCount), int(second.ImageCount)), nil
}
