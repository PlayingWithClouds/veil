package graphql

import (
	"context"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/schema"
)

// categoryEdit maps the API's category input to the library's.
func categoryEdit(input model.CategoryInput) library.CategoryEdit {
	return library.CategoryEdit{
		Name:        input.Name,
		Slug:        input.Slug,
		ParentID:    input.ParentID,
		MatchTagIDs: optionalList(input.MatchTagIds),
		CoverURI:    input.CoverURI,
		Position:    input.Position,
	}
}

// childCategories lists the categories directly inside parentID.
func (r *Resolver) childCategories(ctx context.Context, parentID string) ([]*schema.Category, error) {
	categories, err := r.library.Categories(ctx)
	if err != nil {
		return nil, err
	}
	children := []*schema.Category{}
	for index := range categories {
		if categories[index].Parent != nil && *categories[index].Parent == parentID {
			children = append(children, &categories[index])
		}
	}
	return children, nil
}
