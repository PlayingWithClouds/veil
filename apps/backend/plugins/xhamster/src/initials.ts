import type { ScenePerformer, StudioRef } from "@playingwithclouds/veil-sdk";

const INITIALS_MARKER = "window.initials=";

/** One chip in the video page's tag strip; the flags say what kind of entity it links to. */
export interface VideoTag {
  name: string;
  url?: string;
  isPornstar?: boolean;
  isCategory?: boolean;
  isTag?: boolean;
  isChannel?: boolean;
  isBrand?: boolean;
  isUploader?: boolean;
}

/** The subset of the page's `window.initials` state the scraper reads. */
export interface VideoInitials {
  videoTagsComponent?: { tags?: VideoTag[] };
  videoModel?: { description?: string; created?: number };
}

/**
 * Parse the `window.initials=…;` JSON the video page embeds for client hydration.
 * Returns undefined when the script is missing or not valid JSON.
 */
export function parseInitials(html: string): VideoInitials | undefined {
  const markerIndex = html.indexOf(INITIALS_MARKER);
  if (markerIndex === -1) return undefined;
  const start = markerIndex + INITIALS_MARKER.length;
  const end = html.indexOf("</script>", start);
  if (end === -1) return undefined;
  const json = html.slice(start, end).trim().replace(/;$/, "");
  try {
    return JSON.parse(json) as VideoInitials;
  } catch {
    return undefined;
  }
}

/** The tag strip's chips, or an empty list when the page carries none. */
function videoTags(initials: VideoInitials | undefined): VideoTag[] {
  if (!initials || !initials.videoTagsComponent || !initials.videoTagsComponent.tags) return [];
  return initials.videoTagsComponent.tags;
}

/** Credited pornstars in page order, with their profile URL as identity. */
export function extractPerformers(initials: VideoInitials | undefined): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  for (const tag of videoTags(initials)) {
    if (!tag.isPornstar || !tag.name) continue;
    performers.push({ name: tag.name, order: performers.length, source_url: tag.url });
  }
  return performers;
}

/** Category and tag chip names, de-duplicated case-insensitively, categories first as the page lists them. */
export function extractTags(initials: VideoInitials | undefined): string[] {
  const tags: string[] = [];
  const seen = new Set<string>();
  for (const tag of videoTags(initials)) {
    if (!tag.isCategory && !tag.isTag) continue;
    const name = tag.name.trim();
    if (!name || seen.has(name.toLowerCase())) continue;
    seen.add(name.toLowerCase());
    tags.push(name);
  }
  return tags;
}

/**
 * The video's channel, or failing that its uploader, keyed by the channel/user page URL.
 * Brand chips (e.g. the FapHouse network promo) are not the video's source and are skipped.
 */
export function extractStudio(initials: VideoInitials | undefined): StudioRef | undefined {
  const tags = videoTags(initials);
  const channel = tags.find((tag) => tag.isChannel && !tag.isBrand);
  if (channel) return studioFromTag(channel);
  const uploader = tags.find((tag) => tag.isUploader);
  if (uploader) return studioFromTag(uploader);
  return undefined;
}

/** A StudioRef named after the chip, with its page URL as source_url. */
function studioFromTag(tag: VideoTag): StudioRef {
  return { name: tag.name, source_url: tag.url };
}

/** The uploader's free-text description, if non-empty. */
export function extractDetails(initials: VideoInitials | undefined): string | undefined {
  if (!initials || !initials.videoModel || !initials.videoModel.description) return undefined;
  const description = initials.videoModel.description.trim();
  if (!description) return undefined;
  return description;
}

/** The upload date as ISO YYYY-MM-DD, from the unix-seconds `created` timestamp. */
export function extractDate(initials: VideoInitials | undefined): string | undefined {
  if (!initials || !initials.videoModel) return undefined;
  const created = initials.videoModel.created;
  if (typeof created !== "number" || created <= 0) return undefined;
  return new Date(created * 1000).toISOString().slice(0, 10);
}
