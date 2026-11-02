package library

import (
	"context"
	"fmt"
	"maps"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// mediaWrite is one plugin record of a media kind about to be stored.
type mediaWrite struct {
	kind   schema.RecordKind
	plugin string
	site   siteIdentity
	fields db.Vars
	// defaults are columns a new row needs that the plugin didn't give.
	defaults db.Vars
	// label and image describe the source, so it can be shown as evidence.
	label string
	image *schema.Asset
}

// identityOf builds a media record's site identity or fails for one with
// neither id nor URL.
func identityOf(kind schema.RecordKind, plugin, externalID, url string) (siteIdentity, error) {
	site, ok := identity(plugin, externalID, url)
	if !ok {
		return site, fmt.Errorf("%s returned a %s without id or URL", plugin, kind)
	}
	return site, nil
}

// upsertMedia stores a plugin's media record: a new record for a site
// identity seen the first time, else an update of the record it points at
// (through the recorder, so only fields the source owns change). It returns
// the record and source ids.
func (l *Library) upsertMedia(ctx context.Context, write mediaWrite) (string, string, error) {
	source, err := l.findSource(ctx, write.site)
	if err != nil {
		return "", "", err
	}
	if source == nil {
		return l.createMedia(ctx, write)
	}
	_, err = l.recorder.Update(ctx, write.kind, source.SubjectId, write.fields, provenance.Plugin(write.plugin, source.Id))
	return source.SubjectId, source.Id, err
}

// createMedia inserts a new media record and its first source, which owns
// the fields it filled.
func (l *Library) createMedia(ctx context.Context, write mediaWrite) (string, string, error) {
	filled := columnsOf(write.fields)
	row := db.Vars{"status": string(schema.MediaStatusAvailable), "title": write.site.URL}
	fillListDefaults(row, "performers", "tags")
	maps.Copy(row, write.defaults)
	maps.Copy(row, write.fields)
	id, err := l.database.Insert(ctx, tableFor(write.kind), row)
	if err != nil {
		return "", "", err
	}
	sourceID, err := l.addSource(ctx, write.kind, id, write.site, write.label, write.image)
	if err != nil {
		return "", "", err
	}
	return id, sourceID, l.recorder.Claim(ctx, write.kind, id, filled, provenance.Plugin(write.plugin, sourceID))
}
