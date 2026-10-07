import { parse, HTMLElement } from "node-html-parser";
import { fetchHtml } from "@playingwithclouds/veil-sdk";
import {
  BASE_URL,
  resolveUrl,
  videoIdFromUrl,
  isVideoUrl,
  actressSlugFromUrl,
  isActressUrl,
  categorySlugFromUrl,
  isCategoryUrl,
  parseDurationSeconds,
  relativeDateToISO,
  providerFromUrl,
} from "./http.ts";
import { parseVideoCards } from "./listing.ts";
import type {
  ScrapeResult,
  Scene,
  Performer,
  Tag,
  ScenePerformer,
  Download,
  Image,
  DiscoveredItem,
} from "@playingwithclouds/veil-sdk";

const MAX_RELATED = 40;

// A video page yields a Scene followed by one standalone Performer per credited
// actress; actress and category URLs yield their standalone entity directly.
export async function scrape(url: string): Promise<ScrapeResult[]> {
  const target = resolveUrl(url);
  if (isVideoUrl(target)) return await scrapeVideo(target);
  if (isActressUrl(target)) {
    const performer = await scrapeActress(actressSlugFromUrl(target));
    return [{ type: "performer", performer }];
  }
  if (isCategoryUrl(target)) return [{ type: "tag", tag: categoryTag(target) }];
  throw new Error(`hqporner: unsupported URL ${target}`);
}

async function scrapeVideo(target: string): Promise<ScrapeResult[]> {
  const id = videoIdFromUrl(target);
  const html = await fetchHtml(target);
  const root = parse(html);

  const poster = extractPoster(root);
  const cast = extractCast(root);

  const scene: Scene = {
    type: "scene",
    external_id: `hqporner-${id}`,
    source_url: target,
    title: extractTitle(root),
    poster_path: poster,
    images: poster ? [{ type: "poster", file_path: poster } as Image] : [],
    tags: extractCategories(root),
    performers: cast,
    downloads: extractDownloads(root, html),
    related: extractRelated(root, html, `hqporner-${id}`),
  };
  const duration = extractDurationSeconds(root);
  if (duration > 0) scene.duration_seconds = duration;
  const date = extractDate(root);
  if (date) scene.date = date;

  const results: ScrapeResult[] = [{ type: "scene", scene }];
  for (const performer of await castPerformers(cast)) {
    results.push({ type: "performer", performer });
  }
  return results;
}

// The "Similar HD porn" block reuses the listing card grid; its preload_<id>
// frame arrays sit inside the same section, so the whole page html is passed on.
export function extractRelated(root: HTMLElement, html: string, selfExternalId: string): DiscoveredItem[] {
  const section = findSimilarSection(root);
  if (!section) return [];
  return parseVideoCards(section, html)
    .filter((item) => item.external_id !== selfExternalId)
    .slice(0, MAX_RELATED);
}

function findSimilarSection(root: HTMLElement): HTMLElement | undefined {
  for (const section of root.querySelectorAll("section.features")) {
    const heading = section.querySelector("h2")?.text.toLowerCase() ?? "";
    if (heading.includes("similar")) return section;
  }
  return undefined;
}

// Fetch each credited actress's page so the initial scrape already carries full
// performer records (bio + photo). Falls back to a name-only record per actress.
async function castPerformers(cast: ScenePerformer[]): Promise<Performer[]> {
  const performers: Performer[] = [];
  for (const credit of cast) {
    const slug = actressSlugFromUrl(credit.source_url ?? "");
    if (!slug) continue;
    try {
      performers.push(await scrapeActress(slug));
    } catch {
      performers.push(minimalPerformer(slug, credit.name));
    }
  }
  return performers;
}

async function scrapeActress(slug: string): Promise<Performer> {
  const sourceURL = `${BASE_URL}/actress/${slug}`;
  const root = parse(await fetchHtml(sourceURL));

  const performer = minimalPerformer(slug, extractActressName(root, slug));
  const bio = extractActressBio(root);
  if (bio) performer.details = bio;
  const photo = extractActressPhoto(root, slug);
  if (photo) {
    performer.image_path = photo;
    performer.images = [{ type: "profile", file_path: photo } as Image];
  }
  return performer;
}

function minimalPerformer(slug: string, name: string): Performer {
  return {
    type: "performer",
    external_id: `hqporner-actress-${slug}`,
    source_url: `${BASE_URL}/actress/${slug}`,
    name,
  };
}

// The actress heading renders as `<h1 class="main-h1"><span class="pink">Name</span> porn HD videos</h1>`.
function extractActressName(root: HTMLElement, slug: string): string {
  const highlighted = root.querySelector("h1.main-h1 span")?.text.trim();
  if (highlighted) return highlighted;
  return slug.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

// The "Info" section holds the profile image followed by a bio paragraph.
function extractActressBio(root: HTMLElement): string {
  for (const section of root.querySelectorAll("section")) {
    const heading = section.querySelector("h2")?.text.trim().toLowerCase();
    if (heading !== "info") continue;
    const bio = section.querySelector("p")?.text.trim();
    if (bio) return bio;
  }
  return "";
}

function extractActressPhoto(root: HTMLElement, slug: string): string {
  const img = root.querySelector(`img[src*="porn-categories/${slug}"]`)?.getAttribute("src");
  if (img) return resolveUrl(img);
  return "";
}

function categoryTag(target: string): Tag {
  const slug = categorySlugFromUrl(target);
  return {
    type: "tag",
    external_id: `hqporner-category-${slug}`,
    name: slug.replace(/-/g, " "),
  };
}

function extractTitle(root: HTMLElement): string {
  const heading = root.querySelector("h1.main-h1")?.text.trim();
  if (heading) return heading;
  const pageTitle = root.querySelector("title")?.text ?? "";
  return pageTitle.replace(/\s*-\s*HQporner\.com\s*$/i, "").trim();
}

// The header meta list carries the runtime in an <li class="icon fa-clock-o">.
// Related-video cards reuse fa-clock-o on <span> elements — restrict to <li>.
function extractDurationSeconds(root: HTMLElement): number {
  const text = root.querySelector("li.fa-clock-o")?.text ?? "";
  return parseDurationSeconds(text);
}

// The publish date renders relative ("today", "3 months ago") in the header
// meta list's <li class="icon fa-calendar">.
function extractDate(root: HTMLElement): string {
  const text = root.querySelector("li.fa-calendar")?.text ?? "";
  return relativeDateToISO(text);
}

// Credited actresses link to /actress/<slug>; carry the link as an identity
// hint so ingest can match/create the performer record.
function extractCast(root: HTMLElement): ScenePerformer[] {
  const performers: ScenePerformer[] = [];
  const seen = new Set<string>();
  for (const a of root.querySelectorAll('a[href*="/actress/"]')) {
    const name = a.text.trim();
    const slug = actressSlugFromUrl(a.getAttribute("href") ?? "");
    if (!name || !slug || seen.has(slug)) continue;
    seen.add(slug);
    performers.push({
      name,
      order: performers.length,
      external_id: `hqporner-actress-${slug}`,
      source_url: `${BASE_URL}/actress/${slug}`,
    });
  }
  return performers;
}

// The "This video belongs to the following categories" section lists genre
// chips linking to /category/<slug>.
function extractCategories(root: HTMLElement): string[] {
  const seen = new Set<string>();
  const categories: string[] = [];
  for (const a of root.querySelectorAll('a[href*="/category/"]')) {
    const name = a.text.trim();
    if (!name || seen.has(name)) continue;
    seen.add(name);
    categories.push(name);
  }
  return categories;
}

function extractPoster(root: HTMLElement): string | undefined {
  const meta = root.querySelector('meta[property="og:image"]')?.getAttribute("content");
  if (meta) return resolveUrl(meta);
  return undefined;
}

// The playable source is an embedded player. The site wires it up via
// altPlayer(): `altplayer.php?i=<embedUrl>`. Fall back to the first non-ad
// iframe if that marker is absent.
function extractDownloads(root: HTMLElement, html: string): Download[] {
  const embed = embedUrl(root, html);
  if (!embed) return [];
  return [
    {
      label: providerFromUrl(embed),
      url: embed,
    },
  ];
}

const AD_HOSTS = ["go.", "magsrv", "mayzaent", "mavrtracktor", "exoclick", "juicyads"];

function embedUrl(root: HTMLElement, html: string): string {
  const marker = html.match(/altplayer\.php\?i=([^'"]+)/);
  if (marker) return resolveUrl(marker[1]);

  for (const iframe of root.querySelectorAll("iframe")) {
    const src = iframe.getAttribute("src") ?? "";
    if (!src) continue;
    if (AD_HOSTS.some((host) => src.includes(host))) continue;
    return resolveUrl(src);
  }
  return "";
}
