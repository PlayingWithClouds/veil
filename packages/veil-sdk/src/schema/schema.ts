export type CollectionKind =
  | "user"
  | "watchlist"
  | "queue";

export type SortField =
  | "relevance"
  | "date"
  | "addedAt"
  | "title"
  | "rating"
  | "duration"
  | "viewCount"
  | "lastViewedAt"
  | "random";

export type EventKind =
  | "view"
  | "progress"
  | "finish"
  | "o"
  | "search"
  | "impression"
  | "choice"
  | "heatSpan"
  | "skip"
  | "like";

export type LimitKind =
  | "watchTime"
  | "sessions";

export type LimitWindow =
  | "day"
  | "week";

export type LimitAction =
  | "remind"
  | "lock";

export type AccountStatus =
  | "ok"
  | "expired"
  | "needsLogin"
  | "failed";

export type Platform =
  | "server"
  | "desktop"
  | "android"
  | "tv"
  | "vr";

export type JobKind =
  | "download"
  | "record"
  | "scrape"
  | "followRun"
  | "fingerprint"
  | "embed"
  | "trickplay"
  | "transcode"
  | "libraryScan"
  | "automationRun"
  | "snapshot"
  | "enrich"
  | "speedCheck";

export type JobStatus =
  | "queued"
  | "running"
  | "succeeded"
  | "failed"
  | "cancelled";

export type ContributionKind =
  | "source"
  | "lookup"
  | "stream"
  | "cams"
  | "trigger"
  | "action"
  | "feed"
  | "field"
  | "login"
  | "view";

export type PluginOrigin =
  | "store"
  | "local";

export type RecordKind =
  | "video"
  | "gallery"
  | "image"
  | "audio"
  | "story"
  | "post"
  | "performer"
  | "studio"
  | "camModel"
  | "group"
  | "tag"
  | "category"
  | "collection"
  | "segment"
  | "clip";

export type MediaKind =
  | "video"
  | "gallery"
  | "audio"
  | "story"
  | "post";

export type Actor =
  | "user"
  | "plugin"
  | "automation"
  | "system";

export type MediaStatus =
  | "available"
  | "wanted"
  | "unavailable";

export type VrProjection =
  | "equirectangular180"
  | "equirectangular360"
  | "fisheye"
  | "cubemap";

export type VrStereo =
  | "mono"
  | "sideBySide"
  | "topBottom";

export type FieldType =
  | "text"
  | "number"
  | "boolean"
  | "date"
  | "choice"
  | "choices"
  | "url";

export type TriggerKind =
  | "event"
  | "schedule"
  | "newMedia"
  | "camLive"
  | "fileAdded"
  | "wantedFound"
  | "plugin";

export type RunStatus =
  | "running"
  | "succeeded"
  | "failed"
  | "undone";

export type ChannelKind =
  | "inApp"
  | "push"
  | "ntfy"
  | "webhook";

export type StorageKind =
  | "library"
  | "cache"
  | "archive";

export type SegmentOrigin =
  | "user"
  | "plugin"
  | "ai";

export type SubtitleOrigin =
  | "site"
  | "generated"
  | "translated";

export type ScriptKind =
  | "funscript";

export type FingerprintAlgorithm =
  | "phash"
  | "oshash"
  | "md5"
  | "audio";

export type EmbeddingKind =
  | "face"
  | "visual"
  | "text";

export interface Evidence {
  plugin: string;
  url?: string;
  title?: string;
  image?: Asset;
  detail?: string;
}

export interface Match {
  item: RecordRef;
  score?: number;
  evidence: Evidence[];
}

export interface LookupQuery {
  text?: string;
  imageBase64?: Uint8Array;
  mimeType?: string;
  plugins?: string[];
}

export interface UserState extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  rating?: number;
  favorite: boolean;
  notes?: string;
  blocked: boolean;
  progress?: number;
  viewCount: number;
  lastViewedAt?: string;
  fields: { [key in string]: unknown };
}

export interface Collection extends Entity {
  name: string;
  kind: CollectionKind;
  details?: string;
  cover?: Asset;
}

export interface CollectionItem extends Entity {
  collection: string;
  subjectKind: RecordKind;
  subjectId: string;
  position: number;
}

export interface Sort {
  field: SortField;
  descending: boolean;
}

export interface FieldCondition {
  fieldDefId: string;
  operator: string;
  value?: unknown;
}

export interface Filter {
  kinds: MediaKind[];
  text?: string;
  performerIds?: string[];
  studioIds?: string[];
  categoryIds?: string[];
  includeTagIds?: string[];
  excludeTagIds?: string[];
  sources?: string[];
  dateFrom?: string;
  dateTo?: string;
  minDurationSeconds?: number;
  maxDurationSeconds?: number;
  minRating?: number;
  status?: MediaStatus[];
  inLibrary?: boolean;
  favorite?: boolean;
  unwatched?: boolean;
  vr?: boolean;
  fields?: FieldCondition[];
  sort?: Sort;
  minHeight?: number;
}

export interface SavedFilter extends Entity {
  name: string;
  filter: Filter;
  pinned: boolean;
}

export interface Follow extends Entity {
  name: string;
  filter: Filter;
  targetKind?: RecordKind;
  targetId?: string;
  intervalHours: number;
  enabled: boolean;
  notify: boolean;
  lastRunAt?: string;
  nextRunAt?: string;
  lastError?: string;
}

export interface Event extends Entity {
  at: string;
  kind: EventKind;
  subjectKind?: RecordKind;
  subjectId?: string;
  data?: unknown;
  device?: string;
  sessionId?: string;
}

export interface Limit extends Entity {
  kind: LimitKind;
  window: LimitWindow;
  amount: number;
  action: LimitAction;
  enabled: boolean;
}

export interface Account extends Entity {
  plugin: string;
  label: string;
  status: AccountStatus;
  device: string;
  credentials?: Uint8Array;
  session?: Uint8Array;
  expiresAt?: string;
  lastCheckedAt?: string;
}

export interface Device extends Entity {
  name: string;
  platform: Platform;
  lastSeenAt?: string;
  lastSyncAt?: string;
  current: boolean;
}

export interface Job extends Entity {
  kind: JobKind;
  status: JobStatus;
  subject?: RecordRef;
  payload?: unknown;
  progress: number;
  error?: string;
  attempts: number;
  runAt: string;
  startedAt?: string;
  finishedAt?: string;
  device?: string;
  plugin?: string;
  dedupeKey?: string;
  priority: number;
}

export interface Setting extends Entity {
  plugin: string;
  key: string;
  value?: unknown;
}

export interface CacheEntry {
  key: string;
  value: string;
  expiresAt?: number;
}

export interface ApiToken extends Entity {
  name: string;
  scopes: string[];
  tokenHash: Uint8Array;
  lastUsedAt?: string;
  expiresAt?: string;
}

export interface Contribution {
  kind: ContributionKind;
  name: string;
  mediaKinds?: MediaKind[];
  domains?: string[];
}

export interface Plugin extends Entity {
  name: string;
  displayName?: string;
  version: { major: number; minor: number; patch: number };
  enabled: boolean;
  origin: PluginOrigin;
  contributions: Contribution[];
  settings?: unknown;
}

export interface Entity {
  id: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
}

export interface RecordRef {
  kind: RecordKind;
  id: string;
}

export interface Asset {
  uri: string;
  mimeType?: string;
  width?: number;
  height?: number;
  blurhash?: string;
}

export interface Source extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  plugin: string;
  externalId: string;
  url: string;
  title?: string;
  image?: Asset;
  snapshot?: Asset;
  fetchedAt?: string;
}

export interface Media extends Entity {
  title: string;
  details?: string;
  date?: string;
  cover?: Asset;
  studio?: string;
  performers: string[];
  tags: string[];
  status: MediaStatus;
}

export interface VrFormat {
  projection: VrProjection;
  stereo: VrStereo;
  fieldOfView?: number;
}

export interface Trickplay {
  sheets: Asset[];
  columns: number;
  rows: number;
  tileWidth: number;
  tileHeight: number;
  intervalSeconds: number;
}

export interface Video extends Media {
  durationSeconds?: number;
  width?: number;
  height?: number;
  vr?: VrFormat;
  preview?: Asset;
  trickplay?: Trickplay;
  gallery?: string;
  recordedFrom?: string;
  previewFrames?: Asset[];
}

export interface GroupItem extends Entity {
  group: string;
  video: string;
  position: number;
}

export interface Gallery extends Media {
  imageCount: number;
  video?: string;
}

export interface Image extends Entity {
  gallery: string;
  position: number;
  asset: Asset;
  width?: number;
  height?: number;
}

export interface Audio extends Media {
  durationSeconds?: number;
}

export interface Story extends Media {
  chapterCount: number;
  wordCount?: number;
  complete: boolean;
}

export interface StoryChapter extends Entity {
  story: string;
  position: number;
  title?: string;
  text: string;
  wordCount: number;
  url?: string;
}

export interface Post extends Media {
  creator: string;
  text?: string;
  attachments: RecordRef[];
  locked: boolean;
  priceCents?: number;
  currency?: string;
}

export interface PerformerAttributes {
  gender?: string;
  birthdate?: string;
  deathDate?: string;
  country?: string;
  ethnicity?: string;
  eyeColor?: string;
  hairColor?: string;
  heightCm?: number;
  weightKg?: number;
  measurements?: string;
  fakeTits?: string;
  tattoos?: string;
  piercings?: string;
  careerStart?: number;
  careerEnd?: number;
}

export interface Performer extends Entity {
  name: string;
  aliases: string[];
  image?: Asset;
  details?: string;
  attributes: PerformerAttributes;
  links: string[];
  tags: string[];
  nameKey: string;
}

export interface Studio extends Entity {
  name: string;
  aliases: string[];
  image?: Asset;
  details?: string;
  parent?: string;
  url?: string;
  tags: string[];
  nameKey: string;
}

export interface CamLive {
  viewers: number;
  subject?: string;
  startedAt?: string;
  thumbnail?: Asset;
}

export interface CamModel extends Entity {
  plugin: string;
  username: string;
  displayName: string;
  performer?: string;
  thumbnail?: Asset;
  tags: string[];
  gender?: string;
  live?: CamLive;
  lastSeenOnlineAt?: string;
  autoRecord: boolean;
  recordQuality?: string;
}

export interface Group extends Entity {
  title: string;
  details?: string;
  date?: string;
  cover?: Asset;
  studio?: string;
}

export interface Tag extends Entity {
  name: string;
  aliases: string[];
  nameKey: string;
}

export interface Category extends Entity {
  name: string;
  slug: string;
  parent?: string;
  matchTags: string[];
  cover?: Asset;
  position: number;
}

export interface FieldDef extends Entity {
  name: string;
  kinds: RecordKind[];
  fieldType: FieldType;
  options?: string[];
}

export interface Trigger {
  kind: TriggerKind;
  cron?: string;
  eventKinds?: string[];
  filter?: Filter;
  pluginTrigger?: string;
}

export interface Step {
  action: string;
  args?: unknown;
  condition?: Filter;
  continueOnError: boolean;
}

export interface Automation extends Entity {
  name: string;
  enabled: boolean;
  trigger: Trigger;
  steps: Step[];
  builtIn: boolean;
}

export interface StepResult {
  action: string;
  ok: boolean;
  output?: unknown;
  error?: string;
  touched: RecordRef[];
}

export interface AutomationRun extends Entity {
  automation: string;
  startedAt: string;
  finishedAt?: string;
  status: RunStatus;
  dryRun: boolean;
  triggerData?: unknown;
  steps: StepResult[];
}

export interface NotificationChannel extends Entity {
  kind: ChannelKind;
  name: string;
  config?: unknown;
  enabled: boolean;
}

export interface Notification extends Entity {
  at: string;
  kind: string;
  title: string;
  body?: string;
  subject?: RecordRef;
  read: boolean;
}

export interface Stream extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  source: string;
  url: string;
  mimeType: string;
  quality?: string;
  height?: number;
  headers: { [key in string]: string };
  expiresAt?: string;
  speedKbps?: number;
}

export interface StorageLocation extends Entity {
  device: string;
  kind: StorageKind;
  uri: string;
  quotaBytes?: number;
  usedBytes: number;
}

export interface LibraryRoot extends Entity {
  location: string;
  path: string;
  scanIntervalMinutes?: number;
  lastScanAt?: string;
  organizeTemplate?: string;
}

export interface MediaFile extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  location: string;
  path: string;
  sizeBytes: number;
  mimeType: string;
  quality?: string;
  width?: number;
  height?: number;
  durationSeconds?: number;
  libraryRoot?: string;
  missing: boolean;
}

export interface Segment extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  startSeconds: number;
  endSeconds?: number;
  title?: string;
  tag?: string;
  performers: string[];
  origin: SegmentOrigin;
}

export interface Clip extends Entity {
  video: string;
  startSeconds: number;
  endSeconds: number;
  title?: string;
  cover?: Asset;
  file?: string;
}

export interface Subtitle extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  language: string;
  format: string;
  asset: Asset;
  origin: SubtitleOrigin;
}

export interface Script extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  kind: ScriptKind;
  asset: Asset;
  axis?: string;
  writtenBy: Actor;
}

export interface FieldSource extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  fieldOrdinal: number;
  writtenBy: Actor;
  source?: string;
  locked: boolean;
}

export interface Revision extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  fieldOrdinal: number;
  before?: unknown;
  after?: unknown;
  writtenBy: Actor;
  actorName?: string;
  automationRunId?: string;
}

export interface Fingerprint extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  fileId?: string;
  algorithm: FingerprintAlgorithm;
  value: string;
  durationSeconds?: number;
}

export interface Embedding extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  kind: EmbeddingKind;
  model: string;
  vector: Uint8Array;
  atSeconds?: number;
  region?: unknown;
  scan?: string;
}

export interface FaceScan extends Entity {
  subjectKind: RecordKind;
  subjectId: string;
  model: string;
  picture?: string;
  frames: number;
  faces: number;
}

export interface Alike extends Entity {
  subjectKind: RecordKind;
  firstId: string;
  secondId: string;
  score: number;
  reason?: string;
  dismissed: boolean;
}

export interface Comment extends Entity {
  source: string;
  author?: string;
  text: string;
  postedAt?: string;
  likes?: number;
  replyTo?: string;
  mentions: string[];
}
