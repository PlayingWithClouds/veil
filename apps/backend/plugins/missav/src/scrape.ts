import { parse, HTMLElement } from "node-html-parser";
import { resolveUrl, canonicalUrl, videoCodeFromUrl, decodeEntities, fetchPage } from "./http.ts";
import type {
  ScrapeResult,
  Scene,
  ScenePerformer,
  StudioRef,
  Image,
} from "@playingwithclouds/veil-sdk";

// A link inside an info panel row: its display text and canonical page URL.
interface InfoLink {
  name: string;
  url: string;
}

/**
 * Fetches a missav video page and scrapes it. Every missav item is a single
 * adult JAV title, modelled as one Scene.
 */
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  return parseScenePage(await fetchPage(target), target);
}

/**
 * Parses a video page into the Scene, followed by the Label as a standalone
 * studio under its Maker when the two differ.
 *
 * Scene.related stays unset: the page's "watch next" grid ships as empty
 * Alpine placeholders that the Recombee client fills in the browser, so the
 * HTML carries no related videos to parse.
 */
export function parseScenePage(html: string, target: string): ScrapeResult[] {
  const root = parse(html);
  const rows = infoRows(root);
  const poster = extractPoster(root);
  const maker = infoRowLinks(rows, "Maker:")[0];

  const scene: Scene = {
    type: "scene",
    external_id: `missav-${videoCodeFromUrl(target)}`,
    source_url: target,
    title: extractTitle(root),
    poster_path: poster,
    images: [],
    tags: extractTags(rows),
    performers: extractCast(rows),
    downloads: [{ label: "missav", url: target }],
  };
  if (poster) {
    scene.images = [{ type: "poster", file_path: poster } as Image];
  }
  const releaseDate = infoRowText(rows, "Release date:");
  if (releaseDate) {
    scene.date = releaseDate;
  }
  if (maker) {
    scene.studio = studioRef(maker);
  }

  const results: ScrapeResult[] = [{ type: "scene", scene }];
  const label = labelStudio(rows, maker);
  if (label) {
    results.push(label);
  }
  return results;
}

/** The og:title, falling back to the page heading. */
function extractTitle(root: HTMLElement): string {
  const og = root.querySelector('meta[property="og:title"]')?.getAttribute("content");
  if (og) return decodeEntities(og).trim();
  return root.querySelector("h1")?.text.trim() ?? "";
}

/** The og:image cover as an absolute URL. */
function extractPoster(root: HTMLElement): string | undefined {
  const meta = root.querySelector('meta[property="og:image"]')?.getAttribute("content");
  if (!meta) return undefined;
  return resolveUrl(meta);
}

/**
 * The info panel's genres plus the DVD-style code (e.g. EBON-006), which
 * doubles as a searchable tag.
 */
function extractTags(rows: Map<string, HTMLElement>): string[] {
  const tags = infoRowLinks(rows, "Genre:").map((link) => link.name);
  const dvdCode = infoRowText(rows, "Code:");
  if (dvdCode) {
    tags.push(dvdCode);
  }
  return tags;
}

/** The credited actresses, each identified by their /actresses/ page. */
function extractCast(rows: Map<string, HTMLElement>): ScenePerformer[] {
  return infoRowLinks(rows, "Actress:").map((link, index) => ({
    name: link.name,
    order: index,
    source_url: link.url,
  }));
}

/** A studio reference identified by its /makers/ or /labels/ page. */
function studioRef(link: InfoLink): StudioRef {
  return { name: link.name, source_url: link.url };
}

/**
 * The Label as a standalone studio parented to the Maker. Skipped when either
 * is missing or they share a name (the common case), since parents resolve by
 * name and would point the label at itself.
 */
function labelStudio(rows: Map<string, HTMLElement>, maker: InfoLink | undefined): ScrapeResult | null {
  const label = infoRowLinks(rows, "Label:")[0];
  if (!label || !maker) return null;
  if (label.name === maker.name) return null;
  return {
    type: "studio",
    studio: {
      type: "studio",
      external_id: `missav-label-${label.name}`,
      source_url: label.url,
      name: label.name,
      parent: maker.name,
    },
  };
}

/**
 * The detail info panel's `div.text-secondary` rows, keyed by their opening
 * `<span>Label:</span>`. Scoping lookups to these rows keeps site-wide nav
 * links (e.g. the "VR" genre in the header) out of the scene's data.
 */
function infoRows(root: HTMLElement): Map<string, HTMLElement> {
  const rows = new Map<string, HTMLElement>();
  for (const row of root.querySelectorAll("div.text-secondary")) {
    const key = row.querySelector("span")?.text.trim();
    if (key && !rows.has(key)) {
      rows.set(key, row);
    }
  }
  return rows;
}

/** The value text of an info row, e.g. "2026-06-05" for "Release date:". */
function infoRowText(rows: Map<string, HTMLElement>, label: string): string {
  const row = rows.get(label);
  if (!row) return "";
  return decodeEntities(row.text.replace(label, "").trim());
}

/** The de-duplicated links of an info row, e.g. every genre of "Genre:". */
function infoRowLinks(rows: Map<string, HTMLElement>, label: string): InfoLink[] {
  const row = rows.get(label);
  if (!row) return [];

  const links: InfoLink[] = [];
  const seen = new Set<string>();
  for (const anchor of row.querySelectorAll("a[href]")) {
    const name = decodeEntities(anchor.text.trim());
    if (!name || seen.has(name)) continue;
    seen.add(name);
    links.push({ name, url: canonicalUrl(anchor.getAttribute("href") as string) });
  }
  return links;
}
