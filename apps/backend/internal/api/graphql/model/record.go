package model

// Record is a record of any stored kind (the GraphQL union Record): a
// *schema.Video, *schema.Gallery, *schema.Performer, *schema.Studio or
// *schema.Tag.
type Record interface{}
