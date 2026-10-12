import { parse, type HTMLElement } from "node-html-parser";
import { BROWSER_HEADERS, type ResolveResult } from "@playingwithclouds/veil-sdk";
import { attributeOf, decodeEntities, fetchPage, postIdFromUrl, postUrl, resolveUrl } from "./http.ts";
import { playableUrl } from "./stream-token.ts";

/**
 * Resolves a post URL to the mp4 of its hosted copy. The post page's player
 * script builds the link from a source path in the page, and the link expires, so
 * this runs when the video plays. Posts that only link to other hosts have none.
 */
export async function resolve(url: string): Promise<ResolveResult> {
  const postId = postIdFromUrl(url);
  if (postId === "") throw new Error(`sxyprn: no post in ${url}`);
  const html = await fetchPage(postUrl(postId));
  const sourcePath = hostedSourcePath(parse(html), postId);
  if (sourcePath === "") throw new Error(`sxyprn: post ${postId} has no hosted video, it only links to other hosts`);
  const result: ResolveResult = { url: await finalLocation(playableUrl(sourcePath)), mime_type: "video/mp4" };
  const quality = qualityFromInfo(html);
  if (quality !== "") result.quality = quality;
  return result;
}

/** The source path the page records for a post's hosted copy; empty when the post has none. */
export function hostedSourcePath(root: HTMLElement, postId: string): string {
  const holder = root.querySelector(".vidsnfo");
  if (holder === null) return "";
  const raw = decodeEntities(attributeOf(holder, "data-vnfo"));
  if (raw === "") return "";
  const sourcePaths = JSON.parse(raw) as Record<string, string | undefined>;
  const sourcePath = sourcePaths[postId];
  if (sourcePath === undefined) return "";
  return sourcePath;
}

/**
 * The address a playable URL redirects to, read from the redirect itself. Following
 * the redirect would download the whole video, so it is not followed.
 */
async function finalLocation(url: string): Promise<string> {
  const response = await fetch(url, { headers: BROWSER_HEADERS, redirect: "manual" });
  if (response.status >= 400) throw new Error(`sxyprn: HTTP ${response.status} for the video`);
  const location = response.headers.get("location");
  if (location === null) return url;
  return resolveUrl(location);
}

/** The "720" of the page's "resolution HD720" info line as "720p"; empty when absent. */
export function qualityFromInfo(html: string): string {
  const match = html.match(/resolution:<b>[^<]*<\/b>(\d{3,4})/);
  if (!match) return "";
  return `${match[1]}p`;
}
