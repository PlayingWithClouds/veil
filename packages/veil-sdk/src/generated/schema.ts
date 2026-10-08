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
}

export interface Query {
    vpnStatus: VpnStatus
    stream: (StreamResult | null)
    checkProviderSpeeds: ProviderSpeed[]
    blocklist: BlocklistEntry[]
    changes: Change[]
    collections: Collection[]
    collection: (Collection | null)
    collectionMembers: CollectionMember[]
    collectionIdsForMedia: Scalars['ID'][]
    galleries: Gallery[]
    gallery: (Gallery | null)
    images: Image[]
    image: (Image | null)
    jobs: Job[]
    job: (Job | null)
    downloadedScenes: Scene[]
    mediaCards: MediaCard[]
    oCount: Scalars['Int']
    observation: (Observation | null)
    observations: Observation[]
    performers: Performer[]
    performer: (Performer | null)
    plugins: Plugin[]
    plugin: (Plugin | null)
    pluginPackages: PluginPackage[]
    pluginCategories: PluginCategory[]
    recommendations: RecommendedScene[]
    recommendedRows: RecommendationRow[]
    recommendedCategories: RecommendedCategory[]
    recommendationPair: Scene[]
    savedFilters: SavedFilter[]
    scenes: Scene[]
    scene: (Scene | null)
    randomScenes: Scene[]
    recommendedFeed: Scene[]
    sceneHeat: SceneHeat
    sceneMarkers: SceneMarker[]
    pluginSearch: PluginSearchResult[]
    pluginBrowse: PluginSearchResult[]
    recommendedBrowse: PluginSearchResult[]
    findAlikeSources: AlikeCandidate[]
    /** Suggestions for a partially typed query. Empty query = recent searches plus taste-based picks. */
    searchSuggestions: SearchSuggestion[]
    settings: Settings
    settingEntries: SettingEntry[]
    mediaStreams: Stream[]
    studios: Studio[]
    studio: (Studio | null)
    searchSubscriptions: SearchSubscription[]
    searchSubscription: (SearchSubscription | null)
    subscriptionForTarget: (SearchSubscription | null)
    subscriptionFeed: SubscriptionFeedItem[]
    tags: Tag[]
    tag: (Tag | null)
    userRating: (UserRating | null)
    userRatings: UserRating[]
    watchHistory: WatchHistory[]
    watchHistoryEntry: (WatchHistory | null)
    watchlist: WatchlistItem[]
    __typename: 'Query'
}

export interface Mutation {
    scrape: ScrapePayload
    queueDownload: Scalars['ID']
    addBlock: BlocklistEntry
    removeBlock: Scalars['Boolean']
    createCollection: Collection
    renameCollection: Collection
    deleteCollection: Scalars['Boolean']
    addToCollection: Scalars['Boolean']
    removeFromCollection: Scalars['Boolean']
    reorderCollection: Scalars['Boolean']
    setCollectionTags: Collection
    ensureGalleryImages: (Gallery | null)
    setImageTags: Image
    deleteJob: Scalars['Boolean']
    deleteJobsByKind: Scalars['Int']
    clearJobs: Scalars['Int']
    reorderDownloads: Scalars['Boolean']
    retryJob: Scalars['Boolean']
    incrementOCount: Scalars['Int']
    decrementOCount: Scalars['Int']
    updateObservationStatus: Observation
    setPerformerFavorite: Performer
    enrichPerformer: Scalars['Boolean']
    togglePlugin: Scalars['Boolean']
    updatePluginSettings: Scalars['Boolean']
    installPlugin: Scalars['Boolean']
    uninstallPlugin: Scalars['Boolean']
    updatePlugins: Scalars['String'][]
    recordRecommendationChoice: Scalars['Boolean']
    recordImpressions: Scalars['Int']
    createSavedFilter: SavedFilter
    deleteSavedFilter: Scalars['Boolean']
    recordSceneHeat: Scalars['Boolean']
    setSceneThumbnail: Scalars['Boolean']
    createSceneMarker: SceneMarker
    deleteSceneMarker: Scalars['Boolean']
    resolvePluginResult: Scalars['ID']
    attachAlikeSource: Stream[]
    ensureEntityScenes: Scalars['Int']
    /** Remembers a search the user ran (called when a search is submitted). */
    recordSearch: Scalars['Boolean']
    /** Removes a query from the recent searches. */
    forgetSearch: Scalars['Boolean']
    updateSettings: Settings
    upsertSettingEntry: SettingEntry
    deleteSettingEntry: Scalars['Boolean']
    createStream: Stream
    deleteStream: Scalars['Boolean']
    ensureSceneStreams: Stream[]
    setStudioTags: Studio
    subscribeSearch: SearchSubscription
    subscribe: SearchSubscription
    updateSearchSubscription: SearchSubscription
    unsubscribeSearch: Scalars['Boolean']
    runSearchSubscription: SearchSubscription
    markSearchSubscriptionSeen: SearchSubscription
    upsertUserRating: UserRating
    deleteUserRating: Scalars['Boolean']
    upsertWatchHistory: WatchHistory
    deleteWatchHistory: Scalars['Boolean']
    addToWatchlist: WatchlistItem
    removeFromWatchlist: Scalars['Boolean']
    __typename: 'Mutation'
}

export interface Subscription {
    mediaAdded: MediaEvent
    jobUpdated: Job
    streamsChanged: Stream[]
    relatedChanged: Scene[]
    __typename: 'Subscription'
}

export interface VpnStatus {
    connected: Scalars['Boolean']
    interface: (Scalars['String'] | null)
    __typename: 'VpnStatus'
}

export interface StreamResult {
    url: Scalars['String']
    mimeType: Scalars['String']
    quality: (Scalars['String'] | null)
    headers: StreamHeader[]
    loadTimeMs: (Scalars['Float'] | null)
    speedBps: (Scalars['Float'] | null)
    __typename: 'StreamResult'
}

export interface StreamHeader {
    name: Scalars['String']
    value: Scalars['String']
    __typename: 'StreamHeader'
}

export interface ProviderSpeed {
    url: Scalars['String']
    loadTimeMs: (Scalars['Float'] | null)
    speedBps: (Scalars['Float'] | null)
    error: (Scalars['String'] | null)
    __typename: 'ProviderSpeed'
}

export interface ScrapePayload {
    success: Scalars['Boolean']
    error: (Scalars['String'] | null)
    mediaId: (Scalars['ID'] | null)
    __typename: 'ScrapePayload'
}

export interface MediaEvent {
    id: Scalars['ID']
    type: Scalars['String']
    title: Scalars['String']
    __typename: 'MediaEvent'
}

export interface BlocklistEntry {
    id: Scalars['ID']
    kind: Scalars['String']
    targetId: Scalars['ID']
    label: (Scalars['String'] | null)
    createdAt: Scalars['String']
    __typename: 'BlocklistEntry'
}

export interface Change {
    id: Scalars['ID']
    canonical: Scalars['ID']
    observation: Scalars['ID']
    plugin: Scalars['String']
    key: Scalars['String']
    action: Scalars['String']
    value: (Scalars['String'] | null)
    originalValue: (Scalars['String'] | null)
    iso6391: (Scalars['String'] | null)
    iso31661: (Scalars['String'] | null)
    score: Scalars['Float']
    mode: Scalars['String']
    changedBy: (Scalars['String'] | null)
    createdAt: Scalars['String']
    __typename: 'Change'
}

export interface Collection {
    id: Scalars['ID']
    name: Scalars['String']
    details: (Scalars['String'] | null)
    coverPath: (Scalars['String'] | null)
    itemCount: Scalars['Int']
    origin: Scalars['String']
    sourceUrl: (Scalars['String'] | null)
    externalId: (Scalars['String'] | null)
    tags: Tag[]
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Collection'
}

export interface CollectionMember {
    mediaId: Scalars['ID']
    mediaType: Scalars['String']
    position: Scalars['Int']
    title: Scalars['String']
    posterPath: (Scalars['String'] | null)
    __typename: 'CollectionMember'
}

export interface Gallery {
    id: Scalars['ID']
    externalId: Scalars['String']
    sourceUrl: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    coverPath: (Scalars['String'] | null)
    imageCount: Scalars['Int']
    rating: (Scalars['Float'] | null)
    organized: Scalars['Boolean']
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    images: Image[]
    tagMatch: (Scalars['String'] | null)
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Gallery'
}

export interface Image {
    id: Scalars['ID']
    filePath: Scalars['String']
    title: (Scalars['String'] | null)
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    sourceUrl: (Scalars['String'] | null)
    externalId: (Scalars['String'] | null)
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    position: (Scalars['Int'] | null)
    rating: (Scalars['Float'] | null)
    organized: Scalars['Boolean']
    galleryId: (Scalars['ID'] | null)
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    tagMatch: (Scalars['String'] | null)
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Image'
}

export interface Job {
    id: Scalars['ID']
    kind: Scalars['String']
    status: Scalars['String']
    title: (Scalars['String'] | null)
    description: (Scalars['String'] | null)
    pluginName: Scalars['String']
    target: (Scalars['ID'] | null)
    error: (Scalars['String'] | null)
    attempts: Scalars['Int']
    maxAttempts: Scalars['Int']
    priority: Scalars['Int']
    runAt: (Scalars['String'] | null)
    startedAt: (Scalars['String'] | null)
    finishedAt: (Scalars['String'] | null)
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    downloadTitle: (Scalars['String'] | null)
    downloadUrl: (Scalars['String'] | null)
    downloadProgress: (Scalars['Float'] | null)
    downloadBytesReceived: (Scalars['Float'] | null)
    downloadBytesTotal: (Scalars['Float'] | null)
    scene: (Scene | null)
    __typename: 'Job'
}

export interface MediaCard {
    mediaId: Scalars['ID']
    mediaType: Scalars['String']
    title: Scalars['String']
    posterPath: (Scalars['String'] | null)
    backdropPath: (Scalars['String'] | null)
    __typename: 'MediaCard'
}

export type ObservationStatus = 'PENDING' | 'MERGED' | 'CONFLICTED' | 'REJECTED'

export interface Observation {
    id: Scalars['ID']
    target: Scalars['ID']
    plugin: Scalars['String']
    sourceUrl: Scalars['String']
    confidence: Scalars['Float']
    status: ObservationStatus
    data: (Scalars['String'] | null)
    job: (Scalars['ID'] | null)
    observedAt: Scalars['String']
    __typename: 'Observation'
}

export interface Performer {
    id: Scalars['ID']
    name: Scalars['String']
    aliases: Scalars['String'][]
    details: (Scalars['String'] | null)
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
    careerLength: (Scalars['String'] | null)
    url: (Scalars['String'] | null)
    twitter: (Scalars['String'] | null)
    instagram: (Scalars['String'] | null)
    imagePath: (Scalars['String'] | null)
    favorite: Scalars['Boolean']
    rating: (Scalars['Float'] | null)
    tags: Tag[]
    sceneCount: Scalars['Int']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Performer'
}

export interface PluginPackage {
    name: Scalars['String']
    version: Scalars['String']
    description: (Scalars['String'] | null)
    installedVersion: (Scalars['String'] | null)
    __typename: 'PluginPackage'
}

export interface Plugin {
    id: Scalars['ID']
    name: Scalars['String']
    packageName: (Scalars['String'] | null)
    localBuild: Scalars['Boolean']
    displayName: (Scalars['String'] | null)
    iconUrl: (Scalars['String'] | null)
    description: (Scalars['String'] | null)
    version: Scalars['String']
    capabilities: Scalars['String'][]
    domains: Scalars['String'][]
    enabled: Scalars['Boolean']
    requiresSolver: Scalars['Boolean']
    available: Scalars['Boolean']
    installedAt: Scalars['String']
    updatedAt: Scalars['String']
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

export interface PluginCategory {
    id: Scalars['ID']
    name: Scalars['String']
    url: Scalars['String']
    count: Scalars['Int']
    poster: (Scalars['String'] | null)
    __typename: 'PluginCategory'
}

export interface RecommendedScene {
    scene: Scene
    source: Scalars['String']
    reason: RecommendationReason
    score: Scalars['Float']
    __typename: 'RecommendedScene'
}

export interface RecommendationReason {
    kind: Scalars['String']
    text: Scalars['String']
    entityId: (Scalars['ID'] | null)
    entityName: (Scalars['String'] | null)
    __typename: 'RecommendationReason'
}

export interface RecommendationRow {
    key: Scalars['String']
    title: Scalars['String']
    reason: RecommendationReason
    items: RecommendedScene[]
    __typename: 'RecommendationRow'
}

export interface RecommendedCategory {
    tag: Tag
    scenes: Scene[]
    __typename: 'RecommendedCategory'
}

export interface SceneFilter {
    search: (Scalars['String'] | null)
    studioId: (Scalars['ID'] | null)
    performerId: (Scalars['ID'] | null)
    tagId: (Scalars['ID'] | null)
    minRating: (Scalars['Float'] | null)
    minDuration: (Scalars['Int'] | null)
    maxDuration: (Scalars['Int'] | null)
    dateFrom: (Scalars['String'] | null)
    dateTo: (Scalars['String'] | null)
    sort: (Scalars['String'] | null)
    sources: (Scalars['String'][] | null)
    __typename: 'SceneFilter'
}

export interface SavedFilter {
    id: Scalars['ID']
    name: Scalars['String']
    filter: SceneFilter
    createdAt: Scalars['String']
    __typename: 'SavedFilter'
}

export interface Scene {
    id: Scalars['ID']
    externalId: Scalars['String']
    sourceUrl: Scalars['String']
    title: Scalars['String']
    details: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    durationSeconds: (Scalars['Int'] | null)
    rating: (Scalars['Float'] | null)
    viewCount: Scalars['Int']
    organized: Scalars['Boolean']
    posterPath: (Scalars['String'] | null)
    previewVideo: (Scalars['String'] | null)
    previewImages: Scalars['String'][]
    studio: (Studio | null)
    performers: Performer[]
    tags: Tag[]
    tagMatch: (Scalars['String'] | null)
    related: Scene[]
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Scene'
}

export interface SceneHeat {
    buckets: Scalars['Float'][]
    bestMomentSeconds: (Scalars['Float'] | null)
    thumbnailSeconds: (Scalars['Float'] | null)
    __typename: 'SceneHeat'
}

export interface SceneMarker {
    id: Scalars['ID']
    tag: (Tag | null)
    label: (Scalars['String'] | null)
    seconds: Scalars['Float']
    endSeconds: (Scalars['Float'] | null)
    personal: Scalars['Boolean']
    createdAt: Scalars['String']
    __typename: 'SceneMarker'
}

export interface PluginSearchResult {
    externalId: Scalars['String']
    title: Scalars['String']
    mediaType: Scalars['String']
    plugin: Scalars['String']
    sourceUrl: Scalars['String']
    date: (Scalars['String'] | null)
    posterUrl: (Scalars['String'] | null)
    previewImages: Scalars['String'][]
    previewVideo: (Scalars['String'] | null)
    durationSeconds: (Scalars['Int'] | null)
    __typename: 'PluginSearchResult'
}

export interface AlikeCandidate {
    externalId: Scalars['String']
    title: Scalars['String']
    plugin: Scalars['String']
    sourceUrl: Scalars['String']
    posterUrl: (Scalars['String'] | null)
    date: (Scalars['String'] | null)
    previewImages: Scalars['String'][]
    previewVideo: (Scalars['String'] | null)
    durationSeconds: (Scalars['Int'] | null)
    matchScore: Scalars['Float']
    __typename: 'AlikeCandidate'
}

export type SearchSuggestionKind = 'RECENT' | 'QUERY' | 'TAG' | 'PERFORMER' | 'STUDIO'

export interface SearchSuggestion {
    kind: SearchSuggestionKind
    /** Text to show and to search for when chosen. */
    text: Scalars['String']
    /** Entity id (tag/performer/studio) when the suggestion is an entity, never the raw record id structure — plain string id like other *Id fields. */
    entityId: (Scalars['String'] | null)
    /** Avatar/logo/profile image URL for performers/studios, null otherwise. */
    imageUrl: (Scalars['String'] | null)
    /** Short grey detail line, e.g. '128 videos' or 'Performer'. */
    detail: (Scalars['String'] | null)
    __typename: 'SearchSuggestion'
}

export interface Settings {
    maxConcurrentJobs: Scalars['Int']
    maxJobRetries: Scalars['Int']
    downloadSpeedLimitKBps: Scalars['Int']
    allowDownloadsWhileStreaming: Scalars['Boolean']
    autoEnrichAfterScrape: Scalars['Boolean']
    requireVpn: Scalars['Boolean']
    kindLimits: KindRateLimit[]
    __typename: 'Settings'
}

export interface KindRateLimit {
    kind: Scalars['String']
    maxConcurrent: Scalars['Int']
    retryInitialMs: Scalars['Int']
    retryMultiplier: Scalars['Float']
    retryMaxMs: Scalars['Int']
    __typename: 'KindRateLimit'
}

export interface SettingEntry {
    id: Scalars['ID']
    scope: Scalars['String']
    plugin: (Scalars['ID'] | null)
    key: Scalars['String']
    value: (Scalars['String'] | null)
    description: (Scalars['String'] | null)
    updatedAt: Scalars['String']
    __typename: 'SettingEntry'
}

export interface Stream {
    id: Scalars['ID']
    mediaId: Scalars['ID']
    url: Scalars['String']
    kind: Scalars['String']
    label: (Scalars['String'] | null)
    provider: (Scalars['String'] | null)
    resolution: (Scalars['String'] | null)
    width: (Scalars['Int'] | null)
    height: (Scalars['Int'] | null)
    language: (Scalars['String'] | null)
    format: (Scalars['String'] | null)
    mimeType: (Scalars['String'] | null)
    expectedSpeedBps: (Scalars['Float'] | null)
    fileSizeBytes: (Scalars['Float'] | null)
    verified: Scalars['Boolean']
    pluginName: (Scalars['String'] | null)
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Stream'
}

export interface Studio {
    id: Scalars['ID']
    name: Scalars['String']
    aliases: Scalars['String'][]
    url: (Scalars['String'] | null)
    parent: (Studio | null)
    imagePath: (Scalars['String'] | null)
    details: (Scalars['String'] | null)
    tags: Tag[]
    sceneCount: Scalars['Int']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Studio'
}

export interface SubscriptionFeedItem {
    scene: Scene
    foundAt: Scalars['String']
    subscriptionId: Scalars['ID']
    isNew: Scalars['Boolean']
    __typename: 'SubscriptionFeedItem'
}

export type SubscriptionKind = 'SEARCH' | 'STUDIO' | 'PERFORMER' | 'TAG'

export interface SubscriptionTarget {
    id: Scalars['ID']
    name: Scalars['String']
    imageUrl: (Scalars['String'] | null)
    __typename: 'SubscriptionTarget'
}

export interface SearchSubscription {
    id: Scalars['ID']
    kind: SubscriptionKind
    query: Scalars['String']
    target: (SubscriptionTarget | null)
    sources: Scalars['String'][]
    intervalHours: Scalars['Int']
    enabled: Scalars['Boolean']
    lastRunAt: (Scalars['String'] | null)
    nextRunAt: (Scalars['String'] | null)
    lastError: (Scalars['String'] | null)
    newCount: Scalars['Int']
    totalCount: Scalars['Int']
    createdAt: Scalars['String']
    scenes: Scene[]
    __typename: 'SearchSubscription'
}

export interface Tag {
    id: Scalars['ID']
    name: Scalars['String']
    aliases: Scalars['String'][]
    description: (Scalars['String'] | null)
    category: (Scalars['String'] | null)
    sceneCount: Scalars['Int']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'Tag'
}

export interface UserRating {
    id: Scalars['ID']
    media: Scalars['ID']
    rating: Scalars['Float']
    createdAt: Scalars['String']
    updatedAt: Scalars['String']
    __typename: 'UserRating'
}

export interface WatchHistory {
    id: Scalars['ID']
    media: Scalars['ID']
    scene: (Scene | null)
    startedAt: Scalars['String']
    finishedAt: (Scalars['String'] | null)
    progressSeconds: Scalars['Int']
    maxProgressSeconds: Scalars['Int']
    durationSeconds: (Scalars['Int'] | null)
    completed: Scalars['Boolean']
    updatedAt: Scalars['String']
    __typename: 'WatchHistory'
}

export interface WatchlistItem {
    id: Scalars['ID']
    media: Scalars['ID']
    createdAt: Scalars['String']
    __typename: 'WatchlistItem'
}

export interface QueryGenqlSelection{
    vpnStatus?: VpnStatusGenqlSelection
    stream?: (StreamResultGenqlSelection & { __args: {url: Scalars['String'], pluginName?: (Scalars['String'] | null)} })
    checkProviderSpeeds?: (ProviderSpeedGenqlSelection & { __args: {id: Scalars['ID']} })
    blocklist?: BlocklistEntryGenqlSelection
    changes?: (ChangeGenqlSelection & { __args: {targetId: Scalars['ID'], key?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    collections?: (CollectionGenqlSelection & { __args?: {search?: (Scalars['String'] | null), tagId?: (Scalars['ID'] | null), origin?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    collection?: (CollectionGenqlSelection & { __args: {id: Scalars['ID']} })
    collectionMembers?: (CollectionMemberGenqlSelection & { __args: {collectionId: Scalars['ID']} })
    collectionIdsForMedia?: { __args: {mediaId: Scalars['ID']} }
    galleries?: (GalleryGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null), search?: (Scalars['String'] | null), studioId?: (Scalars['ID'] | null), performerId?: (Scalars['ID'] | null), tagId?: (Scalars['ID'] | null), sources?: (Scalars['String'][] | null)} })
    gallery?: (GalleryGenqlSelection & { __args: {id: Scalars['ID']} })
    images?: (ImageGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null), search?: (Scalars['String'] | null), studioId?: (Scalars['ID'] | null), performerId?: (Scalars['ID'] | null), tagId?: (Scalars['ID'] | null), galleryId?: (Scalars['ID'] | null), sort?: (Scalars['String'] | null)} })
    image?: (ImageGenqlSelection & { __args: {id: Scalars['ID']} })
    jobs?: JobGenqlSelection
    job?: (JobGenqlSelection & { __args: {id: Scalars['ID']} })
    downloadedScenes?: SceneGenqlSelection
    mediaCards?: (MediaCardGenqlSelection & { __args: {ids: Scalars['ID'][]} })
    oCount?: { __args: {mediaId: Scalars['ID']} }
    observation?: (ObservationGenqlSelection & { __args: {id: Scalars['ID']} })
    observations?: (ObservationGenqlSelection & { __args: {targetId: Scalars['ID'], status?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    performers?: (PerformerGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null), search?: (Scalars['String'] | null), tagId?: (Scalars['ID'] | null), sort?: (Scalars['String'] | null)} })
    performer?: (PerformerGenqlSelection & { __args: {id: Scalars['ID']} })
    plugins?: PluginGenqlSelection
    plugin?: (PluginGenqlSelection & { __args: {id: Scalars['ID']} })
    pluginPackages?: (PluginPackageGenqlSelection & { __args?: {query?: (Scalars['String'] | null)} })
    pluginCategories?: (PluginCategoryGenqlSelection & { __args: {plugin: Scalars['String'], limit?: (Scalars['Int'] | null)} })
    recommendations?: (RecommendedSceneGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null), refresh?: (Scalars['Boolean'] | null), sources?: (Scalars['String'][] | null), minDuration?: (Scalars['Int'] | null), maxDuration?: (Scalars['Int'] | null)} })
    recommendedRows?: (RecommendationRowGenqlSelection & { __args?: {rowLimit?: (Scalars['Int'] | null), perRow?: (Scalars['Int'] | null)} })
    recommendedCategories?: (RecommendedCategoryGenqlSelection & { __args?: {categoryLimit?: (Scalars['Int'] | null), perCategory?: (Scalars['Int'] | null)} })
    recommendationPair?: SceneGenqlSelection
    savedFilters?: SavedFilterGenqlSelection
    scenes?: (SceneGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null), search?: (Scalars['String'] | null), studioId?: (Scalars['ID'] | null), performerId?: (Scalars['ID'] | null), tagId?: (Scalars['ID'] | null), sort?: (Scalars['String'] | null), minRating?: (Scalars['Float'] | null), minDuration?: (Scalars['Int'] | null), maxDuration?: (Scalars['Int'] | null), dateFrom?: (Scalars['String'] | null), dateTo?: (Scalars['String'] | null), sources?: (Scalars['String'][] | null)} })
    scene?: (SceneGenqlSelection & { __args: {id: Scalars['ID']} })
    randomScenes?: (SceneGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null)} })
    recommendedFeed?: (SceneGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    sceneHeat?: (SceneHeatGenqlSelection & { __args: {sceneId: Scalars['ID']} })
    sceneMarkers?: (SceneMarkerGenqlSelection & { __args: {mediaId: Scalars['ID']} })
    pluginSearch?: (PluginSearchResultGenqlSelection & { __args: {query: Scalars['String'], limit?: (Scalars['Int'] | null), pluginNames?: (Scalars['String'][] | null)} })
    pluginBrowse?: (PluginSearchResultGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    recommendedBrowse?: (PluginSearchResultGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null)} })
    findAlikeSources?: (AlikeCandidateGenqlSelection & { __args: {sceneId: Scalars['ID'], limit?: (Scalars['Int'] | null)} })
    /** Suggestions for a partially typed query. Empty query = recent searches plus taste-based picks. */
    searchSuggestions?: (SearchSuggestionGenqlSelection & { __args: {query: Scalars['String'], limit?: (Scalars['Int'] | null)} })
    settings?: SettingsGenqlSelection
    settingEntries?: (SettingEntryGenqlSelection & { __args?: {scope?: (Scalars['String'] | null)} })
    mediaStreams?: (StreamGenqlSelection & { __args: {mediaId: Scalars['ID']} })
    studios?: (StudioGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null), search?: (Scalars['String'] | null), tagId?: (Scalars['ID'] | null), sort?: (Scalars['String'] | null)} })
    studio?: (StudioGenqlSelection & { __args: {id: Scalars['ID']} })
    searchSubscriptions?: SearchSubscriptionGenqlSelection
    searchSubscription?: (SearchSubscriptionGenqlSelection & { __args: {id: Scalars['ID']} })
    subscriptionForTarget?: (SearchSubscriptionGenqlSelection & { __args: {targetId: Scalars['ID']} })
    subscriptionFeed?: (SubscriptionFeedItemGenqlSelection & { __args?: {filter?: (SubscriptionFeedFilter | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    tags?: (TagGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null), search?: (Scalars['String'] | null), category?: (Scalars['String'] | null)} })
    tag?: (TagGenqlSelection & { __args: {id: Scalars['ID']} })
    userRating?: (UserRatingGenqlSelection & { __args: {mediaId: Scalars['ID']} })
    userRatings?: (UserRatingGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    watchHistory?: (WatchHistoryGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    watchHistoryEntry?: (WatchHistoryGenqlSelection & { __args: {mediaId: Scalars['ID']} })
    watchlist?: WatchlistItemGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MutationGenqlSelection{
    scrape?: (ScrapePayloadGenqlSelection & { __args: {pluginName: Scalars['String'], url: Scalars['String']} })
    queueDownload?: { __args: {url: Scalars['String'], title: Scalars['String'], pluginName?: (Scalars['String'] | null)} }
    addBlock?: (BlocklistEntryGenqlSelection & { __args: {kind: Scalars['String'], targetId: Scalars['ID'], label?: (Scalars['String'] | null)} })
    removeBlock?: { __args: {targetId: Scalars['ID']} }
    createCollection?: (CollectionGenqlSelection & { __args: {name: Scalars['String']} })
    renameCollection?: (CollectionGenqlSelection & { __args: {collectionId: Scalars['ID'], name: Scalars['String']} })
    deleteCollection?: { __args: {collectionId: Scalars['ID']} }
    addToCollection?: { __args: {collectionId: Scalars['ID'], mediaId: Scalars['ID']} }
    removeFromCollection?: { __args: {collectionId: Scalars['ID'], mediaId: Scalars['ID']} }
    reorderCollection?: { __args: {collectionId: Scalars['ID'], mediaIds: Scalars['ID'][]} }
    setCollectionTags?: (CollectionGenqlSelection & { __args: {collectionId: Scalars['ID'], tagIds: Scalars['ID'][]} })
    ensureGalleryImages?: (GalleryGenqlSelection & { __args: {galleryId: Scalars['ID']} })
    setImageTags?: (ImageGenqlSelection & { __args: {imageId: Scalars['ID'], tagIds: Scalars['ID'][]} })
    deleteJob?: { __args: {id: Scalars['ID']} }
    deleteJobsByKind?: { __args: {kind: Scalars['String']} }
    clearJobs?: boolean | number
    reorderDownloads?: { __args: {jobIds: Scalars['ID'][]} }
    retryJob?: { __args: {id: Scalars['ID']} }
    incrementOCount?: { __args: {mediaId: Scalars['ID']} }
    decrementOCount?: { __args: {mediaId: Scalars['ID']} }
    updateObservationStatus?: (ObservationGenqlSelection & { __args: {id: Scalars['ID'], status: ObservationStatus} })
    setPerformerFavorite?: (PerformerGenqlSelection & { __args: {id: Scalars['ID'], favorite: Scalars['Boolean']} })
    enrichPerformer?: { __args: {id: Scalars['ID']} }
    togglePlugin?: { __args: {name: Scalars['String'], enabled: Scalars['Boolean']} }
    updatePluginSettings?: { __args: {pluginName: Scalars['String'], values: PluginSettingValueInput[]} }
    installPlugin?: { __args: {packageName: Scalars['String']} }
    uninstallPlugin?: { __args: {name: Scalars['String']} }
    updatePlugins?: boolean | number
    recordRecommendationChoice?: { __args: {chosenMediaId: Scalars['ID'], rejectedMediaId: Scalars['ID']} }
    recordImpressions?: { __args: {impressions: ImpressionInput[]} }
    createSavedFilter?: (SavedFilterGenqlSelection & { __args: {name: Scalars['String'], filter: SceneFilterInput} })
    deleteSavedFilter?: { __args: {filterId: Scalars['ID']} }
    recordSceneHeat?: { __args: {input: RecordSceneHeatInput} }
    setSceneThumbnail?: { __args: {sceneId: Scalars['ID'], atSeconds: Scalars['Float'], jpegBase64: Scalars['String']} }
    createSceneMarker?: (SceneMarkerGenqlSelection & { __args: {mediaId: Scalars['ID'], seconds: Scalars['Float'], endSeconds?: (Scalars['Float'] | null), tagName?: (Scalars['String'] | null), label?: (Scalars['String'] | null)} })
    deleteSceneMarker?: { __args: {markerId: Scalars['ID']} }
    resolvePluginResult?: { __args: {pluginName: Scalars['String'], url: Scalars['String'], posterUrl?: (Scalars['String'] | null)} }
    attachAlikeSource?: (StreamGenqlSelection & { __args: {sceneId: Scalars['ID'], pluginName: Scalars['String'], url: Scalars['String']} })
    ensureEntityScenes?: { __args: {entityId: Scalars['ID']} }
    /** Remembers a search the user ran (called when a search is submitted). */
    recordSearch?: { __args: {query: Scalars['String']} }
    /** Removes a query from the recent searches. */
    forgetSearch?: { __args: {query: Scalars['String']} }
    updateSettings?: (SettingsGenqlSelection & { __args: {input: UpdateSettingsInput} })
    upsertSettingEntry?: (SettingEntryGenqlSelection & { __args: {input: UpsertSettingEntryInput} })
    deleteSettingEntry?: { __args: {id: Scalars['ID']} }
    createStream?: (StreamGenqlSelection & { __args: {input: CreateStreamInput} })
    deleteStream?: { __args: {id: Scalars['ID']} }
    ensureSceneStreams?: (StreamGenqlSelection & { __args: {sceneId: Scalars['ID']} })
    setStudioTags?: (StudioGenqlSelection & { __args: {studioId: Scalars['ID'], tagIds: Scalars['ID'][]} })
    subscribeSearch?: (SearchSubscriptionGenqlSelection & { __args: {query: Scalars['String'], sources?: (Scalars['String'][] | null), intervalHours?: (Scalars['Int'] | null)} })
    subscribe?: (SearchSubscriptionGenqlSelection & { __args: {kind: SubscriptionKind, targetId: Scalars['ID'], intervalHours?: (Scalars['Int'] | null)} })
    updateSearchSubscription?: (SearchSubscriptionGenqlSelection & { __args: {id: Scalars['ID'], intervalHours?: (Scalars['Int'] | null), enabled?: (Scalars['Boolean'] | null)} })
    unsubscribeSearch?: { __args: {id: Scalars['ID']} }
    runSearchSubscription?: (SearchSubscriptionGenqlSelection & { __args: {id: Scalars['ID']} })
    markSearchSubscriptionSeen?: (SearchSubscriptionGenqlSelection & { __args: {id: Scalars['ID']} })
    upsertUserRating?: (UserRatingGenqlSelection & { __args: {input: UpsertUserRatingInput} })
    deleteUserRating?: { __args: {mediaId: Scalars['ID']} }
    upsertWatchHistory?: (WatchHistoryGenqlSelection & { __args: {input: UpsertWatchHistoryInput} })
    deleteWatchHistory?: { __args: {mediaId: Scalars['ID']} }
    addToWatchlist?: (WatchlistItemGenqlSelection & { __args: {mediaId: Scalars['ID']} })
    removeFromWatchlist?: { __args: {mediaId: Scalars['ID']} }
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SubscriptionGenqlSelection{
    mediaAdded?: MediaEventGenqlSelection
    jobUpdated?: JobGenqlSelection
    streamsChanged?: (StreamGenqlSelection & { __args: {mediaId: Scalars['ID']} })
    relatedChanged?: (SceneGenqlSelection & { __args: {sceneId: Scalars['ID']} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface VpnStatusGenqlSelection{
    connected?: boolean | number
    interface?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StreamResultGenqlSelection{
    url?: boolean | number
    mimeType?: boolean | number
    quality?: boolean | number
    headers?: StreamHeaderGenqlSelection
    loadTimeMs?: boolean | number
    speedBps?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface StreamHeaderGenqlSelection{
    name?: boolean | number
    value?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ProviderSpeedGenqlSelection{
    url?: boolean | number
    loadTimeMs?: boolean | number
    speedBps?: boolean | number
    error?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ScrapePayloadGenqlSelection{
    success?: boolean | number
    error?: boolean | number
    mediaId?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MediaEventGenqlSelection{
    id?: boolean | number
    type?: boolean | number
    title?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface BlocklistEntryGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    targetId?: boolean | number
    label?: boolean | number
    createdAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ChangeGenqlSelection{
    id?: boolean | number
    canonical?: boolean | number
    observation?: boolean | number
    plugin?: boolean | number
    key?: boolean | number
    action?: boolean | number
    value?: boolean | number
    originalValue?: boolean | number
    iso6391?: boolean | number
    iso31661?: boolean | number
    score?: boolean | number
    mode?: boolean | number
    changedBy?: boolean | number
    createdAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CollectionGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    details?: boolean | number
    coverPath?: boolean | number
    itemCount?: boolean | number
    origin?: boolean | number
    sourceUrl?: boolean | number
    externalId?: boolean | number
    tags?: TagGenqlSelection
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CollectionMemberGenqlSelection{
    mediaId?: boolean | number
    mediaType?: boolean | number
    position?: boolean | number
    title?: boolean | number
    posterPath?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface GalleryGenqlSelection{
    id?: boolean | number
    externalId?: boolean | number
    sourceUrl?: boolean | number
    title?: boolean | number
    details?: boolean | number
    date?: boolean | number
    coverPath?: boolean | number
    imageCount?: boolean | number
    rating?: boolean | number
    organized?: boolean | number
    studio?: StudioGenqlSelection
    performers?: PerformerGenqlSelection
    tags?: TagGenqlSelection
    images?: ImageGenqlSelection
    tagMatch?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ImageGenqlSelection{
    id?: boolean | number
    filePath?: boolean | number
    title?: boolean | number
    details?: boolean | number
    date?: boolean | number
    sourceUrl?: boolean | number
    externalId?: boolean | number
    width?: boolean | number
    height?: boolean | number
    position?: boolean | number
    rating?: boolean | number
    organized?: boolean | number
    galleryId?: boolean | number
    studio?: StudioGenqlSelection
    performers?: PerformerGenqlSelection
    tags?: TagGenqlSelection
    tagMatch?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface JobGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    status?: boolean | number
    title?: boolean | number
    description?: boolean | number
    pluginName?: boolean | number
    target?: boolean | number
    error?: boolean | number
    attempts?: boolean | number
    maxAttempts?: boolean | number
    priority?: boolean | number
    runAt?: boolean | number
    startedAt?: boolean | number
    finishedAt?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    downloadTitle?: boolean | number
    downloadUrl?: boolean | number
    downloadProgress?: boolean | number
    downloadBytesReceived?: boolean | number
    downloadBytesTotal?: boolean | number
    scene?: SceneGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MediaCardGenqlSelection{
    mediaId?: boolean | number
    mediaType?: boolean | number
    title?: boolean | number
    posterPath?: boolean | number
    backdropPath?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ObservationGenqlSelection{
    id?: boolean | number
    target?: boolean | number
    plugin?: boolean | number
    sourceUrl?: boolean | number
    confidence?: boolean | number
    status?: boolean | number
    data?: boolean | number
    job?: boolean | number
    observedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PerformerGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    aliases?: boolean | number
    details?: boolean | number
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
    careerLength?: boolean | number
    url?: boolean | number
    twitter?: boolean | number
    instagram?: boolean | number
    imagePath?: boolean | number
    favorite?: boolean | number
    rating?: boolean | number
    tags?: TagGenqlSelection
    sceneCount?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
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
    id?: boolean | number
    name?: boolean | number
    packageName?: boolean | number
    localBuild?: boolean | number
    displayName?: boolean | number
    iconUrl?: boolean | number
    description?: boolean | number
    version?: boolean | number
    capabilities?: boolean | number
    domains?: boolean | number
    enabled?: boolean | number
    requiresSolver?: boolean | number
    available?: boolean | number
    installedAt?: boolean | number
    updatedAt?: boolean | number
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

export interface PluginCategoryGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    url?: boolean | number
    count?: boolean | number
    poster?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface ImpressionInput {mediaId: Scalars['ID'],source?: (Scalars['String'] | null),surface?: (Scalars['String'] | null),position?: (Scalars['Int'] | null),clicked?: (Scalars['Boolean'] | null)}

export interface RecommendedSceneGenqlSelection{
    scene?: SceneGenqlSelection
    source?: boolean | number
    reason?: RecommendationReasonGenqlSelection
    score?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RecommendationReasonGenqlSelection{
    kind?: boolean | number
    text?: boolean | number
    entityId?: boolean | number
    entityName?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RecommendationRowGenqlSelection{
    key?: boolean | number
    title?: boolean | number
    reason?: RecommendationReasonGenqlSelection
    items?: RecommendedSceneGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RecommendedCategoryGenqlSelection{
    tag?: TagGenqlSelection
    scenes?: SceneGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SceneFilterInput {search?: (Scalars['String'] | null),studioId?: (Scalars['ID'] | null),performerId?: (Scalars['ID'] | null),tagId?: (Scalars['ID'] | null),minRating?: (Scalars['Float'] | null),minDuration?: (Scalars['Int'] | null),maxDuration?: (Scalars['Int'] | null),dateFrom?: (Scalars['String'] | null),dateTo?: (Scalars['String'] | null),sort?: (Scalars['String'] | null),sources?: (Scalars['String'][] | null)}

export interface SceneFilterGenqlSelection{
    search?: boolean | number
    studioId?: boolean | number
    performerId?: boolean | number
    tagId?: boolean | number
    minRating?: boolean | number
    minDuration?: boolean | number
    maxDuration?: boolean | number
    dateFrom?: boolean | number
    dateTo?: boolean | number
    sort?: boolean | number
    sources?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SavedFilterGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    filter?: SceneFilterGenqlSelection
    createdAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SceneGenqlSelection{
    id?: boolean | number
    externalId?: boolean | number
    sourceUrl?: boolean | number
    title?: boolean | number
    details?: boolean | number
    date?: boolean | number
    durationSeconds?: boolean | number
    rating?: boolean | number
    viewCount?: boolean | number
    organized?: boolean | number
    posterPath?: boolean | number
    previewVideo?: boolean | number
    previewImages?: boolean | number
    studio?: StudioGenqlSelection
    performers?: PerformerGenqlSelection
    tags?: TagGenqlSelection
    tagMatch?: boolean | number
    related?: (SceneGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null)} })
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SceneHeatGenqlSelection{
    buckets?: boolean | number
    bestMomentSeconds?: boolean | number
    thumbnailSeconds?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface RecordSceneHeatInput {sceneId: Scalars['ID'],durationSeconds: Scalars['Float'],spans: HeatSpanInput[],scrubs: Scalars['Float'][]}

export interface HeatSpanInput {fromSeconds: Scalars['Float'],toSeconds: Scalars['Float']}

export interface SceneMarkerGenqlSelection{
    id?: boolean | number
    tag?: TagGenqlSelection
    label?: boolean | number
    seconds?: boolean | number
    endSeconds?: boolean | number
    personal?: boolean | number
    createdAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface PluginSearchResultGenqlSelection{
    externalId?: boolean | number
    title?: boolean | number
    mediaType?: boolean | number
    plugin?: boolean | number
    sourceUrl?: boolean | number
    date?: boolean | number
    posterUrl?: boolean | number
    previewImages?: boolean | number
    previewVideo?: boolean | number
    durationSeconds?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface AlikeCandidateGenqlSelection{
    externalId?: boolean | number
    title?: boolean | number
    plugin?: boolean | number
    sourceUrl?: boolean | number
    posterUrl?: boolean | number
    date?: boolean | number
    previewImages?: boolean | number
    previewVideo?: boolean | number
    durationSeconds?: boolean | number
    matchScore?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SearchSuggestionGenqlSelection{
    kind?: boolean | number
    /** Text to show and to search for when chosen. */
    text?: boolean | number
    /** Entity id (tag/performer/studio) when the suggestion is an entity, never the raw record id structure — plain string id like other *Id fields. */
    entityId?: boolean | number
    /** Avatar/logo/profile image URL for performers/studios, null otherwise. */
    imageUrl?: boolean | number
    /** Short grey detail line, e.g. '128 videos' or 'Performer'. */
    detail?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SettingsGenqlSelection{
    maxConcurrentJobs?: boolean | number
    maxJobRetries?: boolean | number
    downloadSpeedLimitKBps?: boolean | number
    allowDownloadsWhileStreaming?: boolean | number
    autoEnrichAfterScrape?: boolean | number
    requireVpn?: boolean | number
    kindLimits?: KindRateLimitGenqlSelection
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UpdateSettingsInput {maxConcurrentJobs?: (Scalars['Int'] | null),maxJobRetries?: (Scalars['Int'] | null),downloadSpeedLimitKBps?: (Scalars['Int'] | null),allowDownloadsWhileStreaming?: (Scalars['Boolean'] | null),autoEnrichAfterScrape?: (Scalars['Boolean'] | null),requireVpn?: (Scalars['Boolean'] | null),kindLimits?: (KindRateLimitInput[] | null)}

export interface KindRateLimitGenqlSelection{
    kind?: boolean | number
    maxConcurrent?: boolean | number
    retryInitialMs?: boolean | number
    retryMultiplier?: boolean | number
    retryMaxMs?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface KindRateLimitInput {kind: Scalars['String'],maxConcurrent?: (Scalars['Int'] | null),retryInitialMs?: (Scalars['Int'] | null),retryMultiplier?: (Scalars['Float'] | null),retryMaxMs?: (Scalars['Int'] | null)}

export interface SettingEntryGenqlSelection{
    id?: boolean | number
    scope?: boolean | number
    plugin?: boolean | number
    key?: boolean | number
    value?: boolean | number
    description?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UpsertSettingEntryInput {scope: Scalars['String'],plugin?: (Scalars['ID'] | null),key: Scalars['String'],value: Scalars['String'],description?: (Scalars['String'] | null)}

export interface StreamGenqlSelection{
    id?: boolean | number
    mediaId?: boolean | number
    url?: boolean | number
    kind?: boolean | number
    label?: boolean | number
    provider?: boolean | number
    resolution?: boolean | number
    width?: boolean | number
    height?: boolean | number
    language?: boolean | number
    format?: boolean | number
    mimeType?: boolean | number
    expectedSpeedBps?: boolean | number
    fileSizeBytes?: boolean | number
    verified?: boolean | number
    pluginName?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface CreateStreamInput {mediaId: Scalars['ID'],url: Scalars['String'],kind?: (Scalars['String'] | null),label?: (Scalars['String'] | null),provider?: (Scalars['String'] | null),resolution?: (Scalars['String'] | null),width?: (Scalars['Int'] | null),height?: (Scalars['Int'] | null),language?: (Scalars['String'] | null),format?: (Scalars['String'] | null),mimeType?: (Scalars['String'] | null),expectedSpeedBps?: (Scalars['Float'] | null),fileSizeBytes?: (Scalars['Float'] | null),verified?: (Scalars['Boolean'] | null),pluginName?: (Scalars['String'] | null)}

export interface StudioGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    aliases?: boolean | number
    url?: boolean | number
    parent?: StudioGenqlSelection
    imagePath?: boolean | number
    details?: boolean | number
    tags?: TagGenqlSelection
    sceneCount?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SubscriptionFeedFilter {subscriptionId?: (Scalars['ID'] | null),kinds?: (SubscriptionKind[] | null),newOnly?: (Scalars['Boolean'] | null),unwatchedOnly?: (Scalars['Boolean'] | null)}

export interface SubscriptionFeedItemGenqlSelection{
    scene?: SceneGenqlSelection
    foundAt?: boolean | number
    subscriptionId?: boolean | number
    isNew?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SubscriptionTargetGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    imageUrl?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface SearchSubscriptionGenqlSelection{
    id?: boolean | number
    kind?: boolean | number
    query?: boolean | number
    target?: SubscriptionTargetGenqlSelection
    sources?: boolean | number
    intervalHours?: boolean | number
    enabled?: boolean | number
    lastRunAt?: boolean | number
    nextRunAt?: boolean | number
    lastError?: boolean | number
    newCount?: boolean | number
    totalCount?: boolean | number
    createdAt?: boolean | number
    scenes?: (SceneGenqlSelection & { __args?: {limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TagGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    aliases?: boolean | number
    description?: boolean | number
    category?: boolean | number
    sceneCount?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UserRatingGenqlSelection{
    id?: boolean | number
    media?: boolean | number
    rating?: boolean | number
    createdAt?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UpsertUserRatingInput {media: Scalars['ID'],rating: Scalars['Float']}

export interface WatchHistoryGenqlSelection{
    id?: boolean | number
    media?: boolean | number
    scene?: SceneGenqlSelection
    startedAt?: boolean | number
    finishedAt?: boolean | number
    progressSeconds?: boolean | number
    maxProgressSeconds?: boolean | number
    durationSeconds?: boolean | number
    completed?: boolean | number
    updatedAt?: boolean | number
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface UpsertWatchHistoryInput {media: Scalars['ID'],progressSeconds: Scalars['Int'],durationSeconds?: (Scalars['Int'] | null),completed?: (Scalars['Boolean'] | null)}

export interface WatchlistItemGenqlSelection{
    id?: boolean | number
    media?: boolean | number
    createdAt?: boolean | number
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
    


    const Subscription_possibleTypes: string[] = ['Subscription']
    export const isSubscription = (obj?: { __typename?: any } | null): obj is Subscription => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSubscription"')
      return Subscription_possibleTypes.includes(obj.__typename)
    }
    


    const VpnStatus_possibleTypes: string[] = ['VpnStatus']
    export const isVpnStatus = (obj?: { __typename?: any } | null): obj is VpnStatus => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isVpnStatus"')
      return VpnStatus_possibleTypes.includes(obj.__typename)
    }
    


    const StreamResult_possibleTypes: string[] = ['StreamResult']
    export const isStreamResult = (obj?: { __typename?: any } | null): obj is StreamResult => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStreamResult"')
      return StreamResult_possibleTypes.includes(obj.__typename)
    }
    


    const StreamHeader_possibleTypes: string[] = ['StreamHeader']
    export const isStreamHeader = (obj?: { __typename?: any } | null): obj is StreamHeader => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStreamHeader"')
      return StreamHeader_possibleTypes.includes(obj.__typename)
    }
    


    const ProviderSpeed_possibleTypes: string[] = ['ProviderSpeed']
    export const isProviderSpeed = (obj?: { __typename?: any } | null): obj is ProviderSpeed => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isProviderSpeed"')
      return ProviderSpeed_possibleTypes.includes(obj.__typename)
    }
    


    const ScrapePayload_possibleTypes: string[] = ['ScrapePayload']
    export const isScrapePayload = (obj?: { __typename?: any } | null): obj is ScrapePayload => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isScrapePayload"')
      return ScrapePayload_possibleTypes.includes(obj.__typename)
    }
    


    const MediaEvent_possibleTypes: string[] = ['MediaEvent']
    export const isMediaEvent = (obj?: { __typename?: any } | null): obj is MediaEvent => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMediaEvent"')
      return MediaEvent_possibleTypes.includes(obj.__typename)
    }
    


    const BlocklistEntry_possibleTypes: string[] = ['BlocklistEntry']
    export const isBlocklistEntry = (obj?: { __typename?: any } | null): obj is BlocklistEntry => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isBlocklistEntry"')
      return BlocklistEntry_possibleTypes.includes(obj.__typename)
    }
    


    const Change_possibleTypes: string[] = ['Change']
    export const isChange = (obj?: { __typename?: any } | null): obj is Change => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isChange"')
      return Change_possibleTypes.includes(obj.__typename)
    }
    


    const Collection_possibleTypes: string[] = ['Collection']
    export const isCollection = (obj?: { __typename?: any } | null): obj is Collection => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCollection"')
      return Collection_possibleTypes.includes(obj.__typename)
    }
    


    const CollectionMember_possibleTypes: string[] = ['CollectionMember']
    export const isCollectionMember = (obj?: { __typename?: any } | null): obj is CollectionMember => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isCollectionMember"')
      return CollectionMember_possibleTypes.includes(obj.__typename)
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
    


    const Job_possibleTypes: string[] = ['Job']
    export const isJob = (obj?: { __typename?: any } | null): obj is Job => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isJob"')
      return Job_possibleTypes.includes(obj.__typename)
    }
    


    const MediaCard_possibleTypes: string[] = ['MediaCard']
    export const isMediaCard = (obj?: { __typename?: any } | null): obj is MediaCard => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isMediaCard"')
      return MediaCard_possibleTypes.includes(obj.__typename)
    }
    


    const Observation_possibleTypes: string[] = ['Observation']
    export const isObservation = (obj?: { __typename?: any } | null): obj is Observation => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isObservation"')
      return Observation_possibleTypes.includes(obj.__typename)
    }
    


    const Performer_possibleTypes: string[] = ['Performer']
    export const isPerformer = (obj?: { __typename?: any } | null): obj is Performer => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPerformer"')
      return Performer_possibleTypes.includes(obj.__typename)
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
    


    const PluginCategory_possibleTypes: string[] = ['PluginCategory']
    export const isPluginCategory = (obj?: { __typename?: any } | null): obj is PluginCategory => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginCategory"')
      return PluginCategory_possibleTypes.includes(obj.__typename)
    }
    


    const RecommendedScene_possibleTypes: string[] = ['RecommendedScene']
    export const isRecommendedScene = (obj?: { __typename?: any } | null): obj is RecommendedScene => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecommendedScene"')
      return RecommendedScene_possibleTypes.includes(obj.__typename)
    }
    


    const RecommendationReason_possibleTypes: string[] = ['RecommendationReason']
    export const isRecommendationReason = (obj?: { __typename?: any } | null): obj is RecommendationReason => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecommendationReason"')
      return RecommendationReason_possibleTypes.includes(obj.__typename)
    }
    


    const RecommendationRow_possibleTypes: string[] = ['RecommendationRow']
    export const isRecommendationRow = (obj?: { __typename?: any } | null): obj is RecommendationRow => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecommendationRow"')
      return RecommendationRow_possibleTypes.includes(obj.__typename)
    }
    


    const RecommendedCategory_possibleTypes: string[] = ['RecommendedCategory']
    export const isRecommendedCategory = (obj?: { __typename?: any } | null): obj is RecommendedCategory => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isRecommendedCategory"')
      return RecommendedCategory_possibleTypes.includes(obj.__typename)
    }
    


    const SceneFilter_possibleTypes: string[] = ['SceneFilter']
    export const isSceneFilter = (obj?: { __typename?: any } | null): obj is SceneFilter => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSceneFilter"')
      return SceneFilter_possibleTypes.includes(obj.__typename)
    }
    


    const SavedFilter_possibleTypes: string[] = ['SavedFilter']
    export const isSavedFilter = (obj?: { __typename?: any } | null): obj is SavedFilter => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSavedFilter"')
      return SavedFilter_possibleTypes.includes(obj.__typename)
    }
    


    const Scene_possibleTypes: string[] = ['Scene']
    export const isScene = (obj?: { __typename?: any } | null): obj is Scene => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isScene"')
      return Scene_possibleTypes.includes(obj.__typename)
    }
    


    const SceneHeat_possibleTypes: string[] = ['SceneHeat']
    export const isSceneHeat = (obj?: { __typename?: any } | null): obj is SceneHeat => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSceneHeat"')
      return SceneHeat_possibleTypes.includes(obj.__typename)
    }
    


    const SceneMarker_possibleTypes: string[] = ['SceneMarker']
    export const isSceneMarker = (obj?: { __typename?: any } | null): obj is SceneMarker => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSceneMarker"')
      return SceneMarker_possibleTypes.includes(obj.__typename)
    }
    


    const PluginSearchResult_possibleTypes: string[] = ['PluginSearchResult']
    export const isPluginSearchResult = (obj?: { __typename?: any } | null): obj is PluginSearchResult => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isPluginSearchResult"')
      return PluginSearchResult_possibleTypes.includes(obj.__typename)
    }
    


    const AlikeCandidate_possibleTypes: string[] = ['AlikeCandidate']
    export const isAlikeCandidate = (obj?: { __typename?: any } | null): obj is AlikeCandidate => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isAlikeCandidate"')
      return AlikeCandidate_possibleTypes.includes(obj.__typename)
    }
    


    const SearchSuggestion_possibleTypes: string[] = ['SearchSuggestion']
    export const isSearchSuggestion = (obj?: { __typename?: any } | null): obj is SearchSuggestion => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSearchSuggestion"')
      return SearchSuggestion_possibleTypes.includes(obj.__typename)
    }
    


    const Settings_possibleTypes: string[] = ['Settings']
    export const isSettings = (obj?: { __typename?: any } | null): obj is Settings => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSettings"')
      return Settings_possibleTypes.includes(obj.__typename)
    }
    


    const KindRateLimit_possibleTypes: string[] = ['KindRateLimit']
    export const isKindRateLimit = (obj?: { __typename?: any } | null): obj is KindRateLimit => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isKindRateLimit"')
      return KindRateLimit_possibleTypes.includes(obj.__typename)
    }
    


    const SettingEntry_possibleTypes: string[] = ['SettingEntry']
    export const isSettingEntry = (obj?: { __typename?: any } | null): obj is SettingEntry => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSettingEntry"')
      return SettingEntry_possibleTypes.includes(obj.__typename)
    }
    


    const Stream_possibleTypes: string[] = ['Stream']
    export const isStream = (obj?: { __typename?: any } | null): obj is Stream => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStream"')
      return Stream_possibleTypes.includes(obj.__typename)
    }
    


    const Studio_possibleTypes: string[] = ['Studio']
    export const isStudio = (obj?: { __typename?: any } | null): obj is Studio => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isStudio"')
      return Studio_possibleTypes.includes(obj.__typename)
    }
    


    const SubscriptionFeedItem_possibleTypes: string[] = ['SubscriptionFeedItem']
    export const isSubscriptionFeedItem = (obj?: { __typename?: any } | null): obj is SubscriptionFeedItem => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSubscriptionFeedItem"')
      return SubscriptionFeedItem_possibleTypes.includes(obj.__typename)
    }
    


    const SubscriptionTarget_possibleTypes: string[] = ['SubscriptionTarget']
    export const isSubscriptionTarget = (obj?: { __typename?: any } | null): obj is SubscriptionTarget => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSubscriptionTarget"')
      return SubscriptionTarget_possibleTypes.includes(obj.__typename)
    }
    


    const SearchSubscription_possibleTypes: string[] = ['SearchSubscription']
    export const isSearchSubscription = (obj?: { __typename?: any } | null): obj is SearchSubscription => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isSearchSubscription"')
      return SearchSubscription_possibleTypes.includes(obj.__typename)
    }
    


    const Tag_possibleTypes: string[] = ['Tag']
    export const isTag = (obj?: { __typename?: any } | null): obj is Tag => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isTag"')
      return Tag_possibleTypes.includes(obj.__typename)
    }
    


    const UserRating_possibleTypes: string[] = ['UserRating']
    export const isUserRating = (obj?: { __typename?: any } | null): obj is UserRating => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isUserRating"')
      return UserRating_possibleTypes.includes(obj.__typename)
    }
    


    const WatchHistory_possibleTypes: string[] = ['WatchHistory']
    export const isWatchHistory = (obj?: { __typename?: any } | null): obj is WatchHistory => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isWatchHistory"')
      return WatchHistory_possibleTypes.includes(obj.__typename)
    }
    


    const WatchlistItem_possibleTypes: string[] = ['WatchlistItem']
    export const isWatchlistItem = (obj?: { __typename?: any } | null): obj is WatchlistItem => {
      if (!obj?.__typename) throw new Error('__typename is missing in "isWatchlistItem"')
      return WatchlistItem_possibleTypes.includes(obj.__typename)
    }
    

export const enumObservationStatus = {
   PENDING: 'PENDING' as const,
   MERGED: 'MERGED' as const,
   CONFLICTED: 'CONFLICTED' as const,
   REJECTED: 'REJECTED' as const
}

export const enumSearchSuggestionKind = {
   RECENT: 'RECENT' as const,
   QUERY: 'QUERY' as const,
   TAG: 'TAG' as const,
   PERFORMER: 'PERFORMER' as const,
   STUDIO: 'STUDIO' as const
}

export const enumSubscriptionKind = {
   SEARCH: 'SEARCH' as const,
   STUDIO: 'STUDIO' as const,
   PERFORMER: 'PERFORMER' as const,
   TAG: 'TAG' as const
}
