import { parse } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { Download, Scene, ScenePerformer, ScrapeResult, StudioRef } from "@playingwithclouds/veil-sdk";
import { attributeOf, decodeEntities, resolveUrl } from "./http.ts";
import { cardsFromRoot, durationFromAlt, externalIdFromPath } from "./cards.ts";

type Root = ReturnType<typeof parse>;

// The headline reads "River Lynn video by Penthouse"; names may be joined by commas or "and".
const HEADLINE_PATTERN = /^(.+?) video by (.+)$/;
const PERFORMER_SEPARATOR = /,\s*|\s+and\s+/;
// The meta description ends with "(Video duration: 00:33:19)" on some videos.
const DESCRIPTION_DURATION_PATTERN = /\s*\(Video duration:[^)]*\)\s*$/;

/** Scrapes one video page into a Scene with its playable mp4, studio, performers and related videos. */
export async function scrapeScene(url: string): Promise<ScrapeResult[]> {
  const pathname = new URL(resolveUrl(url)).pathname;
  if (externalIdFromPath(pathname) === undefined) return [];
  return sceneFromPage(pathname, await fetchHtml(resolveUrl(pathname)));
}

/** Builds the scrape result of a video page's HTML; the page is identified by its URL path. */
export function sceneFromPage(pathname: string, html: string): ScrapeResult[] {
  const externalId = externalIdFromPath(pathname);
  if (externalId === undefined) return [];

  const root = parse(html);
  const headline = parseHeadline(root);
  const scene: Scene = {
    type: "scene",
    external_id: externalId,
    source_url: resolveUrl(pathname),
    title: headlineTitle(headline),
    performers: headline.performerNames.map((name): ScenePerformer => ({ name })),
    related: cardsFromRoot(root),
  };
  if (headline.studio !== undefined) scene.studio = headline.studio;

  addDescriptionFields(scene, root);
  addPlayerFields(scene, root);

  return [{ type: "scene", scene }];
}

/** The page headline split into its performer names, studio link and plain text. */
interface Headline {
  text: string;
  performerNames: string[];
  studio?: StudioRef;
}

/** The scene title: the performer names from the headline, or the whole headline when it names none. */
function headlineTitle(headline: Headline): string {
  if (headline.performerNames.length === 0) return headline.text;
  return headline.performerNames.join(", ");
}

/** Reads the names and studio out of the page's h1. */
function parseHeadline(root: Root): Headline {
  const heading = root.querySelector("h1");
  let text = "";
  if (heading !== null) text = decodeEntities(heading.text.trim());
  const match = text.match(HEADLINE_PATTERN);
  const headline: Headline = { text, performerNames: [] };
  if (match !== null) {
    headline.performerNames = match[1].split(PERFORMER_SEPARATOR).filter((name) => name !== "");
  }
  if (heading !== null) {
    const studioLink = heading.querySelector("a");
    if (studioLink !== null) headline.studio = studioFromLink(studioLink);
  }
  return headline;
}

/** The studio ref of a /destination/<site>/ link in the headline. */
function studioFromLink(link: NonNullable<ReturnType<Root["querySelector"]>>): StudioRef {
  const sourceUrl = resolveUrl(attributeOf(link, "href"));
  const segments = new URL(sourceUrl).pathname.split("/").filter((segment) => segment !== "");
  return {
    name: decodeEntities(link.text.trim()),
    external_id: `porn-star-studio-${segments[1]}`,
    source_url: sourceUrl,
  };
}

/** Sets the description and runtime from the page's meta description. */
function addDescriptionFields(scene: Scene, root: Root): void {
  const meta = root.querySelector('meta[name="description"]');
  if (meta === null) return;
  const description = decodeEntities(attributeOf(meta, "content"));
  const duration = durationFromAlt(description);
  if (duration !== undefined) scene.duration_seconds = duration;
  const details = description.replace(DESCRIPTION_DURATION_PATTERN, "").trim();
  if (details !== "") scene.details = details;
}

/** Sets the poster and the playable mp4 from the page's video player. */
function addPlayerFields(scene: Scene, root: Root): void {
  const player = root.querySelector("video");
  if (player === null) return;
  const poster = attributeOf(player, "poster");
  if (poster !== "") scene.poster_path = resolveUrl(poster);
  const source = player.querySelector("source");
  if (source === null) return;
  const src = attributeOf(source, "src");
  if (src === "") return;
  const download: Download = { label: "mp4", url: resolveUrl(src), format: "mp4" };
  scene.downloads = [download];
}
