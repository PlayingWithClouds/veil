import { resolveUrl } from "@playingwithclouds/veil-sdk";

/** The site's canonical origin; www and other hosts redirect here. */
export const BASE_URL = "https://mat6tube.com";

/** Videos per listing page. */
export const PAGE_SIZE = 24;

/** The canonical page of a video, from its id ("-113656837_456242801"). */
export function watchUrl(id: string): string {
  return `${BASE_URL}/watch/${id}`;
}

/** The external id of a video: mat6tube-<owner>_<video>. */
export function watchExternalId(id: string): string {
  return `mat6tube-${id}`;
}

/**
 * The video id of a /watch/ link (absolute or relative, query and fragment
 * ignored), or undefined when the link is no video page.
 */
export function watchIdFromHref(href: string): string | undefined {
  let link: URL;
  try {
    link = new URL(resolveUrl(BASE_URL, href));
  } catch {
    return undefined;
  }
  if (link.hostname !== "mat6tube.com" && !link.hostname.endsWith(".mat6tube.com")) return undefined;
  const match = link.pathname.match(/^\/watch\/(-?\d+_\d+)\/?$/);
  if (!match) return undefined;
  return match[1];
}

/** Absolute URL of a site path or CDN link. */
export function absoluteUrl(path: string): string {
  return resolveUrl(BASE_URL, path);
}
