// Mirrors internal/plugins/types.go — keep in sync.
// All JSON field names use snake_case.

// Entity-scoped capabilities. "<entity>:find" fetches one full record by URL
// (NDJSON scrape stream: root entity first, then related standalone entities).
// "<entity>:list" returns reference stubs — a keyword search when args.query is
// set, the site's paginated catalog otherwise.
export type Capability =
  | "meta"
  | "enrich"
  | "stream:resolve"
  | "scene:find"
  | "scene:list"
  | "gallery:find"
  | "gallery:list"
  | "performer:find"
  | "performer:list"
  | "studio:find"
  | "studio:list"
  | "tag:find"
  | "tag:list"
  // Declares that sceneList honours args.url: it lists the scenes on one of
  // the site's own pages (a channel, performer or category page), which is how
  // subscriptions follow a studio/performer/tag.
  | "scene:list:page"
  // The site's own search-box completions for a partly typed query
  // (args.query), best first, so the search field can offer things the library
  // doesn't hold yet.
  | "search:suggest"
  // Live webcam rooms: "cam:list" lists the rooms online right now, "cam:stream"
  // resolves one room to its live HLS master playlist.
  | "cam:list"
  | "cam:stream"
  // Ranks the performers who look like the face in a picture (args.image_base64).
  // Only for services whose database holds professional performers or cam
  // models (stage personas), never general face search.
  | "face:identify"
  // One thread of a porn identification forum (args.url): the post and its
  // comments, with the names, links and pictures they give.
  | "post:find";

// The domain is adult content: scenes (videos), galleries (image sets),
// standalone images, the performers/studios/tags that describe them, and
// collections (ordered sets of any of the above). A scene is the atomic unit;
// episodic content is modeled as a collection.
export type MediaType =
  | "scene"
  | "performer"
  | "studio"
  | "tag"
  | "gallery"
  | "image"
  | "collection";

export interface PluginSettingField {
  key: string;
  label: string;
  description?: string;
  type: "string" | "password" | "boolean" | "number";
  required?: boolean;
  default?: string;
}

export interface PluginMeta {
  name: string;
  display_name?: string;
  description?: string;
  /** Emoji or URL */
  icon?: string;
  version: string;
  capabilities: Capability[];
  /** Domains this plugin handles for the resolve capability. Use "*" for wildcard. */
  domains?: string[];
  settings?: PluginSettingField[];
  /**
   * The site answers plain requests with a Cloudflare challenge, so the plugin
   * only works through FlareSolverr. The backend hides it when no solver is
   * configured (e.g. embedded on a phone).
   */
  requires_solver?: boolean;
  /**
   * The site signs poster and preview URLs with an expiry, so the backend
   * stores a copy of each stub's poster and refreshes the preview URL whenever
   * a listing shows the scene again.
   */
  expiring_media?: boolean;
  /**
   * The kinds of content the plugin's sites carry, so a search can be aimed at
   * the sites that have what the user looks for. Omitted: "videos" for a
   * scene lister, "galleries" for a gallery lister.
   */
  content?: ContentKind[];
}

/**
 * A kind of content: studio/amateur/tube videos, recorded live cam shows,
 * Japanese adult video, photo sets.
 */
export type ContentKind = "videos" | "cam-shows" | "jav" | "galleries";

// ---------------------------------------------------------------------------
// Assets attached to a record.

/** A still/profile/cover/logo. Maps to an image record. */
export interface Image {
  /** poster|still|profile|cover|logo|gallery */
  type: string;
  /** full URL or path */
  file_path: string;
  width?: number;
  height?: number;
  aspect_ratio?: number;
  /** ordering within a set */
  position?: number;
}

/** A playback/download source. Observation-only. */
export interface Download {
  label: string;
  url: string;
  quality?: string;
  language?: string;
  format?: string;
  size_bytes?: number;
  speed_bps?: number;
}

// ---------------------------------------------------------------------------
// Core entities emitted by plugins.

/**
 * A studio credit on a scene/gallery/image, resolved to a studio record during
 * ingest (matched by source_url, then external_id, then name).
 */
export interface StudioRef {
  name: string;
  external_id?: string;
  source_url?: string;
  /** channel logo/avatar, set on the studio when it has none */
  image_path?: string;
}

/**
 * A plugin-emitted marker: a tag and/or label at a point (or span, when
 * end_seconds is set) in a scene. Ingested as a global marker credited to the
 * emitting plugin.
 */
export interface SceneMarkerInput {
  seconds: number;
  end_seconds?: number;
  tag?: string;
  label?: string;
}

/** A performer credited on a scene/gallery, with per-appearance detail. */
export interface ScenePerformer {
  name: string;
  /** credited alias for this appearance */
  as?: string;
  /** billing order */
  order?: number;
  /** identity hints so ingest can match/create the performer record */
  external_id?: string;
  source_url?: string;
}

/** The atomic content unit: one adult video. */
export interface Scene {
  type: "scene";
  external_id: string;
  source_url: string;
  title: string;
  /** description */
  details?: string;
  /** ISO 8601 release/publish date */
  date?: string;
  duration_seconds?: number;
  /** 0..10 */
  rating?: number;
  /** the site's own view counter */
  view_count?: number;
  studio?: StudioRef;
  performers?: ScenePerformer[];
  tags?: string[];
  /** markers describing what happens at points/spans in the scene */
  markers?: SceneMarkerInput[];
  /** primary still/cover */
  poster_path?: string;
  /** short hover clip */
  preview_video?: string;
  preview_images?: string[];
  images?: Image[];
  /** playback/download sources */
  downloads?: Download[];
  /** the site's own "related videos" list from the scene page, in site order */
  related?: DiscoveredItem[];
  /**
   * The comments under the video as the page shows them (first page, site
   * order). Viewers often name the performer there; stored per scene and
   * replaced on every scrape.
   */
  comments?: SceneComment[];
}

/** One comment under a video on its site. */
export interface SceneComment {
  /** plain text: HTML stripped, emoticons as their alt text */
  text: string;
  /** the commenter's user name */
  author?: string;
  /** ISO 8601 date-time; relativeDate() turns "3 weeks ago" into one */
  date?: string;
  /** the site's like/vote count */
  likes?: number;
  /** index in Scene.comments of the comment this one answers */
  reply_to?: number;
}

/** A person appearing in scenes/galleries. */
export interface Performer {
  type: "performer";
  external_id: string;
  source_url: string;
  name: string;
  aliases?: string[];
  /** biography */
  details?: string;
  /** female|male|transgender_female|transgender_male|intersex|non_binary */
  gender?: string;
  birthdate?: string;
  death_date?: string;
  country?: string;
  ethnicity?: string;
  eye_color?: string;
  hair_color?: string;
  height_cm?: number;
  weight_kg?: number;
  /** e.g. "34D-24-36" */
  measurements?: string;
  fake_tits?: string;
  tattoos?: string;
  piercings?: string;
  career_length?: string;
  url?: string;
  twitter?: string;
  instagram?: string;
  /** profile image */
  image_path?: string;
  tags?: string[];
  images?: Image[];
}

/** A production studio / source site. */
export interface Studio {
  type: "studio";
  external_id: string;
  source_url: string;
  name: string;
  aliases?: string[];
  url?: string;
  /** parent studio name */
  parent?: string;
  image_path?: string;
  details?: string;
  tags?: string[];
}

/** A free-form category/attribute. */
export interface Tag {
  type: "tag";
  external_id?: string;
  name: string;
  aliases?: string[];
  description?: string;
  category?: string;
}

/** A set of images (photo shoot). */
export interface Gallery {
  type: "gallery";
  external_id: string;
  source_url: string;
  title: string;
  details?: string;
  date?: string;
  studio?: StudioRef;
  performers?: ScenePerformer[];
  tags?: string[];
  cover_path?: string;
  images?: Image[];
  /** photo count when the listing shows it but the images aren't scraped yet */
  image_count?: number;
}

/**
 * A standalone content image (not an asset like a poster or logo): browsable,
 * taggable, performer-linked. Gallery pages are content images too, but arrive
 * via Gallery.images.
 */
export interface ImageContent {
  type: "image";
  external_id?: string;
  source_url: string;
  title?: string;
  details?: string;
  date?: string;
  file_path: string;
  width?: number;
  height?: number;
  studio?: StudioRef;
  performers?: ScenePerformer[];
  tags?: string[];
}

/**
 * A scraper-created ordered set (site series, DVD, channel). Members are
 * referenced by source URL and resolved to records at ingest; unresolved URLs
 * are queued for scraping.
 */
export interface Collection {
  type: "collection";
  external_id?: string;
  source_url: string;
  name: string;
  details?: string;
  cover_path?: string;
  tags?: string[];
  /** ordered member source URLs (scenes/galleries/images) */
  member_urls?: string[];
}

// One item in the NDJSON stream returned by the scrape capability: root first
// (scene/gallery/image/collection), then any related standalone entities.
export type ScrapeResult =
  | { type: "scene"; scene: Scene }
  | { type: "performer"; performer: Performer }
  | { type: "studio"; studio: Studio }
  | { type: "tag"; tag: Tag }
  | { type: "gallery"; gallery: Gallery }
  | { type: "image"; image: ImageContent }
  | { type: "collection"; collection: Collection };

export interface DiscoveredItem {
  /** scene/gallery title or performer/studio/tag name */
  title: string;
  media_type: MediaType;
  source_url: string;
  external_id: string;
  date?: string;
  poster_path?: string;
  /** landscape preview frames, rendered as a cycling thumbnail strip */
  preview_images?: string[];
  /** short preview clip URL, played on hover */
  preview_video?: string;
  /** runtime when the listing shows one (scenes) */
  duration_seconds?: number;
  /** the site's rating (0..10) and view counter when the listing shows them (scenes) */
  rating?: number;
  view_count?: number;
  /** number of items behind this entry (tag/performer listings) */
  count?: number;
  /**
   * channel/uploader when the listing card names it (scenes), so search
   * results carry their channel before the page is visited
   */
  studio?: StudioRef;
  /**
   * Playback sources known at discovery time. Set this when a source can be
   * derived deterministically from the listing (e.g. a fixed embed URL) so the
   * ingested stub gets a stream immediately, with no scrape.
   */
  downloads?: Download[];
}

/** Arguments for an "<entity>:find" capability: one full record by URL. */
export interface FindArgs {
  url: string;
}

/**
 * Arguments for an "<entity>:list" capability. A set query switches the
 * listing from catalog browse to keyword search; a set url (sceneList of
 * plugins declaring "scene:list:page") lists the scenes on that site page.
 */
export interface ListArgs {
  query?: string;
  url?: string;
  limit?: number;
  offset?: number;
  /**
   * The kind of content a search asks for; a plugin carrying several kinds
   * (meta.content) then searches only the sites of that one.
   */
  content?: ContentKind;
}

/** Result of every list capability. */
export interface ItemsResult {
  items: DiscoveredItem[];
}

/** Arguments of the search:suggest capability. */
export interface SuggestArgs {
  /** What the user has typed so far. */
  query: string;
  limit?: number;
}

/** One completion a site's search box offers. */
export interface SiteSuggestion {
  text: string;
  /** "query" when absent. */
  kind?: "query" | "performer" | "tag" | "studio";
}

/** Result of the search:suggest capability, best first. */
export interface SuggestResult {
  suggestions: SiteSuggestion[];
}

/** Arguments of the face:identify capability; the backend downloads the picture itself. */
export interface FaceIdentifyArgs {
  image_base64: string;
  mime_type: string;
}

/** One performer a face-recognition service ranks as a match. */
export interface FaceCandidate {
  /** Stage name. */
  name: string;
  /** 0..1 */
  similarity: number;
  image_url?: string;
  /** The performer's pages on adult sites. */
  profile_urls: string[];
}

/**
 * A video on a tube site whose face looks like the picture's. It names no
 * person: the title is the video's, so it is shown as a link and never credited.
 */
export interface FaceVideoMatch {
  title: string;
  /** 0..1 */
  similarity: number;
  /** The video's page. */
  url: string;
  thumbnail_url?: string;
}

/**
 * A post on a porn identification forum whose picture looks like the searched
 * one. Its title is the asker's; names are the performers its accepted answer names.
 */
export interface FacePostMatch {
  title: string;
  /** The post's page. */
  url: string;
  thumbnail_url?: string;
  /** Whether the post has an accepted answer. */
  solved: boolean;
  /** Stage names of the accepted answer; empty while unsolved. */
  names: string[];
}

/** One comment of a forum thread, as plain text plus what it names and links. */
export interface ForumComment {
  /** The site's id of the comment. */
  id: string;
  author?: string;
  /** ISO 8601. */
  date?: string;
  /** Plain text, line breaks kept. */
  text: string;
  /** Whether this is the thread's accepted answer. */
  accepted: boolean;
  likes?: number;
  /** Performer names the comment tags. */
  names: string[];
  /** Scene or movie titles the comment tags. */
  titles: string[];
  /** Web links the comment gives, unwrapped from the site's redirects, in order. */
  links: string[];
  /** Pictures posted in the comment. */
  images: string[];
  /** Other threads of the same forum the comment points to. */
  posts: ForumPostRef[];
}

/** A thread of the same forum a comment points to. */
export interface ForumPostRef {
  title: string;
  url: string;
  thumbnail_url?: string;
}

/** One thread of a porn identification forum: the question and its comments. */
export interface ForumPost {
  url: string;
  title: string;
  description?: string;
  /** The asked-about picture, or the poster of a video. */
  image_url?: string;
  /** The asked-about clip, when the post is a video or GIF. */
  video_url?: string;
  author?: string;
  /** ISO 8601. */
  date?: string;
  tags: string[];
  /** Whether the thread has an accepted answer. */
  solved: boolean;
  /** Oldest first as the site lists them, the accepted answer marked. */
  comments: ForumComment[];
}

/** Result of the face:identify capability, best first. */
export interface FaceIdentifyResult {
  candidates: FaceCandidate[];
  /** Similar videos, when the service searches video sites too. */
  videos?: FaceVideoMatch[];
  /** Matching forum posts, when the service is a reverse image search of one. */
  posts?: FacePostMatch[];
}

export interface EnrichArgs {
  media_type: MediaType;
  existing: unknown;
  missing: string[];
}

// EnrichResult is a partial record — only set fields are merged onto the
// existing record. Fields span scene + performer + studio (minus discriminators).
export type EnrichResult = Partial<
  Omit<Scene, "type"> & Omit<Performer, "type"> & Omit<Studio, "type">
>;

export interface ResolveArgs {
  url: string;
}

/** Arguments of the cam:list capability. */
export interface CamListArgs {
  /** free-text search over room names and subjects */
  query?: string;
  /** one tag / hashtag the room must carry */
  tag?: string;
  /** site category: "female" | "male" | "couple" | "trans" */
  gender?: string;
  limit?: number;
  offset?: number;
}

/** One live room as a cam site's listing shows it. */
export interface CamRoom {
  /** the room's stable handle on the site; identity together with the plugin */
  username: string;
  display_name?: string;
  /** the room's page on the site */
  source_url: string;
  thumbnail_url?: string;
  viewers?: number;
  tags?: string[];
  /** "female" | "male" | "couple" | "trans" when the site says */
  gender?: string;
  /** the room's current title / subject line */
  subject?: string;
  /** false for rooms in a private or password show, which cannot be watched */
  online: boolean;
}

/** Result of cam:list: one page of rooms and the number the site has in total. */
export interface CamListResult {
  rooms: CamRoom[];
  total?: number;
}

/** Arguments of the cam:stream capability. */
export interface CamStreamArgs {
  username: string;
}

export interface ResolveResult {
  url: string;
  mime_type: "application/x-mpegURL" | "application/dash+xml" | "video/mp4" | "video/webm";
  headers?: Record<string, string>;
  quality?: string;
  load_time_ms?: number;
  speed_bps?: number;
}
