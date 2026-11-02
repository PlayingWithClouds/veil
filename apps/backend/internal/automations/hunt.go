package automations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Searcher asks every plugin for a query and stores the results;
// *pipeline.Pipeline is one.
type Searcher interface {
	Search(ctx context.Context, entity plugins.MediaType, query string, limit int) []schema.RecordRef
}

// WantedHuntID is the fixed id of the built-in wanted hunt, so every
// install ensures the same record.
const WantedHuntID = "5d0f6a3e-8c1b-4f2a-9e7d-3b6c1a2f4e01"

// wantedHuntCron runs the hunt daily at 04:00.
const wantedHuntCron = "0 4 * * *"

// huntSearchLimit is how many results one plugin may return per wanted
// video.
const huntSearchLimit = 20

// huntBatch caps how many wanted videos one run hunts when its input is
// empty.
const huntBatch = 200

// EnsureBuiltIns adds the built-in automations that don't exist yet: the
// daily wanted hunt. One the user disabled stays disabled.
func (s *Service) EnsureBuiltIns(ctx context.Context) error {
	existing, err := s.database.Int(ctx, `SELECT COUNT(*) FROM "Automation" WHERE "id" = $id`, db.Vars{"id": WantedHuntID})
	if err != nil || existing > 0 {
		return err
	}
	cronExpression := wantedHuntCron
	_, err = s.database.Insert(ctx, automationTable, schema.Automation{
		Entity:  schema.Entity{Id: WantedHuntID},
		Name:    "Wanted hunt",
		Enabled: true,
		Trigger: schema.Trigger{Kind: schema.TriggerKindSchedule, Cron: &cronExpression},
		Steps:   []schema.Step{{Action: ActionHunt}},
		BuiltIn: true,
	})
	return err
}

// huntAction searches every site for each wanted video of its input (every
// wanted video, up to huntBatch, when the input is empty). A result that is
// strictly the same scene (alike.SameSceneStrict) gives the wanted video its
// sources and streams and is trashed; the wanted video becomes available,
// wantedFound fires and the user is notified. No args. A dry run only
// reports what it would search, since searching stores results. Output: the
// videos found.
func (s *Service) huntAction(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error) {
	wanted, err := s.wantedVideos(ctx, input)
	if err != nil {
		return Result{}, err
	}
	if run.DryRun || s.searcher == nil {
		return Result{Output: []schema.RecordRef{}, Touched: []schema.RecordRef{}, Report: fmt.Sprintf("would search for %d wanted", len(wanted))}, nil
	}
	result := Result{Output: []schema.RecordRef{}, Touched: []schema.RecordRef{}}
	for _, video := range wanted {
		found, err := s.huntOne(ctx, video)
		if err != nil {
			return result, err
		}
		if found == nil {
			continue
		}
		if err := s.claimFound(ctx, run, video, *found); err != nil {
			return result, err
		}
		wantedRef := schema.RecordRef{Kind: schema.RecordKindVideo, Id: video.Id}
		result.Output = append(result.Output, wantedRef)
		result.Touched = append(result.Touched, wantedRef, schema.RecordRef{Kind: schema.RecordKindVideo, Id: found.Id})
	}
	result.Report = fmt.Sprintf("found %d of %d wanted", len(result.Output), len(wanted))
	return result, nil
}

// wantedVideos loads the wanted videos among refs, or every wanted video
// when refs is empty.
func (s *Service) wantedVideos(ctx context.Context, refs []schema.RecordRef) ([]schema.Video, error) {
	filter := schema.Filter{Status: []schema.MediaStatus{schema.MediaStatusWanted}}
	if len(refs) == 0 {
		return s.library.Videos(ctx, library.MediaQuery{Filter: filter, Limit: huntBatch})
	}
	ids, err := s.library.MatchingIDs(ctx, schema.RecordKindVideo, filter, idsOfKind(refs, schema.RecordKindVideo))
	if err != nil {
		return nil, err
	}
	videos := []schema.Video{}
	for _, id := range ids {
		video, err := s.library.Video(ctx, id)
		if err != nil {
			return nil, err
		}
		if video != nil {
			videos = append(videos, *video)
		}
	}
	return videos, nil
}

// huntOne searches the sites for a wanted video and returns the first
// result that is the same scene, or nil.
func (s *Service) huntOne(ctx context.Context, wanted schema.Video) (*schema.Video, error) {
	target, err := s.candidate(ctx, wanted)
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(wanted.Title + " " + strings.Join(target.Performers, " "))
	for _, ref := range s.searcher.Search(ctx, plugins.MediaTypeScene, query, huntSearchLimit) {
		if ref.Kind != schema.RecordKindVideo || ref.Id == wanted.Id {
			continue
		}
		found, err := s.sameScene(ctx, target, ref.Id)
		if err != nil || found != nil {
			return found, err
		}
	}
	return nil, nil
}

// sameScene loads a search result and returns it when it is an available
// copy of the target scene.
func (s *Service) sameScene(ctx context.Context, target alike.Candidate, videoID string) (*schema.Video, error) {
	video, err := s.library.Video(ctx, videoID)
	if err != nil || video == nil || video.Status == schema.MediaStatusWanted {
		return nil, err
	}
	candidate, err := s.candidate(ctx, *video)
	if err != nil || !alike.SameSceneStrict(target, candidate) {
		return nil, err
	}
	return video, nil
}

// candidate describes a video for the scene matcher.
func (s *Service) candidate(ctx context.Context, video schema.Video) (alike.Candidate, error) {
	performers, err := s.library.PerformersByID(ctx, video.Performers)
	if err != nil {
		return alike.Candidate{}, err
	}
	names := make([]string, 0, len(performers))
	for _, performer := range performers {
		names = append(names, performer.Name)
	}
	candidate := alike.Candidate{Title: video.Title, Performers: names}
	if video.DurationSeconds != nil {
		candidate.DurationSeconds = int(*video.DurationSeconds)
	}
	if video.Date != nil {
		candidate.Date = *video.Date
	}
	return candidate, nil
}

// claimFound moves the found copy's sources and streams onto the wanted
// video, trashes the copy and marks the wanted video available, then fires
// wantedFound and notifies. The status write is dropped when the user
// locked the field.
func (s *Service) claimFound(ctx context.Context, run RunContext, wanted, found schema.Video) error {
	if err := s.library.MoveSources(ctx, schema.RecordKindVideo, found.Id, wanted.Id); err != nil {
		return err
	}
	foundRef := schema.RecordRef{Kind: schema.RecordKindVideo, Id: found.Id}
	if err := s.trash(ctx, foundRef, run.Author); err != nil {
		return err
	}
	if _, err := s.library.Recorder().Update(ctx, schema.RecordKindVideo, wanted.Id, db.Vars{"status": schema.MediaStatusAvailable}, run.Author); err != nil {
		return err
	}
	wantedRef := schema.RecordRef{Kind: schema.RecordKindVideo, Id: wanted.Id}
	if _, err := s.Fire(ctx, schema.TriggerKindWantedFound, map[string]any{"wantedId": wanted.Id, "foundId": found.Id}, []schema.RecordRef{wantedRef}); err != nil {
		log.Printf("automations: fire wantedFound for %s: %v", wanted.Id, err)
	}
	s.notifyFound(ctx, wanted, wantedRef)
	return nil
}

// notifyFound tells the user a wanted video was found.
func (s *Service) notifyFound(ctx context.Context, wanted schema.Video, ref schema.RecordRef) {
	if s.notifier == nil {
		return
	}
	notification := notify.NewNotification{Kind: "wantedFound", Title: "Found: " + wanted.Title, Subject: &ref}
	if _, err := s.notifier.Notify(ctx, notification); err != nil {
		log.Printf("automations: notify found %s: %v", wanted.Id, err)
	}
}
