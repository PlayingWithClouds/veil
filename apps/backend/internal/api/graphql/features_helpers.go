package graphql

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/cams"
	"github.com/playingwithclouds/veil/internal/lookup"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/schema"
)

// related lists the records related to one, 12 by default.
func (r *Resolver) related(ctx context.Context, subject schema.RecordRef, limit *int) ([]*schema.Match, error) {
	size := 12
	if limit != nil && *limit > 0 {
		size, _ = page(limit, nil)
	}
	matches, err := r.recommend.Related(ctx, subject, size)
	return pointers(matches), err
}

// camQuery maps the API's cam query to the cam service's.
func camQuery(query *model.CamQuery) cams.Query {
	if query == nil {
		return cams.Query{}
	}
	camQuery := cams.Query{
		Plugins: query.Plugins, Text: text(query.Text), Tag: text(query.Tag), Gender: text(query.Gender),
		Limit: number(query.Limit), Offset: number(query.Offset),
	}
	if query.Sort != nil {
		camQuery.Sort = string(*query.Sort)
	}
	return camQuery
}

// camAffinity is the taste profile's tags (by name) and performers, as the
// cam service ranks rooms by them.
func camAffinity(taste recommend.Taste) cams.Affinity {
	affinity := cams.Affinity{Tags: map[string]float64{}, Performers: map[string]float64{}}
	for _, entry := range taste.Tags {
		affinity.Tags[entry.Name] = entry.Weight
	}
	for _, entry := range taste.Performers {
		affinity.Performers[entry.Id] = entry.Weight
	}
	return affinity
}

// lookupQuery maps the API's lookup input, decoding the picture.
func lookupQuery(input model.LookupInput) (lookup.Query, error) {
	query := lookup.Query{Text: text(input.Text), MimeType: text(input.MimeType), Plugins: input.Plugins}
	if input.ImageBase64 == nil || *input.ImageBase64 == "" {
		return query, nil
	}
	image, err := base64.StdEncoding.DecodeString(*input.ImageBase64)
	if err != nil {
		return query, fmt.Errorf("the picture isn't base64: %w", err)
	}
	query.Image = image
	return query, nil
}
