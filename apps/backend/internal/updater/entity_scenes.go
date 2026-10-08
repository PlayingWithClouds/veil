package updater

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const (
	// entityScenesInterval is how long a fetched studio or performer is left
	// alone before opening its page searches the sites again.
	entityScenesInterval = 24 * time.Hour
	// entityScenesTimeout bounds one fetch; it runs detached from the request
	// so leaving the page does not abort the ingest halfway.
	entityScenesTimeout = 45 * time.Second
	// entityScenesKeyPrefix keys the last fetch per entity in the kv store.
	entityScenesKeyPrefix = "entity-scenes:"
	// minimumMentionLength skips names too short to recognise in a title.
	minimumMentionLength = 3
)

// FetchMarks remembers when an entity's scenes were last fetched (the kv
// cache in production).
type FetchMarks interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

// EntityScenes fetches a studio's or performer's scenes from every scene
// plugin when its page opens, and credits the matches to it so they show up
// among its scenes.
type EntityScenes struct {
	scheduler *Scheduler
	marks     FetchMarks
	mutex     sync.Mutex
	// inFlight holds the running fetch per entity id, so repeated opens join it.
	inFlight map[string]*entityFetch
}

// entityFetch is one running fetch; done closes when linked and err are set.
type entityFetch struct {
	done   chan struct{}
	linked int
	err    error
}

// NewEntityScenes creates the fetcher on top of the scheduler's discovery.
func NewEntityScenes(scheduler *Scheduler, marks FetchMarks) *EntityScenes {
	return &EntityScenes{scheduler: scheduler, marks: marks, inFlight: map[string]*entityFetch{}}
}

// Ensure fetches the entity's scenes unless that happened within the last
// entityScenesInterval, and returns how many scenes the fetch found credited
// to it (0 when skipped). Concurrent calls for one entity share a fetch.
func (e *EntityScenes) Ensure(ctx context.Context, entityID string) (int, error) {
	target, err := db.ParseRecordID(entityID)
	if err != nil {
		return 0, fmt.Errorf("entity scenes: %w", err)
	}
	if target.Table != "studio" && target.Table != "performer" {
		return 0, fmt.Errorf("entity scenes: %s is not a studio or performer", entityID)
	}
	if e.fetchedRecently(ctx, target.String()) {
		return 0, nil
	}
	fetch := e.startFetch(target.String())
	select {
	case <-fetch.done:
		return fetch.linked, fetch.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// fetchedRecently reports whether the entity was fetched within the interval.
func (e *EntityScenes) fetchedRecently(ctx context.Context, entityID string) bool {
	_, err := e.marks.Get(ctx, entityScenesKeyPrefix+entityID)
	return err == nil
}

// startFetch returns the entity's running fetch, starting one when none runs.
func (e *EntityScenes) startFetch(entityID string) *entityFetch {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if running, found := e.inFlight[entityID]; found {
		return running
	}
	fetch := &entityFetch{done: make(chan struct{})}
	e.inFlight[entityID] = fetch
	go e.runFetch(entityID, fetch)
	return fetch
}

// runFetch fetches the entity's scenes, marks it fetched when any plugin
// answered, and releases the waiters.
func (e *EntityScenes) runFetch(entityID string, fetch *entityFetch) {
	ctx, cancel := context.WithTimeout(context.Background(), entityScenesTimeout)
	defer cancel()
	linked, answered, err := e.fetch(ctx, entityID)
	if err != nil {
		log.Printf("entity scenes %s: %v", entityID, err)
	}
	if answered {
		if markErr := e.marks.Set(ctx, entityScenesKeyPrefix+entityID, time.Now().UTC().Format(time.RFC3339), entityScenesInterval); markErr != nil {
			log.Printf("entity scenes %s: mark fetched: %v", entityID, markErr)
		}
	}
	e.mutex.Lock()
	delete(e.inFlight, entityID)
	e.mutex.Unlock()
	fetch.linked = len(linked)
	fetch.err = err
	close(fetch.done)
}

// fetch lists the entity's own site page when a plugin can, and searches its
// name on every scene plugin. Page results are credited to it as they are;
// search results only when already credited to it or, while uncredited, when
// their title mentions its name or an alias. Returns the credited scene ids
// and whether any plugin answered at all.
func (e *EntityScenes) fetch(ctx context.Context, entityID string) (linked map[string]bool, answered bool, err error) {
	origin, err := e.scheduler.repo.EntityOrigin(ctx, entityID)
	if err != nil {
		return nil, false, err
	}
	linked = map[string]bool{}
	pageIDs, listed, pageErr := e.scheduler.listEntityPage(ctx, origin)
	if listed {
		answered = true
		if _, creditErr := e.scheduler.repo.CreditScenes(ctx, entityID, pageIDs); creditErr != nil {
			return linked, answered, creditErr
		}
		e.addLinked(ctx, entityID, pageIDs, linked)
	}

	searchPlugins := scenePlugins(e.scheduler.discovery.Plugins(nil))
	if len(searchPlugins) == 0 {
		return linked, answered, errors.Join(pageErr, errNoSearchPlugin)
	}
	var found []string
	searchErr := e.scheduler.discovery.Run(ctx, discovery.Options{Query: origin.Name, Limit: resultLimit}, searchPlugins, collectHits(&found))
	if searchErr == nil || len(found) > 0 {
		answered = true
	}
	if err := e.creditMentions(ctx, entityID, append([]string{origin.Name}, origin.Aliases...), found); err != nil {
		return linked, answered, err
	}
	e.addLinked(ctx, entityID, found, linked)
	return linked, answered, errors.Join(pageErr, searchErr)
}

// creditMentions credits the entity on the uncredited scenes among mediaIDs
// whose title mentions one of names.
func (e *EntityScenes) creditMentions(ctx context.Context, entityID string, names []string, mediaIDs []string) error {
	uncredited, err := e.scheduler.repo.UncreditedScenes(ctx, entityID, mediaIDs)
	if err != nil {
		return err
	}
	var matches []string
	for _, scene := range uncredited {
		if mentionsAnyName(scene.Title, names) {
			matches = append(matches, scene.ID)
		}
	}
	_, err = e.scheduler.repo.CreditScenes(ctx, entityID, matches)
	return err
}

// addLinked adds the scenes among mediaIDs credited to the entity to linked.
func (e *EntityScenes) addLinked(ctx context.Context, entityID string, mediaIDs []string, linked map[string]bool) {
	credited, err := e.scheduler.repo.KeepLinked(ctx, entityID, mediaIDs)
	if err != nil {
		log.Printf("entity scenes %s: %v", entityID, err)
		return
	}
	for _, sceneID := range credited {
		linked[sceneID] = true
	}
}

// scenePlugins keeps the plugins that list scenes (dropping gallery-only ones).
func scenePlugins(candidates []*plugins.Plugin) []*plugins.Plugin {
	var kept []*plugins.Plugin
	for _, plugin := range candidates {
		if plugin.Meta.Has(plugins.CapabilitySceneList) {
			kept = append(kept, plugin)
		}
	}
	return kept
}

// mentionsAnyName reports whether title contains one of names as whole words,
// ignoring case and punctuation ("Riley Reid" matches "riley-reid's day").
func mentionsAnyName(title string, names []string) bool {
	paddedTitle := " " + wordsOnly(title) + " "
	for _, name := range names {
		words := wordsOnly(name)
		if len(words) < minimumMentionLength {
			continue
		}
		if strings.Contains(paddedTitle, " "+words+" ") {
			return true
		}
	}
	return false
}

// wordsOnly lowercases text and reduces it to letter/digit words joined by
// single spaces.
func wordsOnly(text string) string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsDigit(character)
	})
	return strings.Join(fields, " ")
}
