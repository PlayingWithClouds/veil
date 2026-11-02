package user

import (
	"context"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// preferenceWindow is how far back picks count towards a site's preference.
const preferenceWindow = 90 * 24 * time.Hour

// preferencePrior is how many impressions a site's rate is pulled towards
// the overall rate with, so a site shown twice and picked once doesn't top
// the list.
const preferencePrior = 50.0

// SitePreference is how likely the user picks a site's videos when shown
// them.
type SitePreference struct {
	Plugin string `json:"plugin"`
	Shown  int    `json:"shown"`
	Chosen int    `json:"chosen"`
	// Score is the smoothed pick rate relative to the overall one: above 1
	// the site is picked more often than average, below 1 less.
	Score float64 `json:"score"`
}

// SitePreferences rates every site whose videos were shown or picked in the
// last 90 days, most preferred first. A video found on several sites counts
// for each of them.
func (s *Service) SitePreferences(ctx context.Context) ([]SitePreference, error) {
	counts, err := db.QueryAs[SitePreference](ctx, s.database,
		`SELECT "source"."plugin" AS "plugin",
		        SUM("event"."kind" = $impression) AS "shown",
		        SUM("event"."kind" = $choice) AS "chosen"
		 FROM "Event" AS "event"
		 JOIN "Source" AS "source" ON "source"."subjectKind" = $video AND "source"."subjectId" = "event"."subjectId"
		      AND "source"."deletedAt" IS NULL
		 WHERE "event"."kind" IN ($impression, $choice) AND "event"."subjectKind" = $video
		   AND "event"."deletedAt" IS NULL AND "event"."at" >= $since
		 GROUP BY "source"."plugin"`,
		db.Vars{
			"impression": schema.EventKindImpression,
			"choice":     schema.EventKindChoice,
			"video":      schema.RecordKindVideo,
			"since":      db.FormatTime(time.Now().Add(-preferenceWindow)),
		})
	if err != nil {
		return nil, err
	}
	scorePreferences(counts)
	return counts, nil
}

// scorePreferences sets each site's score and sorts the most preferred first.
func scorePreferences(counts []SitePreference) {
	shown, chosen := 0, 0
	for _, count := range counts {
		shown += count.Shown
		chosen += count.Chosen
	}
	if shown == 0 || chosen == 0 {
		for index := range counts {
			counts[index].Score = 1
		}
		return
	}
	overall := float64(chosen) / float64(shown)
	for index, count := range counts {
		rate := (float64(count.Chosen) + preferencePrior*overall) / (float64(count.Shown) + preferencePrior)
		counts[index].Score = rate / overall
	}
	sort.SliceStable(counts, func(first, second int) bool {
		return counts[first].Score > counts[second].Score
	})
}
