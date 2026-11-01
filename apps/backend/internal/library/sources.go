package library

import (
	"context"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// siteIdentity names a record on one site. ExternalID falls back to the URL
// for plugins that give no id.
type siteIdentity struct {
	Plugin     string
	ExternalID string
	URL        string
}

// identity builds the site identity of a plugin record; ok is false when the
// record has neither an id nor a URL.
func identity(plugin, externalID, url string) (siteIdentity, bool) {
	if externalID == "" {
		externalID = url
	}
	if externalID == "" {
		return siteIdentity{}, false
	}
	return siteIdentity{Plugin: plugin, ExternalID: externalID, URL: url}, true
}

// findSource returns the source row for a site identity, or nil.
func (l *Library) findSource(ctx context.Context, site siteIdentity) (*schema.Source, error) {
	return db.QueryOneAs[schema.Source](ctx, l.database,
		`SELECT * FROM "Source" WHERE "plugin" = $plugin AND "externalId" = $externalId AND "deletedAt" IS NULL`,
		db.Vars{"plugin": site.Plugin, "externalId": site.ExternalID})
}

// addSource records that the record kind/id was found at site.
func (l *Library) addSource(ctx context.Context, kind schema.RecordKind, id string, site siteIdentity, title string, image *schema.Asset) (string, error) {
	source := schema.Source{
		SubjectKind: kind,
		SubjectId:   id,
		Plugin:      site.Plugin,
		ExternalId:  site.ExternalID,
		Url:         site.URL,
		Title:       optional(title),
		Image:       image,
	}
	return l.database.Insert(ctx, tableSource, source)
}

// markFetched notes that the full page behind a source was read, so visits
// don't fetch it again.
func (l *Library) markFetched(ctx context.Context, sourceID string) error {
	return l.database.Update(ctx, tableSource, sourceID, db.Vars{"fetchedAt": db.Now()})
}

// Sources lists every site a record was found on, oldest first.
func (l *Library) Sources(ctx context.Context, kind schema.RecordKind, id string) ([]schema.Source, error) {
	return db.QueryAs[schema.Source](ctx, l.database,
		`SELECT * FROM "Source" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL ORDER BY "createdAt"`,
		db.Vars{"kind": string(kind), "id": id})
}

// optional returns nil for an empty string, else a pointer to it.
func optional(text string) *string {
	if text == "" {
		return nil
	}
	return &text
}
