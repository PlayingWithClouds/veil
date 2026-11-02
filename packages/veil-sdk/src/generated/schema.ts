// @ts-nocheck
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type Scalars = {
    String: string,
    ID: string,
    Int: number,
    Boolean: boolean,
    Float: number,
    JSON: any,
}

export interface Query {
    play: (Playable | null)
    follows: Follow[]
    follow: (Follow | null)
    followOf: (Follow | null)
    newSince: Record[]
    notifications: Notification[]
    unreadNotificationCount: Scalars['Int']
    notificationChannels: NotificationChannel[]
    limits: Limit[]
    limitStatus: LimitStatus[]
    usage: Usage
    automations: Automation[]
    automation: (Automation | null)
    automationRuns: AutomationRun[]
    automationRun: (AutomationRun | null)
    automationActions: Scalars['String'][]
    camRooms: CamRoomPage
    camModels: CamModel[]
    camModel: (CamModel | null)
    activeRecordings: ActiveRecording[]
    categories: Category[]
    category: (Category | null)
    collections: Collection[]
    collection: (Collection | null)
    builtInCollection: Collection
    collectionsWith: Collection[]
    fieldDefs: FieldDef[]
    recommendations: Match[]
    related: Match[]
    shorts: Short[]
    taste: Taste
    bestMoment: (Scalars['Float'] | null)
    alikes: AlikePair[]
    libraryRoots: LibraryRoot[]
    libraryRoot: (LibraryRoot | null)
    planOrganize: FileMove[]
    downloads: Download[]
    sameMedia: Record[]
    nearest: Match[]
    similar: Match[]
    videos: Video[]
    videoCount: Scalars['Int']
    videoTagCounts: TagCount[]
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
    media: Record[]
    mediaCount: Scalars['Int']
    audios: Audio[]
    audioCount: Scalars['Int']
    audio: (Audio | null)
    stories: Story[]
    storyCount: Scalars['Int']
    story: (Story | null)
    posts: Post[]
    postCount: Scalars['Int']
    post: (Post | null)
    groups: Group[]
    group: (Group | null)
    settings: Scalars['JSON']
    apiTokens: ApiToken[]
    accounts: Account[]
    contributions: PluginContribution[]
    pluginView: PluginView
    clips: Clip[]
    clip: (Clip | null)
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
    followRecord: Follow
    followSearch: Follow
    createFollow: Follow
    updateFollow: Follow
    unfollow: Scalars['Boolean']
    runFollow: Job
    markNotificationsRead: Scalars['Boolean']
    markAllNotificationsRead: Scalars['Boolean']
    createNotificationChannel: NotificationChannel
    updateNotificationChannel: NotificationChannel
    deleteNotificationChannel: Scalars['Boolean']
    createLimit: Limit
    updateLimit: Limit
    deleteLimit: Scalars['Boolean']
    createAutomation: Automation
    updateAutomation: Automation
    deleteAutomation: Scalars['Boolean']
    runAutomation: AutomationRun
    undoAutomationRun: AutomationRun
    watchCam: Playable
    setCamAutoRecord: CamModel
    linkCamPerformer: CamModel
    startRecording: Job
    stopRecording: Scalars['Boolean']
    createCategory: Category
    editCategory: Category
    reorderCategories: Category[]
    createCollection: Collection
    editCollection: Collection
    addToCollection: Collection
    removeFromCollection: Collection
    moveInCollection: Collection
    clearCollection: Collection
    saveFieldDef: FieldDef
    deleteFieldDef: Scalars['Boolean']
    setFieldValue: UserState
    lookup: LookupResult
    merge: (Record | null)
    mergeAlike: (Record | null)
    dismissAlike: Scalars['Boolean']
    addLibraryRoot: LibraryRoot
    updateLibraryRoot: LibraryRoot
    removeLibraryRoot: Scalars['Boolean']
    scanLibraryRoot: Job
    applyOrganize: Scalars['Int']
    trashFile: Scalars['Boolean']
    download: Job
    deleteDownload: Scalars['Boolean']
    addFingerprint: Fingerprint
    fingerprintFile: Job
    snapshotSource: Job
    addEmbedding: Embedding
    removeEmbedding: Scalars['Boolean']
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
    editMedia: (Record | null)
    fetchMedia: (Record | null)
    readChapter: (StoryChapter | null)
    updateSettings: Scalars['JSON']
    createApiToken: CreatedApiToken
    revokeApiToken: Scalars['Boolean']
    loginAccount: Account
    checkAccount: Account
    logoutAccount: Scalars['Boolean']
    submitPluginView: PluginViewResult
    addSegment: Segment
    editSegment: Segment
    addClip: Clip
    editClip: Clip
    attachClipFile: Clip
    addSubtitle: Subtitle
    removeSubtitle: Scalars['Boolean']
    addScript: Script
    removeScript: Scalars['Boolean']
    setTrickplay: (Video | null)
    reportStreamSpeed: (Stream | null)
    refreshSources: (Record | null)
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
    recordEvents: Scalars['Int']
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

export interface Subscription {
    notificationAdded: Notification
    __typename: 'Subscription'
}

export interface Follow {
    id: Scalars['ID']
    name: Scalars['String']
    filter: Filter
    targetKind: (RecordKind | null)
    targetId: (Scalars['ID'] | null)
    target: (Record | null)
    intervalHours: Scalars['Int']
    enabled: Scalars['Boolean']
    notify: Scalars['Boolean']
    lastRunAt: (Scalars['String'] | null)
    nextRunAt: (Scalars['String'] | null)
    lastError: (Scalars['String'] | null)
    __typename: 'Follow'
}

export interface Notification {
    id: Scalars['ID']
    at: Scalars['String']
    kind: Scalars['String']
    title: Scalars['String']
    body: (Scalars['String'] | null)
    subject: (RecordRef | null)
    record: (Record | null)
    read: Scalars['Boolean']
    __typename: 'Notification'
}

export type ChannelKind = 'inApp' | 'push' | 'ntfy' | 'webhook'

export interface NotificationChannel {
    id: Scalars['ID']
    kind: ChannelKind
    name: Scalars['String']
    config: (Scalars['JSON'] | null)
    enabled: Scalars['Boolean']
    delivers: Scalars['Boolean']
    __typename: 'NotificationChannel'
}

export type LimitKind = 'watchTime' | 'sessions'

export type LimitWindow = 'day' | 'week'

export type LimitAction = 'remind' | 'lock'

export interface Limit {
    id: Scalars['ID']
    kind: LimitKind
    window: LimitWindow
    amount: Scalars['Int']
    action: LimitAction
    enabled: Scalars['Boolean']
    __typename: 'Limit'
}

export interface LimitStatus {
    limit: Limit
    used: Scalars['Float']
    reached: Scalars['Boolean']
    __typename: 'LimitStatus'
}

export interface Usage {
    since: Scalars['String']
    until: Scalars['String']
    watchMinutes: Scalars['Float']
    sessions: Scalars['Int']
    __typename: 'Usage'
}

export type TriggerKind = 'event' | 'schedule' | 'newMedia' | 'camLive' | 'fileAdded' | 'wantedFound' | 'plugin'

export interface Trigger {
    kind: TriggerKind
    cron: (Scalars['String'] | null)
    eventKinds: (Scalars['String'][] | null)
    filter: (Filter | null)
    pluginTrigger: (Scalars['String'] | null)
    __typename: 'Trigger'
}

export interface Step {
    action: Scalars['String']
    args: (Scalars['JSON'] | null)
    condition: (Filter | null)
    continueOnError: Scalars['Boolean']
    __typename: 'Step'
}

export interface Automation {
    id: Scalars['ID']
    name: Scalars['String']
    enabled: Scalars['Boolean']
    trigger: Trigger
    steps: Step[]
    builtIn: Scalars['Boolean']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Automation'
}

export type RunStatus = 'running' | 'succeeded' | 'failed' | 'undone'

export interface AutomationRun {
    id: Scalars['ID']
    automation: (Automation | null)
    startedAt: Scalars['String']
    finishedAt: (Scalars['String'] | null)
    status: RunStatus
    dryRun: Scalars['Boolean']
    triggerData: (Scalars['JSON'] | null)
    steps: StepResult[]
    __typename: 'AutomationRun'
}

export interface StepResult {
    action: Scalars['String']
    ok: Scalars['Boolean']
    output: (Scalars['JSON'] | null)
    error: (Scalars['String'] | null)
    touched: RecordRef[]
    __typename: 'StepResult'
}

export interface CamRoomPage {
    rooms: CamModel[]
    total: Scalars['Int']
    __typename: 'CamRoomPage'
}

export interface CamModel {
    id: Scalars['ID']
    plugin: Scalars['String']
    username: Scalars['String']
    displayName: Scalars['String']
    performer: (Performer | null)
    thumbnail: (Asset | null)
    tags: Scalars['String'][]
    gender: (Scalars['String'] | null)
    live: (CamLive | null)
    lastSeenOnlineAt: (Scalars['String'] | null)
    autoRecord: Scalars['Boolean']
    recordQuality: (Scalars['String'] | null)
    recordings: Video[]
    userState: UserState
    __typename: 'CamModel'
}

export interface CamLive {
    viewers: Scalars['Int']
    subject: (Scalars['String'] | null)
    startedAt: (Scalars['String'] | null)
    thumbnail: (Asset | null)
    __typename: 'CamLive'
}

export interface ActiveRecording {
    job: (Job | null)
    camModel: (CamModel | null)
    startedAt: Scalars['String']
    durationSeconds: Scalars['Float']
    sizeBytes: Scalars['Int']
    quality: (Scalars['String'] | null)
    __typename: 'ActiveRecording'
}

export interface Video {
    recordedFrom: (CamModel | null)
    categories: Category[]
    related: Match[]
    files: MediaFile[]
    fingerprints: Fingerprint[]
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    status: MediaStatus
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    durationSeconds: (Scalars['Float'] | null)
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    vr: (VrFormat | null)
    preview: (Asset | null)
    trickplay: (Trickplay | null)
    previewFrames: (Asset[] | null)
    gallery: (Gallery | null)
    sources: Source[]
    streams: Stream[]
    userState: UserState
    oCount: Scalars['Int']
    fieldSources: FieldSource[]
    revisions: Revision[]
    groups: Group[]
    segments: Segment[]
    subtitles: Subtitle[]
    scripts: Script[]
    clips: Clip[]
    comments: Comment[]
    __typename: 'Video'
}

export interface Category {
    id: Scalars['ID']
    name: Scalars['String']
    slug: Scalars['String']
    parent: (Category | null)
    children: Category[]
    matchTags: Tag[]
    cover: (Asset | null)
    position: Scalars['Int']
    videoCount: Scalars['Int']
    galleryCount: Scalars['Int']
    __typename: 'Category'
}

export type CollectionKind = 'user' | 'watchlist' | 'queue'

export interface Collection {
    id: Scalars['ID']
    name: Scalars['String']
    kind: CollectionKind
    details: (Scalars['String'] | null)
    cover: (Asset | null)
    itemCount: Scalars['Int']
    items: CollectionItem[]
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Collection'
}

export interface CollectionItem {
    subjectKind: RecordKind
    subjectId: Scalars['ID']
    record: (Record | null)
    position: Scalars['Int']
    addedAt: Scalars['String']
    __typename: 'CollectionItem'
}

export type FieldType = 'text' | 'number' | 'boolean' | 'date' | 'choice' | 'choices' | 'url'

export interface FieldDef {
    id: Scalars['ID']
    name: Scalars['String']
    kinds: RecordKind[]
    fieldType: FieldType
    options: (Scalars['String'][] | null)
    __typename: 'FieldDef'
}

export interface Gallery {
    categories: Category[]
    related: Match[]
    files: MediaFile[]
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    status: MediaStatus
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

export interface Match {
    item: RecordRef
    record: (Record | null)
    score: (Scalars['Float'] | null)
    evidence: Evidence[]
    __typename: 'Match'
}

export interface Evidence {
    plugin: Scalars['String']
    url: (Scalars['String'] | null)
    title: (Scalars['String'] | null)
    image: (Asset | null)
    detail: (Scalars['String'] | null)
    __typename: 'Evidence'
}

export interface Short {
    match: Match
    video: (Video | null)
    startSeconds: (Scalars['Float'] | null)
    endSeconds: (Scalars['Float'] | null)
    __typename: 'Short'
}

export interface Taste {
    signalCount: Scalars['Int']
    tags: TasteEntry[]
    performers: TasteEntry[]
    studios: TasteEntry[]
    sites: TasteEntry[]
    __typename: 'Taste'
}

export interface TasteEntry {
    id: Scalars['ID']
    name: Scalars['String']
    image: (Asset | null)
    weight: Scalars['Float']
    __typename: 'TasteEntry'
}

export interface LookupResult {
    matches: Match[]
    failures: Scalars['String'][]
    __typename: 'LookupResult'
}

export interface AlikePair {
    id: Scalars['ID']
    subjectKind: RecordKind
    first: (Record | null)
    second: (Record | null)
    score: Scalars['Float']
    reason: (Scalars['String'] | null)
    provablySame: Scalars['Boolean']
    __typename: 'AlikePair'
}

export interface LibraryRoot {
    id: Scalars['ID']
    location: (StorageLocation | null)
    path: Scalars['String']
    scanIntervalMinutes: (Scalars['Int'] | null)
    lastScanAt: (Scalars['String'] | null)
    organizeTemplate: (Scalars['String'] | null)
    __typename: 'LibraryRoot'
}

export interface FileMove {
    fileId: Scalars['ID']
    from: Scalars['String']
    to: Scalars['String']
    __typename: 'FileMove'
}

export interface MediaFile {
    id: Scalars['ID']
    subjectKind: RecordKind
    subjectId: Scalars['ID']
    url: Scalars['String']
    path: Scalars['String']
    location: (StorageLocation | null)
    sizeBytes: Scalars['Int']
    mimeType: Scalars['String']
    quality: (Scalars['String'] | null)
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    durationSeconds: (Scalars['Float'] | null)
    missing: Scalars['Boolean']
    __typename: 'MediaFile'
}

export interface Download {
    job: Job
    file: (MediaFile | null)
    __typename: 'Download'
}

export type FingerprintAlgorithm = 'phash' | 'oshash' | 'md5' | 'audio'

export interface Fingerprint {
    id: Scalars['ID']
    algorithm: FingerprintAlgorithm
    value: Scalars['String']
    fileId: (Scalars['ID'] | null)
    durationSeconds: (Scalars['Float'] | null)
    __typename: 'Fingerprint'
}

export type EmbeddingKind = 'face' | 'visual' | 'text'

export interface Embedding {
    id: Scalars['ID']
    kind: EmbeddingKind
    model: Scalars['String']
    atSeconds: (Scalars['Float'] | null)
    region: (Scalars['JSON'] | null)
    __typename: 'Embedding'
}

export interface Audio {
    files: MediaFile[]
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    status: MediaStatus
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    durationSeconds: (Scalars['Float'] | null)
    sources: Source[]
    streams: Stream[]
    userState: UserState
    fieldSources: FieldSource[]
    revisions: Revision[]
    segments: Segment[]
    subtitles: Subtitle[]
    scripts: Script[]
    comments: Comment[]
    __typename: 'Audio'
}

export interface Source {
    snapshot: (Asset | null)
    id: Scalars['ID']
    plugin: Scalars['String']
    externalId: Scalars['String']
    url: Scalars['String']
    title: (Scalars['String'] | null)
    image: (Asset | null)
    fetchedAt: (Scalars['String'] | null)
    __typename: 'Source'
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

export interface TagCount {
    tag: Tag
    count: Scalars['Int']
    __typename: 'TagCount'
}

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
    url: Scalars['String']
    mimeType: (Scalars['String'] | null)
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    blurhash: (Scalars['String'] | null)
    __typename: 'Asset'
}

export interface VrFormat {
    projection: VrProjection
    stereo: VrStereo
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

export interface Stream {
    id: Scalars['ID']
    url: Scalars['String']
    mimeType: Scalars['String']
    quality: (Scalars['String'] | null)
    height: (Scalars['Int'] | null)
    speedKbps: (Scalars['Int'] | null)
    source: (Source | null)
    expiresAt: (Scalars['String'] | null)
    __typename: 'Stream'
}

export type VrProjection = 'equirectangular180' | 'equirectangular360' | 'fisheye' | 'cubemap'

export type VrStereo = 'mono' | 'sideBySide' | 'topBottom'

export interface Story {
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    status: MediaStatus
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    chapterCount: Scalars['Int']
    wordCount: (Scalars['Int'] | null)
    complete: Scalars['Boolean']
    chapters: StoryChapter[]
    sources: Source[]
    userState: UserState
    fieldSources: FieldSource[]
    revisions: Revision[]
    __typename: 'Story'
}

export interface StoryChapter {
    id: Scalars['ID']
    position: Scalars['Int']
    title: (Scalars['String'] | null)
    text: Scalars['String']
    wordCount: Scalars['Int']
    url: (Scalars['String'] | null)
    __typename: 'StoryChapter'
}

export interface Post {
    id: Scalars['ID']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    title: Scalars['String']
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    status: MediaStatus
    creator: (Performer | null)
    text: (Scalars['String'] | null)
    tags: Tag[]
    attachments: Record[]
    locked: Scalars['Boolean']
    priceCents: (Scalars['Int'] | null)
    currency: (Scalars['String'] | null)
    sources: Source[]
    userState: UserState
    fieldSources: FieldSource[]
    revisions: Revision[]
    __typename: 'Post'
}

export interface Group {
    id: Scalars['ID']
    createdAt: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    cover: (Asset | null)
    studio: (Studio | null)
    videos: Video[]
    sources: Source[]
    __typename: 'Group'
}

export interface ApiToken {
    id: Scalars['ID']
    name: Scalars['String']
    scopes: Scalars['String'][]
    lastUsedAt: (Scalars['String'] | null)
    expiresAt: (Scalars['String'] | null)
    createdAt: Scalars['String']
    __typename: 'ApiToken'
}

export interface CreatedApiToken {
    token: Scalars['String']
    apiToken: ApiToken
    __typename: 'CreatedApiToken'
}

export type AccountStatus = 'ok' | 'expired' | 'needsLogin' | 'failed'

export interface Account {
    id: Scalars['ID']
    plugin: Scalars['String']
    label: Scalars['String']
    status: AccountStatus
    expiresAt: (Scalars['String'] | null)
    lastCheckedAt: (Scalars['String'] | null)
    __typename: 'Account'
}

export interface PluginContribution {
    plugin: Scalars['String']
    kind: Scalars['String']
    name: Scalars['String']
    title: (Scalars['String'] | null)
    mediaKinds: Scalars['String'][]
    domains: Scalars['String'][]
    __typename: 'PluginContribution'
}

export interface PluginView {
    kind: Scalars['String']
    title: Scalars['String']
    body: (Scalars['String'] | null)
    fields: PluginViewField[]
    items: PluginViewItem[]
    actions: PluginViewAction[]
    __typename: 'PluginView'
}

export interface PluginViewField {
    key: Scalars['String']
    label: Scalars['String']
    type: Scalars['String']
    value: (Scalars['String'] | null)
    options: Scalars['String'][]
    required: Scalars['Boolean']
    hint: (Scalars['String'] | null)
    __typename: 'PluginViewField'
}

export interface PluginViewItem {
    title: Scalars['String']
    subtitle: (Scalars['String'] | null)
    imageUrl: (Scalars['String'] | null)
    url: (Scalars['String'] | null)
    record: (Record | null)
    __typename: 'PluginViewItem'
}

export interface PluginViewAction {
    name: Scalars['String']
    label: Scalars['String']
    primary: Scalars['Boolean']
    __typename: 'PluginViewAction'
}

export interface PluginViewResult {
    view: (PluginView | null)
    message: (Scalars['String'] | null)
    __typename: 'PluginViewResult'
}

export interface Plugin {
    contributions: PluginContribution[]
    account: (Account | null)
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

export type SegmentOrigin = 'user' | 'plugin' | 'ai'

export interface Segment {
    id: Scalars['ID']
    subjectKind: RecordKind
    subjectId: Scalars['ID']
    startSeconds: Scalars['Float']
    endSeconds: (Scalars['Float'] | null)
    title: (Scalars['String'] | null)
    tag: (Tag | null)
    performers: Performer[]
    origin: SegmentOrigin
    __typename: 'Segment'
}

export interface Clip {
    id: Scalars['ID']
    createdAt: Scalars['String']
    video: (Video | null)
    startSeconds: Scalars['Float']
    endSeconds: Scalars['Float']
    title: (Scalars['String'] | null)
    cover: (Asset | null)
    file: (MediaFile | null)
    __typename: 'Clip'
}

export type SubtitleOrigin = 'site' | 'generated' | 'translated'

export interface Subtitle {
    id: Scalars['ID']
    language: Scalars['String']
    format: Scalars['String']
    asset: Asset
    origin: SubtitleOrigin
    __typename: 'Subtitle'
}

export type ScriptKind = 'funscript'

export interface Script {
    id: Scalars['ID']
    kind: ScriptKind
    asset: Asset
    axis: (Scalars['String'] | null)
    writtenBy: Actor
    __typename: 'Script'
}

export interface Comment {
    id: Scalars['ID']
    source: (Source | null)
    author: (Scalars['String'] | null)
    text: Scalars['String']
    postedAt: (Scalars['String'] | null)
    likes: (Scalars['Int'] | null)
    replyTo: (Scalars['ID'] | null)
    mentions: Performer[]
    __typename: 'Comment'
}

export interface PluginPackage {
    name: Scalars['String']
    version: Scalars['String']
    description: (Scalars['String'] | null)
    installedVersion: (Scalars['String'] | null)
    __typename: 'PluginPackage'
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

export type Record = (Video | Gallery | Audio | Story | Post | Performer | Studio | Tag | Group | Category | Collection | CamModel | Segment | Clip) & { __isUnion?: true }

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
    minHeight: (Scalars['Int'] | null)
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
    follows?: FollowGenqlSelection
    follow?: (FollowGenqlSelection & { __args: {id: Scalars['ID']} })
    followOf?: (FollowGenqlSelection & { __args: {subject: RecordRefInput} })
    newSince?: (RecordGenqlSelection & { __args: {followId: Scalars['ID'], since?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null)} })
    notifications?: (NotificationGenqlSelection & { __args?: {unreadOnly?: (Scalars['Boolean'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    unreadNotificationCount?: boolean | number
    notificationChannels?: NotificationChannelGenqlSelection
    limits?: LimitGenqlSelection
    limitStatus?: LimitStatusGenqlSelection
    usage?: (UsageGenqlSelection & { __args: {window: LimitWindow} })
    automations?: AutomationGenqlSelection
    automation?: (AutomationGenqlSelection & { __args: {id: Scalars['ID']} })
    automationRuns?: (AutomationRunGenqlSelection & { __args?: {automationId?: (Scalars['ID'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    automationRun?: (AutomationRunGenqlSelection & { __args: {id: Scalars['ID']} })
    automationActions?: boolean | number
    camRooms?: (CamRoomPageGenqlSelection & { __args?: {query?: (CamQuery | null)} })
    camModels?: (CamModelGenqlSelection & { __args?: {live?: (Scalars['Boolean'] | null), autoRecord?: (Scalars['Boolean'] | null), performerId?: (Scalars['ID'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    camModel?: (CamModelGenqlSelection & { __args: {id: Scalars['ID']} })
    activeRecordings?: ActiveRecordingGenqlSelection
    categories?: CategoryGenqlSelection
    category?: (CategoryGenqlSelection & { __args?: {id?: (Scalars['ID'] | null), slug?: (Scalars['String'] | null)} })
    collections?: CollectionGenqlSelection
    collection?: (CollectionGenqlSelection & { __args: {id: Scalars['ID']} })
    builtInCollection?: (CollectionGenqlSelection & { __args: {kind: CollectionKind} })
    collectionsWith?: (CollectionGenqlSelection & { __args: {subject: RecordRefInput} })
    fieldDefs?: (FieldDefGenqlSelection & { __args?: {kind?: (RecordKind | null)} })
    recommendations?: (MatchGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    related?: (MatchGenqlSelection & { __args: {subject: RecordRefInput, limit?: (Scalars['Int'] | null)} })
    shorts?: (ShortGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null)} })
    taste?: TasteGenqlSelection
    bestMoment?: { __args: {videoId: Scalars['ID']} }
    alikes?: (AlikePairGenqlSelection & { __args: {kind: RecordKind, limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    libraryRoots?: LibraryRootGenqlSelection
    libraryRoot?: (LibraryRootGenqlSelection & { __args: {id: Scalars['ID']} })
    planOrganize?: (FileMoveGenqlSelection & { __args: {rootId: Scalars['ID']} })
    downloads?: DownloadGenqlSelection
    sameMedia?: (RecordGenqlSelection & { __args: {subject: RecordRefInput} })
    nearest?: (MatchGenqlSelection & { __args: {kind: EmbeddingKind, model: Scalars['String'], vector: Scalars['Float'][], kinds?: (RecordKind[] | null), limit?: (Scalars['Int'] | null)} })
    similar?: (MatchGenqlSelection & { __args: {subject: RecordRefInput, kind: EmbeddingKind, model: Scalars['String'], limit?: (Scalars['Int'] | null)} })
    videos?: (VideoGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    videoCount?: { __args: {filter?: (FilterInput | null)} } | boolean | number
    videoTagCounts?: (TagCountGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null)} })
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
    media?: (RecordGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    mediaCount?: { __args: {filter?: (FilterInput | null)} } | boolean | number
    audios?: (AudioGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    audioCount?: { __args: {filter?: (FilterInput | null)} } | boolean | number
    audio?: (AudioGenqlSelection & { __args: {id: Scalars['ID']} })
    stories?: (StoryGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    storyCount?: { __args: {filter?: (FilterInput | null)} } | boolean | number
    story?: (StoryGenqlSelection & { __args: {id: Scalars['ID']} })
    posts?: (PostGenqlSelection & { __args?: {filter?: (FilterInput | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    postCount?: { __args: {filter?: (FilterInput | null)} } | boolean | number
    post?: (PostGenqlSelection & { __args: {id: Scalars['ID']} })
    groups?: (GroupGenqlSelection & { __args?: {text?: (Scalars['String'] | null), videoId?: (Scalars['ID'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    group?: (GroupGenqlSelection & { __args: {id: Scalars['ID']} })
    settings?: boolean | number
    apiTokens?: ApiTokenGenqlSelection
    accounts?: AccountGenqlSelection
    contributions?: (PluginContributionGenqlSelection & { __args?: {kind?: (Scalars['String'] | null)} })
    pluginView?: (PluginViewGenqlSelection & { __args: {plugin: Scalars['String'], view: Scalars['String'], input?: (Scalars['JSON'] | null)} })
    clips?: (ClipGenqlSelection & { __args?: {videoId?: (Scalars['ID'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    clip?: (ClipGenqlSelection & { __args: {id: Scalars['ID']} })
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
    followRecord?: (FollowGenqlSelection & { __args: {subject: RecordRefInput, intervalHours?: (Scalars['Int'] | null), notify?: (Scalars['Boolean'] | null)} })
    followSearch?: (FollowGenqlSelection & { __args: {text: Scalars['String'], intervalHours?: (Scalars['Int'] | null), notify?: (Scalars['Boolean'] | null)} })
    createFollow?: (FollowGenqlSelection & { __args: {input: FollowInput} })
    updateFollow?: (FollowGenqlSelection & { __args: {id: Scalars['ID'], input: FollowEditInput} })
    unfollow?: { __args: {id: Scalars['ID']} }
    runFollow?: (JobGenqlSelection & { __args: {id: Scalars['ID']} })
    markNotificationsRead?: { __args: {ids: Scalars['ID'][]} }
    markAllNotificationsRead?: boolean | number
    createNotificationChannel?: (NotificationChannelGenqlSelection & { __args: {kind: ChannelKind, name: Scalars['String'], config?: (Scalars['JSON'] | null), enabled?: (Scalars['Boolean'] | null)} })
    updateNotificationChannel?: (NotificationChannelGenqlSelection & { __args: {id: Scalars['ID'], name?: (Scalars['String'] | null), config?: (Scalars['JSON'] | null), enabled?: (Scalars['Boolean'] | null)} })
    deleteNotificationChannel?: { __args: {id: Scalars['ID']} }
    createLimit?: (LimitGenqlSelection & { __args: {kind: LimitKind, window: LimitWindow, amount: Scalars['Int'], action: LimitAction, enabled?: (Scalars['Boolean'] | null)} })
    updateLimit?: (LimitGenqlSelection & { __args: {id: Scalars['ID'], amount?: (Scalars['Int'] | null), action?: (LimitAction | null), enabled?: (Scalars['Boolean'] | null)} })
    deleteLimit?: { __args: {id: Scalars['ID']} }
    createAutomation?: (AutomationGenqlSelection & { __args: {input: AutomationInput} })
    updateAutomation?: (AutomationGenqlSelection & { __args: {id: Scalars['ID'], input: AutomationEditInput} })
    deleteAutomation?: { __args: {id: Scalars['ID']} }
    runAutomation?: (AutomationRunGenqlSelection & { __args: {id: Scalars['ID'], subjects?: (RecordRefInput[] | null), dryRun?: (Scalars['Boolean'] | null)} })
    undoAutomationRun?: (AutomationRunGenqlSelection & { __args: {id: Scalars['ID']} })
    watchCam?: (PlayableGenqlSelection & { __args: {id: Scalars['ID']} })
    setCamAutoRecord?: (CamModelGenqlSelection & { __args: {id: Scalars['ID'], on: Scalars['Boolean'], quality?: (Scalars['String'] | null)} })
    linkCamPerformer?: (CamModelGenqlSelection & { __args: {id: Scalars['ID'], performerId?: (Scalars['ID'] | null)} })
    startRecording?: (JobGenqlSelection & { __args: {id: Scalars['ID'], quality?: (Scalars['String'] | null), fromSecondsAgo?: (Scalars['Int'] | null)} })
    stopRecording?: { __args: {id: Scalars['ID']} }
    createCategory?: (CategoryGenqlSelection & { __args: {input: CategoryInput} })
    editCategory?: (CategoryGenqlSelection & { __args: {id: Scalars['ID'], input: CategoryInput} })
    reorderCategories?: (CategoryGenqlSelection & { __args: {ids: Scalars['ID'][]} })
    createCollection?: (CollectionGenqlSelection & { __args: {name: Scalars['String'], details?: (Scalars['String'] | null)} })
    editCollection?: (CollectionGenqlSelection & { __args: {id: Scalars['ID'], name?: (Scalars['String'] | null), details?: (Scalars['String'] | null), coverUri?: (Scalars['String'] | null)} })
    addToCollection?: (CollectionGenqlSelection & { __args: {collectionId: Scalars['ID'], subjects: RecordRefInput[], position?: (Scalars['Int'] | null)} })
    removeFromCollection?: (CollectionGenqlSelection & { __args: {collectionId: Scalars['ID'], subjects: RecordRefInput[]} })
    moveInCollection?: (CollectionGenqlSelection & { __args: {collectionId: Scalars['ID'], subject: RecordRefInput, position: Scalars['Int']} })
    clearCollection?: (CollectionGenqlSelection & { __args: {collectionId: Scalars['ID']} })
    saveFieldDef?: (FieldDefGenqlSelection & { __args: {id?: (Scalars['ID'] | null), input: FieldDefInput} })
    deleteFieldDef?: { __args: {id: Scalars['ID']} }
    setFieldValue?: (UserStateGenqlSelection & { __args: {subject: RecordRefInput, fieldDefId: Scalars['ID'], value?: (Scalars['JSON'] | null)} })
    lookup?: (LookupResultGenqlSelection & { __args: {input: LookupInput} })
    merge?: (RecordGenqlSelection & { __args: {kind: RecordKind, keepId: Scalars['ID'], mergeId: Scalars['ID']} })
    mergeAlike?: (RecordGenqlSelection & { __args: {id: Scalars['ID']} })
    dismissAlike?: { __args: {id: Scalars['ID']} }
    addLibraryRoot?: (LibraryRootGenqlSelection & { __args: {locationId: Scalars['ID'], path: Scalars['String'], scanIntervalMinutes?: (Scalars['Int'] | null), organizeTemplate?: (Scalars['String'] | null)} })
    updateLibraryRoot?: (LibraryRootGenqlSelection & { __args: {id: Scalars['ID'], scanIntervalMinutes?: (Scalars['Int'] | null), organizeTemplate?: (Scalars['String'] | null)} })
    removeLibraryRoot?: { __args: {id: Scalars['ID']} }
    scanLibraryRoot?: (JobGenqlSelection & { __args: {id: Scalars['ID']} })
    applyOrganize?: { __args: {rootId: Scalars['ID'], moves: FileMoveInput[]} }
    trashFile?: { __args: {id: Scalars['ID']} }
    download?: (JobGenqlSelection & { __args: {subject: RecordRefInput, streamId?: (Scalars['ID'] | null)} })
    deleteDownload?: { __args: {jobId: Scalars['ID']} }
    addFingerprint?: (FingerprintGenqlSelection & { __args: {subject: RecordRefInput, fileId?: (Scalars['ID'] | null), algorithm: FingerprintAlgorithm, value: Scalars['String'], durationSeconds?: (Scalars['Float'] | null)} })
    fingerprintFile?: (JobGenqlSelection & { __args: {fileId: Scalars['ID']} })
    snapshotSource?: (JobGenqlSelection & { __args: {sourceId: Scalars['ID']} })
    addEmbedding?: (EmbeddingGenqlSelection & { __args: {input: EmbeddingInput} })
    removeEmbedding?: { __args: {id: Scalars['ID']} }
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
    editMedia?: (RecordGenqlSelection & { __args: {subject: RecordRefInput, edit: MediaEdit} })
    fetchMedia?: (RecordGenqlSelection & { __args: {subject: RecordRefInput} })
    readChapter?: (StoryChapterGenqlSelection & { __args: {storyId: Scalars['ID'], position: Scalars['Int']} })
    updateSettings?: { __args: {values: Scalars['JSON']} }
    createApiToken?: (CreatedApiTokenGenqlSelection & { __args: {name: Scalars['String'], scopes: Scalars['String'][], expiresAt?: (Scalars['String'] | null)} })
    revokeApiToken?: { __args: {id: Scalars['ID']} }
    loginAccount?: (AccountGenqlSelection & { __args: {plugin: Scalars['String'], values: Scalars['JSON']} })
    checkAccount?: (AccountGenqlSelection & { __args: {id: Scalars['ID']} })
    logoutAccount?: { __args: {id: Scalars['ID']} }
    submitPluginView?: (PluginViewResultGenqlSelection & { __args: {plugin: Scalars['String'], view: Scalars['String'], action: Scalars['String'], values?: (Scalars['JSON'] | null), input?: (Scalars['JSON'] | null)} })
    addSegment?: (SegmentGenqlSelection & { __args: {subject: RecordRefInput, input: SegmentInput} })
    editSegment?: (SegmentGenqlSelection & { __args: {id: Scalars['ID'], input: SegmentInput} })
    addClip?: (ClipGenqlSelection & { __args: {videoId: Scalars['ID'], input: ClipInput} })
    editClip?: (ClipGenqlSelection & { __args: {id: Scalars['ID'], input: ClipInput} })
    attachClipFile?: (ClipGenqlSelection & { __args: {clipId: Scalars['ID'], fileId: Scalars['ID']} })
    addSubtitle?: (SubtitleGenqlSelection & { __args: {subject: RecordRefInput, input: SubtitleInput} })
    removeSubtitle?: { __args: {id: Scalars['ID']} }
    addScript?: (ScriptGenqlSelection & { __args: {subject: RecordRefInput, uri: Scalars['String'], axis?: (Scalars['String'] | null)} })
    removeScript?: { __args: {id: Scalars['ID']} }
    setTrickplay?: (VideoGenqlSelection & { __args: {videoId: Scalars['ID'], trickplay?: (TrickplayInput | null)} })
    reportStreamSpeed?: (StreamGenqlSelection & { __args: {streamId: Scalars['ID'], kilobitsPerSecond: Scalars['Int']} })
    refreshSources?: (RecordGenqlSelection & { __args: {subject: RecordRefInput} })
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
    recordEvents?: { __args: {events: EventInput[]} }
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

export interface SubscriptionGenqlSelection{
    notificationAdded?: NotificationGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FollowGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    filter?: FilterGenqlSelection
    targetKind?: boolean | number
    targetId?: boolean | number
    target?: RecordGenqlSelection
    intervalHours?: boolean | number
    enabled?: boolean | number
    notify?: boolean | number
    lastRunAt?: boolean | number
    nextRunAt?: boolean | number
    lastError?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FollowInput {name: Scalars['String'],filter: FilterInput,intervalHours?: (Scalars['Int'] | null),notify?: (Scalars['Boolean'] | null)}

export interface FollowEditInput {name?: (Scalars['String'] | null),filter?: (FilterInput | null),intervalHours?: (Scalars['Int'] | null),enabled?: (Scalars['Boolean'] | null),notify?: (Scalars['Boolean'] | null)}

export interface NotificationGenqlSelection{
    id?: boolean | number
    at?: boolean | number
    kind?: boolean | number
    title?: boolean | number
    body?: boolean | number
    subject?: RecordRefGenqlSelection
    record?: RecordGenqlSelection
    read?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface NotificationChannelGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    name?: boolean | number
    config?: boolean | number
    enabled?: boolean | number
    delivers?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface LimitGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    window?: boolean | number
    amount?: boolean | number
    action?: boolean | number
    enabled?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface LimitStatusGenqlSelection{
    limit?: LimitGenqlSelection
    used?: boolean | number
    reached?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UsageGenqlSelection{
    since?: boolean | number
    until?: boolean | number
    watchMinutes?: boolean | number
    sessions?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TriggerGenqlSelection{
    kind?: boolean | number
    cron?: boolean | number
    eventKinds?: boolean | number
    filter?: FilterGenqlSelection
    pluginTrigger?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TriggerInput {kind: TriggerKind,cron?: (Scalars['String'] | null),eventKinds?: (Scalars['String'][] | null),filter?: (FilterInput | null),pluginTrigger?: (Scalars['String'] | null)}

export interface StepGenqlSelection{
    action?: boolean | number
    args?: boolean | number
    condition?: FilterGenqlSelection
    continueOnError?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StepInput {action: Scalars['String'],args?: (Scalars['JSON'] | null),condition?: (FilterInput | null),continueOnError?: (Scalars['Boolean'] | null)}

export interface AutomationGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    enabled?: boolean | number
    trigger?: TriggerGenqlSelection
    steps?: StepGenqlSelection
    builtIn?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface AutomationInput {name: Scalars['String'],enabled?: (Scalars['Boolean'] | null),trigger: TriggerInput,steps: StepInput[]}

export interface AutomationEditInput {name?: (Scalars['String'] | null),enabled?: (Scalars['Boolean'] | null),trigger?: (TriggerInput | null),steps?: (StepInput[] | null)}

export interface AutomationRunGenqlSelection{
    id?: boolean | number
    automation?: AutomationGenqlSelection
    startedAt?: boolean | number
    finishedAt?: boolean | number
    status?: boolean | number
    dryRun?: boolean | number
    triggerData?: boolean | number
    steps?: StepResultGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StepResultGenqlSelection{
    action?: boolean | number
    ok?: boolean | number
    output?: boolean | number
    error?: boolean | number
    touched?: RecordRefGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CamQuery {plugins?: (Scalars['String'][] | null),text?: (Scalars['String'] | null),tag?: (Scalars['String'] | null),gender?: (Scalars['String'] | null),limit?: (Scalars['Int'] | null),offset?: (Scalars['Int'] | null)}

export interface CamRoomPageGenqlSelection{
    rooms?: CamModelGenqlSelection
    total?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CamModelGenqlSelection{
    id?: boolean | number
    plugin?: boolean | number
    username?: boolean | number
    displayName?: boolean | number
    performer?: PerformerGenqlSelection
    thumbnail?: AssetGenqlSelection
    tags?: boolean | number
    gender?: boolean | number
    live?: CamLiveGenqlSelection
    lastSeenOnlineAt?: boolean | number
    autoRecord?: boolean | number
    recordQuality?: boolean | number
    recordings?: VideoGenqlSelection
    userState?: UserStateGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CamLiveGenqlSelection{
    viewers?: boolean | number
    subject?: boolean | number
    startedAt?: boolean | number
    thumbnail?: AssetGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ActiveRecordingGenqlSelection{
    job?: JobGenqlSelection
    camModel?: CamModelGenqlSelection
    startedAt?: boolean | number
    durationSeconds?: boolean | number
    sizeBytes?: boolean | number
    quality?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface VideoGenqlSelection{
    recordedFrom?: CamModelGenqlSelection
    categories?: CategoryGenqlSelection
    related?: (MatchGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null)} })
    files?: MediaFileGenqlSelection
    fingerprints?: FingerprintGenqlSelection
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
    previewFrames?: AssetGenqlSelection
    gallery?: GalleryGenqlSelection
    sources?: SourceGenqlSelection
    streams?: StreamGenqlSelection
    userState?: UserStateGenqlSelection
    oCount?: boolean | number
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    groups?: GroupGenqlSelection
    segments?: SegmentGenqlSelection
    subtitles?: SubtitleGenqlSelection
    scripts?: ScriptGenqlSelection
    clips?: ClipGenqlSelection
    comments?: CommentGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CategoryGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    slug?: boolean | number
    parent?: CategoryGenqlSelection
    children?: CategoryGenqlSelection
    matchTags?: TagGenqlSelection
    cover?: AssetGenqlSelection
    position?: boolean | number
    videoCount?: boolean | number
    galleryCount?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CategoryInput {name?: (Scalars['String'] | null),slug?: (Scalars['String'] | null),parentId?: (Scalars['ID'] | null),matchTagIds?: (Scalars['ID'][] | null),coverUri?: (Scalars['String'] | null),position?: (Scalars['Int'] | null)}

export interface CollectionGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    kind?: boolean | number
    details?: boolean | number
    cover?: AssetGenqlSelection
    itemCount?: boolean | number
    items?: (CollectionItemGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CollectionItemGenqlSelection{
    subjectKind?: boolean | number
    subjectId?: boolean | number
    record?: RecordGenqlSelection
    position?: boolean | number
    addedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FieldDefGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    kinds?: boolean | number
    fieldType?: boolean | number
    options?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FieldDefInput {name: Scalars['String'],kinds: RecordKind[],fieldType: FieldType,options?: (Scalars['String'][] | null)}

export interface GalleryGenqlSelection{
    categories?: CategoryGenqlSelection
    related?: (MatchGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null)} })
    files?: MediaFileGenqlSelection
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

export interface MatchGenqlSelection{
    item?: RecordRefGenqlSelection
    record?: RecordGenqlSelection
    score?: boolean | number
    evidence?: EvidenceGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface EvidenceGenqlSelection{
    plugin?: boolean | number
    url?: boolean | number
    title?: boolean | number
    image?: AssetGenqlSelection
    detail?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ShortGenqlSelection{
    match?: MatchGenqlSelection
    video?: VideoGenqlSelection
    startSeconds?: boolean | number
    endSeconds?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TasteGenqlSelection{
    signalCount?: boolean | number
    tags?: TasteEntryGenqlSelection
    performers?: TasteEntryGenqlSelection
    studios?: TasteEntryGenqlSelection
    sites?: TasteEntryGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TasteEntryGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    image?: AssetGenqlSelection
    weight?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface LookupInput {text?: (Scalars['String'] | null),imageBase64?: (Scalars['String'] | null),mimeType?: (Scalars['String'] | null),plugins?: (Scalars['String'][] | null)}

export interface LookupResultGenqlSelection{
    matches?: MatchGenqlSelection
    failures?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface AlikePairGenqlSelection{
    id?: boolean | number
    subjectKind?: boolean | number
    first?: RecordGenqlSelection
    second?: RecordGenqlSelection
    score?: boolean | number
    reason?: boolean | number
    provablySame?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface LibraryRootGenqlSelection{
    id?: boolean | number
    location?: StorageLocationGenqlSelection
    path?: boolean | number
    scanIntervalMinutes?: boolean | number
    lastScanAt?: boolean | number
    organizeTemplate?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FileMoveGenqlSelection{
    fileId?: boolean | number
    from?: boolean | number
    to?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FileMoveInput {fileId: Scalars['ID'],from: Scalars['String'],to: Scalars['String']}

export interface MediaFileGenqlSelection{
    id?: boolean | number
    subjectKind?: boolean | number
    subjectId?: boolean | number
    url?: boolean | number
    path?: boolean | number
    location?: StorageLocationGenqlSelection
    sizeBytes?: boolean | number
    mimeType?: boolean | number
    quality?: boolean | number
    width?: boolean | number
    height?: boolean | number
    durationSeconds?: boolean | number
    missing?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface DownloadGenqlSelection{
    job?: JobGenqlSelection
    file?: MediaFileGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FingerprintGenqlSelection{
    id?: boolean | number
    algorithm?: boolean | number
    value?: boolean | number
    fileId?: boolean | number
    durationSeconds?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface EmbeddingGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    model?: boolean | number
    atSeconds?: boolean | number
    region?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface EmbeddingInput {subject: RecordRefInput,kind: EmbeddingKind,model: Scalars['String'],vector: Scalars['Float'][],atSeconds?: (Scalars['Float'] | null),region?: (Scalars['JSON'] | null)}

export interface AudioGenqlSelection{
    files?: MediaFileGenqlSelection
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
    sources?: SourceGenqlSelection
    streams?: StreamGenqlSelection
    userState?: UserStateGenqlSelection
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    segments?: SegmentGenqlSelection
    subtitles?: SubtitleGenqlSelection
    scripts?: ScriptGenqlSelection
    comments?: CommentGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SourceGenqlSelection{
    snapshot?: AssetGenqlSelection
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

export interface MediaEdit {title?: (Scalars['String'] | null),details?: (Scalars['String'] | null),date?: (Scalars['String'] | null),studioId?: (Scalars['ID'] | null),performerIds?: (Scalars['ID'][] | null),tagIds?: (Scalars['ID'][] | null),status?: (MediaStatus | null),vr?: (VrFormatInput | null),clearVr?: (Scalars['Boolean'] | null),width?: (Scalars['Int'] | null),height?: (Scalars['Int'] | null),durationSeconds?: (Scalars['Float'] | null)}

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

export interface TagCountGenqlSelection{
    tag?: TagGenqlSelection
    count?: boolean | number
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
    url?: boolean | number
    mimeType?: boolean | number
    width?: boolean | number
    height?: boolean | number
    blurhash?: boolean | number
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

export interface StreamGenqlSelection{
    id?: boolean | number
    url?: boolean | number
    mimeType?: boolean | number
    quality?: boolean | number
    height?: boolean | number
    speedKbps?: boolean | number
    source?: SourceGenqlSelection
    expiresAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface VrFormatInput {projection: VrProjection,stereo: VrStereo,fieldOfView?: (Scalars['Int'] | null)}

export interface StoryGenqlSelection{
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
    chapterCount?: boolean | number
    wordCount?: boolean | number
    complete?: boolean | number
    chapters?: StoryChapterGenqlSelection
    sources?: SourceGenqlSelection
    userState?: UserStateGenqlSelection
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StoryChapterGenqlSelection{
    id?: boolean | number
    position?: boolean | number
    title?: boolean | number
    text?: boolean | number
    wordCount?: boolean | number
    url?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PostGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    title?: boolean | number
    date?: boolean | number
    cover?: AssetGenqlSelection
    status?: boolean | number
    creator?: PerformerGenqlSelection
    text?: boolean | number
    tags?: TagGenqlSelection
    attachments?: RecordGenqlSelection
    locked?: boolean | number
    priceCents?: boolean | number
    currency?: boolean | number
    sources?: SourceGenqlSelection
    userState?: UserStateGenqlSelection
    fieldSources?: FieldSourceGenqlSelection
    revisions?: (RevisionGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface GroupGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    title?: boolean | number
    details?: boolean | number
    date?: boolean | number
    cover?: AssetGenqlSelection
    studio?: StudioGenqlSelection
    videos?: VideoGenqlSelection
    sources?: SourceGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ApiTokenGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    scopes?: boolean | number
    lastUsedAt?: boolean | number
    expiresAt?: boolean | number
    createdAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CreatedApiTokenGenqlSelection{
    token?: boolean | number
    apiToken?: ApiTokenGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface AccountGenqlSelection{
    id?: boolean | number
    plugin?: boolean | number
    label?: boolean | number
    status?: boolean | number
    expiresAt?: boolean | number
    lastCheckedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginContributionGenqlSelection{
    plugin?: boolean | number
    kind?: boolean | number
    name?: boolean | number
    title?: boolean | number
    mediaKinds?: boolean | number
    domains?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginViewGenqlSelection{
    kind?: boolean | number
    title?: boolean | number
    body?: boolean | number
    fields?: PluginViewFieldGenqlSelection
    items?: PluginViewItemGenqlSelection
    actions?: PluginViewActionGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginViewFieldGenqlSelection{
    key?: boolean | number
    label?: boolean | number
    type?: boolean | number
    value?: boolean | number
    options?: boolean | number
    required?: boolean | number
    hint?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginViewItemGenqlSelection{
    title?: boolean | number
    subtitle?: boolean | number
    imageUrl?: boolean | number
    url?: boolean | number
    record?: RecordGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginViewActionGenqlSelection{
    name?: boolean | number
    label?: boolean | number
    primary?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginViewResultGenqlSelection{
    view?: PluginViewGenqlSelection
    message?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginGenqlSelection{
    contributions?: PluginContributionGenqlSelection
    account?: AccountGenqlSelection
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

export interface SegmentGenqlSelection{
    id?: boolean | number
    subjectKind?: boolean | number
    subjectId?: boolean | number
    startSeconds?: boolean | number
    endSeconds?: boolean | number
    title?: boolean | number
    tag?: TagGenqlSelection
    performers?: PerformerGenqlSelection
    origin?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SegmentInput {startSeconds?: (Scalars['Float'] | null),endSeconds?: (Scalars['Float'] | null),clearEnd?: (Scalars['Boolean'] | null),title?: (Scalars['String'] | null),tagId?: (Scalars['ID'] | null),performerIds?: (Scalars['ID'][] | null)}

export interface ClipGenqlSelection{
    id?: boolean | number
    createdAt?: boolean | number
    video?: VideoGenqlSelection
    startSeconds?: boolean | number
    endSeconds?: boolean | number
    title?: boolean | number
    cover?: AssetGenqlSelection
    file?: MediaFileGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ClipInput {startSeconds?: (Scalars['Float'] | null),endSeconds?: (Scalars['Float'] | null),title?: (Scalars['String'] | null),coverUri?: (Scalars['String'] | null)}

export interface SubtitleGenqlSelection{
    id?: boolean | number
    language?: boolean | number
    format?: boolean | number
    asset?: AssetGenqlSelection
    origin?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SubtitleInput {language: Scalars['String'],format: Scalars['String'],uri: Scalars['String'],origin?: (SubtitleOrigin | null)}

export interface ScriptGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    asset?: AssetGenqlSelection
    axis?: boolean | number
    writtenBy?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TrickplayInput {sheetUris: Scalars['String'][],columns: Scalars['Int'],rows: Scalars['Int'],tileWidth: Scalars['Int'],tileHeight: Scalars['Int'],intervalSeconds: Scalars['Float']}

export interface CommentGenqlSelection{
    id?: boolean | number
    source?: SourceGenqlSelection
    author?: boolean | number
    text?: boolean | number
    postedAt?: boolean | number
    likes?: boolean | number
    replyTo?: boolean | number
    mentions?: PerformerGenqlSelection
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
    on_Audio?:AudioGenqlSelection,
    on_Story?:StoryGenqlSelection,
    on_Post?:PostGenqlSelection,
    on_Performer?:PerformerGenqlSelection,
    on_Studio?:StudioGenqlSelection,
    on_Tag?:TagGenqlSelection,
    on_Group?:GroupGenqlSelection,
    on_Category?:CategoryGenqlSelection,
    on_Collection?:CollectionGenqlSelection,
    on_CamModel?:CamModelGenqlSelection,
    on_Segment?:SegmentGenqlSelection,
    on_Clip?:ClipGenqlSelection,
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
    minHeight?: boolean | number
    fields?: FieldConditionGenqlSelection
    sort?: SortGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface FilterInput {kinds?: (MediaKind[] | null),text?: (Scalars['String'] | null),performerIds?: (Scalars['ID'][] | null),studioIds?: (Scalars['ID'][] | null),categoryIds?: (Scalars['ID'][] | null),includeTagIds?: (Scalars['ID'][] | null),excludeTagIds?: (Scalars['ID'][] | null),sources?: (Scalars['String'][] | null),dateFrom?: (Scalars['String'] | null),dateTo?: (Scalars['String'] | null),minDurationSeconds?: (Scalars['Int'] | null),maxDurationSeconds?: (Scalars['Int'] | null),minRating?: (Scalars['Float'] | null),status?: (MediaStatus[] | null),inLibrary?: (Scalars['Boolean'] | null),favorite?: (Scalars['Boolean'] | null),unwatched?: (Scalars['Boolean'] | null),vr?: (Scalars['Boolean'] | null),minHeight?: (Scalars['Int'] | null),fields?: (FieldConditionInput[] | null),sort?: (SortInput | null)}

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
    


    const Subscription_possibleTypes: string[] = ['Subscription']
    export const isSubscription = (obj?: { __typename?: any } | null): obj is Subscription => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSubscription"')
      return Subscription_possibleTypes.includes(obj.__typename)
    }
    


    const Follow_possibleTypes: string[] = ['Follow']
    export const isFollow = (obj?: { __typename?: any } | null): obj is Follow => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFollow"')
      return Follow_possibleTypes.includes(obj.__typename)
    }
    


    const Notification_possibleTypes: string[] = ['Notification']
    export const isNotification = (obj?: { __typename?: any } | null): obj is Notification => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isNotification"')
      return Notification_possibleTypes.includes(obj.__typename)
    }
    


    const NotificationChannel_possibleTypes: string[] = ['NotificationChannel']
    export const isNotificationChannel = (obj?: { __typename?: any } | null): obj is NotificationChannel => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isNotificationChannel"')
      return NotificationChannel_possibleTypes.includes(obj.__typename)
    }
    


    const Limit_possibleTypes: string[] = ['Limit']
    export const isLimit = (obj?: { __typename?: any } | null): obj is Limit => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isLimit"')
      return Limit_possibleTypes.includes(obj.__typename)
    }
    


    const LimitStatus_possibleTypes: string[] = ['LimitStatus']
    export const isLimitStatus = (obj?: { __typename?: any } | null): obj is LimitStatus => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isLimitStatus"')
      return LimitStatus_possibleTypes.includes(obj.__typename)
    }
    


    const Usage_possibleTypes: string[] = ['Usage']
    export const isUsage = (obj?: { __typename?: any } | null): obj is Usage => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isUsage"')
      return Usage_possibleTypes.includes(obj.__typename)
    }
    


    const Trigger_possibleTypes: string[] = ['Trigger']
    export const isTrigger = (obj?: { __typename?: any } | null): obj is Trigger => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTrigger"')
      return Trigger_possibleTypes.includes(obj.__typename)
    }
    


    const Step_possibleTypes: string[] = ['Step']
    export const isStep = (obj?: { __typename?: any } | null): obj is Step => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStep"')
      return Step_possibleTypes.includes(obj.__typename)
    }
    


    const Automation_possibleTypes: string[] = ['Automation']
    export const isAutomation = (obj?: { __typename?: any } | null): obj is Automation => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAutomation"')
      return Automation_possibleTypes.includes(obj.__typename)
    }
    


    const AutomationRun_possibleTypes: string[] = ['AutomationRun']
    export const isAutomationRun = (obj?: { __typename?: any } | null): obj is AutomationRun => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAutomationRun"')
      return AutomationRun_possibleTypes.includes(obj.__typename)
    }
    


    const StepResult_possibleTypes: string[] = ['StepResult']
    export const isStepResult = (obj?: { __typename?: any } | null): obj is StepResult => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStepResult"')
      return StepResult_possibleTypes.includes(obj.__typename)
    }
    


    const CamRoomPage_possibleTypes: string[] = ['CamRoomPage']
    export const isCamRoomPage = (obj?: { __typename?: any } | null): obj is CamRoomPage => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCamRoomPage"')
      return CamRoomPage_possibleTypes.includes(obj.__typename)
    }
    


    const CamModel_possibleTypes: string[] = ['CamModel']
    export const isCamModel = (obj?: { __typename?: any } | null): obj is CamModel => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCamModel"')
      return CamModel_possibleTypes.includes(obj.__typename)
    }
    


    const CamLive_possibleTypes: string[] = ['CamLive']
    export const isCamLive = (obj?: { __typename?: any } | null): obj is CamLive => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCamLive"')
      return CamLive_possibleTypes.includes(obj.__typename)
    }
    


    const ActiveRecording_possibleTypes: string[] = ['ActiveRecording']
    export const isActiveRecording = (obj?: { __typename?: any } | null): obj is ActiveRecording => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isActiveRecording"')
      return ActiveRecording_possibleTypes.includes(obj.__typename)
    }
    


    const Video_possibleTypes: string[] = ['Video']
    export const isVideo = (obj?: { __typename?: any } | null): obj is Video => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isVideo"')
      return Video_possibleTypes.includes(obj.__typename)
    }
    


    const Category_possibleTypes: string[] = ['Category']
    export const isCategory = (obj?: { __typename?: any } | null): obj is Category => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCategory"')
      return Category_possibleTypes.includes(obj.__typename)
    }
    


    const Collection_possibleTypes: string[] = ['Collection']
    export const isCollection = (obj?: { __typename?: any } | null): obj is Collection => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCollection"')
      return Collection_possibleTypes.includes(obj.__typename)
    }
    


    const CollectionItem_possibleTypes: string[] = ['CollectionItem']
    export const isCollectionItem = (obj?: { __typename?: any } | null): obj is CollectionItem => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCollectionItem"')
      return CollectionItem_possibleTypes.includes(obj.__typename)
    }
    


    const FieldDef_possibleTypes: string[] = ['FieldDef']
    export const isFieldDef = (obj?: { __typename?: any } | null): obj is FieldDef => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFieldDef"')
      return FieldDef_possibleTypes.includes(obj.__typename)
    }
    


    const Gallery_possibleTypes: string[] = ['Gallery']
    export const isGallery = (obj?: { __typename?: any } | null): obj is Gallery => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isGallery"')
      return Gallery_possibleTypes.includes(obj.__typename)
    }
    


    const Match_possibleTypes: string[] = ['Match']
    export const isMatch = (obj?: { __typename?: any } | null): obj is Match => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMatch"')
      return Match_possibleTypes.includes(obj.__typename)
    }
    


    const Evidence_possibleTypes: string[] = ['Evidence']
    export const isEvidence = (obj?: { __typename?: any } | null): obj is Evidence => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isEvidence"')
      return Evidence_possibleTypes.includes(obj.__typename)
    }
    


    const Short_possibleTypes: string[] = ['Short']
    export const isShort = (obj?: { __typename?: any } | null): obj is Short => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isShort"')
      return Short_possibleTypes.includes(obj.__typename)
    }
    


    const Taste_possibleTypes: string[] = ['Taste']
    export const isTaste = (obj?: { __typename?: any } | null): obj is Taste => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTaste"')
      return Taste_possibleTypes.includes(obj.__typename)
    }
    


    const TasteEntry_possibleTypes: string[] = ['TasteEntry']
    export const isTasteEntry = (obj?: { __typename?: any } | null): obj is TasteEntry => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTasteEntry"')
      return TasteEntry_possibleTypes.includes(obj.__typename)
    }
    


    const LookupResult_possibleTypes: string[] = ['LookupResult']
    export const isLookupResult = (obj?: { __typename?: any } | null): obj is LookupResult => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isLookupResult"')
      return LookupResult_possibleTypes.includes(obj.__typename)
    }
    


    const AlikePair_possibleTypes: string[] = ['AlikePair']
    export const isAlikePair = (obj?: { __typename?: any } | null): obj is AlikePair => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAlikePair"')
      return AlikePair_possibleTypes.includes(obj.__typename)
    }
    


    const LibraryRoot_possibleTypes: string[] = ['LibraryRoot']
    export const isLibraryRoot = (obj?: { __typename?: any } | null): obj is LibraryRoot => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isLibraryRoot"')
      return LibraryRoot_possibleTypes.includes(obj.__typename)
    }
    


    const FileMove_possibleTypes: string[] = ['FileMove']
    export const isFileMove = (obj?: { __typename?: any } | null): obj is FileMove => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFileMove"')
      return FileMove_possibleTypes.includes(obj.__typename)
    }
    


    const MediaFile_possibleTypes: string[] = ['MediaFile']
    export const isMediaFile = (obj?: { __typename?: any } | null): obj is MediaFile => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMediaFile"')
      return MediaFile_possibleTypes.includes(obj.__typename)
    }
    


    const Download_possibleTypes: string[] = ['Download']
    export const isDownload = (obj?: { __typename?: any } | null): obj is Download => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isDownload"')
      return Download_possibleTypes.includes(obj.__typename)
    }
    


    const Fingerprint_possibleTypes: string[] = ['Fingerprint']
    export const isFingerprint = (obj?: { __typename?: any } | null): obj is Fingerprint => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isFingerprint"')
      return Fingerprint_possibleTypes.includes(obj.__typename)
    }
    


    const Embedding_possibleTypes: string[] = ['Embedding']
    export const isEmbedding = (obj?: { __typename?: any } | null): obj is Embedding => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isEmbedding"')
      return Embedding_possibleTypes.includes(obj.__typename)
    }
    


    const Audio_possibleTypes: string[] = ['Audio']
    export const isAudio = (obj?: { __typename?: any } | null): obj is Audio => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAudio"')
      return Audio_possibleTypes.includes(obj.__typename)
    }
    


    const Source_possibleTypes: string[] = ['Source']
    export const isSource = (obj?: { __typename?: any } | null): obj is Source => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSource"')
      return Source_possibleTypes.includes(obj.__typename)
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
    


    const TagCount_possibleTypes: string[] = ['TagCount']
    export const isTagCount = (obj?: { __typename?: any } | null): obj is TagCount => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTagCount"')
      return TagCount_possibleTypes.includes(obj.__typename)
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
    


    const Stream_possibleTypes: string[] = ['Stream']
    export const isStream = (obj?: { __typename?: any } | null): obj is Stream => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStream"')
      return Stream_possibleTypes.includes(obj.__typename)
    }
    


    const Story_possibleTypes: string[] = ['Story']
    export const isStory = (obj?: { __typename?: any } | null): obj is Story => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStory"')
      return Story_possibleTypes.includes(obj.__typename)
    }
    


    const StoryChapter_possibleTypes: string[] = ['StoryChapter']
    export const isStoryChapter = (obj?: { __typename?: any } | null): obj is StoryChapter => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStoryChapter"')
      return StoryChapter_possibleTypes.includes(obj.__typename)
    }
    


    const Post_possibleTypes: string[] = ['Post']
    export const isPost = (obj?: { __typename?: any } | null): obj is Post => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPost"')
      return Post_possibleTypes.includes(obj.__typename)
    }
    


    const Group_possibleTypes: string[] = ['Group']
    export const isGroup = (obj?: { __typename?: any } | null): obj is Group => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isGroup"')
      return Group_possibleTypes.includes(obj.__typename)
    }
    


    const ApiToken_possibleTypes: string[] = ['ApiToken']
    export const isApiToken = (obj?: { __typename?: any } | null): obj is ApiToken => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isApiToken"')
      return ApiToken_possibleTypes.includes(obj.__typename)
    }
    


    const CreatedApiToken_possibleTypes: string[] = ['CreatedApiToken']
    export const isCreatedApiToken = (obj?: { __typename?: any } | null): obj is CreatedApiToken => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCreatedApiToken"')
      return CreatedApiToken_possibleTypes.includes(obj.__typename)
    }
    


    const Account_possibleTypes: string[] = ['Account']
    export const isAccount = (obj?: { __typename?: any } | null): obj is Account => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAccount"')
      return Account_possibleTypes.includes(obj.__typename)
    }
    


    const PluginContribution_possibleTypes: string[] = ['PluginContribution']
    export const isPluginContribution = (obj?: { __typename?: any } | null): obj is PluginContribution => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginContribution"')
      return PluginContribution_possibleTypes.includes(obj.__typename)
    }
    


    const PluginView_possibleTypes: string[] = ['PluginView']
    export const isPluginView = (obj?: { __typename?: any } | null): obj is PluginView => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginView"')
      return PluginView_possibleTypes.includes(obj.__typename)
    }
    


    const PluginViewField_possibleTypes: string[] = ['PluginViewField']
    export const isPluginViewField = (obj?: { __typename?: any } | null): obj is PluginViewField => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginViewField"')
      return PluginViewField_possibleTypes.includes(obj.__typename)
    }
    


    const PluginViewItem_possibleTypes: string[] = ['PluginViewItem']
    export const isPluginViewItem = (obj?: { __typename?: any } | null): obj is PluginViewItem => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginViewItem"')
      return PluginViewItem_possibleTypes.includes(obj.__typename)
    }
    


    const PluginViewAction_possibleTypes: string[] = ['PluginViewAction']
    export const isPluginViewAction = (obj?: { __typename?: any } | null): obj is PluginViewAction => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginViewAction"')
      return PluginViewAction_possibleTypes.includes(obj.__typename)
    }
    


    const PluginViewResult_possibleTypes: string[] = ['PluginViewResult']
    export const isPluginViewResult = (obj?: { __typename?: any } | null): obj is PluginViewResult => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginViewResult"')
      return PluginViewResult_possibleTypes.includes(obj.__typename)
    }
    


    const Plugin_possibleTypes: string[] = ['Plugin']
    export const isPlugin = (obj?: { __typename?: any } | null): obj is Plugin => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPlugin"')
      return Plugin_possibleTypes.includes(obj.__typename)
    }
    


    const Segment_possibleTypes: string[] = ['Segment']
    export const isSegment = (obj?: { __typename?: any } | null): obj is Segment => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSegment"')
      return Segment_possibleTypes.includes(obj.__typename)
    }
    


    const Clip_possibleTypes: string[] = ['Clip']
    export const isClip = (obj?: { __typename?: any } | null): obj is Clip => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isClip"')
      return Clip_possibleTypes.includes(obj.__typename)
    }
    


    const Subtitle_possibleTypes: string[] = ['Subtitle']
    export const isSubtitle = (obj?: { __typename?: any } | null): obj is Subtitle => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSubtitle"')
      return Subtitle_possibleTypes.includes(obj.__typename)
    }
    


    const Script_possibleTypes: string[] = ['Script']
    export const isScript = (obj?: { __typename?: any } | null): obj is Script => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isScript"')
      return Script_possibleTypes.includes(obj.__typename)
    }
    


    const Comment_possibleTypes: string[] = ['Comment']
    export const isComment = (obj?: { __typename?: any } | null): obj is Comment => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isComment"')
      return Comment_possibleTypes.includes(obj.__typename)
    }
    


    const PluginPackage_possibleTypes: string[] = ['PluginPackage']
    export const isPluginPackage = (obj?: { __typename?: any } | null): obj is PluginPackage => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginPackage"')
      return PluginPackage_possibleTypes.includes(obj.__typename)
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
    


    const Record_possibleTypes: string[] = ['Video','Gallery','Audio','Story','Post','Performer','Studio','Tag','Group','Category','Collection','CamModel','Segment','Clip']
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
    

export const enumChannelKind = {
   inApp: 'inApp' as const,
   push: 'push' as const,
   ntfy: 'ntfy' as const,
   webhook: 'webhook' as const
}

export const enumLimitKind = {
   watchTime: 'watchTime' as const,
   sessions: 'sessions' as const
}

export const enumLimitWindow = {
   day: 'day' as const,
   week: 'week' as const
}

export const enumLimitAction = {
   remind: 'remind' as const,
   lock: 'lock' as const
}

export const enumTriggerKind = {
   event: 'event' as const,
   schedule: 'schedule' as const,
   newMedia: 'newMedia' as const,
   camLive: 'camLive' as const,
   fileAdded: 'fileAdded' as const,
   wantedFound: 'wantedFound' as const,
   plugin: 'plugin' as const
}

export const enumRunStatus = {
   running: 'running' as const,
   succeeded: 'succeeded' as const,
   failed: 'failed' as const,
   undone: 'undone' as const
}

export const enumCollectionKind = {
   user: 'user' as const,
   watchlist: 'watchlist' as const,
   queue: 'queue' as const
}

export const enumFieldType = {
   text: 'text' as const,
   number: 'number' as const,
   boolean: 'boolean' as const,
   date: 'date' as const,
   choice: 'choice' as const,
   choices: 'choices' as const,
   url: 'url' as const
}

export const enumFingerprintAlgorithm = {
   phash: 'phash' as const,
   oshash: 'oshash' as const,
   md5: 'md5' as const,
   audio: 'audio' as const
}

export const enumEmbeddingKind = {
   face: 'face' as const,
   visual: 'visual' as const,
   text: 'text' as const
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

export const enumVrProjection = {
   equirectangular180: 'equirectangular180' as const,
   equirectangular360: 'equirectangular360' as const,
   fisheye: 'fisheye' as const,
   cubemap: 'cubemap' as const
}

export const enumVrStereo = {
   mono: 'mono' as const,
   sideBySide: 'sideBySide' as const,
   topBottom: 'topBottom' as const
}

export const enumAccountStatus = {
   ok: 'ok' as const,
   expired: 'expired' as const,
   needsLogin: 'needsLogin' as const,
   failed: 'failed' as const
}

export const enumSegmentOrigin = {
   user: 'user' as const,
   plugin: 'plugin' as const,
   ai: 'ai' as const
}

export const enumSubtitleOrigin = {
   site: 'site' as const,
   generated: 'generated' as const,
   translated: 'translated' as const
}

export const enumScriptKind = {
   funscript: 'funscript' as const
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
