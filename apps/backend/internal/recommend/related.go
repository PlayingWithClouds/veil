package recommend

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// relatedItem collects why one item is related to the target.
type relatedItem struct {
	ref        schema.RecordRef
	score      float64
	companion  string
	performers []string
	studio     string
	tags       []string
	similar    string
}

// relatedSet gathers the related items by reference.
type relatedSet map[schema.RecordRef]*relatedItem

// entry returns the item's entry, creating it.
func (set relatedSet) entry(ref schema.RecordRef) *relatedItem {
	found, exists := set[ref]
	if !exists {
		found = &relatedItem{ref: ref}
		set[ref] = found
	}
	return found
}

// Related lists the videos and galleries related to item, best first: its
// gallery or video (linked, or likely shot together), items sharing its
// performers, studio or rare tags, and items whose embeddings are close.
// Nothing for other kinds, or when the item is trashed or blocked.
func (e *Engine) Related(ctx context.Context, item schema.RecordRef, limit int) ([]schema.Match, error) {
	if !isMediaKind(item.Kind) || limit <= 0 {
		return []schema.Match{}, nil
	}
	loaded, err := e.loadFeatures(ctx, []schema.RecordRef{item})
	if err != nil {
		return nil, err
	}
	target, found := loaded[item]
	if !found {
		return []schema.Match{}, nil
	}
	set, err := e.gatherRelated(ctx, target, limit)
	if err != nil {
		return nil, err
	}
	ranked, err := e.rankRelated(ctx, target, set)
	if err != nil {
		return nil, err
	}
	return e.relatedMatches(ctx, ranked[:min(limit, len(ranked))])
}

// gatherRelated runs every relation and collects what it found.
func (e *Engine) gatherRelated(ctx context.Context, target *mediaFeatures, limit int) (relatedSet, error) {
	stats, err := e.libraryStats(ctx)
	if err != nil {
		return nil, err
	}
	set := relatedSet{}
	steps := []func() error{
		func() error { return e.addCompanions(ctx, target, set) },
		func() error { return e.addSharedCredits(ctx, target, stats.rarity, set) },
		func() error { return e.addSimilar(ctx, target, limit, set) },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, err
		}
	}
	delete(set, target.ref)
	return set, nil
}

// addCompanions adds the target's gallery or video: the linked one, the ones
// linking to it, and the likely one (same studio or cast, close release date,
// alike.Score agreeing).
func (e *Engine) addCompanions(ctx context.Context, target *mediaFeatures, set relatedSet) error {
	otherKind := schema.RecordKindGallery
	detail := "Photo set of this video"
	linkColumn := "video"
	if target.ref.Kind == schema.RecordKindGallery {
		otherKind, detail, linkColumn = schema.RecordKindVideo, "Video of this set", "gallery"
	}
	if companion, linked := target.companionRef(); linked {
		addCompanion(set, companion, detail, relatedWeightLinked)
	}
	condition := `("media"."` + linkColumn + `" = $targetId OR ` + companionCondition(target) + `)`
	candidates, err := e.queryFeatures(ctx, otherKind, schema.Filter{}, condition, `LIMIT $companionLimit`, companionVars(target))
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		if candidate.companion == target.ref.Id {
			addCompanion(set, candidate.ref, detail, relatedWeightLinked)
			continue
		}
		if likelyCompanion(target, candidate) {
			addCompanion(set, candidate.ref, detail, relatedWeightCompanion)
		}
	}
	return nil
}

// addCompanion records a companion once, with the weight of how sure the
// link is.
func addCompanion(set relatedSet, ref schema.RecordRef, detail string, weight float64) {
	found := set.entry(ref)
	if found.companion == "" {
		found.companion = detail
		found.score += weight
	}
}

// companionCondition narrows the other kind to items that could have been
// shot with the target: same studio or a shared performer, released within
// companionDateSlackDays when the target has a date.
func companionCondition(target *mediaFeatures) string {
	credits := `("media"."studio" = $targetStudio OR EXISTS (SELECT 1 FROM json_each("media"."performers")
		WHERE value IN (SELECT value FROM json_each($targetPerformers))))`
	if target.publishedAt.IsZero() {
		return credits
	}
	return `(` + credits + ` AND "media"."date" >= $companionFrom AND "media"."date" <= $companionTo)`
}

// companionVars binds companionCondition's parameters.
func companionVars(target *mediaFeatures) db.Vars {
	slack := companionDateSlackDays * 24 * time.Hour
	return db.Vars{
		"targetId":         target.ref.Id,
		"targetStudio":     target.studio,
		"targetPerformers": target.performers,
		"companionFrom":    target.publishedAt.Add(-slack).Format("2006-01-02"),
		"companionTo":      target.publishedAt.Add(slack).Format("2006-01-02") + "~",
		"companionLimit":   companionCandidates,
	}
}

// likelyCompanion reports whether a gallery and a video look like one shoot.
func likelyCompanion(target, candidate *mediaFeatures) bool {
	score := alike.Score(
		alike.Target{Performers: target.performers, Title: target.title, Date: target.date},
		alike.Candidate{Performers: candidate.performers, Title: candidate.title, Date: candidate.date},
	)
	return score >= companionMinimumScore
}

// addSharedCredits adds the newest items sharing the target's performers,
// its studio, or its rare tags.
func (e *Engine) addSharedCredits(ctx context.Context, target *mediaFeatures, rarity *tagRarity, set relatedSet) error {
	byPerformer, err := e.entityItems(ctx, rankedKinds, "performers", target.performers, itemsPerEntity, false)
	if err != nil {
		return err
	}
	for _, performerID := range target.performers {
		for _, ref := range byPerformer[performerID] {
			found := set.entry(ref)
			found.performers = append(found.performers, performerID)
			found.score += relatedWeightPerformer / float64(len(target.performers))
		}
	}
	if err := e.addSharedStudio(ctx, target, set); err != nil {
		return err
	}
	return e.addSharedTags(ctx, target, rarity, set)
}

// addSharedStudio adds the newest items of the target's studio.
func (e *Engine) addSharedStudio(ctx context.Context, target *mediaFeatures, set relatedSet) error {
	if target.studio == "" {
		return nil
	}
	byStudio, err := e.entityItems(ctx, rankedKinds, "studio", []string{target.studio}, itemsPerEntity, false)
	if err != nil {
		return err
	}
	for _, ref := range byStudio[target.studio] {
		found := set.entry(ref)
		found.studio = target.studio
		found.score += relatedWeightStudio
	}
	return nil
}

// addSharedTags adds the newest items carrying the target's rare tags.
func (e *Engine) addSharedTags(ctx context.Context, target *mediaFeatures, rarity *tagRarity, set relatedSet) error {
	tags := rareTags(target.tags, rarity)
	byTag, err := e.entityItems(ctx, rankedKinds, "tags", tags, itemsPerEntity, false)
	if err != nil {
		return err
	}
	for _, tagID := range tags {
		for _, ref := range byTag[tagID] {
			found := set.entry(ref)
			found.tags = append(found.tags, tagID)
			found.score += relatedWeightTag / float64(len(tags))
		}
	}
	return nil
}

// addSimilar adds the items whose embeddings are closest to the target's.
func (e *Engine) addSimilar(ctx context.Context, target *mediaFeatures, limit int, set relatedSet) error {
	found, err := e.similarTo(ctx, []schema.RecordRef{target.ref}, limit)
	if err != nil {
		return err
	}
	for _, item := range found {
		entry := set.entry(item.item)
		entry.similar = "Looks similar"
		if item.kind == schema.EmbeddingKindText {
			entry.similar = "Similar description"
		}
		entry.score += relatedWeightSimilar * item.similarity
	}
	return nil
}

// rankRelated keeps the eligible related items that aren't copies of the
// target or of each other, best first.
func (e *Engine) rankRelated(ctx context.Context, target *mediaFeatures, set relatedSet) ([]*relatedItem, error) {
	refs := make([]schema.RecordRef, 0, len(set))
	for ref := range set {
		refs = append(refs, ref)
	}
	features, err := e.loadFeatures(ctx, refs)
	if err != nil {
		return nil, err
	}
	ranked := make([]*relatedItem, 0, len(features))
	for ref, item := range set {
		if features[ref] != nil {
			ranked = append(ranked, item)
		}
	}
	sort.Slice(ranked, func(left, right int) bool {
		if ranked[left].score != ranked[right].score {
			return ranked[left].score > ranked[right].score
		}
		return ranked[left].ref.Id < ranked[right].ref.Id
	})
	return e.withoutCopies(ctx, target, ranked, features)
}

// withoutCopies drops the items that are the target on another site, and
// the lower-ranked copies among the rest.
func (e *Engine) withoutCopies(ctx context.Context, target *mediaFeatures, ranked []*relatedItem, features map[schema.RecordRef]*mediaFeatures) ([]*relatedItem, error) {
	items := []*mediaFeatures{target}
	for _, item := range ranked {
		items = append(items, features[item.ref])
	}
	found, err := e.loadDuplicates(ctx, items)
	if err != nil {
		return nil, err
	}
	found.keep(target)
	kept := make([]*relatedItem, 0, len(ranked))
	for _, item := range ranked {
		if found.copiesKept(features[item.ref]) {
			continue
		}
		found.keep(features[item.ref])
		kept = append(kept, item)
	}
	return kept, nil
}

// relatedMatches turns related items into matches, naming the shared
// performers, studio and tags.
func (e *Engine) relatedMatches(ctx context.Context, items []*relatedItem) ([]schema.Match, error) {
	ids := map[string]bool{}
	for _, item := range items {
		for _, id := range append(append([]string{item.studio}, item.performers...), item.tags...) {
			ids[id] = true
		}
	}
	names, err := e.entityNames(ctx, setKeys(ids))
	if err != nil {
		return nil, err
	}
	out := make([]schema.Match, 0, len(items))
	for _, item := range items {
		out = append(out, schema.Match{Item: item.ref, Score: float32Pointer(min(1, item.score)), Evidence: relatedEvidence(item, names)})
	}
	return out, nil
}

// relatedEvidence lists why an item is related, strongest relation first.
func relatedEvidence(item *relatedItem, names map[string]string) []schema.Evidence {
	evidence := []schema.Evidence{}
	if item.companion != "" {
		evidence = append(evidence, veilEvidence(item.companion))
	}
	if item.similar != "" {
		evidence = append(evidence, veilEvidence(item.similar))
	}
	if len(item.performers) > 0 {
		evidence = append(evidence, veilEvidence("Same performer: "+joinNames(item.performers, names)))
	}
	if item.studio != "" {
		evidence = append(evidence, veilEvidence("Same studio: "+names[item.studio]))
	}
	if len(item.tags) > 0 {
		evidence = append(evidence, veilEvidence("Shared tags: "+joinNames(item.tags, names)))
	}
	return evidence
}

// joinNames lists the names of the ids that have one, comma-separated.
func joinNames(ids []string, names map[string]string) string {
	named := []string{}
	for _, id := range ids {
		if names[id] != "" {
			named = append(named, names[id])
		}
	}
	return strings.Join(named, ", ")
}
