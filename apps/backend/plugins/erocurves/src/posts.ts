import type { DiscoveredItem, Gallery, Image, ScenePerformer } from "@playingwithclouds/veil-sdk";
import { BASE_URL, decodeEntities, stripTags } from "./http.ts";

// Full-size gallery images are linked from the post body; only the site's own
// upload CDN counts, so banner and ad links are skipped.
const IMAGE_LINK_PATTERN = /<a\s[^>]*href=['"](https:\/\/cdn\.erocurves\.com\/wp-content\/uploads\/[^'"]+\.(?:jpe?g|png|webp))['"]/gi;
// The excerpt names the studio: "Get More Of This Babe At Met Art".
const STUDIO_PATTERN = /Get more of this babe at\s+(.+?)\.?$/i;

// The WordPress REST term entry: a category or a tag.
export interface WordPressTerm {
  id: number;
  name: string;
  taxonomy: string;
}

// One post from /wp-json/wp/v2/posts. Only the fields the plugin reads are typed.
export interface WordPressPost {
  id: number;
  date?: string;
  slug: string;
  link: string;
  title: { rendered: string };
  content: { rendered: string };
  excerpt?: { rendered: string };
  _embedded?: { "wp:term"?: WordPressTerm[][] };
}

/** The full-size image URLs linked in a post body, in page order and without repeats. */
export function imageUrlsFromContent(contentHtml: string): string[] {
  const urls: string[] = [];
  const seen = new Set<string>();
  for (const match of contentHtml.matchAll(IMAGE_LINK_PATTERN)) {
    const url = match[1];
    if (seen.has(url)) continue;
    seen.add(url);
    urls.push(url);
  }
  return urls;
}

/** The studio named in a post excerpt, or undefined when it names none. */
export function studioFromExcerpt(excerptHtml: string): string | undefined {
  const match = stripTags(excerptHtml).match(STUDIO_PATTERN);
  if (!match) return undefined;
  return decodeEntities(match[1].trim());
}

/** The excerpt HTML of a post, empty when the response carries none. */
export function excerptHtml(post: WordPressPost): string {
  if (post.excerpt === undefined) return "";
  return post.excerpt.rendered;
}

/** The post's terms of one taxonomy, from the _embed response. */
export function termsOf(post: WordPressPost, taxonomy: string): string[] {
  const names: string[] = [];
  if (post._embedded === undefined || post._embedded["wp:term"] === undefined) return names;
  const termGroups = post._embedded["wp:term"];
  for (const group of termGroups) {
    for (const term of group) {
      if (term.taxonomy === taxonomy) names.push(decodeEntities(term.name));
    }
  }
  return names;
}

/** Turns a post with its embedded terms into a Gallery record. */
export function galleryFromPost(post: WordPressPost): Gallery {
  const images: Image[] = imageUrlsFromContent(post.content.rendered).map((url, position) => ({
    type: "gallery",
    file_path: url,
    position,
  }));

  const studioName = studioFromExcerpt(excerptHtml(post));
  // Post tags name both the models and, sometimes, the studio; the studio is not a performer.
  const performers: ScenePerformer[] = termsOf(post, "post_tag")
    .filter((name) => name !== studioName)
    .map((name) => ({ name }));

  const gallery: Gallery = {
    type: "gallery",
    external_id: externalIdOf(post.id),
    source_url: sourceUrlOf(post),
    title: decodeEntities(post.title.rendered),
    images,
    tags: termsOf(post, "category"),
    performers,
  };
  if (images.length > 0) gallery.cover_path = images[0].file_path;
  if (post.date !== undefined) gallery.date = post.date;
  if (studioName !== undefined) gallery.studio = { name: studioName };
  return gallery;
}

/** The DiscoveredItem stub for a post in a listing. */
export function stubFromPost(post: WordPressPost): DiscoveredItem {
  const imageUrls = imageUrlsFromContent(post.content.rendered);
  const stub: DiscoveredItem = {
    title: decodeEntities(post.title.rendered),
    media_type: "gallery",
    source_url: sourceUrlOf(post),
    external_id: externalIdOf(post.id),
  };
  if (imageUrls.length > 0) {
    stub.poster_path = imageUrls[0];
    stub.count = imageUrls.length;
  }
  return stub;
}

/**
 * The erocurves page of a post. Some posts carry a permalink to another site, so
 * the URL is built from the slug on this site instead.
 */
export function sourceUrlOf(post: WordPressPost): string {
  return `${BASE_URL}/${post.slug}/`;
}

/** The stable external id of a post, shared by the listing and the page scrape. */
export function externalIdOf(postId: number): string {
  return `erocurves-${postId}`;
}
