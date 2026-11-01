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
}

export interface Query {
    play: (Playable | null)
    videos: Video[]
    video: (Video | null)
    galleries: Gallery[]
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
    __typename: 'Query'
}

export interface Mutation {
    scrape: Scraped
    fetchVideo: (Video | null)
    fetchGallery: (Gallery | null)
    browse: Listed
    search: Listed
    setPluginEnabled: Scalars['Boolean']
    updatePluginSettings: Scalars['Boolean']
    installPlugin: Scalars['Boolean']
    uninstallPlugin: Scalars['Boolean']
    updatePlugins: Scalars['String'][]
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
    __typename: 'Studio'
}

export interface Tag {
    id: Scalars['ID']
    name: Scalars['String']
    aliases: Scalars['String'][]
    videoCount: Scalars['Int']
    galleryCount: Scalars['Int']
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

export interface QueryGenqlSelection{
    play?: (PlayableGenqlSelection & { __args: {url: Scalars['String']} })
    videos?: (VideoGenqlSelection & { __args?: {filter?: (MediaFilter | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    video?: (VideoGenqlSelection & { __args: {id: Scalars['ID']} })
    galleries?: (GalleryGenqlSelection & { __args?: {filter?: (MediaFilter | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
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
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface MutationGenqlSelection{
    scrape?: (ScrapedGenqlSelection & { __args: {plugin: Scalars['String'], url: Scalars['String']} })
    fetchVideo?: (VideoGenqlSelection & { __args: {id: Scalars['ID']} })
    fetchGallery?: (GalleryGenqlSelection & { __args: {id: Scalars['ID']} })
    browse?: (ListedGenqlSelection & { __args: {plugin: Scalars['String'], kind: ListKind, query?: (Scalars['String'] | null), limit?: (Scalars['Int'] | null), offset?: (Scalars['Int'] | null)} })
    search?: (ListedGenqlSelection & { __args: {kind: ListKind, query: Scalars['String'], limit?: (Scalars['Int'] | null)} })
    setPluginEnabled?: { __args: {name: Scalars['String'], enabled: Scalars['Boolean']} }
    updatePluginSettings?: { __args: {plugin: Scalars['String'], values: PluginSettingValueInput[]} }
    installPlugin?: { __args: {packageName: Scalars['String']} }
    uninstallPlugin?: { __args: {name: Scalars['String']} }
    updatePlugins?: boolean | number
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

export interface MediaFilter {text?: (Scalars['String'] | null),performerId?: (Scalars['ID'] | null),studioId?: (Scalars['ID'] | null),tagId?: (Scalars['ID'] | null)}

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
    __typename?: boolean | number
    __scalar?: boolean | number
}

export interface TagGenqlSelection{
    id?: boolean | number
    name?: boolean | number
    aliases?: boolean | number
    videoCount?: boolean | number
    galleryCount?: boolean | number
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
    

export const enumListKind = {
   VIDEO: 'VIDEO' as const,
   GALLERY: 'GALLERY' as const,
   PERFORMER: 'PERFORMER' as const,
   STUDIO: 'STUDIO' as const,
   TAG: 'TAG' as const
}
