// Every supported Kernel Video Sharing site, as configuration. KVS sites share
// the engine (flashvars/<source> media behind /get_file/, path-paginated
// listings, async get_block pagination) but each runs its own theme, so what
// differs per site is captured here: URL shapes, listing paths and the page
// blocks holding a video's credits and related list. Adding a KVS tube is
// adding an entry to SITES.

export interface SiteConfig {
  /** Short stable key; namespaces external ids as kvs-<key>-<video>. */
  key: string;
  /** Display name, also the studio credited when a video names no channel. */
  name: string;
  /** Canonical origin every URL is rewritten to, e.g. https://www.pornhat.com. */
  origin: string;
  /** Further hostnames serving the same catalog (mirrors). Subdomains of the origin host always match. */
  mirrors?: string[];
  /** Matches a video page's pathname; group 1 is the video's key (id or slug). */
  videoPath: RegExp;
  /** First page of the newest-videos listing. */
  latestPath: string;
  /** First page of a search: {query} is the dash-joined query, {q} the form-encoded one. */
  searchPath: string;
  /** Index of the site's categories/tags, listed by tag:list. */
  tagIndexPath?: string;
  /** Index of the site's models/pornstars, listed by performer:list. */
  modelIndexPath?: string;
  /** Path prefixes of tag/category pages, e.g. "/categories/". */
  tagPrefixes: string[];
  /** Path prefixes of model/pornstar pages. */
  modelPrefixes: string[];
  /** Path prefixes of channel/site pages, most specific first. */
  channelPrefixes: string[];
  /** The block on a video page that holds its credits (models, tags, channel). */
  infoSelector: string;
  /** The block on a video page that lists related videos. */
  relatedSelector: string;
  /** Element on a video page showing its runtime, when no metadata carries it. */
  durationSelector?: string;
  /** Element holding a listing card's title, when the anchor/img carry none. */
  cardTitleSelector?: string;
  /** All qualities redirect to one adaptive HLS master, so one download is offered. */
  adaptiveStreams?: boolean;
  /** The site answers plain requests with a Cloudflare challenge. */
  requiresSolver?: boolean;
}

/** Path prefixes most KVS themes use; sites override what differs. */
const DEFAULT_TAG_PREFIXES = ["/categories/", "/tags/"];
const DEFAULT_MODEL_PREFIXES = ["/models/", "/pornstars/"];
const DEFAULT_CHANNEL_PREFIXES = ["/channels/", "/sites/"];

/**
 * The stock KVS related block ids (list_videos_related_videos, custom
 * variants) plus the plain class some themes use instead.
 */
const DEFAULT_RELATED_SELECTOR = '[id*="related_videos"], [id*="related_sphinx"], .related-videos, .related';

type SiteOptions = Partial<SiteConfig> &
  Pick<SiteConfig, "key" | "name" | "origin" | "videoPath" | "latestPath" | "searchPath" | "infoSelector">;

/** A site entry with the common KVS defaults filled in. */
function kvsSite(options: SiteOptions): SiteConfig {
  return {
    tagIndexPath: "/categories/",
    modelIndexPath: "/models/",
    tagPrefixes: DEFAULT_TAG_PREFIXES,
    modelPrefixes: DEFAULT_MODEL_PREFIXES,
    channelPrefixes: DEFAULT_CHANNEL_PREFIXES,
    relatedSelector: DEFAULT_RELATED_SELECTOR,
    ...options,
  };
}

// ok.xxx, pornhat.com and perfectgirls.xxx run one build: the home page is the
// newest listing, credits sit in .video-info, every quality redirects to one
// HLS master on privatehost.com. ok.xxx and pornhat file paysites under a
// network (/paysite/ under /sites/).
const PRIVATEHOST_NETWORK = {
  latestPath: "/",
  searchPath: "/search/{query}/",
  tagIndexPath: "/tags/",
  modelIndexPath: "/models/",
  tagPrefixes: ["/tags/", "/categories/"],
  channelPrefixes: ["/paysite/", "/sites/", "/channels/", "/label/"],
  infoSelector: ".video-info",
  adaptiveStreams: true,
};

// porngo.com and xxxfiles.com share a theme (xxxfiles serves porngo's catalog).
const PORNGO_NETWORK = {
  latestPath: "/latest-updates/",
  searchPath: "/search/{query}/",
  tagPrefixes: ["/categories/"],
  infoSelector: ".video-links",
};

// w1mp.com, fuqster.com and fapnfuck.com share a theme: credits in the
// desktop player toolbar (a mobile copy precedes it).
const FQ_NETWORK = {
  videoPath: /^\/video\/(\d+)\//,
  latestPath: "/latest-updates/",
  searchPath: "/search/{query}/",
  infoSelector: ".player-tools.desc",
};

// The stock KVS theme: newest at /latest-updates/, credits in .block-details.
const STOCK_THEME = {
  latestPath: "/latest-updates/",
  searchPath: "/search/{query}/",
  infoSelector: ".block-details",
};

// A widespread commercial theme: credits in the .top-options bar under the
// player, sometimes with a categories/tags tab below it.
const TOP_OPTIONS_THEME = {
  latestPath: "/latest-updates/",
  searchPath: "/search/{query}/",
  infoSelector: ".top-options, .video-taxonomy, #tabcat",
};

// gig.sex and xxxi.porn: localized builds with singular entity paths and
// /new-porn for the newest videos.
const LOCALIZED_THEME = {
  latestPath: "/new-porn/",
  searchPath: "/search/{query}/",
  tagPrefixes: ["/category/", "/tag/"],
  modelPrefixes: ["/pornstar/", "/model/"],
  channelPrefixes: ["/channel/"],
  infoSelector: ".info-content",
};

export const SITES: SiteConfig[] = [
  kvsSite({
    ...PRIVATEHOST_NETWORK,
    key: "okxxx",
    name: "OK.XXX",
    origin: "https://ok.xxx",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...PRIVATEHOST_NETWORK,
    key: "pornhat",
    name: "PornHat",
    origin: "https://www.pornhat.com",
    // PornHat's video URLs carry only the slug.
    videoPath: /^\/video\/([a-z0-9][a-z0-9-]*)\//,
  }),
  kvsSite({
    ...PRIVATEHOST_NETWORK,
    key: "perfectgirls",
    name: "PerfectGirls",
    origin: "https://www.perfectgirls.xxx",
    videoPath: /^\/video\/(\d+)\//,
    modelIndexPath: "/pornstars/",
  }),
  kvsSite({
    ...PORNGO_NETWORK,
    key: "porngo",
    name: "PornGO",
    origin: "https://www.porngo.com",
    videoPath: /^\/videos\/(\d+)\//,
  }),
  kvsSite({
    ...PORNGO_NETWORK,
    key: "xxxfiles",
    name: "XXXFiles",
    origin: "https://www.xxxfiles.com",
    videoPath: /^\/videos\/(\d+)\//,
  }),
  kvsSite({
    key: "anysex",
    name: "AnySex",
    origin: "https://anysex.com",
    videoPath: /^\/video\/(\d+)\//,
    latestPath: "/videos/new/",
    searchPath: "/search/?q={q}",
    tagIndexPath: "/videos/categories/",
    tagPrefixes: ["/videos/categories/"],
    channelPrefixes: ["/channels/"],
    infoSelector: ".video-info",
  }),
  kvsSite({
    key: "xcafe",
    name: "xCafe",
    origin: "https://xcafe.com",
    videoPath: /^\/(\d+)\/$/,
    latestPath: "/latest-updates/",
    searchPath: "/videos/{query}/",
    // Categories are keyword listings under /videos/.
    tagPrefixes: ["/videos/"],
    channelPrefixes: ["/channels/"],
    infoSelector: ".video_meta",
    relatedSelector: ".related-section",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "anyporn",
    name: "AnyPorn",
    origin: "https://anyporn.com",
    videoPath: /^\/(\d+)\/$/,
    latestPath: "/newest/",
    // The models index renders empty.
    modelIndexPath: undefined,
  }),
  kvsSite({
    key: "xcum",
    name: "xCum",
    origin: "https://xcum.com",
    videoPath: /^\/v\/(\d+)\//,
    latestPath: "/",
    searchPath: "/q/{query}/",
    tagIndexPath: undefined,
    modelIndexPath: undefined,
    tagPrefixes: ["/t/"],
    modelPrefixes: [],
    // Sponsor links (/s/) only carry a "join" call to action, not a name.
    channelPrefixes: [],
    infoSelector: ".video-info",
    relatedSelector: ".thumbs",
    // Card images are alt-labelled with the performers, not the title.
    cardTitleSelector: ".inf > span",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "analdin",
    name: "Analdin",
    origin: "https://www.analdin.com",
    videoPath: /^\/videos\/(\d+)\//,
    durationSelector: ".hp-duration",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "xozilla",
    name: "Xozilla",
    origin: "https://www.xozilla.com",
    videoPath: /^\/videos\/(\d+)\//,
  }),
  kvsSite({ ...FQ_NETWORK, key: "w1mp", name: "W1mp", origin: "https://w1mp.com" }),
  kvsSite({ ...FQ_NETWORK, key: "fuqster", name: "Fuqster", origin: "https://fuqster.com" }),
  kvsSite({ ...FQ_NETWORK, key: "fapnfuck", name: "FapNFuck", origin: "https://fapnfuck.com" }),
  kvsSite({
    ...TOP_OPTIONS_THEME,
    key: "yesporn",
    name: "YesPorn",
    origin: "https://yesporn.vip",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...TOP_OPTIONS_THEME,
    key: "justporn",
    name: "JustPorn",
    origin: "https://www.justporn.com",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...TOP_OPTIONS_THEME,
    key: "pornve",
    name: "PornVE",
    origin: "https://pornve.com",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...TOP_OPTIONS_THEME,
    key: "saintporn",
    name: "SaintPorn",
    origin: "https://saintporn.com",
    videoPath: /^\/video\/([a-z0-9][a-z0-9-]*)\//,
  }),
  kvsSite({
    ...TOP_OPTIONS_THEME,
    key: "fapgoat",
    name: "FapGoat",
    origin: "https://fapgoat.com",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "severeporn",
    name: "SeverePorn",
    origin: "https://severeporn.com",
    videoPath: /^\/video\/(\d+)\//,
    infoSelector: ".tags-video",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "watchporn",
    name: "WatchPorn",
    origin: "https://watchporn.to",
    videoPath: /^\/video\/(\d+)\//,
    infoSelector: ".single__info",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "neporn",
    name: "NEPorn",
    origin: "https://neporn.com",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "mygoodporn",
    name: "MyGoodPorn",
    origin: "https://mygoodporn.tv",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "megatube",
    name: "MegaTube",
    origin: "https://www.megatube.xxx",
    videoPath: /^\/videos\/(\d+)\//,
    infoSelector: ".video-info",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "fapnado",
    name: "Fapnado",
    origin: "https://www.fapnado.com",
    videoPath: /^\/videos\/(\d+)\//,
    infoSelector: "#tab_video_info",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "cherrygasp",
    name: "CherryGasp",
    origin: "https://cherrygasp.com",
    videoPath: /^\/video\/([a-z0-9][a-z0-9-]*)\//,
    infoSelector: ".clean-info-container",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "pornwex",
    name: "PornWex",
    origin: "https://www.pornwex.tv",
    videoPath: /^\/video\/(\d+)\//,
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "xmegadrive",
    name: "xMegaDrive",
    origin: "https://www.xmegadrive.com",
    videoPath: /^\/videos\/([a-z0-9][a-z0-9-]*)\//,
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "3movs",
    name: "3Movs",
    origin: "https://www.3movs.com",
    videoPath: /^\/videos\/(\d+)\//,
    searchPath: "/search_videos/?q={q}",
    modelIndexPath: "/pornstars/",
    infoSelector: ".list_info_user",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "porntrex",
    name: "PornTrex",
    origin: "https://www.porntrex.com",
    // Video URLs have no trailing slash.
    videoPath: /^\/video\/(\d+)(?:\/|$)/,
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "whoreshub",
    name: "WhoresHub",
    origin: "https://www.whoreshub.com",
    videoPath: /^\/videos\/(\d+)\//,
    infoSelector: ".info-wrap",
  }),
  kvsSite({
    ...STOCK_THEME,
    key: "trahkino",
    name: "TrahKino",
    origin: "https://trahkino.cc",
    videoPath: /^\/video\/(\d+)\//,
    infoSelector: ".info-content",
  }),
  kvsSite({
    key: "inxxx",
    name: "InXXX",
    origin: "https://www.inxxx.com",
    // Video pages are /v/<slug>.xxx-video (older ones /v/<hash>.xxx-video).
    videoPath: /^\/v\/([a-z0-9][a-z0-9-]*)\.xxx-video$/,
    latestPath: "/",
    searchPath: "/s/{query}/",
    tagIndexPath: "/tags/",
    modelIndexPath: "/top-pornstars/",
    tagPrefixes: ["/tags/"],
    modelPrefixes: ["/pornstars/"],
    channelPrefixes: [],
    infoSelector: ".primary-tags-content",
  }),
  kvsSite({
    ...LOCALIZED_THEME,
    key: "gigsex",
    name: "Gig.sex",
    origin: "https://www.gig.sex",
    videoPath: /^\/video\/([a-z0-9][a-z0-9-]*)\//,
    tagIndexPath: "/porn-categories/",
    modelIndexPath: "/pornstars/",
  }),
  kvsSite({
    ...LOCALIZED_THEME,
    key: "xxxi",
    name: "XXXi.porn",
    origin: "https://xxxi.porn",
    // Paths have no trailing slash on this site.
    videoPath: /^\/video\/(\d+)(?:\/|$)/,
    latestPath: "/new-porn",
    searchPath: "/search/{query}",
    tagIndexPath: "/tags",
    modelIndexPath: "/albums/models",
  }),
];

/** The hostname without a leading "www.". */
export function bareHost(hostname: string): string {
  return hostname.toLowerCase().replace(/^www\./, "");
}

/** The site's own hostname without "www.", e.g. "pornhat.com". */
export function siteHost(site: SiteConfig): string {
  return bareHost(new URL(site.origin).hostname);
}

/** Hostnames routed to this plugin (meta.domains); subdomains match too. */
export function supportedDomains(): string[] {
  const domains: string[] = [];
  for (const site of SITES) {
    domains.push(siteHost(site));
    for (const mirror of site.mirrors ?? []) domains.push(bareHost(mirror));
  }
  return domains;
}

/** The site serving `url` (its host, a mirror, or a subdomain of either), or undefined. */
export function siteForUrl(url: string): SiteConfig | undefined {
  let hostname: string;
  try {
    hostname = bareHost(new URL(url).hostname);
  } catch {
    return undefined;
  }
  return SITES.find((site) => {
    const hosts = [siteHost(site), ...(site.mirrors ?? []).map(bareHost)];
    return hosts.some((host) => hostname === host || hostname.endsWith("." + host));
  });
}

/** The site with this key, or undefined. */
export function siteByKey(key: string): SiteConfig | undefined {
  return SITES.find((site) => site.key === key);
}

/**
 * The sites search and browse span: KVS_SITES (comma-separated keys or
 * hosts) when set, else every site. Sites needing FlareSolverr drop out when
 * no solver is configured.
 */
export function enabledSites(): SiteConfig[] {
  let sites = SITES;
  const selection = (process.env.KVS_SITES ?? "").trim();
  if (selection) {
    const wanted = selection.split(",").map((entry) => bareHost(entry.trim()));
    sites = SITES.filter((site) => wanted.includes(site.key) || wanted.includes(siteHost(site)));
  }
  const solverAvailable = Boolean((process.env.FLARESOLVERR_URL ?? "").trim());
  return sites.filter((site) => solverAvailable || !site.requiresSolver);
}
