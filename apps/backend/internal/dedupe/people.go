package dedupe

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/schema"
)

// peopleBatch is how many performers and how many studios one pass checks.
const peopleBatch = 150

// cursorKey prefixes the cache entry holding how far the walk through a kind
// got (a createdAt); it starts over at the end.
const cursorKey = "dedupe-cursor:"

// portraitKey prefixes the cache entry holding a performer's or studio's
// picture hash (hex, "0" when it can't be read).
const portraitKey = "portrait:"

// mergePeople walks the next batch of performers and studios and merges each
// into the records sharing a name or alias with it that are provably the same
// (alike.SamePerson, alike.SameStudio). The one credited on more videos
// survives. Weaker matches stay suggestions (performerDuplicates).
func (s *Sweeper) mergePeople(ctx context.Context) error {
	for _, kind := range []schema.RecordKind{schema.RecordKindPerformer, schema.RecordKindStudio} {
		if err := s.mergeKind(ctx, kind); err != nil {
			return err
		}
	}
	return nil
}

// mergeKind checks the next batch of one kind and advances the cursor.
func (s *Sweeper) mergeKind(ctx context.Context, kind schema.RecordKind) error {
	cursor, err := s.cache.Get(ctx, cursorKey+string(kind))
	if errors.Is(err, cache.ErrMiss) {
		cursor, err = "", nil
	}
	if err != nil {
		return err
	}
	ids, stamps, err := s.library.PeopleAfter(ctx, kind, cursor, peopleBatch)
	if err != nil {
		return err
	}
	next := ""
	if len(stamps) == peopleBatch {
		next = stamps[len(stamps)-1]
	}
	merged := map[string]bool{}
	for _, id := range ids {
		if merged[id] {
			continue
		}
		if err := s.mergeDuplicatesOf(ctx, kind, id, merged); err != nil {
			log.Printf("dedupe: %s %s: %v", kind, id, err)
		}
	}
	return s.cache.Set(ctx, cursorKey+string(kind), next, 0)
}

// mergeDuplicatesOf merges the provable duplicates of one record, marking
// the merged ids.
func (s *Sweeper) mergeDuplicatesOf(ctx context.Context, kind schema.RecordKind, id string, merged map[string]bool) error {
	candidates, err := s.duplicateIDs(ctx, kind, id)
	if err != nil {
		return err
	}
	for _, other := range candidates {
		same, err := s.provablySamePerson(ctx, kind, id, other)
		if err != nil {
			return err
		}
		if !same {
			continue
		}
		keep, gone, err := s.survivor(ctx, kind, id, other)
		if err != nil {
			return err
		}
		log.Printf("dedupe: merging %s %s into %s (same name, tied)", kind, gone, keep)
		if err := s.library.Merge(ctx, kind, keep, gone); err != nil {
			return err
		}
		merged[gone] = true
		if gone == id {
			return nil
		}
	}
	return nil
}

// duplicateIDs lists the records sharing a name or alias with one.
func (s *Sweeper) duplicateIDs(ctx context.Context, kind schema.RecordKind, id string) ([]string, error) {
	ids := []string{}
	if kind == schema.RecordKindStudio {
		studios, err := s.library.StudioDuplicates(ctx, id)
		for _, studio := range studios {
			ids = append(ids, studio.Id)
		}
		return ids, err
	}
	performers, err := s.library.PerformerDuplicates(ctx, id)
	for _, performer := range performers {
		ids = append(ids, performer.Id)
	}
	return ids, err
}

// provablySamePerson runs the merge rule of the kind on two records.
func (s *Sweeper) provablySamePerson(ctx context.Context, kind schema.RecordKind, firstID, secondID string) (bool, error) {
	if kind == schema.RecordKindStudio {
		first, err := s.studioFacts(ctx, firstID)
		if err != nil || first == nil {
			return false, err
		}
		second, err := s.studioFacts(ctx, secondID)
		if err != nil || second == nil {
			return false, err
		}
		return alike.SameStudio(*first, *second), nil
	}
	first, err := s.personFacts(ctx, firstID)
	if err != nil || first == nil {
		return false, err
	}
	second, err := s.personFacts(ctx, secondID)
	if err != nil || second == nil {
		return false, err
	}
	shared, err := s.library.SharedVideoCount(ctx, firstID, secondID)
	if err != nil {
		return false, err
	}
	return alike.SamePerson(*first, *second, shared), nil
}

// personFacts reads what tells a performer apart; nil when it's gone.
func (s *Sweeper) personFacts(ctx context.Context, id string) (*alike.PersonFacts, error) {
	performer, err := s.library.Performer(ctx, id)
	if err != nil || performer == nil {
		return nil, err
	}
	facts := &alike.PersonFacts{Birthdate: text(performer.Attributes.Birthdate), Gender: text(performer.Attributes.Gender)}
	if performer.Image != nil {
		facts.PictureHash = s.portraitHash(ctx, id, performer.Image.Uri)
	}
	return facts, nil
}

// studioFacts reads what ties a studio; nil when it's gone.
func (s *Sweeper) studioFacts(ctx context.Context, id string) (*alike.StudioFacts, error) {
	studio, err := s.library.Studio(ctx, id)
	if err != nil || studio == nil {
		return nil, err
	}
	facts := &alike.StudioFacts{URL: text(studio.Url), Parent: text(studio.Parent)}
	if studio.Image != nil {
		facts.PictureHash = s.portraitHash(ctx, id, studio.Image.Uri)
	}
	return facts, nil
}

// portraitHash is the cached hash of a record's picture, hashed on first use.
func (s *Sweeper) portraitHash(ctx context.Context, id, uri string) uint64 {
	key := portraitKey + id + ":" + uri
	stored, err := s.cache.Get(ctx, key)
	if err == nil {
		hash, _ := strconv.ParseUint(stored, 16, 64)
		return hash
	}
	hash := s.hash(ctx, uri)
	if err := s.cache.Set(ctx, key, strconv.FormatUint(hash, 16), 0); err != nil {
		log.Printf("dedupe: remember picture of %s: %v", id, err)
	}
	return hash
}

// survivor picks which of two records survives a merge: the one credited on
// more videos, the first given on a tie.
func (s *Sweeper) survivor(ctx context.Context, kind schema.RecordKind, firstID, secondID string) (string, string, error) {
	firstCount, err := s.library.CreditCount(ctx, kind, firstID)
	if err != nil {
		return "", "", err
	}
	secondCount, err := s.library.CreditCount(ctx, kind, secondID)
	if err != nil {
		return "", "", err
	}
	if secondCount > firstCount {
		return secondID, firstID, nil
	}
	return firstID, secondID, nil
}

// text is a string pointer's value, "" for nil.
func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
