package pipeline

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Notifier tells the user something happened (notify.Service).
type Notifier interface {
	Notify(ctx context.Context, notification notify.NewNotification) (*schema.Notification, error)
}

// collection tallies a favourite's collection, for the notification when it ends.
type collection struct {
	mu    sync.Mutex
	added int
	sites []string
}

// Collecting is what's being fetched about one performer now.
type Collecting struct {
	PerformerID string
	// Reading: their profiles are being read and the sites asked (after favouriting).
	Reading bool
	// Listing: plugins whose list of their videos is queued or running.
	Listing []string
}

// SetNotifier sets who hears when a favourite's collection is done.
func (p *Pipeline) SetNotifier(notifier Notifier) {
	p.notifier = notifier
}

// Collecting lists every performer something is being fetched about now.
func (p *Pipeline) Collecting(ctx context.Context) ([]Collecting, error) {
	byPerformer := map[string]*Collecting{}
	order := []string{}
	entry := func(id string) *Collecting {
		if found, ok := byPerformer[id]; ok {
			return found
		}
		byPerformer[id] = &Collecting{PerformerID: id, Listing: []string{}}
		order = append(order, id)
		return byPerformer[id]
	}
	p.reading.Range(func(key, _ any) bool {
		entry(key.(string)).Reading = true
		return true
	})
	listings, err := p.activeListings(ctx, "", "")
	if err != nil {
		return nil, err
	}
	for _, job := range listings {
		current := entry(job.Subject.Id)
		if !slices.Contains(current.Listing, job.PluginName) {
			current.Listing = append(current.Listing, job.PluginName)
		}
	}
	result := make([]Collecting, 0, len(order))
	for _, id := range order {
		result = append(result, *byPerformer[id])
	}
	return result, nil
}

// activeListings returns the queued and running performer-videos jobs, of
// one performer when performerID is set, leaving out exceptJobID.
func (p *Pipeline) activeListings(ctx context.Context, performerID, exceptJobID string) ([]jobs.Job, error) {
	if p.queue == nil {
		return nil, nil
	}
	active, err := p.queue.List(ctx, jobs.Query{
		Statuses: []schema.JobStatus{jobs.StatusQueued, jobs.StatusRunning},
		Kinds:    []schema.JobKind{jobs.KindScrape},
		Limit:    activeJobsLimit,
	})
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(active, func(job jobs.Job) bool {
		if job.ID == exceptJobID || job.Subject == nil || !IsPerformerVideosJob(&job) {
			return true
		}
		return performerID != "" && job.Subject.Id != performerID
	}), nil
}

// tally counts videos a favourite's collection added from a site.
func (p *Pipeline) tally(performerID, pluginName string, added int) {
	value, ok := p.collections.Load(performerID)
	if !ok {
		return
	}
	current := value.(*collection)
	current.mu.Lock()
	defer current.mu.Unlock()
	current.added += added
	if !slices.Contains(current.sites, pluginName) {
		current.sites = append(current.sites, pluginName)
	}
}

// finishCollection notifies once a favourite's collection is done: nothing
// read any more and no listing left besides exceptJobID, the one finishing.
func (p *Pipeline) finishCollection(ctx context.Context, performerID, exceptJobID string) {
	if _, reading := p.reading.Load(performerID); reading {
		return
	}
	remaining, err := p.activeListings(ctx, performerID, exceptJobID)
	if err != nil || len(remaining) > 0 {
		return
	}
	value, ok := p.collections.LoadAndDelete(performerID)
	if !ok || p.notifier == nil {
		return
	}
	p.notifyCollected(ctx, performerID, value.(*collection))
}

// notifyCollected tells the user how many videos a favourite's collection added.
func (p *Pipeline) notifyCollected(ctx context.Context, performerID string, done *collection) {
	performer, err := p.library.Performer(ctx, performerID)
	if err != nil || performer == nil {
		return
	}
	done.mu.Lock()
	added, sites := done.added, p.siteNames(done.sites)
	done.mu.Unlock()
	notification := notify.NewNotification{
		Kind:    "performerCollected",
		Title:   fmt.Sprintf("Collected %d new videos of %s", added, performer.Name),
		Subject: &schema.RecordRef{Kind: schema.RecordKindPerformer, Id: performerID},
	}
	if len(sites) > 0 {
		notification.Body = "From " + strings.Join(sites, ", ")
	}
	if _, err := p.notifier.Notify(ctx, notification); err != nil {
		log.Printf("notify collected %s: %v", performerID, err)
	}
}

// siteNames are the plugins' display names, else their names.
func (p *Pipeline) siteNames(pluginNames []string) []string {
	shown := make([]string, 0, len(pluginNames))
	for _, name := range pluginNames {
		plugin, found := p.registry.Get(name)
		if found && plugin.Meta.DisplayName != "" {
			name = plugin.Meta.DisplayName
		}
		shown = append(shown, name)
	}
	return shown
}

// Unassign takes a performer off a video they aren't in. When one of the
// performer's profiles put them there, that profile is someone else's of the
// same name: it moves to a new performer with the videos only it credited,
// and what it filled in on the performer is cleared and read again from
// their other profiles. Returns the new performer's id, "" when none.
func (p *Pipeline) Unassign(ctx context.Context, videoID, performerID string) (string, error) {
	sourceID, err := p.library.CreditSourceID(ctx, videoID)
	if err != nil {
		return "", err
	}
	otherID := ""
	if p.isProfileOf(ctx, performerID, sourceID) {
		otherID, err = p.library.SplitProfile(ctx, performerID, sourceID)
		if err != nil {
			return "", err
		}
		p.rereadAfterSplit(performerID, otherID)
	}
	return otherID, p.library.UncreditPerformer(ctx, videoID, performerID)
}

// isProfileOf reports whether the source is one of the performer's profiles.
func (p *Pipeline) isProfileOf(ctx context.Context, performerID, sourceID string) bool {
	if sourceID == "" {
		return false
	}
	sources, err := p.library.Sources(ctx, schema.RecordKindPerformer, performerID)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(sources, func(source schema.Source) bool { return source.Id == sourceID })
}

// rereadAfterSplit refills the performer from their remaining profiles and
// fills the new one from the profile it took, in the background.
func (p *Pipeline) rereadAfterSplit(performerID, otherID string) {
	go func() {
		ctx := context.Background()
		if err := p.Refresh(ctx, schema.RecordKindPerformer, performerID); err != nil {
			log.Printf("unassign: reread %s: %v", performerID, err)
		}
		if err := p.Refresh(ctx, schema.RecordKindPerformer, otherID); err != nil {
			log.Printf("unassign: read %s: %v", otherID, err)
		}
	}()
}
