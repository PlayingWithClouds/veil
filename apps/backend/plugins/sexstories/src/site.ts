import type { ScenePerformer } from "@playingwithclouds/veil-sdk";
import { fetchHtml } from "@playingwithclouds/veil-sdk";

export const SITE_URL = "https://www.sexstories.com";

/** Fetches a page with a plain request; the site needs no cookie, age gate or solver. */
export function fetchPage(url: string): Promise<string> {
  return fetchHtml(url);
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

/** The id out of a sexstories.com/story/<id>[/<slug>] address, or undefined for other addresses. */
export function storyIdOf(url: string): string | undefined {
  const match = url.match(/^https?:\/\/(?:www\.)?sexstories\.com\/story\/(\d+)/i);
  if (!match) return undefined;
  return match[1];
}

/** The slug of a story address, or an empty string when it has none. */
export function storySlugOf(url: string): string {
  const match = url.match(/\/story\/\d+\/([^/?#]+)/);
  if (!match) return "";
  return match[1];
}

/** The page of one story; the site ignores the slug, so it is optional. */
export function storyUrl(storyId: string, slug: string): string {
  if (slug === "") return `${SITE_URL}/story/${storyId}/`;
  return `${SITE_URL}/story/${storyId}/${slug}`;
}

/** The identity a story carries. */
export function externalIdOf(storyId: string): string {
  return `sexstories-${storyId}`;
}

/**
 * The author a profile link names, or undefined for anonymous stories (profile 0)
 * and links without a name. Story pages and listing cards credit through this, so
 * both resolve to the same performer.
 */
export function authorOfProfile(name: string, href: string): ScenePerformer | undefined {
  const profileId = profileIdOf(href);
  const cleanName = normalizeText(name);
  if (profileId === undefined || profileId === "0" || cleanName === "") return undefined;
  return {
    name: cleanName,
    external_id: `sexstories-author-${profileId}`,
    source_url: absoluteUrl(encodeURI(href)),
  };
}

/** The account id out of a /profile<id>/<name> path, or undefined for other paths. */
export function profileIdOf(href: string): string | undefined {
  const match = href.match(/\/profile(\d+)/);
  if (!match) return undefined;
  return match[1];
}
