import { NeedsLoginError, sessionFetch } from "@playingwithclouds/veil-sdk";
import type { ScenePerformer } from "@playingwithclouds/veil-sdk";
import { parse } from "node-html-parser";

export const SITE_URL = "https://chyoa.com";
export const LOGIN_URL = `${SITE_URL}/auth/login`;

const BROWSER_USER_AGENT = "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0";

/** Headers of a browser reading pages in English. */
export function pageHeaders(): Record<string, string> {
  return {
    "User-Agent": BROWSER_USER_AGENT,
    Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "Accept-Language": "en",
  };
}

/** Collapses whitespace runs (non-breaking spaces included) and trims. */
export function normalizeText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}

/** Makes a site-relative path absolute; other addresses stay as they are. */
export function absoluteUrl(href: string): string {
  if (href.startsWith("/")) return `${SITE_URL}${href}`;
  return href;
}

/**
 * The author a profile link names, or undefined without a name. Story pages and
 * listing cards credit through this, so both resolve to the same performer.
 */
export function authorOfLink(link: { text: string; getAttribute: (name: string) => string | undefined }): ScenePerformer | undefined {
  const name = normalizeText(link.text);
  if (name === "") return undefined;
  return {
    name,
    external_id: `chyoa-user-${name.toLowerCase()}`,
    source_url: absoluteUrl(link.getAttribute("href") ?? "").replace(/(?:%20)+$/, ""),
  };
}

/** The id out of a chyoa.com/story/<slug>.<id> address, or undefined for other addresses. */
export function storyIdOf(url: string): string | undefined {
  const match = url.match(/^https?:\/\/(?:www\.)?chyoa\.com\/story\/[^?#]*\.(\d+)(?:[/?#]|$)/i);
  if (!match) return undefined;
  return match[1];
}

/** The id out of a chyoa.com/chapter/<slug>.<id> address, or undefined for other addresses. */
export function chapterIdOf(url: string): string | undefined {
  const match = url.match(/^https?:\/\/(?:www\.)?chyoa\.com\/chapter\/[^?#]*\.(\d+)(?:[/?#]|$)/i);
  if (!match) return undefined;
  return match[1];
}

/** The page of one story; the site ignores the slug and redirects to the real one. */
export function storyUrl(storyId: string): string {
  return `${SITE_URL}/story/x.${storyId}`;
}

/** The identity a story carries. */
export function externalIdOf(storyId: string): string {
  return `chyoa-${storyId}`;
}

/** Whether a page is the login form the site shows instead of a page that needs an account. */
export function isLoginPage(html: string): boolean {
  return parse(html).querySelector('form[action$="/auth/login"]') !== null;
}

/** Whether a response was sent from the login page, after the redirects were followed. */
function endedAtLogin(response: Response): boolean {
  if (response.url === "") return false;
  return new URL(response.url).pathname === "/auth/login";
}

/** Fetches a page that is the same for everyone (listings), with the account's cookies if there are any. */
export async function fetchPublicPage(url: string): Promise<string> {
  const response = await sessionFetch(url, { headers: pageHeaders() });
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  return response.text();
}

/**
 * Fetches a story or chapter page. Logged out, the site redirects these to
 * the login page, which is a login needed rather than an error.
 */
export async function fetchReaderPage(url: string): Promise<string> {
  const response = await sessionFetch(url, { headers: pageHeaders() });
  if (endedAtLogin(response)) throw new NeedsLoginError("Log in to Chyoa to read stories");
  if (response.status === 404) throw new Error(`No chapter at ${url}`);
  if (!response.ok) throw new Error(`HTTP ${response.status} fetching ${url}`);
  const html = await response.text();
  if (isLoginPage(html)) throw new NeedsLoginError("Log in to Chyoa to read stories");
  return html;
}
