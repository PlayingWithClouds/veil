package recommend

// Reason kinds: why an item was recommended, and what reason.entityID points at.
const (
	reasonRelated   = "related"   // seed item id; entityName is its title
	reasonPerformer = "performer" // performer id
	reasonStudio    = "studio"    // studio id
	reasonTag       = "tag"       // tag id
	reasonExplore   = "explore"   // tag id adjacent to the top tags
	reasonLookalike = "lookalike" // performer id
	reasonRandom    = "random"    // none
	reasonSearch    = "search"    // none; entityName is the query
	reasonNewest    = "newest"    // none
)

// reason explains one recommendation, e.g. "Because you watched X".
type reason struct {
	kind string
	// entityID is the item, performer, studio or tag behind the reason; empty
	// for kinds without one.
	entityID string
	// entityName is the entity's display name, filled in by describe's caller
	// for performers, studios and tags.
	entityName string
}

// named reports whether the reason's entity name is read from the
// performer, studio and tag tables.
func (why reason) named() bool {
	switch why.kind {
	case reasonPerformer, reasonStudio, reasonTag, reasonExplore, reasonLookalike:
		return true
	}
	return false
}

// text is the human-readable explanation.
func (why reason) text() string {
	switch why.kind {
	case reasonRelated:
		if why.entityName == "" {
			return "Similar to what you watched"
		}
		return "Because you watched " + why.entityName
	case reasonPerformer:
		return "More with " + why.entityName
	case reasonStudio:
		return "More from " + why.entityName
	case reasonTag:
		return "Because you like " + why.entityName
	case reasonExplore:
		return "Explore " + why.entityName
	case reasonLookalike:
		return "Discover " + why.entityName
	case reasonRandom:
		return "Something different"
	case reasonSearch:
		return "From your search “" + why.entityName + "”"
	}
	return "Newly added"
}
