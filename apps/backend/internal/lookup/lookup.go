// Package lookup answers "who is this, which video is this": a text or a
// picture goes to every identification plugin at once (lookup, face:identify,
// identification forums), and what they name comes back as library records,
// each with a score and the evidence behind it. Named people only come from
// plugins over databases of professional performers and cam models (see the
// plugin rules); similar tube videos are records of their own, never credits.
package lookup

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// pluginTimeout is how long one plugin may take to answer.
	pluginTimeout = 45 * time.Second
	// maxPictureBytes is the biggest picture sent; services accept about 6 MB.
	maxPictureBytes = 5 << 20
	// forumThreads is how many matching forum threads are read for names.
	forumThreads = 3
	// answerScore is the score a solved forum answer gives the name it gives.
	answerScore = 0.6
	// corroboration is added to a match's score for every further piece of
	// evidence, up to 1.
	corroboration = 0.1
	// maxMatches bounds the merged answer.
	maxMatches = 40
)

// Query is a lookup input: a text, a picture, or both. Plugins narrows it
// to some plugins (all when empty).
type Query struct {
	Text     string
	Image    []byte
	MimeType string
	Plugins  []string
}

// Result is a lookup's merged answer, best first, and the plugins that
// failed.
type Result struct {
	Matches  []schema.Match
	Failures []string
}

// Service runs lookups. Safe for concurrent use.
type Service struct {
	registry *plugins.Registry
	runner   *plugins.Runner
	library  *library.Library
}

// New returns the lookup service.
func New(registry *plugins.Registry, runner *plugins.Runner, lib *library.Library) *Service {
	return &Service{registry: registry, runner: runner, library: lib}
}

// finding is one plugin's answer before merging: a stored record and why.
type finding struct {
	ref      schema.RecordRef
	score    float64
	evidence schema.Evidence
}

// Lookup asks every identification plugin about the query at once and
// merges their answers.
func (s *Service) Lookup(ctx context.Context, query Query) (Result, error) {
	if query.Text == "" && len(query.Image) == 0 {
		return Result{}, fmt.Errorf("a lookup needs a text or a picture")
	}
	if len(query.Image) > maxPictureBytes {
		return Result{}, fmt.Errorf("the picture is larger than %d MB", maxPictureBytes>>20)
	}
	askers := s.askers(query)
	answers := make([][]finding, len(askers))
	failures := make([]string, len(askers))
	var group sync.WaitGroup
	for index, ask := range askers {
		group.Go(func() {
			askContext, cancel := context.WithTimeout(ctx, pluginTimeout)
			defer cancel()
			found, err := ask.run(askContext, query)
			answers[index] = found
			if err != nil {
				log.Printf("lookup %s: %v", ask.plugin.Meta.Name, err)
				failures[index] = ask.plugin.Meta.Name + ": " + err.Error()
			}
		})
	}
	group.Wait()
	return Result{Matches: merge(answers), Failures: nonEmpty(failures)}, nil
}

// asker is one plugin capability a query goes to.
type asker struct {
	plugin *plugins.Plugin
	run    func(ctx context.Context, query Query) ([]finding, error)
}

// askers lists the plugin capabilities that can answer the query.
func (s *Service) askers(query Query) []asker {
	askers := []asker{}
	for _, plugin := range s.registry.WithAnyCapability(plugins.CapabilityLookup, plugins.CapabilityFaceIdentify, plugins.CapabilityForumSearch, plugins.CapabilityLegacyForumSearch) {
		if !wanted(query.Plugins, plugin.Meta.Name) {
			continue
		}
		askers = append(askers, s.askersOf(plugin, query)...)
	}
	return askers
}

// askersOf lists what one plugin is asked for this query.
func (s *Service) askersOf(plugin *plugins.Plugin, query Query) []asker {
	askers := []asker{}
	if plugin.Serves(plugins.CapabilityLookup) {
		askers = append(askers, asker{plugin: plugin, run: func(ctx context.Context, query Query) ([]finding, error) {
			return s.askLookup(ctx, plugin, query)
		}})
	}
	if len(query.Image) > 0 && plugin.Serves(plugins.CapabilityFaceIdentify) {
		askers = append(askers, asker{plugin: plugin, run: func(ctx context.Context, query Query) ([]finding, error) {
			return s.askFace(ctx, plugin, query)
		}})
	}
	if query.Text != "" && plugin.ServesForum() {
		askers = append(askers, asker{plugin: plugin, run: func(ctx context.Context, query Query) ([]finding, error) {
			return s.askForum(ctx, plugin, query.Text)
		}})
	}
	return askers
}

// wanted reports whether a plugin is among those asked for (all when none).
func wanted(names []string, name string) bool {
	return len(names) == 0 || slices.Contains(names, name)
}

// askLookup runs a plugin's lookup capability and stores what it names.
func (s *Service) askLookup(ctx context.Context, plugin *plugins.Plugin, query Query) ([]finding, error) {
	result, err := s.runner.Lookup(ctx, plugin, plugins.LookupArgs{
		Text: query.Text, ImageBase64: encodeImage(query.Image), MimeType: query.MimeType,
	})
	if err != nil {
		return nil, err
	}
	found := []finding{}
	for _, match := range result.Matches {
		ref, err := s.library.IngestListed(ctx, plugin.Meta.Name, match.Item)
		if err != nil || ref.Id == "" {
			continue
		}
		for _, evidence := range match.Evidence {
			found = append(found, finding{ref: ref, score: match.Score, evidence: evidenceOf(plugin.Meta.Name, evidence)})
		}
		if len(match.Evidence) == 0 {
			found = append(found, finding{ref: ref, score: match.Score, evidence: schema.Evidence{Plugin: plugin.Meta.Name}})
		}
	}
	return found, nil
}

// encodeImage is the picture as base64, "" for none.
func encodeImage(image []byte) string {
	if len(image) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(image)
}

// evidenceOf maps a plugin's evidence to the schema's.
func evidenceOf(plugin string, evidence plugins.LookupEvidence) schema.Evidence {
	mapped := schema.Evidence{Plugin: plugin, Url: optional(evidence.URL), Title: optional(evidence.Title), Detail: optional(evidence.Detail)}
	if evidence.ImageURL != "" {
		mapped.Image = &schema.Asset{Uri: evidence.ImageURL}
	}
	return mapped
}

// optional returns nil for "", else a pointer to the text.
func optional(text string) *string {
	if text == "" {
		return nil
	}
	return &text
}

// merge combines every plugin's findings into one match per record: the
// best score, raised a little by each further piece of evidence.
func merge(answers [][]finding) []schema.Match {
	byRecord := map[schema.RecordRef]*schema.Match{}
	order := []schema.RecordRef{}
	for _, found := range answers {
		for _, item := range found {
			order = addFinding(byRecord, order, item)
		}
	}
	matches := make([]schema.Match, 0, len(order))
	for _, ref := range order {
		matches = append(matches, *byRecord[ref])
	}
	sort.SliceStable(matches, func(first, second int) bool { return scoreOf(matches[first]) > scoreOf(matches[second]) })
	if len(matches) > maxMatches {
		matches = matches[:maxMatches]
	}
	return matches
}

// addFinding folds one finding into the merged matches.
func addFinding(byRecord map[schema.RecordRef]*schema.Match, order []schema.RecordRef, item finding) []schema.RecordRef {
	match, known := byRecord[item.ref]
	if !known {
		score := float32(item.score)
		byRecord[item.ref] = &schema.Match{Item: item.ref, Score: &score, Evidence: []schema.Evidence{item.evidence}}
		return append(order, item.ref)
	}
	best := max(float64(*match.Score), item.score)
	raised := float32(min(1, best+corroboration))
	match.Score = &raised
	match.Evidence = append(match.Evidence, item.evidence)
	return order
}

// scoreOf is a match's score, 0 when it has none.
func scoreOf(match schema.Match) float32 {
	if match.Score == nil {
		return 0
	}
	return *match.Score
}

// nonEmpty drops the empty entries.
func nonEmpty(values []string) []string {
	kept := []string{}
	for _, value := range values {
		if value != "" {
			kept = append(kept, value)
		}
	}
	return kept
}
