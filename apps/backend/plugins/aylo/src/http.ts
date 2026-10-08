import { parse, HTMLElement } from "node-html-parser";
import type { ScenePerformer, StudioRef } from "@playingwithclouds/veil-sdk";
import type { Site } from "./sites.ts";

export const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

// Pornhub shows an age gate to cookie-less clients; the other sites ignore these.
const SITE_COOKIE = "age_verified=1; platform=pc; accessAgeDisclaimerPH=1; accessAgeDisclaimerUK=1";

/** Request headers that make the sites serve their English desktop pages. */
export function pageHeaders(): Record<string, string> {
  return {
    "User-Agent": UA,
    Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "Accept-Language": "en-US,en;q=0.9",
    Cookie: SITE_COOKIE,
  };
}

/** Headers the CDNs expect on stream requests. */
export function streamHeaders(site: Site): Record<string, string> {
  return { Referer: `https://${site.host}/`, "User-Agent": UA };
}

/** Fetches a page as text, throwing on any non-2xx so callers see geo-blocks and 404s. */
export async function fetchText(url: string): Promise<string> {
  const response = await fetch(url, {
    headers: pageHeaders(),
    redirect: "follow",
    signal: AbortSignal.timeout(20_000),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  return response.text();
}

/**
 * Fetches a scene page and also returns its `ss` session cookie, which
 * Pornhub's quality-list endpoint requires (the signed URL is bound to it).
 * Headers.get joins repeated Set-Cookie values, so the cookie is matched by name.
 */
export async function fetchPageWithSession(url: string): Promise<{ html: string; sessionCookie?: string }> {
  const response = await fetch(url, {
    headers: pageHeaders(),
    redirect: "follow",
    signal: AbortSignal.timeout(20_000),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  const setCookie = response.headers.get("set-cookie") || "";
  const session = setCookie.match(/(?:^|,\s*)ss=([^;,\s]+)/);
  const html = await response.text();
  if (!session) return { html };
  return { html, sessionCookie: `ss=${session[1]}` };
}

/** Fetches a page and parses it. */
export async function fetchDocument(url: string): Promise<HTMLElement> {
  return parse(await fetchText(url));
}

/** Resolves an href (absolute, protocol-relative or root-relative) against the site. */
export function absoluteUrl(site: Site, href: string): string {
  if (href.startsWith("http")) return href;
  if (href.startsWith("//")) return `https:${href}`;
  if (href.startsWith("/")) return `https://${site.host}${href}`;
  return `https://${site.host}/${href}`;
}

/** Returns the URL with its `page` query parameter set; page 1 leaves the URL untouched. */
export function withPageParam(url: string, page: number): string {
  const hashIndex = url.indexOf("#");
  let base = url;
  if (hashIndex !== -1) base = url.slice(0, hashIndex);
  base = base.replace(/([?&])page=\d+&?/, "$1").replace(/[?&]$/, "");
  if (page <= 1) return base;
  if (base.includes("?")) return `${base}&page=${page}`;
  return `${base}?page=${page}`;
}

/** Points a URL at the site's canonical host and, for entity pages, its full video list. */
export function listingUrl(site: Site, url: string): string {
  const parsed = new URL(url);
  const pathname = site.videosPagePath(parsed.pathname);
  return `https://${site.host}${pathname}${parsed.search}`;
}

/** Parses "12:48", "05:45" or "1:02:03" into seconds, 0 when unparseable. */
export function parseClockDuration(text: string): number {
  const parts = text
    .trim()
    .split(":")
    .map((part) => parseInt(part, 10));
  if (parts.length < 2 || parts.some((part) => !Number.isFinite(part))) return 0;
  return parts.reduce((total, part) => total * 60 + part, 0);
}

/** Parses an ISO-8601 duration such as "PT00H12M48S" or "PT1217S" into seconds. */
export function parseIsoDuration(text: string | undefined): number {
  if (!text) return 0;
  const match = text.match(/^P(?:(\d+)D)?T?(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?$/);
  if (!match) return 0;
  const days = parseInt(match[1] || "0", 10);
  const hours = parseInt(match[2] || "0", 10);
  const minutes = parseInt(match[3] || "0", 10);
  const seconds = Math.round(parseFloat(match[4] || "0"));
  return ((days * 24 + hours) * 60 + minutes) * 60 + seconds;
}

/** Last path segment of a page link, URL-decoded ("/pornstar/katty+west" gives "katty west"). */
export function slugFromHref(href: string): string {
  const path = href.split(/[?#]/)[0].replace(/\/+$/, "");
  const segment = path.slice(path.lastIndexOf("/") + 1);
  try {
    return decodeURIComponent(segment.replace(/\+/g, " "));
  } catch {
    return segment;
  }
}

/**
 * Builds the StudioRef for a card/page uploader link when it points at a
 * channel or amateur profile. Pornstar and model pages are performers, not
 * studios, so they yield undefined.
 */
export function studioFromLink(site: Site, anchor: HTMLElement | null | undefined): StudioRef | undefined {
  if (!anchor) return undefined;
  const href = anchor.getAttribute("href");
  const name = anchor.text.trim();
  if (!href || !name) return undefined;

  let pathname = href;
  if (href.startsWith("http")) {
    pathname = new URL(href).pathname;
  }
  const match = pathname.match(site.studioPathPattern);
  if (!match) return undefined;

  return {
    name,
    external_id: `${site.key}-${match[1]}-${match[2]}`,
    source_url: `https://${site.host}${pathname}`,
  };
}

/** Builds a ScenePerformer from a /pornstar/ link. */
export function performerFromLink(site: Site, href: string, name: string): ScenePerformer | undefined {
  const cleanName = name.replace(/,\s*$/, "").trim();
  if (!cleanName || !href) return undefined;
  const pathname = href.startsWith("http") ? new URL(href).pathname : href.split(/[?#]/)[0];
  if (!pathname.includes("/pornstar/") && !pathname.includes("/model/")) return undefined;
  return {
    name: cleanName,
    external_id: `${site.key}-ps-${slugFromHref(pathname).replace(/\s+/g, "-")}`,
    source_url: `https://${site.host}${pathname}`,
  };
}
