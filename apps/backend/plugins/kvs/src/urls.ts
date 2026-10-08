import { siteForUrl, type SiteConfig } from "./sites.ts";

/** `href` as an absolute URL, resolved against the site's origin. */
export function absoluteUrl(site: SiteConfig, href: string): string {
  const trimmed = href.trim();
  if (trimmed.startsWith("//")) return "https:" + trimmed;
  try {
    return new URL(trimmed, site.origin + "/").toString();
  } catch {
    return "";
  }
}

/**
 * `url` on the site's canonical origin with query and fragment dropped, so
 * the same page reached via www/language subdomains or tracking parameters
 * maps to one URL.
 */
export function canonicalUrl(site: SiteConfig, url: string): string {
  const pathname = pathnameOf(site, url);
  if (!pathname) return "";
  return site.origin + pathname;
}

/** The pathname of `href` when it points at the site (relative or absolute), else "". */
export function pathnameOf(site: SiteConfig, href: string): string {
  const absolute = absoluteUrl(site, href);
  if (!absolute) return "";
  if (siteForUrl(absolute) !== site) return "";
  return new URL(absolute).pathname;
}

/** The video key (id or slug) of a video page URL on the site, or "". */
export function videoKey(site: SiteConfig, href: string): string {
  const pathname = pathnameOf(site, href);
  if (!pathname) return "";
  const match = pathname.match(site.videoPath);
  if (!match) return "";
  return match[1];
}

/** The external id of a site's video: kvs-<site>-<key>. */
export function videoExternalId(site: SiteConfig, key: string): string {
  return `kvs-${site.key}-${key}`;
}

/** The external id of a site's tag/model/channel page. */
export function entityExternalId(site: SiteConfig, kind: string, slug: string): string {
  return `kvs-${site.key}-${kind}-${slug}`;
}

/**
 * The slug of a page under one of `prefixes` (e.g. "/models/<slug>/"), or ""
 * when `href` is no such page. Purely numeric slugs are pagination
 * ("/models/2/"), not entities.
 */
export function entitySlug(site: SiteConfig, href: string, prefixes: string[]): string {
  const pathname = pathnameOf(site, href);
  if (!pathname) return "";
  for (const prefix of prefixes) {
    if (!pathname.startsWith(prefix)) continue;
    const slug = pathname.slice(prefix.length).split("/")[0];
    if (!slug || /^\d+$/.test(slug)) return "";
    return decodeURIComponent(slug);
  }
  return "";
}

/** The page URL of an entity slug under its prefix on the site. */
export function entityUrl(site: SiteConfig, prefix: string, slug: string): string {
  return `${site.origin}${prefix}${encodeURIComponent(slug)}/`;
}

/** The first page of a search on the site. */
export function searchUrl(site: SiteConfig, query: string): string {
  const words = query.trim().split(/\s+/);
  const dashed = words.map((word) => encodeURIComponent(word)).join("-");
  const formEncoded = words.map((word) => encodeURIComponent(word)).join("+");
  const path = site.searchPath.replace("{query}", dashed).replace("{q}", formEncoded);
  return site.origin + path;
}

/**
 * Page `page` (1-based) of a path-paginated listing: KVS themes address it as
 * "<listing path><page>/", keeping any query (e.g. "/search/2/?q=x").
 */
export function pathPageUrl(firstPageUrl: string, page: number): string {
  if (page <= 1) return firstPageUrl;
  const url = new URL(firstPageUrl);
  let pathname = url.pathname;
  if (!pathname.endsWith("/")) pathname += "/";
  url.pathname = `${pathname}${page}/`;
  return url.toString();
}
