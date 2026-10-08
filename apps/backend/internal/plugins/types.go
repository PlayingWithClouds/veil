package plugins

import (
	"encoding/json"
	"net/url"
	"strings"
)

type Capability string

const (
	CapabilityEnrich        Capability = "enrich"
	CapabilityStreamResolve Capability = "stream:resolve"

	// Entity-scoped capabilities. "<entity>:find" fetches one full record by URL
	// (the NDJSON scrape stream: root entity first, then related standalone
	// entities). "<entity>:list" returns reference stubs — a keyword search when
	// args.query is set, the site's paginated catalog otherwise.
	CapabilitySceneFind     Capability = "scene:find"
	CapabilitySceneList     Capability = "scene:list"
	CapabilityGalleryFind   Capability = "gallery:find"
	CapabilityGalleryList   Capability = "gallery:list"
	CapabilityPerformerFind Capability = "performer:find"
	CapabilityPerformerList Capability = "performer:list"
	CapabilityStudioFind    Capability = "studio:find"
	CapabilityStudioList    Capability = "studio:list"
	CapabilityTagFind       Capability = "tag:find"
	CapabilityTagList       Capability = "tag:list"

	// CapabilitySceneListPage declares that scene:list honours args.url: it
	// lists the scenes on one of the site's own pages (a channel, performer or
	// category page), which is how subscriptions follow a studio/performer/tag.
	CapabilitySceneListPage Capability = "scene:list:page"
)

// Content entities a plugin can surface in search/browse feeds, in preference
// order: video plugins expose scenes, image-set plugins expose galleries.
var contentEntities = []MediaType{MediaTypeScene, MediaTypeGallery}

// ContentFindEntity returns the content entity type this plugin can fetch in
// full by URL (scene before gallery), or false if it supports neither.
func (p *Plugin) ContentFindEntity() (MediaType, bool) {
	for _, entity := range contentEntities {
		if p.Meta.Has(FindCapability(entity)) {
			return entity, true
		}
	}
	return "", false
}

// ContentListEntity returns the content entity type this plugin can list
// paginated (scene before gallery), or false if it supports neither.
func (p *Plugin) ContentListEntity() (MediaType, bool) {
	for _, entity := range contentEntities {
		if p.Meta.Has(ListCapability(entity)) {
			return entity, true
		}
	}
	return "", false
}

// ContentListEntities returns every content entity type this plugin can list
// (scene and/or gallery), so search can pull both from a plugin that exposes them.
func (p *Plugin) ContentListEntities() []MediaType {
	var entities []MediaType
	for _, entity := range contentEntities {
		if p.Meta.Has(ListCapability(entity)) {
			entities = append(entities, entity)
		}
	}
	return entities
}

// searchListEntities are the entity types a keyword search lists from plugins:
// content (scenes, galleries) plus the performers and studios that describe it.
var searchListEntities = []MediaType{MediaTypeScene, MediaTypeGallery, MediaTypePerformer, MediaTypeStudio}

// SearchListEntities returns every entity type this plugin can list that a
// keyword search should pull (scene, gallery, performer, studio).
func (p *Plugin) SearchListEntities() []MediaType {
	var entities []MediaType
	for _, entity := range searchListEntities {
		if p.Meta.Has(ListCapability(entity)) {
			entities = append(entities, entity)
		}
	}
	return entities
}

// ServesURL reports whether rawURL's host is one of the plugin's meta.domains
// (or a subdomain of one). The "*" wildcard does not count.
func (p *Plugin) ServesURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return false
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	for _, domain := range p.Meta.Domains {
		if domain == "*" {
			continue
		}
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

// FindCapability returns the "<entity>:find" capability for a media type.
func FindCapability(entity MediaType) Capability {
	return Capability(string(entity) + ":find")
}

// ListCapability returns the "<entity>:list" capability for a media type.
func ListCapability(entity MediaType) Capability {
	return Capability(string(entity) + ":list")
}

type MediaType string

// The domain is adult content: scenes (videos), galleries (image sets),
// standalone images, the performers/studios/tags that describe them, and
// collections (ordered sets of any of the above). There is no
// series/season/episode hierarchy — a scene is the atomic unit; episodic
// content is modeled as a collection.
const (
	MediaTypeScene      MediaType = "scene"
	MediaTypePerformer  MediaType = "performer"
	MediaTypeStudio     MediaType = "studio"
	MediaTypeTag        MediaType = "tag"
	MediaTypeGallery    MediaType = "gallery"
	MediaTypeImage      MediaType = "image"
	MediaTypeCollection MediaType = "collection"
)

// PluginSettingField describes one configurable value a plugin needs.
// The value is injected into the plugin subprocess as an env var named by Key.
type PluginSettingField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"` // "string" | "password" | "boolean" | "number"
	Required    bool   `json:"required,omitempty"`
	Default     string `json:"default,omitempty"`
}

// PluginMeta is returned by a plugin when invoked with capability "meta".
type PluginMeta struct {
	Name         string       `json:"name"`
	DisplayName  string       `json:"display_name,omitempty"`
	Description  string       `json:"description,omitempty"`
	Icon         string       `json:"icon,omitempty"` // emoji or URL
	Version      string       `json:"version"`
	Capabilities []Capability `json:"capabilities"`
	// Domains lists provider hostnames this plugin can resolve (resolve capability only).
	// Use "*" to act as a wildcard fallback for any domain.
	Domains  []string             `json:"domains,omitempty"`
	Settings []PluginSettingField `json:"settings,omitempty"`
	// RequiresSolver marks plugins whose site sits behind a Cloudflare
	// challenge and only works through FlareSolverr.
	RequiresSolver bool `json:"requires_solver,omitempty"`
}

// ResolveArgs is passed to a plugin's resolve capability.
type ResolveArgs struct {
	URL string `json:"url"`
}

// ResolveResult is returned by a plugin's resolve capability.
type ResolveResult struct {
	URL        string            `json:"url"`
	MimeType   string            `json:"mime_type"`
	Headers    map[string]string `json:"headers,omitempty"`
	Quality    string            `json:"quality,omitempty"`
	LoadTimeMs int64             `json:"load_time_ms,omitempty"`
	SpeedBps   float64           `json:"speed_bps,omitempty"`
}

func (m PluginMeta) Has(c Capability) bool {
	for _, cap := range m.Capabilities {
		if cap == c {
			return true
		}
	}
	return false
}

// Plugin is a loaded npm package plugin.
type Plugin struct {
	Meta       PluginMeta
	Dir        string // absolute path to plugin directory (contains package.json)
	EntryPoint string // relative entry file resolved from package.json "main"
	Package    string // npm package name from package.json
	LocalBuild bool   // bundled from this checkout by the dev bundler, not installed from npm
}

// ---------------------------------------------------------------------------
// Core entities emitted by plugins. Field names mirror the SurrealDB schema in
// sql/migrations so ingest can map an observation onto a canonical record.

// Scene is the atomic content unit: one adult video. Only ExternalID, SourceURL
// and Title are required; the rest are filled opportunistically and completed
// later by enrich plugins.
type Scene struct {
	Type       MediaType `json:"type"` // "scene"
	ExternalID string    `json:"external_id"`
	SourceURL  string    `json:"source_url"`
	Title      string    `json:"title"`
	Details    string    `json:"details,omitempty"` // description
	Date       string    `json:"date,omitempty"`    // ISO 8601 release/publish date
	Duration   int       `json:"duration_seconds,omitempty"`
	Rating     float64   `json:"rating,omitempty"` // 0..10
	// ViewCount is the site's own view counter, not Veil's watch count.
	ViewCount int `json:"view_count,omitempty"`
	// Related entities by name/reference (resolved to records during ingest).
	Studio     *StudioRef       `json:"studio,omitempty"`
	Performers []ScenePerformer `json:"performers,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
	// Markers describing what happens at points/spans in the scene. Ingested
	// as global markers credited to the emitting plugin.
	Markers []SceneMarkerInput `json:"markers,omitempty"`
	// Media assets.
	PosterPath    string   `json:"poster_path,omitempty"`   // primary still/cover
	PreviewVideo  string   `json:"preview_video,omitempty"` // short hover clip
	PreviewImages []string `json:"preview_images,omitempty"`
	Images        []Image  `json:"images,omitempty"`
	// Playback/download sources. Observation-only.
	Downloads []Download `json:"downloads,omitempty"`
	// Related is the site's own "related videos" list from the scene page, in
	// the site's order. Ingested as stubs linked to this scene.
	Related []DiscoveredItem `json:"related,omitempty"`
}

// StudioRef is a studio credit on a scene/gallery/image, resolved to a studio
// record during ingest (matched by source_url, then external_id, then name).
type StudioRef struct {
	Name       string `json:"name"`
	ExternalID string `json:"external_id,omitempty"`
	SourceURL  string `json:"source_url,omitempty"`
	// ImagePath is the channel's logo/avatar, set on the studio when it has none.
	ImagePath string `json:"image_path,omitempty"`
}

// SceneMarkerInput is a plugin-emitted marker: a tag and/or label at a point
// (or span, when EndSeconds is set) in a scene.
type SceneMarkerInput struct {
	Seconds    float64  `json:"seconds"`
	EndSeconds *float64 `json:"end_seconds,omitempty"`
	Tag        string   `json:"tag,omitempty"`
	Label      string   `json:"label,omitempty"`
}

// ScenePerformer is a performer credited on a scene, with per-appearance detail.
type ScenePerformer struct {
	Name  string `json:"name"`
	As    string `json:"as,omitempty"`    // credited alias for this appearance
	Order int    `json:"order,omitempty"` // billing order
	// Optional identity hints so ingest can match/create the performer record.
	ExternalID string `json:"external_id,omitempty"`
	SourceURL  string `json:"source_url,omitempty"`
}

// Performer is a person appearing in scenes/galleries.
type Performer struct {
	Type         MediaType `json:"type"` // "performer"
	ExternalID   string    `json:"external_id"`
	SourceURL    string    `json:"source_url"`
	Name         string    `json:"name"`
	Aliases      []string  `json:"aliases,omitempty"`
	Details      string    `json:"details,omitempty"` // biography
	Gender       string    `json:"gender,omitempty"`
	Birthdate    string    `json:"birthdate,omitempty"`
	DeathDate    string    `json:"death_date,omitempty"`
	Country      string    `json:"country,omitempty"`
	Ethnicity    string    `json:"ethnicity,omitempty"`
	EyeColor     string    `json:"eye_color,omitempty"`
	HairColor    string    `json:"hair_color,omitempty"`
	HeightCm     int       `json:"height_cm,omitempty"`
	WeightKg     int       `json:"weight_kg,omitempty"`
	Measurements string    `json:"measurements,omitempty"`
	FakeTits     string    `json:"fake_tits,omitempty"`
	Tattoos      string    `json:"tattoos,omitempty"`
	Piercings    string    `json:"piercings,omitempty"`
	CareerLength string    `json:"career_length,omitempty"`
	URL          string    `json:"url,omitempty"`
	Twitter      string    `json:"twitter,omitempty"`
	Instagram    string    `json:"instagram,omitempty"`
	ImagePath    string    `json:"image_path,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	Images       []Image   `json:"images,omitempty"`
}

// Studio is a production studio / source site.
type Studio struct {
	Type       MediaType `json:"type"` // "studio"
	ExternalID string    `json:"external_id"`
	SourceURL  string    `json:"source_url"`
	Name       string    `json:"name"`
	Aliases    []string  `json:"aliases,omitempty"`
	URL        string    `json:"url,omitempty"`
	Parent     string    `json:"parent,omitempty"` // parent studio name
	ImagePath  string    `json:"image_path,omitempty"`
	Details    string    `json:"details,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
}

// Tag is a free-form category/attribute.
type Tag struct {
	Type        MediaType `json:"type"` // "tag"
	ExternalID  string    `json:"external_id,omitempty"`
	Name        string    `json:"name"`
	Aliases     []string  `json:"aliases,omitempty"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category,omitempty"`
}

// Gallery is a set of images (photo shoot).
type Gallery struct {
	Type       MediaType        `json:"type"` // "gallery"
	ExternalID string           `json:"external_id"`
	SourceURL  string           `json:"source_url"`
	Title      string           `json:"title"`
	Details    string           `json:"details,omitempty"`
	Date       string           `json:"date,omitempty"`
	Studio     *StudioRef       `json:"studio,omitempty"`
	Performers []ScenePerformer `json:"performers,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
	CoverPath  string           `json:"cover_path,omitempty"`
	Images     []Image          `json:"images,omitempty"`
}

// ImageContent is a standalone content image (not an asset like a poster or
// logo): browsable, taggable, performer-linked. Gallery pages are content
// images too, but arrive via Gallery.Images.
type ImageContent struct {
	Type       MediaType        `json:"type"` // "image"
	ExternalID string           `json:"external_id,omitempty"`
	SourceURL  string           `json:"source_url"`
	Title      string           `json:"title,omitempty"`
	Details    string           `json:"details,omitempty"`
	Date       string           `json:"date,omitempty"`
	FilePath   string           `json:"file_path"`
	Width      int              `json:"width,omitempty"`
	Height     int              `json:"height,omitempty"`
	Studio     *StudioRef       `json:"studio,omitempty"`
	Performers []ScenePerformer `json:"performers,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
}

// Collection is a scraper-created ordered set (site series, DVD, channel).
// Members are referenced by source URL and resolved to records at ingest;
// unresolved URLs are queued for scraping.
type Collection struct {
	Type       MediaType `json:"type"` // "collection"
	ExternalID string    `json:"external_id,omitempty"`
	SourceURL  string    `json:"source_url"`
	Name       string    `json:"name"`
	Details    string    `json:"details,omitempty"`
	CoverPath  string    `json:"cover_path,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	// Ordered member source URLs (scenes/galleries/images).
	MemberURLs []string `json:"member_urls,omitempty"`
}

// ScrapeResult is one item in the NDJSON stream returned by an "<entity>:find"
// capability. Plugins emit one item per line: root first (scene/gallery/image/
// collection), then any related standalone entities (performer, studio, tag).
// Go reads lines until EOF.
type ScrapeResult struct {
	Type       MediaType     `json:"type"`
	Scene      *Scene        `json:"scene,omitempty"`
	Performer  *Performer    `json:"performer,omitempty"`
	Studio     *Studio       `json:"studio,omitempty"`
	Tag        *Tag          `json:"tag,omitempty"`
	Gallery    *Gallery      `json:"gallery,omitempty"`
	Image      *ImageContent `json:"image,omitempty"`
	Collection *Collection   `json:"collection,omitempty"`
}

// DiscoveredItem is a minimal reference returned by the find/list capabilities.
type DiscoveredItem struct {
	Title      string    `json:"title"` // scene/gallery title or performer/studio/tag name
	MediaType  MediaType `json:"media_type"`
	SourceURL  string    `json:"source_url"`
	ExternalID string    `json:"external_id"`
	Date       string    `json:"date,omitempty"`
	PosterPath string    `json:"poster_path,omitempty"`
	// PreviewImages holds landscape preview frames, rendered as a cycling
	// thumbnail strip on the banner card.
	PreviewImages []string `json:"preview_images,omitempty"`
	// PreviewVideo is an optional short preview clip URL, played on hover.
	PreviewVideo string `json:"preview_video,omitempty"`
	// DurationSeconds is the item's runtime when the listing shows one (scenes).
	DurationSeconds int `json:"duration_seconds,omitempty"`
	// Rating (0..10) and ViewCount are the site's numbers when the listing shows them (scenes).
	Rating    float64 `json:"rating,omitempty"`
	ViewCount int     `json:"view_count,omitempty"`
	// Count is the number of items behind this entry (tag/performer listings).
	Count int `json:"count,omitempty"`
	// Studio is the channel/uploader when the listing card names it (scenes),
	// so search results carry their channel before the page is visited.
	Studio *StudioRef `json:"studio,omitempty"`
	// Downloads are playback sources known at discovery time. When a plugin can
	// derive a source deterministically from the listing (e.g. a fixed embed URL),
	// setting this attaches a stream to the stub immediately — no scrape needed.
	Downloads []Download `json:"downloads,omitempty"`
}

// FindArgs is passed to an "<entity>:find" capability.
type FindArgs struct {
	URL string `json:"url"`
}

// ListArgs is passed to an "<entity>:list" capability. A set Query switches
// the listing from catalog browse to keyword search; a set URL (scene:list of
// plugins declaring scene:list:page) lists the scenes on that site page.
type ListArgs struct {
	Query  string `json:"query,omitempty"`
	URL    string `json:"url,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// ItemsResult is returned by every find/list capability.
type ItemsResult struct {
	Items []DiscoveredItem `json:"items"`
}

// EnrichArgs is passed to a plugin's enrich capability.
type EnrichArgs struct {
	MediaType MediaType       `json:"media_type"`
	Existing  json.RawMessage `json:"existing"`
	Missing   []string        `json:"missing"`
}

// EnrichResult is a partial record — only set fields are merged. Field names
// mirror Scene/Performer/Studio so any entity type can be enriched.
type EnrichResult struct {
	// Scalars — pointers so "unset" is distinct from zero value.
	Title        *string    `json:"title,omitempty"`
	Name         *string    `json:"name,omitempty"`
	Details      *string    `json:"details,omitempty"`
	Date         *string    `json:"date,omitempty"`
	Duration     *int       `json:"duration_seconds,omitempty"`
	Rating       *float64   `json:"rating,omitempty"`
	PosterPath   *string    `json:"poster_path,omitempty"`
	PreviewVideo *string    `json:"preview_video,omitempty"`
	Studio       *StudioRef `json:"studio,omitempty"`
	// Performer-specific scalars.
	Gender       *string `json:"gender,omitempty"`
	Birthdate    *string `json:"birthdate,omitempty"`
	DeathDate    *string `json:"death_date,omitempty"`
	Country      *string `json:"country,omitempty"`
	Ethnicity    *string `json:"ethnicity,omitempty"`
	EyeColor     *string `json:"eye_color,omitempty"`
	HairColor    *string `json:"hair_color,omitempty"`
	HeightCm     *int    `json:"height_cm,omitempty"`
	WeightKg     *int    `json:"weight_kg,omitempty"`
	Measurements *string `json:"measurements,omitempty"`
	FakeTits     *string `json:"fake_tits,omitempty"`
	Tattoos      *string `json:"tattoos,omitempty"`
	Piercings    *string `json:"piercings,omitempty"`
	CareerLength *string `json:"career_length,omitempty"`
	URL          *string `json:"url,omitempty"`
	Twitter      *string `json:"twitter,omitempty"`
	Instagram    *string `json:"instagram,omitempty"`
	ImagePath    *string `json:"image_path,omitempty"`
	// Related entities / assets.
	Aliases       []string         `json:"aliases,omitempty"`
	Tags          []string         `json:"tags,omitempty"`
	Performers    []ScenePerformer `json:"performers,omitempty"`
	PreviewImages []string         `json:"preview_images,omitempty"`
	Images        []Image          `json:"images,omitempty"`
	Downloads     []Download       `json:"downloads,omitempty"`
}

// ---------------------------------------------------------------------------
// Assets attached to a record.

// Image is a still/profile/cover/logo. Maps to an image record.
type Image struct {
	Type        string  `json:"type"`      // poster|still|profile|cover|logo|gallery
	FilePath    string  `json:"file_path"` // full URL or path
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	AspectRatio float64 `json:"aspect_ratio,omitempty"`
	Position    int     `json:"position,omitempty"` // ordering within a set
}

// Download is a playback/download source. Observation-only.
type Download struct {
	Label     string  `json:"label"`
	URL       string  `json:"url"`
	Quality   string  `json:"quality,omitempty"`
	Language  string  `json:"language,omitempty"`
	Format    string  `json:"format,omitempty"`
	SizeBytes int64   `json:"size_bytes,omitempty"`
	SpeedBps  float64 `json:"speed_bps,omitempty"`
}
