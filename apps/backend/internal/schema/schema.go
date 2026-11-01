package schema

type CollectionKind int32

const (
	CollectionKindUser      CollectionKind = 1
	CollectionKindWatchlist CollectionKind = 2
	CollectionKindQueue     CollectionKind = 3
)

type SortField int32

const (
	SortFieldRelevance    SortField = 1
	SortFieldDate         SortField = 2
	SortFieldAddedAt      SortField = 3
	SortFieldTitle        SortField = 4
	SortFieldRating       SortField = 5
	SortFieldDuration     SortField = 6
	SortFieldViewCount    SortField = 7
	SortFieldLastViewedAt SortField = 8
	SortFieldRandom       SortField = 9
)

type EventKind int32

const (
	EventKindView       EventKind = 1
	EventKindProgress   EventKind = 2
	EventKindFinish     EventKind = 3
	EventKindO          EventKind = 4
	EventKindSearch     EventKind = 5
	EventKindImpression EventKind = 6
	EventKindChoice     EventKind = 7
	EventKindHeatSpan   EventKind = 8
	EventKindSkip       EventKind = 9
	EventKindLike       EventKind = 10
)

type LimitKind int32

const (
	LimitKindWatchTime LimitKind = 1
	LimitKindSessions  LimitKind = 2
)

type LimitWindow int32

const (
	LimitWindowDay  LimitWindow = 1
	LimitWindowWeek LimitWindow = 2
)

type LimitAction int32

const (
	LimitActionRemind LimitAction = 1
	LimitActionLock   LimitAction = 2
)

type AccountStatus int32

const (
	AccountStatusOk         AccountStatus = 1
	AccountStatusExpired    AccountStatus = 2
	AccountStatusNeedsLogin AccountStatus = 3
	AccountStatusFailed     AccountStatus = 4
)

type Platform int32

const (
	PlatformServer  Platform = 1
	PlatformDesktop Platform = 2
	PlatformAndroid Platform = 3
	PlatformTv      Platform = 4
	PlatformVr      Platform = 5
)

type JobKind int32

const (
	JobKindDownload      JobKind = 1
	JobKindRecord        JobKind = 2
	JobKindScrape        JobKind = 3
	JobKindFollowRun     JobKind = 4
	JobKindFingerprint   JobKind = 5
	JobKindEmbed         JobKind = 6
	JobKindTrickplay     JobKind = 7
	JobKindTranscode     JobKind = 8
	JobKindLibraryScan   JobKind = 9
	JobKindAutomationRun JobKind = 10
	JobKindSnapshot      JobKind = 11
)

type JobStatus int32

const (
	JobStatusQueued    JobStatus = 1
	JobStatusRunning   JobStatus = 2
	JobStatusSucceeded JobStatus = 3
	JobStatusFailed    JobStatus = 4
	JobStatusCancelled JobStatus = 5
)

type ContributionKind int32

const (
	ContributionKindSource  ContributionKind = 1
	ContributionKindLookup  ContributionKind = 2
	ContributionKindStream  ContributionKind = 3
	ContributionKindCams    ContributionKind = 4
	ContributionKindTrigger ContributionKind = 5
	ContributionKindAction  ContributionKind = 6
	ContributionKindFeed    ContributionKind = 7
	ContributionKindField   ContributionKind = 8
	ContributionKindLogin   ContributionKind = 9
	ContributionKindView    ContributionKind = 10
)

type PluginOrigin int32

const (
	PluginOriginStore PluginOrigin = 1
	PluginOriginLocal PluginOrigin = 2
)

type RecordKind int32

const (
	RecordKindVideo      RecordKind = 1
	RecordKindGallery    RecordKind = 2
	RecordKindImage      RecordKind = 3
	RecordKindAudio      RecordKind = 4
	RecordKindStory      RecordKind = 5
	RecordKindPost       RecordKind = 6
	RecordKindPerformer  RecordKind = 7
	RecordKindStudio     RecordKind = 8
	RecordKindCamModel   RecordKind = 9
	RecordKindGroup      RecordKind = 10
	RecordKindTag        RecordKind = 11
	RecordKindCategory   RecordKind = 12
	RecordKindCollection RecordKind = 13
	RecordKindSegment    RecordKind = 14
	RecordKindClip       RecordKind = 15
)

type MediaKind int32

const (
	MediaKindVideo   MediaKind = 1
	MediaKindGallery MediaKind = 2
	MediaKindAudio   MediaKind = 3
	MediaKindStory   MediaKind = 4
	MediaKindPost    MediaKind = 5
)

type Actor int32

const (
	ActorUser       Actor = 1
	ActorPlugin     Actor = 2
	ActorAutomation Actor = 3
	ActorSystem     Actor = 4
)

type MediaStatus int32

const (
	MediaStatusAvailable   MediaStatus = 1
	MediaStatusWanted      MediaStatus = 2
	MediaStatusUnavailable MediaStatus = 3
)

type VrProjection int32

const (
	VrProjectionEquirectangular180 VrProjection = 1
	VrProjectionEquirectangular360 VrProjection = 2
	VrProjectionFisheye            VrProjection = 3
	VrProjectionCubemap            VrProjection = 4
)

type VrStereo int32

const (
	VrStereoMono       VrStereo = 1
	VrStereoSideBySide VrStereo = 2
	VrStereoTopBottom  VrStereo = 3
)

type FieldType int32

const (
	FieldTypeText    FieldType = 1
	FieldTypeNumber  FieldType = 2
	FieldTypeBoolean FieldType = 3
	FieldTypeDate    FieldType = 4
	FieldTypeChoice  FieldType = 5
	FieldTypeChoices FieldType = 6
	FieldTypeUrl     FieldType = 7
)

type TriggerKind int32

const (
	TriggerKindEvent       TriggerKind = 1
	TriggerKindSchedule    TriggerKind = 2
	TriggerKindNewMedia    TriggerKind = 3
	TriggerKindCamLive     TriggerKind = 4
	TriggerKindFileAdded   TriggerKind = 5
	TriggerKindWantedFound TriggerKind = 6
	TriggerKindPlugin      TriggerKind = 7
)

type RunStatus int32

const (
	RunStatusRunning   RunStatus = 1
	RunStatusSucceeded RunStatus = 2
	RunStatusFailed    RunStatus = 3
	RunStatusUndone    RunStatus = 4
)

type ChannelKind int32

const (
	ChannelKindInApp   ChannelKind = 1
	ChannelKindPush    ChannelKind = 2
	ChannelKindNtfy    ChannelKind = 3
	ChannelKindWebhook ChannelKind = 4
)

type StorageKind int32

const (
	StorageKindLibrary StorageKind = 1
	StorageKindCache   StorageKind = 2
	StorageKindArchive StorageKind = 3
)

type SegmentOrigin int32

const (
	SegmentOriginUser   SegmentOrigin = 1
	SegmentOriginPlugin SegmentOrigin = 2
	SegmentOriginAi     SegmentOrigin = 3
)

type SubtitleOrigin int32

const (
	SubtitleOriginSite       SubtitleOrigin = 1
	SubtitleOriginGenerated  SubtitleOrigin = 2
	SubtitleOriginTranslated SubtitleOrigin = 3
)

type ScriptKind int32

const (
	ScriptKindFunscript ScriptKind = 1
)

type FingerprintAlgorithm int32

const (
	FingerprintAlgorithmPhash  FingerprintAlgorithm = 1
	FingerprintAlgorithmOshash FingerprintAlgorithm = 2
	FingerprintAlgorithmMd5    FingerprintAlgorithm = 3
	FingerprintAlgorithmAudio  FingerprintAlgorithm = 4
)

type EmbeddingKind int32

const (
	EmbeddingKindFace   EmbeddingKind = 1
	EmbeddingKindVisual EmbeddingKind = 2
	EmbeddingKindText   EmbeddingKind = 3
)

type Evidence struct {
	Plugin string  `json:"plugin"`
	Url    *string `json:"url,omitempty"`
	Title  *string `json:"title,omitempty"`
	Image  *Asset  `json:"image,omitempty"`
	Detail *string `json:"detail,omitempty"`
}

type Match struct {
	Item     RecordRef  `json:"item"`
	Score    *float32   `json:"score,omitempty"`
	Evidence []Evidence `json:"evidence"`
}

type LookupQuery struct {
	Text        *string   `json:"text,omitempty"`
	ImageBase64 *[]byte   `json:"imageBase64,omitempty"`
	MimeType    *string   `json:"mimeType,omitempty"`
	Plugins     *[]string `json:"plugins,omitempty"`
}

type UserState struct {
	Entity
	SubjectKind  RecordKind     `json:"subjectKind"`
	SubjectId    string         `json:"subjectId"`
	Rating       *float32       `json:"rating,omitempty"`
	Favorite     bool           `json:"favorite"`
	Notes        *string        `json:"notes,omitempty"`
	Blocked      bool           `json:"blocked"`
	Progress     *float64       `json:"progress,omitempty"`
	ViewCount    int32          `json:"viewCount"`
	LastViewedAt *string        `json:"lastViewedAt,omitempty"`
	Fields       map[string]any `json:"fields"`
}

type Collection struct {
	Entity
	Name    string         `json:"name"`
	Kind    CollectionKind `json:"kind"`
	Details *string        `json:"details,omitempty"`
	Cover   *Asset         `json:"cover,omitempty"`
}

type CollectionItem struct {
	Entity
	Collection  string     `json:"collection"`
	SubjectKind RecordKind `json:"subjectKind"`
	SubjectId   string     `json:"subjectId"`
	Position    int32      `json:"position"`
}

type Sort struct {
	Field      SortField `json:"field"`
	Descending bool      `json:"descending"`
}

type FieldCondition struct {
	FieldDefId string `json:"fieldDefId"`
	Operator   string `json:"operator"`
	Value      *any   `json:"value,omitempty"`
}

type Filter struct {
	Kinds              []MediaKind       `json:"kinds"`
	Text               *string           `json:"text,omitempty"`
	PerformerIds       *[]string         `json:"performerIds,omitempty"`
	StudioIds          *[]string         `json:"studioIds,omitempty"`
	CategoryIds        *[]string         `json:"categoryIds,omitempty"`
	IncludeTagIds      *[]string         `json:"includeTagIds,omitempty"`
	ExcludeTagIds      *[]string         `json:"excludeTagIds,omitempty"`
	Sources            *[]string         `json:"sources,omitempty"`
	DateFrom           *string           `json:"dateFrom,omitempty"`
	DateTo             *string           `json:"dateTo,omitempty"`
	MinDurationSeconds *int32            `json:"minDurationSeconds,omitempty"`
	MaxDurationSeconds *int32            `json:"maxDurationSeconds,omitempty"`
	MinRating          *float32          `json:"minRating,omitempty"`
	Status             *[]MediaStatus    `json:"status,omitempty"`
	InLibrary          *bool             `json:"inLibrary,omitempty"`
	Favorite           *bool             `json:"favorite,omitempty"`
	Unwatched          *bool             `json:"unwatched,omitempty"`
	Vr                 *bool             `json:"vr,omitempty"`
	Fields             *[]FieldCondition `json:"fields,omitempty"`
	Sort               *Sort             `json:"sort,omitempty"`
}

type SavedFilter struct {
	Entity
	Name   string `json:"name"`
	Filter Filter `json:"filter"`
	Pinned bool   `json:"pinned"`
}

type Follow struct {
	Entity
	Name          string      `json:"name"`
	Filter        Filter      `json:"filter"`
	TargetKind    *RecordKind `json:"targetKind,omitempty"`
	TargetId      *string     `json:"targetId,omitempty"`
	IntervalHours int32       `json:"intervalHours"`
	Enabled       bool        `json:"enabled"`
	Notify        bool        `json:"notify"`
	LastRunAt     *string     `json:"lastRunAt,omitempty"`
	NextRunAt     *string     `json:"nextRunAt,omitempty"`
	LastError     *string     `json:"lastError,omitempty"`
}

type Event struct {
	Entity
	At          string      `json:"at"`
	Kind        EventKind   `json:"kind"`
	SubjectKind *RecordKind `json:"subjectKind,omitempty"`
	SubjectId   *string     `json:"subjectId,omitempty"`
	Data        *any        `json:"data,omitempty"`
	Device      *string     `json:"device,omitempty"`
	SessionId   *string     `json:"sessionId,omitempty"`
}

type Limit struct {
	Entity
	Kind    LimitKind   `json:"kind"`
	Window  LimitWindow `json:"window"`
	Amount  int32       `json:"amount"`
	Action  LimitAction `json:"action"`
	Enabled bool        `json:"enabled"`
}

type Account struct {
	Entity
	Plugin        string        `json:"plugin"`
	Label         string        `json:"label"`
	Status        AccountStatus `json:"status"`
	Device        string        `json:"device"`
	Credentials   *[]byte       `json:"credentials,omitempty"`
	Session       *[]byte       `json:"session,omitempty"`
	ExpiresAt     *string       `json:"expiresAt,omitempty"`
	LastCheckedAt *string       `json:"lastCheckedAt,omitempty"`
}

type Device struct {
	Entity
	Name       string   `json:"name"`
	Platform   Platform `json:"platform"`
	LastSeenAt *string  `json:"lastSeenAt,omitempty"`
	LastSyncAt *string  `json:"lastSyncAt,omitempty"`
	Current    bool     `json:"current"`
}

type Job struct {
	Entity
	Kind       JobKind    `json:"kind"`
	Status     JobStatus  `json:"status"`
	Subject    *RecordRef `json:"subject,omitempty"`
	Payload    *any       `json:"payload,omitempty"`
	Progress   float32    `json:"progress"`
	Error      *string    `json:"error,omitempty"`
	Attempts   int32      `json:"attempts"`
	RunAt      string     `json:"runAt"`
	StartedAt  *string    `json:"startedAt,omitempty"`
	FinishedAt *string    `json:"finishedAt,omitempty"`
	Device     *string    `json:"device,omitempty"`
}

type ApiToken struct {
	Entity
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes"`
	TokenHash  []byte   `json:"tokenHash"`
	LastUsedAt *string  `json:"lastUsedAt,omitempty"`
	ExpiresAt  *string  `json:"expiresAt,omitempty"`
}

type Contribution struct {
	Kind       ContributionKind `json:"kind"`
	Name       string           `json:"name"`
	MediaKinds *[]MediaKind     `json:"mediaKinds,omitempty"`
	Domains    *[]string        `json:"domains,omitempty"`
}

type Plugin struct {
	Entity
	Name          string         `json:"name"`
	DisplayName   *string        `json:"displayName,omitempty"`
	Version       string         `json:"version"`
	Enabled       bool           `json:"enabled"`
	Origin        PluginOrigin   `json:"origin"`
	Contributions []Contribution `json:"contributions"`
	Settings      *any           `json:"settings,omitempty"`
}

type Entity struct {
	Id        string  `json:"id"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	DeletedAt *string `json:"deletedAt,omitempty"`
}

type RecordRef struct {
	Kind RecordKind `json:"kind"`
	Id   string     `json:"id"`
}

type Asset struct {
	Uri      string  `json:"uri"`
	MimeType *string `json:"mimeType,omitempty"`
	Width    *int32  `json:"width,omitempty"`
	Height   *int32  `json:"height,omitempty"`
	Blurhash *string `json:"blurhash,omitempty"`
}

type Source struct {
	Entity
	SubjectKind RecordKind `json:"subjectKind"`
	SubjectId   string     `json:"subjectId"`
	Plugin      string     `json:"plugin"`
	ExternalId  string     `json:"externalId"`
	Url         string     `json:"url"`
	Title       *string    `json:"title,omitempty"`
	Image       *Asset     `json:"image,omitempty"`
	Snapshot    *Asset     `json:"snapshot,omitempty"`
	FetchedAt   *string    `json:"fetchedAt,omitempty"`
}

type Media struct {
	Entity
	Title      string      `json:"title"`
	Details    *string     `json:"details,omitempty"`
	Date       *string     `json:"date,omitempty"`
	Cover      *Asset      `json:"cover,omitempty"`
	Studio     *string     `json:"studio,omitempty"`
	Performers []string    `json:"performers"`
	Tags       []string    `json:"tags"`
	Status     MediaStatus `json:"status"`
}

type VrFormat struct {
	Projection  VrProjection `json:"projection"`
	Stereo      VrStereo     `json:"stereo"`
	FieldOfView *int32       `json:"fieldOfView,omitempty"`
}

type Trickplay struct {
	Sheets          []Asset `json:"sheets"`
	Columns         int32   `json:"columns"`
	Rows            int32   `json:"rows"`
	TileWidth       int32   `json:"tileWidth"`
	TileHeight      int32   `json:"tileHeight"`
	IntervalSeconds float32 `json:"intervalSeconds"`
}

type Video struct {
	Media
	DurationSeconds *float64   `json:"durationSeconds,omitempty"`
	Width           *int32     `json:"width,omitempty"`
	Height          *int32     `json:"height,omitempty"`
	Vr              *VrFormat  `json:"vr,omitempty"`
	Preview         *Asset     `json:"preview,omitempty"`
	Trickplay       *Trickplay `json:"trickplay,omitempty"`
	Gallery         *string    `json:"gallery,omitempty"`
	RecordedFrom    *string    `json:"recordedFrom,omitempty"`
}

type GroupItem struct {
	Entity
	Group    string `json:"group"`
	Video    string `json:"video"`
	Position int32  `json:"position"`
}

type Gallery struct {
	Media
	ImageCount int32   `json:"imageCount"`
	Video      *string `json:"video,omitempty"`
}

type Image struct {
	Entity
	Gallery  string `json:"gallery"`
	Position int32  `json:"position"`
	Asset    Asset  `json:"asset"`
	Width    *int32 `json:"width,omitempty"`
	Height   *int32 `json:"height,omitempty"`
}

type Audio struct {
	Media
	DurationSeconds *float64 `json:"durationSeconds,omitempty"`
}

type Story struct {
	Media
	ChapterCount int32  `json:"chapterCount"`
	WordCount    *int32 `json:"wordCount,omitempty"`
	Complete     bool   `json:"complete"`
}

type StoryChapter struct {
	Entity
	Story     string  `json:"story"`
	Position  int32   `json:"position"`
	Title     *string `json:"title,omitempty"`
	Text      string  `json:"text"`
	WordCount int32   `json:"wordCount"`
}

type Post struct {
	Media
	Creator     string      `json:"creator"`
	Text        *string     `json:"text,omitempty"`
	Attachments []RecordRef `json:"attachments"`
	Locked      bool        `json:"locked"`
	PriceCents  *int64      `json:"priceCents,omitempty"`
	Currency    *string     `json:"currency,omitempty"`
}

type PerformerAttributes struct {
	Gender       *string `json:"gender,omitempty"`
	Birthdate    *string `json:"birthdate,omitempty"`
	DeathDate    *string `json:"deathDate,omitempty"`
	Country      *string `json:"country,omitempty"`
	Ethnicity    *string `json:"ethnicity,omitempty"`
	EyeColor     *string `json:"eyeColor,omitempty"`
	HairColor    *string `json:"hairColor,omitempty"`
	HeightCm     *int32  `json:"heightCm,omitempty"`
	WeightKg     *int32  `json:"weightKg,omitempty"`
	Measurements *string `json:"measurements,omitempty"`
	FakeTits     *string `json:"fakeTits,omitempty"`
	Tattoos      *string `json:"tattoos,omitempty"`
	Piercings    *string `json:"piercings,omitempty"`
	CareerStart  *int32  `json:"careerStart,omitempty"`
	CareerEnd    *int32  `json:"careerEnd,omitempty"`
}

type Performer struct {
	Entity
	Name       string              `json:"name"`
	Aliases    []string            `json:"aliases"`
	Image      *Asset              `json:"image,omitempty"`
	Details    *string             `json:"details,omitempty"`
	Attributes PerformerAttributes `json:"attributes"`
	Links      []string            `json:"links"`
	Tags       []string            `json:"tags"`
}

type Studio struct {
	Entity
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	Image   *Asset   `json:"image,omitempty"`
	Details *string  `json:"details,omitempty"`
	Parent  *string  `json:"parent,omitempty"`
	Url     *string  `json:"url,omitempty"`
	Tags    []string `json:"tags"`
}

type CamLive struct {
	Viewers   int32   `json:"viewers"`
	Subject   *string `json:"subject,omitempty"`
	StartedAt *string `json:"startedAt,omitempty"`
	Thumbnail *Asset  `json:"thumbnail,omitempty"`
}

type CamModel struct {
	Entity
	Plugin           string   `json:"plugin"`
	Username         string   `json:"username"`
	DisplayName      string   `json:"displayName"`
	Performer        *string  `json:"performer,omitempty"`
	Thumbnail        *Asset   `json:"thumbnail,omitempty"`
	Tags             []string `json:"tags"`
	Gender           *string  `json:"gender,omitempty"`
	Live             *CamLive `json:"live,omitempty"`
	LastSeenOnlineAt *string  `json:"lastSeenOnlineAt,omitempty"`
	AutoRecord       bool     `json:"autoRecord"`
	RecordQuality    *string  `json:"recordQuality,omitempty"`
}

type Group struct {
	Entity
	Title   string  `json:"title"`
	Details *string `json:"details,omitempty"`
	Date    *string `json:"date,omitempty"`
	Cover   *Asset  `json:"cover,omitempty"`
	Studio  *string `json:"studio,omitempty"`
}

type Tag struct {
	Entity
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

type Category struct {
	Entity
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	Parent    *string  `json:"parent,omitempty"`
	MatchTags []string `json:"matchTags"`
	Cover     *Asset   `json:"cover,omitempty"`
	Position  int32    `json:"position"`
}

type FieldDef struct {
	Entity
	Name      string       `json:"name"`
	Kinds     []RecordKind `json:"kinds"`
	FieldType FieldType    `json:"fieldType"`
	Options   *[]string    `json:"options,omitempty"`
}

type Trigger struct {
	Kind          TriggerKind `json:"kind"`
	Cron          *string     `json:"cron,omitempty"`
	EventKinds    *[]string   `json:"eventKinds,omitempty"`
	Filter        *Filter     `json:"filter,omitempty"`
	PluginTrigger *string     `json:"pluginTrigger,omitempty"`
}

type Step struct {
	Action          string  `json:"action"`
	Args            *any    `json:"args,omitempty"`
	Condition       *Filter `json:"condition,omitempty"`
	ContinueOnError bool    `json:"continueOnError"`
}

type Automation struct {
	Entity
	Name    string  `json:"name"`
	Enabled bool    `json:"enabled"`
	Trigger Trigger `json:"trigger"`
	Steps   []Step  `json:"steps"`
	BuiltIn bool    `json:"builtIn"`
}

type StepResult struct {
	Action  string      `json:"action"`
	Ok      bool        `json:"ok"`
	Output  *any        `json:"output,omitempty"`
	Error   *string     `json:"error,omitempty"`
	Touched []RecordRef `json:"touched"`
}

type AutomationRun struct {
	Entity
	Automation  string       `json:"automation"`
	StartedAt   string       `json:"startedAt"`
	FinishedAt  *string      `json:"finishedAt,omitempty"`
	Status      RunStatus    `json:"status"`
	DryRun      bool         `json:"dryRun"`
	TriggerData *any         `json:"triggerData,omitempty"`
	Steps       []StepResult `json:"steps"`
}

type NotificationChannel struct {
	Entity
	Kind    ChannelKind `json:"kind"`
	Name    string      `json:"name"`
	Config  *any        `json:"config,omitempty"`
	Enabled bool        `json:"enabled"`
}

type Notification struct {
	Entity
	At      string     `json:"at"`
	Kind    string     `json:"kind"`
	Title   string     `json:"title"`
	Body    *string    `json:"body,omitempty"`
	Subject *RecordRef `json:"subject,omitempty"`
	Read    bool       `json:"read"`
}

type Stream struct {
	Entity
	SubjectKind RecordKind        `json:"subjectKind"`
	SubjectId   string            `json:"subjectId"`
	Source      string            `json:"source"`
	Url         string            `json:"url"`
	MimeType    string            `json:"mimeType"`
	Quality     *string           `json:"quality,omitempty"`
	Height      *int32            `json:"height,omitempty"`
	Headers     map[string]string `json:"headers"`
	ExpiresAt   *string           `json:"expiresAt,omitempty"`
	SpeedKbps   *int32            `json:"speedKbps,omitempty"`
}

type StorageLocation struct {
	Entity
	Device     string      `json:"device"`
	Kind       StorageKind `json:"kind"`
	Uri        string      `json:"uri"`
	QuotaBytes *int64      `json:"quotaBytes,omitempty"`
	UsedBytes  int64       `json:"usedBytes"`
}

type LibraryRoot struct {
	Entity
	Location            string  `json:"location"`
	Path                string  `json:"path"`
	ScanIntervalMinutes *int32  `json:"scanIntervalMinutes,omitempty"`
	LastScanAt          *string `json:"lastScanAt,omitempty"`
	OrganizeTemplate    *string `json:"organizeTemplate,omitempty"`
}

type MediaFile struct {
	Entity
	SubjectKind     RecordKind `json:"subjectKind"`
	SubjectId       string     `json:"subjectId"`
	Location        string     `json:"location"`
	Path            string     `json:"path"`
	SizeBytes       int64      `json:"sizeBytes"`
	MimeType        string     `json:"mimeType"`
	Quality         *string    `json:"quality,omitempty"`
	Width           *int32     `json:"width,omitempty"`
	Height          *int32     `json:"height,omitempty"`
	DurationSeconds *float64   `json:"durationSeconds,omitempty"`
	LibraryRoot     *string    `json:"libraryRoot,omitempty"`
	Missing         bool       `json:"missing"`
}

type Segment struct {
	Entity
	SubjectKind  RecordKind    `json:"subjectKind"`
	SubjectId    string        `json:"subjectId"`
	StartSeconds float64       `json:"startSeconds"`
	EndSeconds   *float64      `json:"endSeconds,omitempty"`
	Title        *string       `json:"title,omitempty"`
	Tag          *string       `json:"tag,omitempty"`
	Performers   []string      `json:"performers"`
	Origin       SegmentOrigin `json:"origin"`
}

type Clip struct {
	Entity
	Video        string  `json:"video"`
	StartSeconds float64 `json:"startSeconds"`
	EndSeconds   float64 `json:"endSeconds"`
	Title        *string `json:"title,omitempty"`
	Cover        *Asset  `json:"cover,omitempty"`
	File         *string `json:"file,omitempty"`
}

type Subtitle struct {
	Entity
	SubjectKind RecordKind     `json:"subjectKind"`
	SubjectId   string         `json:"subjectId"`
	Language    string         `json:"language"`
	Format      string         `json:"format"`
	Asset       Asset          `json:"asset"`
	Origin      SubtitleOrigin `json:"origin"`
}

type Script struct {
	Entity
	SubjectKind RecordKind `json:"subjectKind"`
	SubjectId   string     `json:"subjectId"`
	Kind        ScriptKind `json:"kind"`
	Asset       Asset      `json:"asset"`
	Axis        *string    `json:"axis,omitempty"`
	WrittenBy   Actor      `json:"writtenBy"`
}

type FieldSource struct {
	Entity
	SubjectKind  RecordKind `json:"subjectKind"`
	SubjectId    string     `json:"subjectId"`
	FieldOrdinal uint32     `json:"fieldOrdinal"`
	WrittenBy    Actor      `json:"writtenBy"`
	Source       *string    `json:"source,omitempty"`
	Locked       bool       `json:"locked"`
}

type Revision struct {
	Entity
	SubjectKind     RecordKind `json:"subjectKind"`
	SubjectId       string     `json:"subjectId"`
	FieldOrdinal    uint32     `json:"fieldOrdinal"`
	Before          *any       `json:"before,omitempty"`
	After           *any       `json:"after,omitempty"`
	WrittenBy       Actor      `json:"writtenBy"`
	ActorName       *string    `json:"actorName,omitempty"`
	AutomationRunId *string    `json:"automationRunId,omitempty"`
}

type Fingerprint struct {
	Entity
	SubjectKind     RecordKind           `json:"subjectKind"`
	SubjectId       string               `json:"subjectId"`
	FileId          *string              `json:"fileId,omitempty"`
	Algorithm       FingerprintAlgorithm `json:"algorithm"`
	Value           string               `json:"value"`
	DurationSeconds *float64             `json:"durationSeconds,omitempty"`
}

type Embedding struct {
	Entity
	SubjectKind RecordKind    `json:"subjectKind"`
	SubjectId   string        `json:"subjectId"`
	Kind        EmbeddingKind `json:"kind"`
	Model       string        `json:"model"`
	Vector      []byte        `json:"vector"`
	AtSeconds   *float64      `json:"atSeconds,omitempty"`
	Region      *any          `json:"region,omitempty"`
}

type Alike struct {
	Entity
	SubjectKind RecordKind `json:"subjectKind"`
	FirstId     string     `json:"firstId"`
	SecondId    string     `json:"secondId"`
	Score       float32    `json:"score"`
	Reason      *string    `json:"reason,omitempty"`
	Dismissed   bool       `json:"dismissed"`
}

type Comment struct {
	Entity
	Source   string   `json:"source"`
	Author   *string  `json:"author,omitempty"`
	Text     string   `json:"text"`
	PostedAt *string  `json:"postedAt,omitempty"`
	Likes    *int32   `json:"likes,omitempty"`
	ReplyTo  *string  `json:"replyTo,omitempty"`
	Mentions []string `json:"mentions"`
}
