package user

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// fieldDefTable is the table custom field definitions live in.
const fieldDefTable = "FieldDef"

// FieldDefs lists the custom fields, all of them or those for kind.
func (s *Service) FieldDefs(ctx context.Context, kind *schema.RecordKind) ([]schema.FieldDef, error) {
	if kind == nil {
		return db.QueryAs[schema.FieldDef](ctx, s.database,
			`SELECT * FROM "FieldDef" WHERE "deletedAt" IS NULL ORDER BY "name" COLLATE NOCASE`, nil)
	}
	return db.QueryAs[schema.FieldDef](ctx, s.database,
		`SELECT * FROM "FieldDef" WHERE "deletedAt" IS NULL
		 AND EXISTS (SELECT 1 FROM json_each("kinds") WHERE value = $kind) ORDER BY "name" COLLATE NOCASE`,
		db.Vars{"kind": *kind})
}

// FieldDef loads one custom field definition.
func (s *Service) FieldDef(ctx context.Context, id string) (*schema.FieldDef, error) {
	return db.Get[schema.FieldDef](ctx, s.database, fieldDefTable, id)
}

// SaveFieldDef creates a custom field (id empty) or changes one. Choice
// fields need options.
func (s *Service) SaveFieldDef(ctx context.Context, id string, definition schema.FieldDef) (*schema.FieldDef, error) {
	if strings.TrimSpace(definition.Name) == "" {
		return nil, fmt.Errorf("a field needs a name")
	}
	if len(definition.Kinds) == 0 {
		return nil, fmt.Errorf("a field needs at least one kind of record")
	}
	isChoice := definition.FieldType == schema.FieldTypeChoice || definition.FieldType == schema.FieldTypeChoices
	if isChoice && len(definition.Options) == 0 {
		return nil, fmt.Errorf("a %s field needs options", definition.FieldType)
	}
	fields := db.Vars{"name": strings.TrimSpace(definition.Name), "kinds": definition.Kinds, "fieldType": definition.FieldType, "options": definition.Options}
	if id == "" {
		created, err := s.database.Insert(ctx, fieldDefTable, fields)
		if err != nil {
			return nil, err
		}
		return s.FieldDef(ctx, created)
	}
	if err := s.database.Update(ctx, fieldDefTable, id, fields); err != nil {
		return nil, err
	}
	return s.FieldDef(ctx, id)
}

// DeleteFieldDef trashes a custom field. Values already set stay in the
// user states until the field is purged.
func (s *Service) DeleteFieldDef(ctx context.Context, id string) error {
	return s.database.Trash(ctx, fieldDefTable, id)
}

// SetFieldValue sets a custom field's value on a record, checked against the
// field's type; nil clears it.
func (s *Service) SetFieldValue(ctx context.Context, kind schema.RecordKind, id, fieldDefID string, value any) (*schema.UserState, error) {
	definition, err := s.FieldDef(ctx, fieldDefID)
	if err != nil {
		return nil, err
	}
	if definition == nil {
		return nil, fmt.Errorf("no custom field %s", fieldDefID)
	}
	if !slices.Contains(definition.Kinds, kind) {
		return nil, fmt.Errorf("%s isn't a field of a %s", definition.Name, kind)
	}
	if err := checkFieldValue(*definition, value); err != nil {
		return nil, err
	}
	path := `$."` + fieldDefID + `"`
	err = s.database.Tx(ctx, func(tx *db.DB) error {
		stateID, err := ensureState(ctx, tx, kind, id)
		if err != nil {
			return err
		}
		vars := db.Vars{"id": stateID, "path": path, "value": encodeValue(value), "now": db.Now()}
		if value == nil {
			_, err = tx.Exec(ctx, `UPDATE "UserState" SET "fields" = json_remove("fields", $path), "updatedAt" = $now WHERE "id" = $id`, vars)
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE "UserState" SET "fields" = json_set("fields", $path, json($value)), "updatedAt" = $now WHERE "id" = $id`, vars)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, kind, id)
}

// encodeValue renders a field value as JSON text for json_set.
func encodeValue(value any) string {
	encoded := db.StoredValue([]any{value})
	text, _ := encoded.(string)
	return strings.TrimSuffix(strings.TrimPrefix(text, "["), "]")
}

// isoDate matches a YYYY-MM-DD date.
var isoDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// checkFieldValue fails when value doesn't fit the field's type.
func checkFieldValue(definition schema.FieldDef, value any) error {
	if value == nil {
		return nil
	}
	if fieldValueFits(definition, value) {
		return nil
	}
	return fmt.Errorf("%v isn't a valid %s value for %s", value, definition.FieldType, definition.Name)
}

// fieldValueFits reports whether a non-nil value fits the field's type.
func fieldValueFits(definition schema.FieldDef, value any) bool {
	text, isText := value.(string)
	switch definition.FieldType {
	case schema.FieldTypeText:
		return isText
	case schema.FieldTypeNumber:
		_, isNumber := asNumber(value)
		return isNumber
	case schema.FieldTypeBoolean:
		_, isBool := value.(bool)
		return isBool
	case schema.FieldTypeDate:
		return isText && isoDate.MatchString(text)
	case schema.FieldTypeUrl:
		parsed, err := url.Parse(text)
		return isText && err == nil && parsed.Scheme != "" && parsed.Host != ""
	case schema.FieldTypeChoice:
		return isText && slices.Contains(definition.Options, text)
	case schema.FieldTypeChoices:
		return allChoices(definition.Options, value)
	}
	return false
}

// allChoices reports whether value is a list of the given options.
func allChoices(options []string, value any) bool {
	list, isList := value.([]any)
	if !isList {
		return false
	}
	for _, element := range list {
		text, isText := element.(string)
		if !isText || !slices.Contains(options, text) {
			return false
		}
	}
	return true
}
