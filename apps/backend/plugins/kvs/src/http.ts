import { BROWSER_HEADERS, fetchHtmlSmart } from "@playingwithclouds/veil-sdk";
import type { SiteConfig } from "./sites.ts";

const PAGE_TIMEOUT_MS = 20_000;

/** The Referer KVS expects on its pages and /get_file/ media. */
export function siteReferer(site: SiteConfig): string {
  return site.origin + "/";
}

/**
 * Request headers for the site's pages: a browser, same-site Referer, and
 * English, since several sites translate titles to the requested language.
 */
function pageHeaders(site: SiteConfig): Record<string, string> {
  return { ...BROWSER_HEADERS, "Accept-Language": "en-US,en;q=0.9", Referer: siteReferer(site) };
}

/** A fetched page and the URL it was finally served from. */
export interface FetchedPage {
  html: string;
  url: string;
}

/** Fetches one of the site's HTML pages. */
export async function fetchPage(site: SiteConfig, url: string): Promise<string> {
  return (await fetchPageAt(site, url)).html;
}

/**
 * Fetches one of the site's HTML pages along with its final URL, which
 * differs when the site redirects to a mirror (redirects drop the query, so
 * later pages must be built on it). Sites behind Cloudflare go through
 * FlareSolverr when the plain request is refused.
 */
export async function fetchPageAt(site: SiteConfig, url: string): Promise<FetchedPage> {
  if (site.requiresSolver) {
    return { html: await fetchHtmlSmart(url, pageHeaders(site)), url };
  }
  const response = await fetch(url, {
    headers: pageHeaders(site),
    signal: AbortSignal.timeout(PAGE_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  let finalUrl = url;
  if (response.url) finalUrl = response.url;
  return { html: await response.text(), url: finalUrl };
}
