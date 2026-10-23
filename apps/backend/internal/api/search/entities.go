package search

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// entityIndexTTL is how long the index of performer, studio and tag names is
// reused before the library is read again; a name added meanwhile is
// recognised in queries a minute or two later.
const entityIndexTTL = 2 * time.Minute

// entityDirectory keeps the index of the library's performer, studio and tag
// names, rebuilt when it is older than entityIndexTTL.
type entityDirectory struct {
	ingestSvc *ingest.Service
	now       func() time.Time

	mutex    sync.Mutex
	index    *searchrank.EntityIndex
	loadedAt time.Time
}

// newEntityDirectory creates a directory reading names through ingestSvc.
func newEntityDirectory(ingestSvc *ingest.Service) *entityDirectory {
	return &entityDirectory{ingestSvc: ingestSvc, now: time.Now}
}

// Index returns the current name index, rebuilding it when stale. On a failed
// rebuild the old index (possibly nil, which finds nothing) stays in use.
func (directory *entityDirectory) Index(ctx context.Context) *searchrank.EntityIndex {
	directory.mutex.Lock()
	defer directory.mutex.Unlock()
	if directory.index != nil && directory.now().Sub(directory.loadedAt) < entityIndexTTL {
		return directory.index
	}
	index, err := directory.build(ctx)
	if err != nil {
		return directory.index
	}
	directory.index, directory.loadedAt = index, directory.now()
	return index
}

// build reads every performer, studio and tag into a fresh index: names
// first, so no alias takes a spelling that is another record's name.
func (directory *entityDirectory) build(ctx context.Context) (*searchrank.EntityIndex, error) {
	index := searchrank.NewEntityIndex()
	kinds := []searchrank.EntityKind{searchrank.KindPerformer, searchrank.KindStudio, searchrank.KindTag}
	aliasesByID := map[string][]string{}
	for _, kind := range kinds {
		records, err := directory.ingestSvc.NamedRecords(ctx, string(kind))
		if err != nil {
			return nil, fmt.Errorf("entity index: %w", err)
		}
		for _, record := range records {
			id := recordID(record)
			name, _ := record["name"].(string)
			index.AddName(searchrank.Entity{Kind: kind, ID: id, Name: name})
			aliasesByID[id] = stringList(record["aliases"])
		}
	}
	for id, aliases := range aliasesByID {
		for _, alias := range aliases {
			index.AddAlias(id, alias)
		}
	}
	return index, nil
}

// stringList reads a JSON array of strings.
func stringList(value any) []string {
	items, _ := value.([]any)
	list := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			list = append(list, text)
		}
	}
	return list
}
