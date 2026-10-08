export type SiteKey = "pornhub" | "redtube" | "youporn" | "tube8";

/**
 * Per-site configuration. The four sites share CDNs and a similar page
 * structure, so everything that differs (URL shapes, CSS selectors) lives here
 * and the rest of the plugin is written once against this shape.
 */
export interface Site {
  key: SiteKey;
  displayName: string;
  /** Bare registrable domain, used to route URLs to the site. */
  domain: string;
  /** Canonical host that pages are fetched from. */
  host: string;

  /** Extracts the video id from a page, embed or canonical URL, "" when none. */
  idFromUrl: (url: string) => string;
  /** Canonical scene page URL for a video id. */
  sceneUrl: (id: string) => string;
  /** Keyword search listing, first page. */
  searchUrl: (query: string) => string;
  /** The newest-videos feed, first page. */
  newestUrl: () => string;
  /** Page listing the site's categories. */
  categoriesUrl: string;
  /** Matches a category link's path, capturing nothing; used to harvest the category index. */
  categoryLinkSelector: string;
  /** Matches a channel/uploader path, capturing [kind, slug]. */
  studioPathPattern: RegExp;
  /** Rewrites a channel/performer page URL to the one that lists its videos. */
  videosPagePath: (pathname: string) => string;

  /** One selector per video card in a grid. */
  cardSelector: string;
  /** Ancestors whose cards are navigation chrome, not results. */
  cardExcludedAncestors: string[];
  titleLinkSelector: string;
  thumbnailSelector: string;
  durationSelector: string;
  uploaderLinkSelector: string;

  /** Containers on a scene page holding the related-videos grid. */
  relatedContainerSelectors: string[];
  /** Links to the scene's categories and tags. */
  tagLinkSelector: string;
  /** Link to the scene's uploader/channel. */
  uploaderPageSelector: string;
  /** Performer credit links on a scene page; a performer may match more than once. */
  performerCreditSelector: string;
}

const PORNHUB: Site = {
  key: "pornhub",
  displayName: "Pornhub",
  domain: "pornhub.com",
  host: "www.pornhub.com",
  idFromUrl: (url) => {
    const viewKey = url.match(/[?&]viewkey=(\w+)/);
    if (viewKey) return viewKey[1];
    const embed = url.match(/pornhub\.com\/embed\/(\w+)/);
    if (embed) return embed[1];
    return "";
  },
  sceneUrl: (id) => `https://www.pornhub.com/view_video.php?viewkey=${id}`,
  searchUrl: (query) => `https://www.pornhub.com/video/search?search=${encodeURIComponent(query)}`,
  newestUrl: () => "https://www.pornhub.com/video?o=cm",
  categoriesUrl: "https://www.pornhub.com/categories",
  categoryLinkSelector: "li.catPic .categoryTitleWrapper a",
  studioPathPattern: /^\/(channels|users)\/([^/?#]+)/,
  // A pornstar/channel root mixes several sections; /videos is the paginated full list.
  videosPagePath: (pathname) => {
    if (/^\/(pornstar|channels|model|users)\/[^/]+\/?$/.test(pathname)) {
      return pathname.replace(/\/$/, "") + "/videos";
    }
    return pathname;
  },

  cardSelector: "li.pcVideoListItem",
  cardExcludedAncestors: ["#hottestMenuSection", "#playListHeaderSection"],
  titleLinkSelector: "span.title a",
  thumbnailSelector: ".phimage img",
  durationSelector: "var.duration",
  uploaderLinkSelector: ".usernameWrap a",

  relatedContainerSelectors: ["#relatedVideosListing"],
  tagLinkSelector: ".categoriesWrapper a.item, .tagsWrapper a.item",
  uploaderPageSelector: ".video-detailed-info .userInfoBlock .usernameWrap a",
  performerCreditSelector: ".pornstarsWrapper a.pstar-list-btn",
};

const REDTUBE: Site = {
  key: "redtube",
  displayName: "RedTube",
  domain: "redtube.com",
  host: "www.redtube.com",
  idFromUrl: (url) => {
    const embed = url.match(/embed\.redtube\.com\/\?id=(\d+)/);
    if (embed) return embed[1];
    const page = url.match(/redtube\.com\/(\d+)/);
    if (page) return page[1];
    return "";
  },
  sceneUrl: (id) => `https://www.redtube.com/${id}`,
  searchUrl: (query) => `https://www.redtube.com/?search=${encodeURIComponent(query)}`,
  newestUrl: () => "https://www.redtube.com/newest",
  categoriesUrl: "https://www.redtube.com/categories",
  categoryLinkSelector: 'a[href^="/redtube/"]',
  studioPathPattern: /^\/(channels|amateur)\/([^/?#]+)/,
  videosPagePath: (pathname) => pathname,

  cardSelector: "li.videoblock_list",
  cardExcludedAncestors: ["#trending_videos_block"],
  titleLinkSelector: "a.video-title-text",
  thumbnailSelector: "img.js_thumbImageTag",
  durationSelector: ".tm_video_duration",
  uploaderLinkSelector: ".author-title-text",

  relatedContainerSelectors: ["#related_videos_center"],
  tagLinkSelector: "#video_tags_carousel a.video_carousel_item",
  uploaderPageSelector: ".video-infobox-uploader-name a.video-infobox-link",
  // The popup repeats each credit as image/name links; the first match per performer is the plain credit.
  performerCreditSelector: '.video-infobox-ps a[href^="/pornstar/"]',
};

/** YouPorn and Tube8 run the same front-end; only paths and ids differ. */
const YOUPORN: Site = {
  key: "youporn",
  displayName: "YouPorn",
  domain: "youporn.com",
  host: "www.youporn.com",
  idFromUrl: (url) => {
    const match = url.match(/youporn\.com\/(?:watch|embed)\/(\d+)/);
    if (match) return match[1];
    return "";
  },
  sceneUrl: (id) => `https://www.youporn.com/watch/${id}/`,
  searchUrl: (query) => `https://www.youporn.com/search/?query=${encodeURIComponent(query)}`,
  newestUrl: () => "https://www.youporn.com/browse/time/",
  categoriesUrl: "https://www.youporn.com/categories/",
  categoryLinkSelector: 'a[href^="/category/"]',
  studioPathPattern: /^\/(channel|amateur)\/([^/?#]+)/,
  videosPagePath: (pathname) => pathname,

  cardSelector: "article.video-box",
  cardExcludedAncestors: [],
  titleLinkSelector: "a.video-title-text",
  thumbnailSelector: "img.thumb-image",
  durationSelector: ".tm_video_duration",
  uploaderLinkSelector: ".author-title-text",

  relatedContainerSelectors: ["#relatedVideosWrapper", "#show-next"],
  tagLinkSelector: ".videoTags a.tm_carousel_tag",
  uploaderPageSelector: ".submitByLink a",
  performerCreditSelector: "a.metaDataPornstarLink",
};

const TUBE8: Site = {
  ...YOUPORN,
  key: "tube8",
  displayName: "Tube8",
  domain: "tube8.com",
  host: "www.tube8.com",
  idFromUrl: (url) => {
    const match = url.match(/tube8\.com\/(?:porn-video|embed)\/(\d+)/);
    if (match) return match[1];
    return "";
  },
  sceneUrl: (id) => `https://www.tube8.com/porn-video/${id}/`,
  searchUrl: (query) => `https://www.tube8.com/searches.html/?q=${encodeURIComponent(query)}`,
  newestUrl: () => "https://www.tube8.com/newest.html/",
  categoriesUrl: "https://www.tube8.com/categories.html/",
  categoryLinkSelector: 'a[href^="/cat/"]',
};

export const ALL_SITES: Site[] = [PORNHUB, REDTUBE, YOUPORN, TUBE8];

export const SITE_DOMAINS: string[] = ALL_SITES.map((site) => site.domain);

/** The site a URL belongs to, by hostname (subdomains such as de.pornhub.com included). */
export function siteForUrl(url: string): Site | undefined {
  let hostname = "";
  try {
    hostname = new URL(url).hostname;
  } catch {
    return undefined;
  }
  for (const site of ALL_SITES) {
    if (hostname === site.domain || hostname.endsWith("." + site.domain)) return site;
  }
  return undefined;
}

/** Like siteForUrl but throws, for handlers that cannot continue without a site. */
export function requireSite(url: string): Site {
  const site = siteForUrl(url);
  if (!site) throw new Error(`aylo: not an Aylo site URL: ${url}`);
  return site;
}

/**
 * The sites that search/browse listings cover, from the AYLO_SITES setting
 * (comma-separated site keys). Unset or unrecognised means all four.
 */
export function enabledSites(): Site[] {
  const setting = process.env.AYLO_SITES;
  if (!setting || !setting.trim()) return ALL_SITES;
  const wanted = setting.split(",").map((key) => key.trim().toLowerCase());
  const selected = ALL_SITES.filter((site) => wanted.includes(site.key));
  if (selected.length === 0) return ALL_SITES;
  return selected;
}

/** External id of a scene, `<site>-<id>`. */
export function sceneExternalId(site: Site, id: string): string {
  return `${site.key}-${id}`;
}
