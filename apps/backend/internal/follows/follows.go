// Package follows keeps the user's follows: a performer, studio, tag,
// category or search, each a filter that runs on the sites every few hours
// and reports the media it brings in.
//
// A run lists the target's pages on its sites (each Source of the record,
// through plugins declaring scene:list:page) or, for a search, asks every
// plugin; then it counts the media matching the filter that arrived since
// the previous run, notifies and fires OnNewMedia.
package follows

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// followTable is the table follows are stored in.
const followTable = "Follow"

// defaultIntervalHours is how often a follow runs when it names no interval.
const defaultIntervalHours = 24

// Discoverer brings site listings into the library; *pipeline.Pipeline is
// one. ListPage returns an error wrapping pipeline.ErrNoPageListing for a
// plugin that can't list pages.
type Discoverer interface {
	ListPage(ctx context.Context, pluginName, url string, limit int) ([]schema.RecordRef, error)
	Search(ctx context.Context, entity plugins.MediaType, query string, limit int) []schema.RecordRef
}

// NewMediaHook is called after a run found new media for a follow.
type NewMediaHook func(ctx context.Context, follow schema.Follow, media []schema.RecordRef)

// Options are what the service works with. Notifier and OnNewMedia may be
// nil.
type Options struct {
	Library    *library.Library
	Queue      *jobs.Queue
	Discoverer Discoverer
	Notifier   *notify.Service
	OnNewMedia NewMediaHook
}

// Service stores and runs follows.
type Service struct {
	database   *db.DB
	library    *library.Library
	queue      *jobs.Queue
	discoverer Discoverer
	notifier   *notify.Service
	onNewMedia NewMediaHook
	now        func() time.Time
}

// New returns a follow service.
func New(options Options) *Service {
	return &Service{
		database:   options.Library.DB(),
		library:    options.Library,
		queue:      options.Queue,
		discoverer: options.Discoverer,
		notifier:   options.Notifier,
		onNewMedia: options.OnNewMedia,
		now:        time.Now,
	}
}

// NewFollow is a follow to add. TargetKind and TargetID name the followed
// record, when it is one; IntervalHours defaults to defaultIntervalHours.
type NewFollow struct {
	Name          string
	Filter        schema.Filter
	TargetKind    *schema.RecordKind
	TargetID      *string
	IntervalHours int
	Notify        bool
}

// FollowEdit changes a follow; nil fields stay as they are.
type FollowEdit struct {
	Name          *string
	Filter        *schema.Filter
	IntervalHours *int
	Enabled       *bool
	Notify        *bool
}

// Follow adds an enabled follow, due to run at once.
func (s *Service) Follow(ctx context.Context, follow NewFollow) (*schema.Follow, error) {
	name := strings.TrimSpace(follow.Name)
	if name == "" {
		return nil, fmt.Errorf("a follow needs a name")
	}
	if (follow.TargetKind == nil) != (follow.TargetID == nil) {
		return nil, fmt.Errorf("a follow's target needs both kind and id")
	}
	interval := follow.IntervalHours
	if interval <= 0 {
		interval = defaultIntervalHours
	}
	now := db.FormatTime(s.now())
	id, err := s.database.Insert(ctx, followTable, schema.Follow{
		Name:          name,
		Filter:        follow.Filter,
		TargetKind:    follow.TargetKind,
		TargetId:      follow.TargetID,
		IntervalHours: int32(interval),
		Enabled:       true,
		Notify:        follow.Notify,
		NextRunAt:     &now,
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Update applies an edit to a follow. A new interval moves the next run to
// that long after the last one.
func (s *Service) Update(ctx context.Context, id string, edit FollowEdit) (*schema.Follow, error) {
	follow, err := s.Get(ctx, id)
	if err != nil || follow == nil {
		return nil, missingFollow(id, err)
	}
	fields, err := editFields(*follow, edit, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.database.Update(ctx, followTable, id, fields); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// editFields turns an edit into the columns it changes.
func editFields(follow schema.Follow, edit FollowEdit, now time.Time) (db.Vars, error) {
	fields := db.Vars{}
	if edit.Name != nil {
		name := strings.TrimSpace(*edit.Name)
		if name == "" {
			return nil, fmt.Errorf("a follow needs a name")
		}
		fields["name"] = name
	}
	if edit.Filter != nil {
		fields["filter"] = *edit.Filter
	}
	if edit.Enabled != nil {
		fields["enabled"] = *edit.Enabled
	}
	if edit.Notify != nil {
		fields["notify"] = *edit.Notify
	}
	if edit.IntervalHours != nil {
		if *edit.IntervalHours <= 0 {
			return nil, fmt.Errorf("a follow's interval must be positive")
		}
		fields["intervalHours"] = *edit.IntervalHours
		fields["nextRunAt"] = db.FormatTime(lastRunOrNow(follow, now).Add(time.Duration(*edit.IntervalHours) * time.Hour))
	}
	return fields, nil
}

// lastRunOrNow is when the follow last ran, or now when it never did.
func lastRunOrNow(follow schema.Follow, now time.Time) time.Time {
	if follow.LastRunAt == nil {
		return now
	}
	lastRun, err := time.Parse(db.TimeFormat, *follow.LastRunAt)
	if err != nil {
		return now
	}
	return lastRun
}

// Unfollow moves a follow to the trash.
func (s *Service) Unfollow(ctx context.Context, id string) error {
	return s.database.Trash(ctx, followTable, id)
}

// Get loads one follow; nil when there is none.
func (s *Service) Get(ctx context.Context, id string) (*schema.Follow, error) {
	return db.Get[schema.Follow](ctx, s.database, followTable, id)
}

// Follows lists every follow, by name.
func (s *Service) Follows(ctx context.Context) ([]schema.Follow, error) {
	return db.QueryAs[schema.Follow](ctx, s.database,
		`SELECT * FROM "Follow" WHERE "deletedAt" IS NULL ORDER BY "name" COLLATE NOCASE, rowid`, nil)
}

// missingFollow is the error for a follow that couldn't be loaded.
func missingFollow(id string, err error) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("no follow %s", id)
}
