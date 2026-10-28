import { parse } from "node-html-parser";
import type { FacePostMatch } from "@playingwithclouds/veil-sdk";

export const SITE_ORIGIN = "https://namethatporn.com";

/** Resolves a site path ("/post/1-who-is-she.html") to an absolute URL; anything off-site is dropped. */
function absoluteUrl(path: string | undefined): string | undefined {
  if (!path) return undefined;
  if (path.startsWith(`${SITE_ORIGIN}/`)) return path;
  if (!path.startsWith("/")) return undefined;
  return SITE_ORIGIN + path;
}

/**
 * The thumbnail of a result. The page lazy-loads it: data-dyn holds the
 * picture's address without extension and the site's script appends one.
 */
function thumbnailUrl(dynamicSource: string | undefined): string | undefined {
  if (!dynamicSource || !dynamicSource.startsWith(`${SITE_ORIGIN}/`)) return undefined;
  return `${dynamicSource}.jpg`;
}

/**
 * Parses the result page of an image search: the matching posts, best first.
 * Names stay empty here; they come from each post's page (parseAnswerNames).
 */
export function parsePosts(html: string): FacePostMatch[] {
  const posts: FacePostMatch[] = [];
  for (const link of parse(html).querySelectorAll("#nsw_i_r a.nsw_i_r_w")) {
    const url = absoluteUrl(link.getAttribute("href"));
    if (!url) continue;
    const image = link.querySelector("img");
    const post: FacePostMatch = {
      title: image?.getAttribute("alt")?.trim() || "Untitled post",
      url,
      solved: link.querySelector(".slvd") !== null,
      names: [],
    };
    const thumbnail = thumbnailUrl(image?.getAttribute("data-dyn"));
    if (thumbnail) post.thumbnail_url = thumbnail;
    posts.push(post);
  }
  return posts;
}

/** Reads the performer name tags of a post's accepted answer, in order and without repeats. */
export function parseAnswerNames(html: string): string[] {
  const names: string[] = [];
  for (const tag of parse(html).querySelectorAll("#correct_answer .name_syntax .name")) {
    const name = tag.text.trim();
    if (name === "" || names.includes(name)) continue;
    names.push(name);
  }
  return names;
}
