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

// camAffinity is the cam taste as the cam service ranks rooms by it.
func camAffinity(taste recommend.CamTaste) cams.Affinity {
	return cams.Affinity{
		Tags: taste.Tags, Performers: taste.Performers, PerformerNames: taste.PerformerNames,
		Sites: taste.Sites, Genders: taste.Genders, Rooms: taste.Rooms,
	}
}

// camRoomsForYou lists one page of rooms ranked by the user's taste.
func (r *queryResolver) camRoomsForYou(ctx context.Context, listing cams.Query) (*model.CamRoomPage, error) {
	taste, err := r.recommend.CamTaste(ctx)
	if err != nil {
		return nil, err
	}
	ranked, total, err := r.cams.ForYou(ctx, listing, camAffinity(taste))
	if err != nil {
		return nil, err
	}
	return camRecommendationPage(ranked, total), nil
}

// camRecommendationPage is a page of recommended rooms with their reasons.
func camRecommendationPage(recommendations []cams.Recommendation, total int) *model.CamRoomPage {
	listing := &model.CamRoomPage{Rooms: []*schema.CamModel{}, Total: total, Reasons: []string{}}
	for index := range recommendations {
		listing.Rooms = append(listing.Rooms, &recommendations[index].Room)
		listing.Reasons = append(listing.Reasons, recommendations[index].Reason)
	}
	return listing
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
