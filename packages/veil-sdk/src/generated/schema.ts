// @ts-nocheck
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type Scalars = {
    String: string,
    Int: number,
    ID: string,
    Boolean: boolean,
    Float: number,
    JSON: any,
}

export interface Query {
    play: (Playable | null)
    videos: Video[]
    videoCount: Scalars['Int']
    video: (Video | null)
    galleries: Gallery[]
    galleryCount: Scalars['Int']
    gallery: (Gallery | null)
    performers: Performer[]
    performer: (Performer | null)
    studios: Studio[]
    studio: (Studio | null)
    tags: Tag[]
    tag: (Tag | null)
    plugins: Plugin[]
    plugin: (Plugin | null)
    pluginPackages: PluginPackage[]
    trash: TrashedRecord[]
    jobs: Job[]
    job: (Job | null)
    devices: Device[]
    currentDevice: Device
    storageLocations: StorageLocation[]
    userStates: UserState[]
    history: UserState[]
    events: Event[]
    eventCount: Scalars['Int']
    recentSearches: Scalars['String'][]
    savedFilters: SavedFilter[]
    savedFilter: (SavedFilter | null)
    __typename: 'Query'
}

export interface Mutation {
    scrape: Scraped
    fetchVideo: (Video | null)
    fetchGallery: (Gallery | null)
    browse: Listed
    search: Listed
    editVideo: (Video | null)
    editGallery: (Gallery | null)
    editPerformer: (Performer | null)
    editStudio: (Studio | null)
    editTag: (Tag | null)
    lockField: FieldSource[]
    revertRevision: Scalars['Boolean']
    setPluginEnabled: Scalars['Boolean']
    updatePluginSettings: Scalars['Boolean']
    installPlugin: Scalars['Boolean']
    uninstallPlugin: Scalars['Boolean']
    updatePlugins: Scalars['String'][]
    trashRecord: Scalars['Boolean']
    restoreRecord: Scalars['Boolean']
    purgeRecord: Scalars['Boolean']
    emptyTrash: Scalars['Int']
    queueFetch: Job
    cancelJob: (Job | null)
    retryJob: (Job | null)
    clearFinishedJobs: Scalars['Int']
    renameDevice: Device
    addStorageLocation: StorageLocation
    setStorageQuota: StorageLocation
    removeStorageLocation: Scalars['Boolean']
    rate: UserState
    setFavorite: UserState
    setBlocked: UserState
    setNotes: UserState
    recordEvent: Event
    clearEvents: Scalars['Int']
    saveFilter: SavedFilter
    deleteSavedFilter: Scalars['Boolean']
    __typename: 'Mutation'
}

export interface Playable {
    url: Scalars['String']
    mimeType: Scalars['String']
    quality: (Scalars['String'] | null)
    __typename: 'Playable'
}

export interface Scraped {
    kind: Scalars['String']
    id: Scalars['ID']
    __typename: 'Scraped'
}

export interface FieldSource {
    field: Scalars['String']
    writtenBy: Actor
    source: (Source | null)
    locked: Scalars['Boolean']
    updatedAt: Scalars['String']
    __typename: 'FieldSource'
}

export interface Revision {
    id: Scalars['ID']
    createdAt: Scalars['String']
    subjectKind: RecordKind
    subjectId: Scalars['ID']
    field: Scalars['String']
    before: (Scalars['JSON'] | null)
    after: (Scalars['JSON'] | null)
    writtenBy: Actor
    actorName: (Scalars['String'] | null)
    __typename: 'Revision'
}

export type Actor = 'user' | 'plugin' | 'automation' | 'system'

export type ListKind = 'VIDEO' | 'GALLERY' | 'PERFORMER' | 'STUDIO' | 'TAG'

export interface Listed {
    videos: Video[]
    galleries: Gallery[]
    performers: Performer[]
    studios: Studio[]
    tags: Tag[]
    __typename: 'Listed'
}

export interface Asset {
    uri: Scalars['String']
    mimeType: (Scalars['String'] | null)
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    blurhash: (Scalars['String'] | null)
    __typename: 'Asset'
}

export interface Video {
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    status: Scalars['String']
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    durationSeconds: (Scalars['Float'] | null)
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    vr: (VrFormat | null)
    preview: (Asset | null)
    trickplay: (Trickplay | null)
    gallery: (Gallery | null)
    sources: Source[]
    streams: Stream[]
    userState: UserState
    oCount: Scalars['Int']
    fieldSources: FieldSource[]
    revisions: Revision[]
    __typename: 'Video'
}

export interface VrFormat {
    projection: Scalars['String']
    stereo: Scalars['String']
    fieldOfView: (Scalars['Int'] | null)
    __typename: 'VrFormat'
}

export interface Trickplay {
    sheets: Asset[]
    columns: Scalars['Int']
    rows: Scalars['Int']
    tileWidth: Scalars['Int']
    tileHeight: Scalars['Int']
    intervalSeconds: Scalars['Float']
    __typename: 'Trickplay'
}

export interface Gallery {
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    status: Scalars['String']
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    imageCount: Scalars['Int']
    images: Image[]
    video: (Video | null)
    sources: Source[]
    userState: UserState
    oCount: Scalars['Int']
    fieldSources: FieldSource[]
    revisions: Revision[]
    __typename: 'Gallery'
}

export interface Image {
    id: Scalars['ID']
    position: Scalars['Int']
    asset: Asset
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    __typename: 'Image'
}

export interface Performer {
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    name: Scalars['String']
    aliases: Scalars['String'][]
    image: (Asset | null)
    details: (Scalars['String'] | null)
    attributes: PerformerAttributes
    links: Scalars['String'][]
    tags: Tag[]
    sources: Source[]
    videoCount: Scalars['Int']
    galleryCount: Scalars['Int']
    userState: UserState
    fieldSources: FieldSource[]
    revisions: Revision[]
    __typename: 'Performer'
}

export interface PerformerAttributes {
    gender: (Scalars['String'] | null)
    birthdate: (Scalars['String'] | null)
    deathDate: (Scalars['String'] | null)
    country: (Scalars['String'] | null)
    ethnicity: (Scalars['String'] | null)
    eyeColor: (Scalars['String'] | null)
    hairColor: (Scalars['String'] | null)
    heightCm: (Scalars['Int'] | null)
    weightKg: (Scalars['Int'] | null)
    measurements: (Scalars['String'] | null)
    fakeTits: (Scalars['String'] | null)
    tattoos: (Scalars['String'] | null)
    piercings: (Scalars['String'] | null)
    careerStart: (Scalars['Int'] | null)
    careerEnd: (Scalars['Int'] | null)
    __typename: 'PerformerAttributes'
}

export interface Studio {
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    name: Scalars['String']
    aliases: Scalars['String'][]
    image: (Asset | null)
    details: (Scalars['String'] | null)
    url: (Scalars['String'] | null)
    parent: (Studio | null)
    tags: Tag[]
    sources: Source[]
    videoCount: Scalars['Int']
    galleryCount: Scalars['Int']
    userState: UserState
    fieldSources: FieldSource[]
    revisions: Revision[]
    __typename: 'Studio'
}

export interface Tag {
    id: Scalars['ID']
    name: Scalars['String']
    aliases: Scalars['String'][]
    videoCount: Scalars['Int']
    galleryCount: Scalars['Int']
    userState: UserState
    revisions: Revision[]
    __typename: 'Tag'
}

export interface Source {
    id: Scalars['ID']
    plugin: Scalars['String']
    externalId: Scalars['String']
    url: Scalars['String']
    title: (Scalars['String'] | null)
    image: (Asset | null)
    fetchedAt: (Scalars['String'] | null)
    __typename: 'Source'
}

export interface Stream {
    id: Scalars['ID']
    url: Scalars['String']
    mimeType: Scalars['String']
    quality: (Scalars['String'] | null)
    height: (Scalars['Int'] | null)
    speedKbps: (Scalars['Int'] | null)
    source: (Source | null)
    __typename: 'Stream'
}

export interface PluginPackage {
    name: Scalars['String']
    version: Scalars['String']
    description: (Scalars['String'] | null)
    installedVersion: (Scalars['String'] | null)
    __typename: 'PluginPackage'
}

export interface Plugin {
    name: Scalars['String']
    packageName: (Scalars['String'] | null)
    localBuild: Scalars['Boolean']
    displayName: (Scalars['String'] | null)
    iconUrl: (Scalars['String'] | null)
    description: (Scalars['String'] | null)
    version: Scalars['String']
    capabilities: Scalars['String'][]
    domains: Scalars['String'][]
    content: Scalars['String'][]
    enabled: Scalars['Boolean']
    requiresSolver: Scalars['Boolean']
    available: Scalars['Boolean']
    settings: PluginSettingField[]
    settingValues: PluginSettingValue[]
    __typename: 'Plugin'
}

export interface PluginSettingField {
    key: Scalars['String']
    label: Scalars['String']
    description: (Scalars['String'] | null)
    type: Scalars['String']
    required: Scalars['Boolean']
    default: (Scalars['String'] | null)
    __typename: 'PluginSettingField'
}

export interface PluginSettingValue {
    key: Scalars['String']
    value: Scalars['String']
    __typename: 'PluginSettingValue'
}

export interface TrashedRecord {
    kind: RecordKind
    id: Scalars['ID']
    label: Scalars['String']
    deletedAt: Scalars['String']
    __typename: 'TrashedRecord'
}

export type JobKind = 'download' | 'record' | 'scrape' | 'followRun' | 'fingerprint' | 'embed' | 'trickplay' | 'transcode' | 'libraryScan' | 'automationRun' | 'snapshot' | 'enrich' | 'speedCheck'

export type JobStatus = 'queued' | 'running' | 'succeeded' | 'failed' | 'cancelled'

export interface Job {
    id: Scalars['ID']
    kind: JobKind
    status: JobStatus
    subject: (RecordRef | null)
    plugin: (Scalars['String'] | null)
    payload: (Scalars['JSON'] | null)
    progress: Scalars['Float']
    error: (Scalars['String'] | null)
    attempts: Scalars['Int']
    priority: Scalars['Int']
    device: (Scalars['ID'] | null)
    createdAt: Scalars['String']
    runAt: (Scalars['String'] | null)
    startedAt: (Scalars['String'] | null)
    finishedAt: (Scalars['String'] | null)
    __typename: 'Job'
}

export type Platform = 'server' | 'desktop' | 'android' | 'tv' | 'vr'

export interface Device {
    id: Scalars['ID']
    name: Scalars['String']
    platform: Platform
    lastSeenAt: (Scalars['String'] | null)
    lastSyncAt: (Scalars['String'] | null)
    current: Scalars['Boolean']
    createdAt: Scalars['String']
    __typename: 'Device'
}

export type StorageKind = 'library' | 'cache' | 'archive'

export interface StorageLocation {
    id: Scalars['ID']
    device: (Device | null)
    kind: StorageKind
    uri: Scalars['String']
    quotaBytes: (Scalars['Int'] | null)
    usedBytes: Scalars['Int']
    __typename: 'StorageLocation'
}

export type RecordKind = 'video' | 'gallery' | 'image' | 'audio' | 'story' | 'post' | 'performer' | 'studio' | 'camModel' | 'group' | 'tag' | 'category' | 'collection' | 'segment' | 'clip'

export type MediaKind = 'video' | 'gallery' | 'audio' | 'story' | 'post'

export type MediaStatus = 'available' | 'wanted' | 'unavailable'

export interface RecordRef {
    kind: RecordKind
    id: Scalars['ID']
    __typename: 'RecordRef'
}

export type Record = (Video | Gallery | Performer | Studio | Tag) & { __isUnion?: true }

export interface UserState {
    subjectKind: RecordKind
    subjectId: Scalars['ID']
    record: (Record | null)
    rating: (Scalars['Float'] | null)
    favorite: Scalars['Boolean']
    notes: (Scalars['String'] | null)
    blocked: Scalars['Boolean']
    progress: (Scalars['Float'] | null)
    viewCount: Scalars['Int']
    lastViewedAt: (Scalars['String'] | null)
    fields: Scalars['JSON']
    updatedAt: (Scalars['String'] | null)
    __typename: 'UserState'
}

export type EventKind = 'view' | 'progress' | 'finish' | 'o' | 'search' | 'impression' | 'choice' | 'heatSpan' | 'skip' | 'like'

export interface Event {
    id: Scalars['ID']
    at: Scalars['String']
    kind: EventKind
    subjectKind: (RecordKind | null)
    subjectId: (Scalars['ID'] | null)
    data: (Scalars['JSON'] | null)
    device: (Scalars['ID'] | null)
    sessionId: (Scalars['ID'] | null)
    __typename: 'Event'
}

export type SortField = 'relevance' | 'date' | 'addedAt' | 'title' | 'rating' | 'duration' | 'viewCount' | 'lastViewedAt' | 'random'

export interface Sort {
    field: SortField
    descending: Scalars['Boolean']
    __typename: 'Sort'
}

export interface FieldCondition {
    fieldDefId: Scalars['ID']
    operator: Scalars['String']
    value: (Scalars['JSON'] | null)
    __typename: 'FieldCondition'
}

export interface Filter {
    kinds: MediaKind[]
    text: (Scalars['String'] | null)
    performerIds: (Scalars['ID'][] | null)
    studioIds: (Scalars['ID'][] | null)
    categoryIds: (Scalars['ID'][] | null)
    includeTagIds: (Scalars['ID'][] | null)
    excludeTagIds: (Scalars['ID'][] | null)
    sources: (Scalars['String'][] | null)
    dateFrom: (Scalars['String'] | null)
    dateTo: (Scalars['String'] | null)
    minDurationSeconds: (Scalars['Int'] | null)
    maxDurationSeconds: (Scalars['Int'] | null)
    minRating: (Scalars['Float'] | null)
    status: (MediaStatus[] | null)
    inLibrary: (Scalars['Boolean'] | null)
    favorite: (Scalars['Boolean'] | null)
    unwatched: (Scalars['Boolean'] | null)
    vr: (Scalars['Boolean'] | null)
    fields: (FieldCondition[] | null)
    sort: (Sort | null)
    __typename: 'Filter'
}

export interface SavedFilter {
    id: Scalars['ID']
    name: Scalars['String']
    filter: Filter
    pinned: Scalars['Boolean']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'SavedFilter'
}

export interface QueryGenqlSelection{
    play?: (PlayableGenqlSelection & { __args: {url: Scalars['String']} })
    videos?: (VideoGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    videoCount?: { __args: {filter?: (FilterInput | null)} } | boolean | number
    video?: (VideoGenqlSelection & { __args: {id: Scalars['ID']} })
    galleries?: (GalleryGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    galleryCount?: { __args: {filter?: (FilterInput | null)} } | boolean | number
    gallery?: (GalleryGenqlSelection & { __args: {id: Scalars['ID']} })
    performers?: (PerformerGenqlSelection & { __args?: {text?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    performer?: (PerformerGenqlSelection & { __args: {id: Scalars['ID']} })
    studios?: (StudioGenqlSelection & { __args?: {text?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    studio?: (StudioGenqlSelection & { __args: {id: Scalars['ID']} })
    tags?: (TagGenqlSelection & { __args?: {text?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    tag?: (TagGenqlSelection & { __args: {id: Scalars['ID']} })
    plugins?: PluginGenqlSelection
    plugin?: (PluginGenqlSelection & { __args: {name: Scalars['String']} })
    pluginPackages?: (PluginPackageGenqlSelection & { __args?: {query?: (Scalars['String'] | null)} })
    trash?: (TrashedRecordGenqlSelection & { __args: {kind: RecordKind, limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    jobs?: (JobGenqlSelection & { __args?: {statuses?: (JobStatus[] | null), kinds?: (JobKind[] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    job?: (JobGenqlSelection & { __args: {id: Scalars['ID']} })
    devices?: DeviceGenqlSelection
    currentDevice?: DeviceGenqlSelection
    storageLocations?: (StorageLocationGenqlSelection & { __args?: {deviceId?: (Scalars['ID'] | null)} })
    userStates?: (UserStateGenqlSelection & { __args?: {filter?: (UserStateFilter | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    history?: (UserStateGenqlSelection & { __args?: {kind?: (RecordKind | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    events?: (EventGenqlSelection & { __args?: {filter?: (EventFilter | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    eventCount?: { __args: {filter?: (EventFilter | null)} } | boolean | number
    recentSearches?: { __args: {limit?: (Scalars['Int'] | null)} } | boolean | number
    savedFilters?: SavedFilterGenqlSelection
    savedFilter?: (SavedFilterGenqlSelection & { __args: {id: Scalars['ID']} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MutationGenqlSelection{
    scrape?: (ScrapedGenqlSelection & { __args: {plugin: Scalars['String'], url: Scalars['String']} })
    fetchVideo?: (VideoGenqlSelection & { __args: {id: Scalars['ID']} })
    fetchGallery?: (GalleryGenqlSelection & { __args: {id: Scalars['ID']} })
    browse?: (ListedGenqlSelection & { __args: {plugin: Scalars['String'], kind: ListKind, query?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    search?: (ListedGenqlSelection & { __args: {kind: ListKind, query: Scalars['String'], limit?: (Scalars['Int'] | null)} })
    editVideo?: (VideoGenqlSelection & { __args: {id: Scalars['ID'], edit: MediaEdit} })
    editGallery?: (GalleryGenqlSelection & { __args: {id: Scalars['ID'], edit: MediaEdit} })
    editPerformer?: (PerformerGenqlSelection & { __args: {id: Scalars['ID'], edit: PerformerEdit} })
    editStudio?: (StudioGenqlSelection & { __args: {id: Scalars['ID'], edit: StudioEdit} })
    editTag?: (TagGenqlSelection & { __args: {id: Scalars['ID'], edit: TagEdit} })
    lockField?: (FieldSourceGenqlSelection & { __args: {subject: RecordRefInput, field: Scalars['String'], locked: Scalars['Boolean']} })
    revertRevision?: { __args: {id: Scalars['ID']} }
    setPluginEnabled?: { __args: {name: Scalars['String'], enabled: Scalars['Boolean']} }
    updatePluginSettings?: { __args: {plugin: Scalars['String'], values: PluginSettingValueInput[]} }
    installPlugin?: { __args: {packageName: Scalars['String']} }
    uninstallPlugin?: { __args: {name: Scalars['String']} }
    updatePlugins?: boolean | number
    trashRecord?: { __args: {subject: RecordRefInput} }
    restoreRecord?: { __args: {subject: RecordRefInput} }
    purgeRecord?: { __args: {subject: RecordRefInput} }
    emptyTrash?: boolean | number
    queueFetch?: (JobGenqlSelection & { __args: {subject: RecordRefInput} })
    cancelJob?: (JobGenqlSelection & { __args: {id: Scalars['ID']} })
    retryJob?: (JobGenqlSelection & { __args: {id: Scalars['ID']} })
    clearFinishedJobs?: boolean | number
    renameDevice?: (DeviceGenqlSelection & { __args: {id: Scalars['ID'], name: Scalars['String']} })
    addStorageLocation?: (StorageLocationGenqlSelection & { __args: {kind: StorageKind, uri: Scalars['String'], quotaBytes?: (Scalars['Int'] | null)} })
    setStorageQuota?: (StorageLocationGenqlSelection & { __args: {id: Scalars['ID'], quotaBytes?: (Scalars['Int'] | null)} })
    removeStorageLocation?: { __args: {id: Scalars['ID']} }
    rate?: (UserStateGenqlSelection & { __args: {subject: RecordRefInput, rating?: (Scalars['Float'] | null)} })
    setFavorite?: (UserStateGenqlSelection & { __args: {subject: RecordRefInput, favorite: Scalars['Boolean']} })
    setBlocked?: (UserStateGenqlSelection & { __args: {subject: RecordRefInput, blocked: Scalars['Boolean']} })
    setNotes?: (UserStateGenqlSelection & { __args: {subject: RecordRefInput, notes: Scalars['String']} })
    recordEvent?: (EventGenqlSelection & { __args: {event: EventInput} })
    clearEvents?: { __args: {kinds?: (EventKind[] | null)} } | boolean | number
    saveFilter?: (SavedFilterGenqlSelection & { __args: {id?: (Scalars['ID'] | null), name: Scalars['String'], filter: FilterInput, pinned?: (Scalars['Boolean'] | null)} })
    deleteSavedFilter?: { __args: {id: Scalars['ID']} }
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PlayableGenqlSelection{
    url?: boolean | number
    mimeType?: boolean | number
    quality?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ScrapedGenqlSelection{
    kind?: boolean | number
    id?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MediaEdit {title?: (Scalars['String'] | null),details?: (Scalars['String'] | null),date?: (Scalars['String'] | null),studioId?: (Scalars['ID'] | null),performerIds?: (Scalars['ID'][] | null),tagIds?: (Scalars['ID'][] | null)}

export interface PerformerEdit {name?: (Scalars['String'] | null),aliases?: (Scalars['String'][] | null),details?: (Scalars['String'] | null),links?: (Scalars['String'][] | null)}

export interface StudioEdit {name?: (Scalars['String'] | null),aliases?: (Scalars['String'][] | null),details?: (Scalars['String'] | null),url?: (Scalars['String'] | null),parentId?: (Scalars['ID'] | null)}

export interface TagEdit {name?: (Scalars['String'] | null),aliases?: (Scalars['String'][] | null)}

export interface FieldSourceGenqlSelection{
    field?: boolean | number
    writtenBy?: boolean | number
    source?: SourceGenqlSelection
    locked?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RevisionGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    subjectKind?: boolean | number
    subjectId?: boolean | number
    field?: boolean | number
    before?: boolean | number
    after?: boolean | number
    writtenBy?: boolean | number
    actorName?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ListedGenqlSelection{
    videos?: VideoGenqlSelection
    galleries?: GalleryGenqlSelection
    performers?: PerformerGenqlSelection
    studios?: StudioGenqlSelection
    tags?: TagGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface AssetGenqlSelection{
    uri?: boolean | number
    mimeType?: boolean | number
    width?: boolean | number
    height?: boolean | number
    blurhash?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface VideoGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    title?: boolean | number
    details?: boolean | number
    date?: boolean | number
    cover?: AssetGenqlSelection
    status?: boolean | number
    studio?: StudioGenqlSelection
    performers?: PerformerGenqlSelection
    tags?: TagGenqlSelection
    durationSeconds?: boolean | number
    width?: boolean | number
    height?: boolean | number
    vr?: VrFormatGenqlSelection
    preview?: AssetGenqlSelection
    trickplay?: TrickplayGenqlSelection
    gallery?: GalleryGenqlSelection
    sources?: SourceGenqlSelection
    streams?: StreamGenqlSelection
    userState?: UserStateGenqlSelection
    oCount?: boolean | number
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface VrFormatGenqlSelection{
    projection?: boolean | number
    stereo?: boolean | number
    fieldOfView?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TrickplayGenqlSelection{
    sheets?: AssetGenqlSelection
    columns?: boolean | number
    rows?: boolean | number
    tileWidth?: boolean | number
    tileHeight?: boolean | number
    intervalSeconds?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface GalleryGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    title?: boolean | number
    details?: boolean | number
    date?: boolean | number
    cover?: AssetGenqlSelection
    status?: boolean | number
    studio?: StudioGenqlSelection
    performers?: PerformerGenqlSelection
    tags?: TagGenqlSelection
    imageCount?: boolean | number
    images?: ImageGenqlSelection
    video?: VideoGenqlSelection
    sources?: SourceGenqlSelection
    userState?: UserStateGenqlSelection
    oCount?: boolean | number
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ImageGenqlSelection{
    id?: boolean | number
    position?: boolean | number
    asset?: AssetGenqlSelection
    width?: boolean | number
    height?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PerformerGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    name?: boolean | number
    aliases?: boolean | number
    image?: AssetGenqlSelection
    details?: boolean | number
    attributes?: PerformerAttributesGenqlSelection
    links?: boolean | number
    tags?: TagGenqlSelection
    sources?: SourceGenqlSelection
    videoCount?: boolean | number
    galleryCount?: boolean | number
    userState?: UserStateGenqlSelection
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PerformerAttributesGenqlSelection{
    gender?: boolean | number
    birthdate?: boolean | number
    deathDate?: boolean | number
    country?: boolean | number
    ethnicity?: boolean | number
    eyeColor?: boolean | number
    hairColor?: boolean | number
    heightCm?: boolean | number
    weightKg?: boolean | number
    measurements?: boolean | number
    fakeTits?: boolean | number
    tattoos?: boolean | number
    piercings?: boolean | number
    careerStart?: boolean | number
    careerEnd?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StudioGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    name?: boolean | number
    aliases?: boolean | number
    image?: AssetGenqlSelection
    details?: boolean | number
    url?: boolean | number
    parent?: StudioGenqlSelection
    tags?: TagGenqlSelection
    sources?: SourceGenqlSelection
    videoCount?: boolean | number
    galleryCount?: boolean | number
    userState?: UserStateGenqlSelection
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TagGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    aliases?: boolean | number
    videoCount?: boolean | number
    galleryCount?: boolean | number
    userState?: UserStateGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SourceGenqlSelection{
    id?: boolean | number
    plugin?: boolean | number
    externalId?: boolean | number
    url?: boolean | number
    title?: boolean | number
    image?: AssetGenqlSelection
    fetchedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StreamGenqlSelection{
    id?: boolean | number
    url?: boolean | number
    mimeType?: boolean | number
    quality?: boolean | number
    height?: boolean | number
    speedKbps?: boolean | number
    source?: SourceGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginPackageGenqlSelection{
    name?: boolean | number
    version?: boolean | number
    description?: boolean | number
    installedVersion?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginGenqlSelection{
    name?: boolean | number
    packageName?: boolean | number
    localBuild?: boolean | number
    displayName?: boolean | number
    iconUrl?: boolean | number
    description?: boolean | number
    version?: boolean | number
    capabilities?: boolean | number
    domains?: boolean | number
    content?: boolean | number
    enabled?: boolean | number
    requiresSolver?: boolean | number
    available?: boolean | number
    settings?: PluginSettingFieldGenqlSelection
    settingValues?: PluginSettingValueGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginSettingFieldGenqlSelection{
    key?: boolean | number
    label?: boolean | number
    description?: boolean | number
    type?: boolean | number
    required?: boolean | number
    default?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginSettingValueGenqlSelection{
    key?: boolean | number
    value?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginSettingValueInput {key: Scalars['String'],value: Scalars['String']}

export interface TrashedRecordGenqlSelection{
    kind?: boolean | number
    id?: boolean | number
    label?: boolean | number
    deletedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface JobGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    status?: boolean | number
    subject?: RecordRefGenqlSelection
    plugin?: boolean | number
    payload?: boolean | number
    progress?: boolean | number
    error?: boolean | number
    attempts?: boolean | number
    priority?: boolean | number
    device?: boolean | number
    createdAt?: boolean | number
    runAt?: boolean | number
    startedAt?: boolean | number
    finishedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface DeviceGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    platform?: boolean | number
    lastSeenAt?: boolean | number
    lastSyncAt?: boolean | number
    current?: boolean | number
    createdAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StorageLocationGenqlSelection{
    id?: boolean | number
    device?: DeviceGenqlSelection
    kind?: boolean | number
    uri?: boolean | number
    quotaBytes?: boolean | number
    usedBytes?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RecordRefInput {kind: RecordKind,id: Scalars['ID']}

export interface RecordRefGenqlSelection{
    kind?: boolean | number
    id?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RecordGenqlSelection{
    on_Video?:VideoGenqlSelection,
    on_Gallery?:GalleryGenqlSelection,
    on_Performer?:PerformerGenqlSelection,
    on_Studio?:StudioGenqlSelection,
    on_Tag?:TagGenqlSelection,
    __typename?: boolean | number
}

export interface UserStateGenqlSelection{
    subjectKind?: boolean | number
    subjectId?: boolean | number
    record?: RecordGenqlSelection
    rating?: boolean | number
    favorite?: boolean | number
    notes?: boolean | number
    blocked?: boolean | number
    progress?: boolean | number
    viewCount?: boolean | number
    lastViewedAt?: boolean | number
    fields?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UserStateFilter {kind?: (RecordKind | null),favorite?: (Scalars['Boolean'] | null),blocked?: (Scalars['Boolean'] | null),rated?: (Scalars['Boolean'] | null),viewed?: (Scalars['Boolean'] | null)}

export interface EventGenqlSelection{
    id?: boolean | number
    at?: boolean | number
    kind?: boolean | number
    subjectKind?: boolean | number
    subjectId?: boolean | number
    data?: boolean | number
    device?: boolean | number
    sessionId?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface EventInput {kind: EventKind,subject?: (RecordRefInput | null),data?: (Scalars['JSON'] | null),sessionId?: (Scalars['ID'] | null)}

export interface EventFilter {kinds?: (EventKind[] | null),subject?: (RecordRefInput | null),since?: (Scalars['String'] | null),until?: (Scalars['String'] | null)}

export interface SortGenqlSelection{
    field?: boolean | number
    descending?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SortInput {field: SortField,descending?: (Scalars['Boolean'] | null)}

export interface FieldConditionGenqlSelection{
    fieldDefId?: boolean | number
    operator?: boolean | number
    value?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FieldConditionInput {fieldDefId: Scalars['ID'],operator: Scalars['String'],value?: (Scalars['JSON'] | null)}

export interface FilterGenqlSelection{
    kinds?: boolean | number
    text?: boolean | number
    performerIds?: boolean | number
    studioIds?: boolean | number
    categoryIds?: boolean | number
    includeTagIds?: boolean | number
    excludeTagIds?: boolean | number
    sources?: boolean | number
    dateFrom?: boolean | number
    dateTo?: boolean | number
    minDurationSeconds?: boolean | number
    maxDurationSeconds?: boolean | number
    minRating?: boolean | number
    status?: boolean | number
    inLibrary?: boolean | number
    favorite?: boolean | number
    unwatched?: boolean | number
    vr?: boolean | number
    fields?: FieldConditionGenqlSelection
    sort?: SortGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FilterInput {kinds?: (MediaKind[] | null),text?: (Scalars['String'] | null),performerIds?: (Scalars['ID'][] | null),studioIds?: (Scalars['ID'][] | null),categoryIds?: (Scalars['ID'][] | null),includeTagIds?: (Scalars['ID'][] | null),excludeTagIds?: (Scalars['ID'][] | null),sources?: (Scalars['String'][] | null),dateFrom?: (Scalars['String'] | null),dateTo?: (Scalars['String'] | null),minDurationSeconds?: (Scalars['Int'] | null),maxDurationSeconds?: (Scalars['Int'] | null),minRating?: (Scalars['Float'] | null),status?: (MediaStatus[] | null),inLibrary?: (Scalars['Boolean'] | null),favorite?: (Scalars['Boolean'] | null),unwatched?: (Scalars['Boolean'] | null),vr?: (Scalars['Boolean'] | null),fields?: (FieldConditionInput[] | null),sort?: (SortInput | null)}

export interface SavedFilterGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    filter?: FilterGenqlSelection
    pinned?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}


    const Query_possibleTypes: string[] = ['Query']
    export const isQuery = (obj?: { __typename?: any } | null): obj is Query => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isQuery"')
      return Query_possibleTypes.includes(obj.__typename)
    }
    


    const Mutation_possibleTypes: string[] = ['Mutation']
    export const isMutation = (obj?: { __typename?: any } | null): obj is Mutation => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMutation"')
      return Mutation_possibleTypes.includes(obj.__typename)
    }
    


    const Playable_possibleTypes: string[] = ['Playable']
    export const isPlayable = (obj?: { __typename?: any } | null): obj is Playable => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPlayable"')
      return Playable_possibleTypes.includes(obj.__typename)
    }
    


    const Scraped_possibleTypes: string[] = ['Scraped']
    export const isScraped = (obj?: { __typename?: any } | null): obj is Scraped => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isScraped"')
      return Scraped_possibleTypes.includes(obj.__typename)
    }
    


    const FieldSource_possibleTypes: string[] = ['FieldSource']
    export const isFieldSource = (obj?: { __typename?: any } | null): obj is FieldSource => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFieldSource"')
      return FieldSource_possibleTypes.includes(obj.__typename)
    }
    


    const Revision_possibleTypes: string[] = ['Revision']
    export const isRevision = (obj?: { __typename?: any } | null): obj is Revision => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRevision"')
      return Revision_possibleTypes.includes(obj.__typename)
    }
    


    const Listed_possibleTypes: string[] = ['Listed']
    export const isListed = (obj?: { __typename?: any } | null): obj is Listed => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isListed"')
      return Listed_possibleTypes.includes(obj.__typename)
    }
    


    const Asset_possibleTypes: string[] = ['Asset']
    export const isAsset = (obj?: { __typename?: any } | null): obj is Asset => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAsset"')
      return Asset_possibleTypes.includes(obj.__typename)
    }
    


    const Video_possibleTypes: string[] = ['Video']
    export const isVideo = (obj?: { __typename?: any } | null): obj is Video => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isVideo"')
      return Video_possibleTypes.includes(obj.__typename)
    }
    


    const VrFormat_possibleTypes: string[] = ['VrFormat']
    export const isVrFormat = (obj?: { __typename?: any } | null): obj is VrFormat => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isVrFormat"')
      return VrFormat_possibleTypes.includes(obj.__typename)
    }
    


    const Trickplay_possibleTypes: string[] = ['Trickplay']
    export const isTrickplay = (obj?: { __typename?: any } | null): obj is Trickplay => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTrickplay"')
      return Trickplay_possibleTypes.includes(obj.__typename)
    }
    


    const Gallery_possibleTypes: string[] = ['Gallery']
    export const isGallery = (obj?: { __typename?: any } | null): obj is Gallery => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isGallery"')
      return Gallery_possibleTypes.includes(obj.__typename)
    }
    


    const Image_possibleTypes: string[] = ['Image']
    export const isImage = (obj?: { __typename?: any } | null): obj is Image => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isImage"')
      return Image_possibleTypes.includes(obj.__typename)
    }
    


    const Performer_possibleTypes: string[] = ['Performer']
    export const isPerformer = (obj?: { __typename?: any } | null): obj is Performer => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPerformer"')
      return Performer_possibleTypes.includes(obj.__typename)
    }
    


    const PerformerAttributes_possibleTypes: string[] = ['PerformerAttributes']
    export const isPerformerAttributes = (obj?: { __typename?: any } | null): obj is PerformerAttributes => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPerformerAttributes"')
      return PerformerAttributes_possibleTypes.includes(obj.__typename)
    }
    


    const Studio_possibleTypes: string[] = ['Studio']
    export const isStudio = (obj?: { __typename?: any } | null): obj is Studio => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStudio"')
      return Studio_possibleTypes.includes(obj.__typename)
    }
    


    const Tag_possibleTypes: string[] = ['Tag']
    export const isTag = (obj?: { __typename?: any } | null): obj is Tag => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTag"')
      return Tag_possibleTypes.includes(obj.__typename)
    }
    


    const Source_possibleTypes: string[] = ['Source']
    export const isSource = (obj?: { __typename?: any } | null): obj is Source => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSource"')
      return Source_possibleTypes.includes(obj.__typename)
    }
    


    const Stream_possibleTypes: string[] = ['Stream']
    export const isStream = (obj?: { __typename?: any } | null): obj is Stream => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStream"')
      return Stream_possibleTypes.includes(obj.__typename)
    }
    


    const PluginPackage_possibleTypes: string[] = ['PluginPackage']
    export const isPluginPackage = (obj?: { __typename?: any } | null): obj is PluginPackage => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginPackage"')
      return PluginPackage_possibleTypes.includes(obj.__typename)
    }
    


    const Plugin_possibleTypes: string[] = ['Plugin']
    export const isPlugin = (obj?: { __typename?: any } | null): obj is Plugin => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPlugin"')
      return Plugin_possibleTypes.includes(obj.__typename)
    }
    


    const PluginSettingField_possibleTypes: string[] = ['PluginSettingField']
    export const isPluginSettingField = (obj?: { __typename?: any } | null): obj is PluginSettingField => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginSettingField"')
      return PluginSettingField_possibleTypes.includes(obj.__typename)
    }
    


    const PluginSettingValue_possibleTypes: string[] = ['PluginSettingValue']
    export const isPluginSettingValue = (obj?: { __typename?: any } | null): obj is PluginSettingValue => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginSettingValue"')
      return PluginSettingValue_possibleTypes.includes(obj.__typename)
    }
    


    const TrashedRecord_possibleTypes: string[] = ['TrashedRecord']
    export const isTrashedRecord = (obj?: { __typename?: any } | null): obj is TrashedRecord => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTrashedRecord"')
      return TrashedRecord_possibleTypes.includes(obj.__typename)
    }
    


    const Job_possibleTypes: string[] = ['Job']
    export const isJob = (obj?: { __typename?: any } | null): obj is Job => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isJob"')
      return Job_possibleTypes.includes(obj.__typename)
    }
    


    const Device_possibleTypes: string[] = ['Device']
    export const isDevice = (obj?: { __typename?: any } | null): obj is Device => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isDevice"')
      return Device_possibleTypes.includes(obj.__typename)
    }
    


    const StorageLocation_possibleTypes: string[] = ['StorageLocation']
    export const isStorageLocation = (obj?: { __typename?: any } | null): obj is StorageLocation => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStorageLocation"')
      return StorageLocation_possibleTypes.includes(obj.__typename)
    }
    


    const RecordRef_possibleTypes: string[] = ['RecordRef']
    export const isRecordRef = (obj?: { __typename?: any } | null): obj is RecordRef => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecordRef"')
      return RecordRef_possibleTypes.includes(obj.__typename)
    }
    


    const Record_possibleTypes: string[] = ['Video','Gallery','Performer','Studio','Tag']
    export const isRecord = (obj?: { __typename?: any } | null): obj is Record => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecord"')
      return Record_possibleTypes.includes(obj.__typename)
    }
    


    const UserState_possibleTypes: string[] = ['UserState']
    export const isUserState = (obj?: { __typename?: any } | null): obj is UserState => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isUserState"')
      return UserState_possibleTypes.includes(obj.__typename)
    }
    


    const Event_possibleTypes: string[] = ['Event']
    export const isEvent = (obj?: { __typename?: any } | null): obj is Event => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isEvent"')
      return Event_possibleTypes.includes(obj.__typename)
    }
    


    const Sort_possibleTypes: string[] = ['Sort']
    export const isSort = (obj?: { __typename?: any } | null): obj is Sort => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSort"')
      return Sort_possibleTypes.includes(obj.__typename)
    }
    


    const FieldCondition_possibleTypes: string[] = ['FieldCondition']
    export const isFieldCondition = (obj?: { __typename?: any } | null): obj is FieldCondition => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFieldCondition"')
      return FieldCondition_possibleTypes.includes(obj.__typename)
    }
    


    const Filter_possibleTypes: string[] = ['Filter']
    export const isFilter = (obj?: { __typename?: any } | null): obj is Filter => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFilter"')
      return Filter_possibleTypes.includes(obj.__typename)
    }
    


    const SavedFilter_possibleTypes: string[] = ['SavedFilter']
    export const isSavedFilter = (obj?: { __typename?: any } | null): obj is SavedFilter => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSavedFilter"')
      return SavedFilter_possibleTypes.includes(obj.__typename)
    }
    

export const enumActor = {
   user: 'user' as const,
   plugin: 'plugin' as const,
   automation: 'automation' as const,
   system: 'system' as const
}

export const enumListKind = {
   VIDEO: 'VIDEO' as const,
   GALLERY: 'GALLERY' as const,
   PERFORMER: 'PERFORMER' as const,
   STUDIO: 'STUDIO' as const,
   TAG: 'TAG' as const
}

export const enumJobKind = {
   download: 'download' as const,
   record: 'record' as const,
   scrape: 'scrape' as const,
   followRun: 'followRun' as const,
   fingerprint: 'fingerprint' as const,
   embed: 'embed' as const,
   trickplay: 'trickplay' as const,
   transcode: 'transcode' as const,
   libraryScan: 'libraryScan' as const,
   automationRun: 'automationRun' as const,
   snapshot: 'snapshot' as const,
   enrich: 'enrich' as const,
   speedCheck: 'speedCheck' as const
}

export const enumJobStatus = {
   queued: 'queued' as const,
   running: 'running' as const,
   succeeded: 'succeeded' as const,
   failed: 'failed' as const,
   cancelled: 'cancelled' as const
}

export const enumPlatform = {
   server: 'server' as const,
   desktop: 'desktop' as const,
   android: 'android' as const,
   tv: 'tv' as const,
   vr: 'vr' as const
}

export const enumStorageKind = {
   library: 'library' as const,
   cache: 'cache' as const,
   archive: 'archive' as const
}

export const enumRecordKind = {
   video: 'video' as const,
   gallery: 'gallery' as const,
   image: 'image' as const,
   audio: 'audio' as const,
   story: 'story' as const,
   post: 'post' as const,
   performer: 'performer' as const,
   studio: 'studio' as const,
   camModel: 'camModel' as const,
   group: 'group' as const,
   tag: 'tag' as const,
   category: 'category' as const,
   collection: 'collection' as const,
   segment: 'segment' as const,
   clip: 'clip' as const
}

export const enumMediaKind = {
   video: 'video' as const,
   gallery: 'gallery' as const,
   audio: 'audio' as const,
   story: 'story' as const,
   post: 'post' as const
}

export const enumMediaStatus = {
   available: 'available' as const,
   wanted: 'wanted' as const,
   unavailable: 'unavailable' as const
}

export const enumEventKind = {
   view: 'view' as const,
   progress: 'progress' as const,
   finish: 'finish' as const,
   o: 'o' as const,
   search: 'search' as const,
   impression: 'impression' as const,
   choice: 'choice' as const,
   heatSpan: 'heatSpan' as const,
   skip: 'skip' as const,
   like: 'like' as const
}

export const enumSortField = {
   relevance: 'relevance' as const,
   date: 'date' as const,
   addedAt: 'addedAt' as const,
   title: 'title' as const,
   rating: 'rating' as const,
   duration: 'duration' as const,
   viewCount: 'viewCount' as const,
   lastViewedAt: 'lastViewedAt' as const,
   random: 'random' as const
}
