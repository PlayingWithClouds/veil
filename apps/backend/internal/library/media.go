package library

import (
	"context"
	"fmt"
	"maps"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
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
	// listedCredits gives the performers a listing card names, taken only
	// while the record credits nobody: the page read owns the credits.
	listedCredits func(context.Context) (db.Vars, error)
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

// listedCredits credits a listing card's performers to a record that
// credits nobody yet (see mediaWrite.listedCredits); nil for a page read or
// a card naming nobody.
func (l *Library) listedCredits(plugin string, credits []plugins.ScenePerformer, fetched bool) func(context.Context) (db.Vars, error) {
	if fetched || len(credits) == 0 {
		return nil
	}
	return func(ctx context.Context) (db.Vars, error) {
		ids, err := l.performerIDs(ctx, plugin, credits)
		return db.Vars{"performers": ids}, err
	}
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
	if err := l.creditListedIfUncredited(ctx, &write, source.SubjectId); err != nil {
		return "", "", err
	}
	_, err = l.recorder.Update(ctx, write.kind, source.SubjectId, write.fields, provenance.Plugin(write.plugin, source.Id))
	return source.SubjectId, source.Id, err
}

// creditListedIfUncredited adds a listing card's credits to the write when
// the record it updates credits nobody yet.
func (l *Library) creditListedIfUncredited(ctx context.Context, write *mediaWrite, id string) error {
	if write.listedCredits == nil {
		return nil
	}
	uncredited, err := l.creditsNobody(ctx, write.kind, id)
	if err != nil || !uncredited {
		return err
	}
	return addListedCredits(ctx, write)
}

// creditsNobody reports whether a media record has no performers.
func (l *Library) creditsNobody(ctx context.Context, kind schema.RecordKind, id string) (bool, error) {
	counts, err := l.database.Strings(ctx, `SELECT CAST(json_array_length("performers") AS TEXT) FROM `+db.Quote(tableFor(kind))+` WHERE "id" = $id`, db.Vars{"id": id})
	if err != nil {
		return false, err
	}
	return len(counts) == 1 && counts[0] == "0", nil
}

// addListedCredits adds the listing card's credits to the write's fields.
func addListedCredits(ctx context.Context, write *mediaWrite) error {
	if write.listedCredits == nil {
		return nil
	}
	credits, err := write.listedCredits(ctx)
	if err != nil {
		return err
	}
	write.fields = maps.Clone(write.fields)
	maps.Copy(write.fields, credits)
	return nil
}

// createMedia inserts a new media record and its first source, which owns
// the fields it filled.
func (l *Library) createMedia(ctx context.Context, write mediaWrite) (string, string, error) {
	if err := addListedCredits(ctx, &write); err != nil {
		return "", "", err
	}
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
