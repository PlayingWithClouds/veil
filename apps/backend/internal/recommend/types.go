package recommend

import "strings"

// Candidate source names. Every recommended item carries the one it was
// served from, and impressions echo it back.
const (
	SourceRelatedSite      = "related:site"
	SourceRelatedPerformer = "related:performer"
	SourceRelatedTag       = "related:tag"
	SourceRelatedTitle     = "related:title"
	SourceSubscription     = "subscription"
	SourcePerformer        = "affinity:performer"
	SourceStudio           = "affinity:studio"
	SourceTag              = "affinity:tag"
	SourceAdjacentTag      = "explore:adjacent"
	SourceRandom           = "explore:random"
	SourceSearch           = "search"
	SourceNewest           = "newest"
)

// Reason kinds: why an item was recommended, and what Reason.EntityID points at.
const (
	ReasonRelated      = "related"      // seed scene id
	ReasonPerformer    = "performer"    // performer id
	ReasonStudio       = "studio"       // studio id
	ReasonTag          = "tag"          // tag id
	ReasonExplore      = "explore"      // tag id adjacent to the top tags
	ReasonRandom       = "random"       // none
	ReasonSubscription = "subscription" // search subscription id
	ReasonSearch       = "search"       // none; EntityName is the query
	ReasonNewest       = "newest"       // none
)

// Reason explains one recommendation, e.g. "Because you watched X".
type Reason struct {
	Kind string
	// EntityID is the scene/performer/studio/tag/subscription behind the
	// reason; empty for kinds without one.
	EntityID string
	// EntityName is the entity's display name (seed title, performer name,
	// subscription query, ...).
	EntityName string
	// Text is the human-readable explanation.
	Text string
}

// Key identifies the reason's group, used to build rows.
func (reason Reason) Key() string {
	if reason.EntityID == "" {
		return reason.Kind
	}
	return reason.Kind + ":" + reason.EntityID
}

// Item is one ranked recommendation.
type Item struct {
	SceneID string
	Source  string
	Reason  Reason
	Score   float64
}

// Row is a titled group of items sharing a reason.
type Row struct {
	Key    string
	Title  string
	Reason Reason
	Items  []Item
}

// Category is a top-affinity tag with its best-scored scenes.
type Category struct {
	TagID string
	Items []Item
}

// Impression is one feed item the user was shown, or opened.
type Impression struct {
	SceneID  string
	Source   string
	Surface  string
	Position *int
	Clicked  bool
}

// isExploration reports whether items from source fill the reserved
// exploration slots.
func isExploration(source string) bool {
	return strings.HasPrefix(source, "explore:")
}

// describe fills the reason's human-readable text from its kind and entity name.
func (reason *Reason) describe() {
	switch reason.Kind {
	case ReasonRelated:
		reason.Text = "Because you watched " + reason.EntityName
	case ReasonPerformer, ReasonStudio:
		reason.Text = "More from " + reason.EntityName
	case ReasonTag:
		reason.Text = "Because you like " + reason.EntityName
	case ReasonExplore:
		reason.Text = "Explore " + reason.EntityName
	case ReasonRandom:
		reason.Text = "Something different"
	case ReasonSubscription:
		reason.Text = "New for “" + reason.EntityName + "”"
	case ReasonSearch:
		reason.Text = "From your search “" + reason.EntityName + "”"
	default:
		reason.Text = "Newly added"
	}
}
