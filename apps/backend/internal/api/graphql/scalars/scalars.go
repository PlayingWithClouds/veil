// Package scalars marshals Go types the schema uses that gqlgen has no
// built-in mapping for.
package scalars

import (
	"github.com/99designs/gqlgen/graphql"
)

// MarshalFloat32 writes an f32 schema field as a GraphQL Float.
func MarshalFloat32(value float32) graphql.Marshaler {
	return graphql.MarshalFloat(float64(value))
}

// UnmarshalFloat32 reads a GraphQL Float into an f32 schema field.
func UnmarshalFloat32(value any) (float32, error) {
	number, err := graphql.UnmarshalFloat(value)
	return float32(number), err
}
