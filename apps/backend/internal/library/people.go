package library

import (
	"context"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// upsertPerformer stores a performer a plugin described: matched by its site
// identity, else by name, else created. Returns "" for a nameless one.
func (l *Library) upsertPerformer(ctx context.Context, plugin string, performer *plugins.Performer) (string, error) {
	id, err := l.namedRecordID(ctx, schema.RecordKindPerformer, tablePerformer, plugin, performer.Name, performer.ExternalID, performer.SourceURL, performer.ImagePath)
	if err != nil || id == "" {
		return id, err
	}
	return id, l.database.Update(ctx, tablePerformer, id, performerFields(performer))
}

// upsertStudio stores a studio a plugin described, like upsertPerformer.
func (l *Library) upsertStudio(ctx context.Context, plugin string, studio *plugins.Studio) (string, error) {
	id, err := l.namedRecordID(ctx, schema.RecordKindStudio, tableStudio, plugin, studio.Name, studio.ExternalID, studio.SourceURL, studio.ImagePath)
	if err != nil || id == "" {
		return id, err
	}
	fields := db.Vars{}
	setText(fields, "details", studio.Details)
	setText(fields, "url", studio.URL)
	setAsset(fields, "image", studio.ImagePath)
	if len(studio.Aliases) > 0 {
		fields["aliases"] = studio.Aliases
	}
	return id, l.database.Update(ctx, tableStudio, id, fields)
}

// upsertTag stores a tag by name and keeps the aliases a plugin gives.
func (l *Library) upsertTag(ctx context.Context, tag *plugins.Tag) (string, error) {
	id, err := l.findOrCreateNamed(ctx, tableTag, tag.Name)
	if err != nil || id == "" || len(tag.Aliases) == 0 {
		return id, err
	}
	return id, l.database.Update(ctx, tableTag, id, db.Vars{"aliases": tag.Aliases})
}

// namedRecordID returns the performer or studio a plugin record means: the
// one its site identity already points at, else the one with its name
// (which then gains the source), else a new one.
func (l *Library) namedRecordID(ctx context.Context, kind schema.RecordKind, table, plugin, name, externalID, url, image string) (string, error) {
	site, hasSite := identity(plugin, externalID, url)
	if hasSite {
		source, err := l.findSource(ctx, site)
		if err != nil || source != nil {
			return sourceSubject(source), err
		}
	}
	id, err := l.findOrCreateNamed(ctx, table, name)
	if err != nil || id == "" || !hasSite {
		return id, err
	}
	_, err = l.addSource(ctx, kind, id, site, name, assetOf(image))
	return id, err
}

// sourceSubject is the id a source points at; "" for none.
func sourceSubject(source *schema.Source) string {
	if source == nil {
		return ""
	}
	return source.SubjectId
}

// performerFields are the columns a performer page fills. The name stays as
// first stored; a different spelling becomes an alias.
func performerFields(performer *plugins.Performer) db.Vars {
	fields := db.Vars{}
	setText(fields, "details", performer.Details)
	setAsset(fields, "image", performer.ImagePath)
	if len(performer.Aliases) > 0 {
		fields["aliases"] = performer.Aliases
	}
	if links := performerLinks(performer); len(links) > 0 {
		fields["links"] = links
	}
	if attributes := performerAttributes(performer); attributes != nil {
		fields["attributes"] = attributes
	}
	return fields
}

// performerLinks collects a performer's own sites and socials.
func performerLinks(performer *plugins.Performer) []string {
	links := []string{}
	for _, link := range []string{performer.URL, performer.Twitter, performer.Instagram} {
		if link != "" {
			links = append(links, link)
		}
	}
	return links
}

// performerAttributes maps a plugin's body and career facts; nil when the
// page gave none.
func performerAttributes(performer *plugins.Performer) *schema.PerformerAttributes {
	attributes := schema.PerformerAttributes{
		Gender:       optional(performer.Gender),
		Birthdate:    optional(isoDate(performer.Birthdate)),
		DeathDate:    optional(isoDate(performer.DeathDate)),
		Country:      optional(performer.Country),
		Ethnicity:    optional(performer.Ethnicity),
		EyeColor:     optional(performer.EyeColor),
		HairColor:    optional(performer.HairColor),
		Measurements: optional(performer.Measurements),
		FakeTits:     optional(performer.FakeTits),
		Tattoos:      optional(performer.Tattoos),
		Piercings:    optional(performer.Piercings),
		HeightCm:     optionalCount(performer.HeightCm),
		WeightKg:     optionalCount(performer.WeightKg),
	}
	if attributes == (schema.PerformerAttributes{}) {
		return nil
	}
	return &attributes
}

// optionalCount returns nil for zero, else a pointer to the value.
func optionalCount(value int) *int32 {
	if value <= 0 {
		return nil
	}
	count := int32(value)
	return &count
}
