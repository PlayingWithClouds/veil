package files

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// templatePlaceholders are the placeholders an organize template may use.
var templatePlaceholders = map[string]bool{"studio": true, "date": true, "title": true, "performers": true, "ext": true}

// placeholderPattern finds {name} placeholders.
var placeholderPattern = regexp.MustCompile(`\{([A-Za-z]+)\}`)

// unsafeCharacters can't appear in file names on every platform.
var unsafeCharacters = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

// spaceRuns matches runs of whitespace.
var spaceRuns = regexp.MustCompile(`\s+`)

// Move is one planned rename of a library file; paths are relative to the
// root and slash-separated.
type Move struct {
	FileID string
	From   string
	To     string
}

// validateTemplate rejects templates with unknown placeholders or that
// climb out of the root.
func validateTemplate(template string) error {
	for _, match := range placeholderPattern.FindAllStringSubmatch(template, -1) {
		if !templatePlaceholders[match[1]] {
			return fmt.Errorf("unknown placeholder {%s}; use {studio}, {date}, {title}, {performers} or {ext}", match[1])
		}
	}
	for _, segment := range strings.Split(template, "/") {
		if segment == ".." {
			return fmt.Errorf("an organize template stays inside its root")
		}
	}
	return nil
}

// PlanOrganize works out where the root's template puts each of its video
// and audio files, without moving anything. Files already in place are left
// out; clashing targets get a " (2)", " (3)" … suffix.
func (s *Service) PlanOrganize(ctx context.Context, rootID string) ([]Move, error) {
	opened, err := s.openRoot(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if opened.root.OrganizeTemplate == nil {
		return nil, fmt.Errorf("library root %s has no organize template", rootID)
	}
	files, err := s.organizableFiles(ctx, rootID)
	if err != nil {
		return nil, err
	}
	taken := map[string]bool{}
	moves := []Move{}
	for _, file := range files {
		values, err := s.templateValues(ctx, file)
		if err != nil {
			return nil, err
		}
		from := opened.rootRelative(file.Path)
		to := renderTemplate(*opened.root.OrganizeTemplate, values)
		if to == from {
			taken[to] = true
			continue
		}
		to = freeTarget(opened, to, taken)
		taken[to] = true
		moves = append(moves, Move{FileID: file.Id, From: from, To: to})
	}
	return moves, nil
}

// organizableFiles lists the present video and audio files of a root.
func (s *Service) organizableFiles(ctx context.Context, rootID string) ([]schema.MediaFile, error) {
	return db.QueryAs[schema.MediaFile](ctx, s.database,
		`SELECT * FROM "MediaFile" WHERE "libraryRoot" = $root AND "deletedAt" IS NULL AND "missing" = 0
		 AND "subjectKind" IN ('video', 'audio') ORDER BY "path"`, db.Vars{"root": rootID})
}

// templateValues are the placeholder values of a file's record.
func (s *Service) templateValues(ctx context.Context, file schema.MediaFile) (map[string]string, error) {
	media, err := db.QueryOneAs[schema.Media](ctx, s.database,
		`SELECT * FROM `+db.Quote(mediaTable(file.SubjectKind))+` WHERE "id" = $id`, db.Vars{"id": file.SubjectId})
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, fmt.Errorf("file %s has no record", file.Id)
	}
	studio, err := s.studioName(ctx, media.Studio)
	if err != nil {
		return nil, err
	}
	performers, err := s.database.Strings(ctx,
		`SELECT "Performer"."name" FROM json_each($ids) JOIN "Performer" ON "Performer"."id" = json_each.value
		 ORDER BY json_each.key`, db.Vars{"ids": media.Performers})
	if err != nil {
		return nil, err
	}
	values := map[string]string{
		"title":      media.Title,
		"studio":     studio,
		"performers": strings.Join(performers, ", "),
		"ext":        strings.TrimPrefix(strings.ToLower(path.Ext(file.Path)), "."),
	}
	if media.Date != nil {
		values["date"] = *media.Date
	}
	return values, nil
}

// studioName is the name of the studio id points at; "" for none.
func (s *Service) studioName(ctx context.Context, studioID *string) (string, error) {
	if studioID == nil {
		return "", nil
	}
	names, err := s.database.Strings(ctx, `SELECT "name" FROM "Studio" WHERE "id" = $id`, db.Vars{"id": *studioID})
	if err != nil || len(names) == 0 {
		return "", err
	}
	return names[0], nil
}

// renderTemplate fills a template with filename-safe values and tidies the
// result: empty folders dropped, spaces collapsed, and the extension added
// when the template has none.
func renderTemplate(template string, values map[string]string) string {
	filled := placeholderPattern.ReplaceAllStringFunc(template, func(placeholder string) string {
		return safeName(values[strings.Trim(placeholder, "{}")])
	})
	segments := []string{}
	for _, segment := range strings.Split(filled, "/") {
		segment = tidySegment(segment)
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	rendered := strings.Join(segments, "/")
	extension := values["ext"]
	if extension != "" && !strings.HasSuffix(strings.ToLower(rendered), "."+extension) {
		rendered += "." + extension
	}
	return rendered
}

// safeName replaces the characters a file name can't hold.
func safeName(value string) string {
	return unsafeCharacters.ReplaceAllString(value, "_")
}

// tidySegment collapses whitespace and trims spaces and dots a path segment
// can't start or end with; separators left by empty placeholders go too.
func tidySegment(segment string) string {
	segment = spaceRuns.ReplaceAllString(segment, " ")
	return strings.Trim(segment, " .-_")
}

// freeTarget is to, or to with a numbered suffix when another planned move
// or an existing file has it.
func freeTarget(opened *openedRoot, to string, taken map[string]bool) string {
	extension := path.Ext(to)
	stem := strings.TrimSuffix(to, extension)
	candidate := to
	for number := 2; isTaken(opened, candidate, taken); number++ {
		candidate = fmt.Sprintf("%s (%d)%s", stem, number, extension)
	}
	return candidate
}

// isTaken reports whether a root-relative path is planned or on disk.
func isTaken(opened *openedRoot, rootRelative string, taken map[string]bool) bool {
	if taken[rootRelative] {
		return true
	}
	target, err := insidePath(opened.directory, rootRelative)
	return err != nil || pathExists(target)
}

// ApplyOrganize carries out planned moves of a root's files and returns how
// many it moved. It never overwrites: a move whose target exists, or whose
// file is no longer where the plan found it, stops with an error.
func (s *Service) ApplyOrganize(ctx context.Context, rootID string, moves []Move) (int, error) {
	opened, err := s.openRoot(ctx, rootID)
	if err != nil {
		return 0, err
	}
	for index, move := range moves {
		if err := s.applyMove(ctx, opened, move); err != nil {
			return index, err
		}
	}
	return len(moves), nil
}

// applyMove renames one file on disk and in its row.
func (s *Service) applyMove(ctx context.Context, opened *openedRoot, move Move) error {
	file, err := s.File(ctx, move.FileID)
	if err != nil {
		return err
	}
	if file.LibraryRoot == nil || *file.LibraryRoot != opened.root.Id || opened.rootRelative(file.Path) != move.From {
		return fmt.Errorf("file %s is no longer at %s", move.FileID, move.From)
	}
	source, err := opened.absolute(file.Path)
	if err != nil {
		return err
	}
	target, err := insidePath(opened.directory, move.To)
	if err != nil {
		return err
	}
	if pathExists(target) {
		return fmt.Errorf("%s already exists", move.To)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		return err
	}
	newPath := opened.locationPath(move.To)
	return s.database.Tx(ctx, func(tx *db.DB) error {
		if err := clearTrashedRow(ctx, tx, file.Location, newPath); err != nil {
			return err
		}
		return tx.Update(ctx, tableMediaFile, file.Id, db.Vars{"path": newPath})
	})
}
