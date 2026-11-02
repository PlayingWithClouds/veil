package recommend

import (
	"context"
	"fmt"
	"math"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// durationBucketRatio is the runtime ratio one duration bucket spans. It
// matches alike's same-scene tolerance (0.95), so two copies alike.SameScene
// can match always sit in the same or a neighbouring bucket.
const durationBucketRatio = 0.95

// duplicates decides which items are the same media on different sites. It
// only hides copies, never merges them, so it may use the loose
// alike.SameScene; pairs recorded in "Alike" (and not dismissed) count too.
type duplicates struct {
	identities map[schema.RecordRef]alike.Candidate
	pairs      map[schema.RecordRef][]schema.RecordRef
	// kept indexes the items kept so far by kind and identity bucket, so an
	// item is only compared with the kept items it could match.
	kept    map[identityBucket][]*mediaFeatures
	keptRef map[schema.RecordRef]bool
}

// identityBucket groups items that may be copies: videos by runtime band,
// galleries by image count.
type identityBucket struct {
	kind   schema.RecordKind
	bucket int
}

// bucketOf is the item's identity bucket; false when it has too little to be
// matched by content (unknown runtime, no images).
func bucketOf(item *mediaFeatures) (identityBucket, bool) {
	if item.ref.Kind == schema.RecordKindGallery {
		return identityBucket{kind: item.ref.Kind, bucket: item.imageCount}, item.imageCount > 0
	}
	if item.durationSeconds <= 0 {
		return identityBucket{}, false
	}
	band := int(math.Log(float64(item.durationSeconds)) / math.Log(1/durationBucketRatio))
	return identityBucket{kind: item.ref.Kind, bucket: band}, true
}

// neighbours are the buckets whose items item may copy.
func neighbours(item *mediaFeatures, bucket identityBucket) []identityBucket {
	if item.ref.Kind == schema.RecordKindGallery {
		return []identityBucket{bucket}
	}
	return []identityBucket{
		{kind: bucket.kind, bucket: bucket.bucket - 1},
		bucket,
		{kind: bucket.kind, bucket: bucket.bucket + 1},
	}
}

// sameAs reports whether item is a copy of other by content.
func (found *duplicates) sameAs(item, other *mediaFeatures) bool {
	if item.ref.Kind == schema.RecordKindGallery {
		return alike.SameGallery(found.identities[item.ref], found.identities[other.ref], item.imageCount, other.imageCount)
	}
	return alike.SameScene(found.identities[item.ref], found.identities[other.ref])
}

// copiesKept reports whether item is a copy of an item kept before.
func (found *duplicates) copiesKept(item *mediaFeatures) bool {
	for _, partner := range found.pairs[item.ref] {
		if found.keptRef[partner] {
			return true
		}
	}
	bucket, matchable := bucketOf(item)
	if !matchable {
		return false
	}
	for _, neighbour := range neighbours(item, bucket) {
		if found.copiesAny(item, found.kept[neighbour]) {
			return true
		}
	}
	return false
}

// copiesAny reports whether item is a copy of one of others.
func (found *duplicates) copiesAny(item *mediaFeatures, others []*mediaFeatures) bool {
	for _, other := range others {
		if found.sameAs(item, other) {
			return true
		}
	}
	return false
}

// keep records item as kept.
func (found *duplicates) keep(item *mediaFeatures) {
	found.keptRef[item.ref] = true
	if bucket, matchable := bucketOf(item); matchable {
		found.kept[bucket] = append(found.kept[bucket], item)
	}
}

// loadDuplicates prepares the matcher for the items: their identities, with
// the performers a title names for items crediting no one (which can only
// keep two copies apart, never join them), and their recorded Alike pairs.
func (e *Engine) loadDuplicates(ctx context.Context, items []*mediaFeatures) (*duplicates, error) {
	found := &duplicates{
		identities: map[schema.RecordRef]alike.Candidate{},
		pairs:      map[schema.RecordRef][]schema.RecordRef{},
		kept:       map[identityBucket][]*mediaFeatures{},
		keptRef:    map[schema.RecordRef]bool{},
	}
	finder, err := e.finderFor(ctx, items)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		identity := alike.Candidate{Title: item.title, Performers: item.performers, DurationSeconds: item.durationSeconds, Date: item.date}
		if len(item.performers) == 0 {
			identity.Named = finder.Find(item.title)
		}
		found.identities[item.ref] = identity
	}
	return found, e.loadAlikePairs(ctx, items, found)
}

// finderFor loads the performer finder when some item credits no one; an
// empty finder otherwise.
func (e *Engine) finderFor(ctx context.Context, items []*mediaFeatures) (*names.Finder, error) {
	for _, item := range items {
		if len(item.performers) == 0 {
			return e.performerFinder(ctx)
		}
	}
	return names.NewFinder(), nil
}

// loadAlikePairs reads the recorded, undismissed Alike pairs among the items.
func (e *Engine) loadAlikePairs(ctx context.Context, items []*mediaFeatures, found *duplicates) error {
	if len(items) == 0 {
		return nil
	}
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, refKey(item.ref))
	}
	rows, err := e.database.Query(ctx,
		`SELECT "subjectKind", "firstId", "secondId" FROM "Alike"
		 WHERE "dismissed" = 0 AND "deletedAt" IS NULL
		   AND "subjectKind" || ':' || "firstId" IN (SELECT value FROM json_each($keys))
		   AND "subjectKind" || ':' || "secondId" IN (SELECT value FROM json_each($keys))`,
		db.Vars{"keys": keys})
	if err != nil {
		return fmt.Errorf("read alike pairs: %w", err)
	}
	for _, row := range rows {
		first, second := rowRef(row, "subjectKind", "firstId"), rowRef(row, "subjectKind", "secondId")
		found.pairs[first] = append(found.pairs[first], second)
		found.pairs[second] = append(found.pairs[second], first)
	}
	return nil
}

// performerFinder finds the performers a title names, by their names and
// aliases.
func (e *Engine) performerFinder(ctx context.Context) (*names.Finder, error) {
	rows, err := e.database.Query(ctx, `SELECT "id", "name", "aliases" FROM "Performer" WHERE "deletedAt" IS NULL`, nil)
	if err != nil {
		return nil, fmt.Errorf("load performer names: %w", err)
	}
	finder := names.NewFinder()
	for _, row := range rows {
		id := rowString(row, "id")
		finder.Add(id, rowString(row, "name"))
		for _, alias := range rowStrings(row, "aliases") {
			finder.Add(id, alias)
		}
	}
	return finder, nil
}

// dropDuplicates keeps one item per media that several sites carry. scored is
// best first, so the copy that ranks highest for this user stays.
func (e *Engine) dropDuplicates(ctx context.Context, scored []scoredItem) ([]scoredItem, error) {
	items := make([]*mediaFeatures, 0, len(scored))
	for _, item := range scored {
		items = append(items, item.features)
	}
	found, err := e.loadDuplicates(ctx, items)
	if err != nil {
		return nil, err
	}
	kept := make([]scoredItem, 0, len(scored))
	for _, item := range scored {
		if found.copiesKept(item.features) {
			continue
		}
		found.keep(item.features)
		kept = append(kept, item)
	}
	return kept, nil
}
