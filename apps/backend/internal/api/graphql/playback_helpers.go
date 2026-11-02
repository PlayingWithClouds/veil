package graphql

import (
	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/schema"
)

// segmentEdit maps the API's segment input to the library's.
func segmentEdit(input model.SegmentInput) library.SegmentEdit {
	return library.SegmentEdit{
		StartSeconds: input.StartSeconds,
		EndSeconds:   input.EndSeconds,
		ClearEnd:     input.ClearEnd != nil && *input.ClearEnd,
		Title:        input.Title,
		TagID:        input.TagID,
		PerformerIDs: optionalList(input.PerformerIds),
	}
}

// clipEdit maps the API's clip input to the library's.
func clipEdit(input model.ClipInput) library.ClipEdit {
	return library.ClipEdit{StartSeconds: input.StartSeconds, EndSeconds: input.EndSeconds, Title: input.Title, CoverURI: input.CoverURI}
}

// trickplayOf maps the API's trickplay input to the schema's; nil stays nil.
func trickplayOf(input *model.TrickplayInput) *schema.Trickplay {
	if input == nil {
		return nil
	}
	sheets := make([]schema.Asset, 0, len(input.SheetUris))
	for _, uri := range input.SheetUris {
		sheets = append(sheets, schema.Asset{Uri: uri})
	}
	return &schema.Trickplay{
		Sheets:          sheets,
		Columns:         int32(input.Columns),
		Rows:            int32(input.Rows),
		TileWidth:       int32(input.TileWidth),
		TileHeight:      int32(input.TileHeight),
		IntervalSeconds: float32(input.IntervalSeconds),
	}
}
