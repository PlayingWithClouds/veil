import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { Image, Performer, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { BASE_URL, decodeEntities, resolveUrl } from "./http.ts";

/** Pictures of credited performers live under this path; members and unset avatars never do. */
const PORNSTAR_PICTURE_PATH = "/pics/alpha/pornstars/";

/** Profile URL of a slug. */
export function profileUrl(slug: string): string {
  return `${BASE_URL}/profile/${slug}`;
}

/** The slug of a /profile/<slug> URL, or "" when the URL is another kind of page. */
export function profileSlugFromUrl(url: string): string {
  const match = url.match(/^https?:\/\/(?:www\.)?tnaflix\.com\/profile\/([^/?#]+)\/?(?:[?#].*)?$/i);
  if (!match) return "";
  return match[1];
}

/** Whether a URL is a profile page. */
export function isProfileUrl(url: string): boolean {
  return profileSlugFromUrl(url) !== "";
}

/** The slug tnaflix derives from a performer name: lower case, runs of other characters become one hyphen. */
export function slugFromName(name: string): string {
  return name
    .normalize("NFKD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

/** Whether the profile's picture is a performer photo, which only credited performers have. */
export function hasPornstarPicture(performer: Performer): boolean {
  if (!performer.image_path) return false;
  return performer.image_path.includes(PORNSTAR_PICTURE_PATH);
}

/** The profile's own URL from its active tab, falling back to the URL asked for. */
function canonicalUrl(root: ReturnType<typeof parse>, requestedUrl: string): string {
  const href = root.querySelector(".ph-tabs a.active")?.getAttribute("href");
  if (href && isProfileUrl(resolveUrl(href))) return resolveUrl(href);
  return requestedUrl;
}

/** The avatar URL when it is a real picture rather than the site's placeholder. */
function avatarOf(root: ReturnType<typeof parse>): string | undefined {
  const src = root.querySelector(".ph-avatar img")?.getAttribute("src");
  if (!src || src.includes("/assets/img/")) return undefined;
  return resolveUrl(src);
}

/**
 * Parses a profile page into a Performer, undefined when the page has no
 * profile header. The page shows only the name and picture; no bio or stats.
 */
export function parseProfilePage(html: string, requestedUrl: string): Performer | undefined {
  const root = parse(html);
  const name = root.querySelector(".profile-header .ph-title h1")?.text.trim();
  if (!name) return undefined;

  const url = canonicalUrl(root, requestedUrl);
  const slug = profileSlugFromUrl(url);
  if (slug === "") return undefined;

  const performer: Performer = {
    type: "performer",
    external_id: `tnaflix-profile-${slug}`,
    source_url: profileUrl(slug),
    name: decodeEntities(name),
  };
  const avatar = avatarOf(root);
  if (avatar) {
    performer.image_path = avatar;
    const image: Image = { type: "profile", file_path: avatar };
    performer.images = [image];
  }
  return performer;
}

/** Fetches a profile page; undefined when tnaflix answers 404 (no such profile). */
export async function fetchProfilePage(url: string): Promise<string | undefined> {
  try {
    return await fetchHtml(url);
  } catch (error) {
    if (String(error).includes("HTTP 404")) return undefined;
    throw error;
  }
}

/** performer:find: one profile by URL. */
export async function performerFind(url: string): Promise<ScrapeResult[]> {
  const html = await fetchHtml(url);
  const performer = parseProfilePage(html, url);
  if (!performer) throw new Error(`no performer profile at ${url}`);
  return [{ type: "performer", performer }];
}
