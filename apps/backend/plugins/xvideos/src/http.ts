import { BROWSER_HEADERS } from "@playingwithclouds/veil-sdk";

/** The sites this plugin reads: xvideos and its sister site xnxx, which share player, CDN and comments. */
export type Site = "xvideos" | "xnxx";

/** Each site's English origin; localized subdomains (de.xvideos.com) are read through it. */
export const ORIGINS: Record<Site, string> = {
  xvideos: "https://www.xvideos.com",
  xnxx: "https://www.xnxx.com",
};

const REQUEST_TIMEOUT_MS = 20_000;

/** The site a URL belongs to, or undefined for any other host. */
export function siteOf(url: string): Site | undefined {
  let host: string;
  try {
    host = new URL(url).hostname.toLowerCase();
  } catch {
    return undefined;
  }
  if (host === "xvideos.com" || host.endsWith(".xvideos.com")) return "xvideos";
  if (host === "xnxx.com" || host.endsWith(".xnxx.com")) return "xnxx";
  return undefined;
}

/** The URL on the site's English www host, keeping path and query; site-relative paths are resolved against it. */
export function canonicalUrl(url: string, site: Site = "xvideos"): string {
  if (url.startsWith("/")) return ORIGINS[site] + url;
  const parsed = new URL(url);
  const owner = siteOf(url);
  if (!owner) return url;
  return ORIGINS[owner] + parsed.pathname + parsed.search;
}

/**
 * Fetches a page or JSON document of the sites as a browser would. English is
 * asked for: the sites machine-translate titles into the browser's language.
 */
export async function fetchText(url: string, extraHeaders: Record<string, string> = {}): Promise<string> {
  const response = await fetch(url, {
    headers: { ...BROWSER_HEADERS, "Accept-Language": "en-US,en;q=0.9", ...extraHeaders },
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  return response.text();
}

/** Fetches one of the sites' JSON endpoints, which answer only requests marked as XHR. */
export async function fetchJson(url: string, referer: string): Promise<any> {
  const text = await fetchText(url, {
    Accept: "application/json, text/javascript, */*; q=0.01",
    "X-Requested-With": "XMLHttpRequest",
    Referer: referer,
  });
  return JSON.parse(text);
}

/** "931.8k" / "4.3M" / "1,107" as a number; undefined when unreadable. */
export function countFromText(text: string | undefined): number | undefined {
  if (!text) return undefined;
  const match = text.trim().replace(/,/g, "").match(/^([\d.]+)\s*([kKmMbB]?)/);
  if (!match) return undefined;
  const value = parseFloat(match[1]);
  if (!Number.isFinite(value)) return undefined;
  const multipliers: Record<string, number> = { k: 1e3, m: 1e6, b: 1e9 };
  const unit = match[2].toLowerCase();
  if (unit === "") return Math.round(value);
  return Math.round(value * multipliers[unit]);
}

/** The cards' coarse runtime ("5 min", "8min", "1 h 12 min", "40 sec") in seconds; 0 when unreadable. */
export function secondsFromLabel(text: string | undefined): number {
  if (!text) return 0;
  const hours = text.match(/(\d+)\s*h/);
  const minutes = text.match(/(\d+)\s*min/);
  const seconds = text.match(/(\d+)\s*sec/);
  let total = 0;
  if (hours) total += parseInt(hours[1], 10) * 3600;
  if (minutes) total += parseInt(minutes[1], 10) * 60;
  if (seconds) total += parseInt(seconds[1], 10);
  return total;
}

/** "99%" as the 0..10 rating; undefined when unreadable. */
export function ratingFromPercent(text: string | undefined): number | undefined {
  if (!text) return undefined;
  const percent = parseFloat(text);
  if (!Number.isFinite(percent)) return undefined;
  return Math.round(Math.min(Math.max(percent, 0), 100)) / 10;
}

/** Named entities the sites use in titles: markup characters and accented letters. */
const NAMED_ENTITIES: Record<string, string> = {
  amp: "&", lt: "<", gt: ">", quot: '"', apos: "'", nbsp: " ",
  auml: "ä", ouml: "ö", uuml: "ü", Auml: "Ä", Ouml: "Ö", Uuml: "Ü", szlig: "ß",
  aacute: "á", eacute: "é", iacute: "í", oacute: "ó", uacute: "ú", Eacute: "É",
  agrave: "à", egrave: "è", ograve: "ò", acirc: "â", ecirc: "ê", ocirc: "ô",
  ccedil: "ç", ntilde: "ñ", Ntilde: "Ñ", hellip: "…", rsquo: "’", lsquo: "‘", rdquo: "”", ldquo: "“",
};

/** Decodes the HTML entities the sites leave in attributes and JSON titles; unknown ones stay as written. */
export function decodeEntities(text: string): string {
  return text
    .replace(/&#(\d+);/g, (_, code) => String.fromCodePoint(parseInt(code, 10)))
    .replace(/&#x([0-9a-f]+);/gi, (_, code) => String.fromCodePoint(parseInt(code, 16)))
    .replace(/&([a-zA-Z]+);/g, (entity, name: string) => NAMED_ENTITIES[name] ?? entity);
}
