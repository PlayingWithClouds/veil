package follows

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// listLimit is how many items one site page or search may bring in.
const listLimit = 40

// newMediaLimit caps how many new media one run reports.
const newMediaLimit = 200

// scheduleInterval is how often Run looks for due follows.
const scheduleInterval = time.Minute

// mediaKinds are the kinds a follow's filter is counted on.
var mediaKinds = []schema.RecordKind{schema.RecordKindVideo, schema.RecordKindGallery}

// payloadFollowID is the job payload key naming the follow to run.
const payloadFollowID = "followId"

// RunTask is the task of followRun jobs: it runs the follow the payload
// names. Site errors end up in the follow's lastError, not in a retry.
func (s *Service) RunTask() jobs.Task {
	return func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		followID, _ := job.Payload[payloadFollowID].(string)
		if followID == "" {
			return fmt.Errorf("followRun job without a %s", payloadFollowID)
		}
		_, err := s.RunFollow(ctx, followID)
		return err
	}
}

// RunFollow runs one follow now and returns the new media it found. A
// disabled or trashed follow doesn't run.
func (s *Service) RunFollow(ctx context.Context, followID string) ([]schema.RecordRef, error) {
	follow, err := s.Get(ctx, followID)
	if err != nil || follow == nil || !follow.Enabled {
		return nil, err
	}
	if isCamFollow(*follow) {
		// The cam watcher (internal/cams) checks followed rooms going live;
		// a room has no media to list.
		return nil, s.recordRun(ctx, *follow, s.now(), nil)
	}
	startedAt := s.now()
	listed, listErr := s.discover(ctx, *follow)
	// Runs count [previous finish, this finish): media created in the
	// finishing millisecond itself are counted by the next run, never lost.
	finishedAt := s.now()
	since := db.FormatTime(startedAt)
	if follow.LastRunAt != nil {
		since = *follow.LastRunAt
	}
	media, err := s.newMedia(ctx, *follow, listed, since, db.FormatTime(finishedAt))
	if err != nil {
		return nil, err
	}
	if err := s.recordRun(ctx, *follow, finishedAt, listErr); err != nil {
		return nil, err
	}
	s.announce(ctx, *follow, media)
	return media, nil
}

// isCamFollow reports whether the follow is of a cam room.
func isCamFollow(follow schema.Follow) bool {
	return follow.TargetKind != nil && *follow.TargetKind == schema.RecordKindCamModel
}

// discover brings the follow's listings into the library: the target's
// pages on each of its sites, or the search on every site. Sites that can't
// list pages are skipped; failures are joined into one error.
func (s *Service) discover(ctx context.Context, follow schema.Follow) ([]schema.RecordRef, error) {
	if s.discoverer == nil {
		return nil, nil
	}
	if follow.TargetKind == nil {
		return s.search(ctx, follow), nil
	}
	sources, err := s.library.Sources(ctx, *follow.TargetKind, *follow.TargetId)
	if err != nil {
		return nil, err
	}
	listed := []schema.RecordRef{}
	var failures []error
	for _, source := range sources {
		refs, err := s.discoverer.ListPage(ctx, source.Plugin, source.Url, listLimit)
		if errors.Is(err, pipeline.ErrNoPageListing) {
			continue
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", source.Plugin, err))
			continue
		}
		listed = append(listed, refs...)
	}
	return listed, errors.Join(failures...)
}

// search runs a text follow's search on every site.
func (s *Service) search(ctx context.Context, follow schema.Follow) []schema.RecordRef {
	if follow.Filter.Text == nil || strings.TrimSpace(*follow.Filter.Text) == "" {
		return nil
	}
	return s.discoverer.Search(ctx, plugins.MediaTypeScene, *follow.Filter.Text, listLimit)
}

// newMedia is the media that arrived in [since, until): everything matching
// the follow's filter, plus what the run listed on the target's own pages,
// which carries no credits yet (see withoutTarget).
func (s *Service) newMedia(ctx context.Context, follow schema.Follow, listed []schema.RecordRef, since, until string) ([]schema.RecordRef, error) {
	media := []schema.RecordRef{}
	for _, kind := range kindsOf(follow.Filter) {
		ids, err := s.newIDs(ctx, follow, kind, listed, since, until)
		if err != nil {
			return nil, err
		}
		media = appendRefs(media, kind, ids)
	}
	if len(media) > newMediaLimit {
		media = media[:newMediaLimit]
	}
	return media, nil
}

// appendRefs appends the ids of kind as refs, skipping ones already there.
func appendRefs(refs []schema.RecordRef, kind schema.RecordKind, ids []string) []schema.RecordRef {
	for _, id := range ids {
		ref := schema.RecordRef{Kind: kind, Id: id}
		if !slices.Contains(refs, ref) {
			refs = append(refs, ref)
		}
	}
	return refs
}

// newIDs is newMedia for one kind.
func (s *Service) newIDs(ctx context.Context, follow schema.Follow, kind schema.RecordKind, listed []schema.RecordRef, since, until string) ([]string, error) {
	ids, err := s.library.AddedBetween(ctx, kind, follow.Filter, since, until, newMediaLimit)
	if err != nil || follow.TargetKind == nil {
		return ids, err
	}
	created, err := s.createdBetween(ctx, kind, idsOfKind(listed, kind), since, until)
	if err != nil {
		return nil, err
	}
	matching, err := s.library.MatchingIDs(ctx, kind, withoutTarget(follow.Filter, *follow.TargetKind), created)
	if err != nil {
		return nil, err
	}
	return append(ids, matching...), nil
}

// createdBetween keeps the ids of kind that were created in [since, until).
func (s *Service) createdBetween(ctx context.Context, kind schema.RecordKind, ids []string, since, until string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	table := "Video"
	if kind == schema.RecordKindGallery {
		table = "Gallery"
	}
	return s.database.Strings(ctx,
		`SELECT "id" FROM `+db.Quote(table)+` WHERE "id" IN (SELECT value FROM json_each($ids))
		 AND "createdAt" >= $since AND "createdAt" < $until`,
		db.Vars{"ids": ids, "since": since, "until": until})
}

// idsOfKind picks the ids of refs of kind.
func idsOfKind(refs []schema.RecordRef, kind schema.RecordKind) []string {
	ids := []string{}
	for _, ref := range refs {
		if ref.Kind == kind {
			ids = append(ids, ref.Id)
		}
	}
	return ids
}

// kindsOf is the media kinds a filter asks for; all of them when it names
// none.
func kindsOf(filter schema.Filter) []schema.RecordKind {
	if len(filter.Kinds) == 0 {
		return mediaKinds
	}
	kinds := []schema.RecordKind{}
	for _, kind := range mediaKinds {
		if slices.Contains(filter.Kinds, schema.MediaKind(kind)) {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

// recordRun stamps the run on the follow: when it ran, when it runs next,
// and the site error, if any.
func (s *Service) recordRun(ctx context.Context, follow schema.Follow, finishedAt time.Time, listErr error) error {
	fields := db.Vars{
		"lastRunAt": db.FormatTime(finishedAt),
		"nextRunAt": db.FormatTime(finishedAt.Add(time.Duration(follow.IntervalHours) * time.Hour)),
		"lastError": nil,
	}
	if listErr != nil {
		fields["lastError"] = listErr.Error()
	}
	return s.database.Update(ctx, followTable, follow.Id, fields)
}

// announce notifies about new media, when the follow asks for it, and
// fires the new-media hook.
func (s *Service) announce(ctx context.Context, follow schema.Follow, media []schema.RecordRef) {
	if len(media) == 0 {
		return
	}
	if follow.Notify && s.notifier != nil {
		notification := notify.NewNotification{Kind: "newMedia", Title: fmt.Sprintf("%d new for %s", len(media), follow.Name), Subject: targetOf(follow)}
		if _, err := s.notifier.Notify(ctx, notification); err != nil {
			log.Printf("follows: notify %s: %v", follow.Name, err)
		}
	}
	if s.onNewMedia != nil {
		s.onNewMedia(ctx, follow, media)
	}
}

// targetOf is the followed record, or nil for a search follow.
func targetOf(follow schema.Follow) *schema.RecordRef {
	if follow.TargetKind == nil {
		return nil
	}
	return &schema.RecordRef{Kind: *follow.TargetKind, Id: *follow.TargetId}
}

// Run enqueues due follows every minute until ctx ends.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(scheduleInterval)
	defer ticker.Stop()
	for {
		if _, err := s.EnqueueDue(ctx); err != nil {
			log.Printf("follows: enqueue due: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// EnqueueDue queues a followRun job for every enabled follow whose next run
// has come, one per follow, and returns how many it queued.
func (s *Service) EnqueueDue(ctx context.Context) (int, error) {
	due, err := s.database.Strings(ctx,
		`SELECT "id" FROM "Follow" WHERE "deletedAt" IS NULL AND "enabled" = 1
		 AND ("nextRunAt" IS NULL OR "nextRunAt" <= $now)`,
		db.Vars{"now": db.FormatTime(s.now())})
	if err != nil {
		return 0, err
	}
	for _, id := range due {
		if _, err := s.Enqueue(ctx, id); err != nil {
			return 0, err
		}
	}
	return len(due), nil
}

// Enqueue queues a run of one follow now and returns the job id; a run
// already queued is reused.
func (s *Service) Enqueue(ctx context.Context, followID string) (string, error) {
	return s.queue.Enqueue(ctx, jobs.Job{
		Kind:      schema.JobKindFollowRun,
		Payload:   map[string]any{payloadFollowID: followID},
		DedupeKey: "followRun:" + followID,
	})
}

// NewSince lists the media matching a follow's filter that arrived after
// since (db.TimeFormat), newest first. The schema keeps no "seen" mark per
// follow, so the client passes when it last looked; "" means since the
// follow was made.
func (s *Service) NewSince(ctx context.Context, followID, since string, limit int) ([]schema.RecordRef, error) {
	follow, err := s.Get(ctx, followID)
	if err != nil || follow == nil {
		return nil, missingFollow(followID, err)
	}
	if since == "" {
		since = follow.CreatedAt
	}
	until := db.FormatTime(s.now().Add(time.Millisecond))
	media := []schema.RecordRef{}
	for _, kind := range kindsOf(follow.Filter) {
		ids, err := s.library.AddedBetween(ctx, kind, follow.Filter, since, until, limit)
		if err != nil {
			return nil, err
		}
		media = appendRefs(media, kind, ids)
	}
	return media, nil
}
