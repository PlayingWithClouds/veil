import { parse, type HTMLElement } from "node-html-parser";
import type {
  DiscoveredItem,
  Scene,
  ScenePerformer,
  ScrapeResult,
  StudioRef,
} from "@playingwithclouds/veil-sdk";
import { parseVideoCards, titleFromPostText } from "./cards.ts";
import {
  attributeOf,
  blogIdFromUrl,
  decodeEntities,
  externalIdOf,
  fetchPage,
  normalizeText,
  postIdFromUrl,
  postUrl,
  resolveUrl,
} from "./http.ts";
import { hostedSourcePath } from "./resolve.ts";
import { isoDurationSeconds, viewsFromText } from "./values.ts";

/** Fetches a post page and returns its scene as the scrape stream's root record. */
export async function scrapeScene(url: string): Promise<ScrapeResult[]> {
  const postId = postIdFromUrl(url);
  if (postId === "") throw new Error(`sxyprn: no post in ${url}`);
  const root = parse(await fetchPage(postUrl(postId)));
  return [{ type: "scene", scene: sceneFromPostPage(root, postId) }];
}

/**
 * The scene record of a post page. A post with a hosted copy lists its post URL
 * as the download: the playable link carries a token that expires, so the
 * resolve capability builds it from the post page when the video plays.
 */
export function sceneFromPostPage(root: HTMLElement, postId: string): Scene {
  const mainPost = root.querySelector("div.post_el_post");
  if (mainPost === null) throw new Error(`sxyprn: post ${postId} has no post on its page`);
  const scene: Scene = {
    type: "scene",
    external_id: externalIdOf(postId),
    source_url: postUrl(postId),
    title: titleOf(mainPost),
    tags: tagsOf(mainPost),
    related: relatedFrom(root, postId),
  };
  const performers = performersOf(mainPost);
  if (performers.length > 0) scene.performers = performers;
  const studio = studioOf(mainPost);
  if (studio !== undefined) scene.studio = studio;
  const date = dateOf(root);
  if (date !== "") scene.date = date;
  const duration = isoDurationSeconds(metaContent(root, "duration"));
  if (duration > 0) scene.duration_seconds = duration;
  const views = viewsOf(root, postId);
  if (views !== undefined) scene.view_count = views;
  if (hostedSourcePath(root, postId) !== "") {
    scene.downloads = [{ label: "sxyprn", url: postUrl(postId), format: "mp4" }];
  }
  const poster = metaContent(root, "thumbnailUrl");
  if (poster !== "") scene.poster_path = resolveUrl(poster);
  return scene;
}

/** The content of a schema.org meta tag such as itemprop="duration"; empty when missing. */
function metaContent(root: HTMLElement, itemProperty: string): string {
  const meta = root.querySelector(`meta[itemprop="${itemProperty}"]`);
  if (meta === null) return "";
  return decodeEntities(attributeOf(meta, "content"));
}

/** The post's description as its title. */
function titleOf(mainPost: HTMLElement): string {
  const postText = mainPost.querySelector(".post_text");
  if (postText === null) return "";
  return titleFromPostText(postText.innerHTML);
}

/** The post's hashtags plus its category, without duplicates. */
function tagsOf(mainPost: HTMLElement): string[] {
  const tags = mainPost.querySelectorAll("a.hash_link").map(tagNameOf);
  const category = categoryOf(mainPost);
  if (category !== "") tags.push(category);
  return uniqueNonEmpty(tags);
}

/** A hashtag's name: its label attribute, else its text without the "#". */
function tagNameOf(hashLink: HTMLElement): string {
  const label = attributeOf(hashLink, "label");
  if (label !== "") return label;
  return normalizeText(hashLink.text).replace(/^#/, "");
}

/** The category the poster filed the post under, such as "MYLF" or "Only Fans". */
function categoryOf(mainPost: HTMLElement): string {
  const category = mainPost.querySelector("b.sub_cat_s");
  if (category === null) return "";
  return normalizeText(decodeEntities(attributeOf(category, "data-subkey")).replace(/\+/g, " "));
}

/** The performers the description names; each links to a keyword page of the site. */
function performersOf(mainPost: HTMLElement): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const link of mainPost.querySelectorAll("a.ps_link")) {
    const name = normalizeText(link.text);
    if (name === "" || seen.has(name)) continue;
    seen.add(name);
    performers.push({ name, source_url: resolveUrl(attributeOf(link, "href")) });
  }
  return performers;
}

/**
 * The uploader of the post as its studio: the blog that posted it, identified by
 * its blog id. Undefined when the post names no uploader.
 */
function studioOf(mainPost: HTMLElement): StudioRef | undefined {
  const link = mainPost.querySelector(".pes_author_div a[href^='/blog/']");
  if (link === null) return undefined;
  const nameElement = link.querySelector(".a_name");
  if (nameElement === null) return undefined;
  const href = attributeOf(link, "href");
  return {
    name: normalizeText(nameElement.text),
    external_id: blogIdFromUrl(href),
    source_url: resolveUrl(href),
  };
}

/** The upload day of a post, as YYYY-MM-DD; empty when the page has none. */
function dateOf(root: HTMLElement): string {
  const match = metaContent(root, "uploadDate").match(/^\d{4}-\d{2}-\d{2}/);
  if (!match) return "";
  return match[0];
}

/** The view count shown on the post's own control line; undefined when absent. */
function viewsOf(root: HTMLElement, postId: string): number | undefined {
  const timeElement = root.querySelector(`div.post_control[data-postid='${postId}'] .post_control_time`);
  if (timeElement === null) return undefined;
  return viewsFromText(timeElement.text);
}

/** The page's other posts, without the post itself. */
function relatedFrom(root: HTMLElement, postId: string): DiscoveredItem[] {
  const own = externalIdOf(postId);
  return parseVideoCards(root).filter((item) => item.external_id !== own);
}

/** The values with duplicates and empty strings removed, in their first-seen order. */
function uniqueNonEmpty(values: string[]): string[] {
  const seen = new Set<string>();
  const result: string[] = [];
  for (const value of values) {
    if (value === "" || seen.has(value)) continue;
    seen.add(value);
    result.push(value);
  }
  return result;
}
