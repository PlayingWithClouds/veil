import { parse, HTMLElement } from "node-html-parser";
import type { Performer, Image } from "@playingwithclouds/veil-sdk";
import { SITE, UA } from "./http.ts";

// Builds an eporner pornstar page URL from a performer name: lowercased, spaces
// and punctuation collapsed to single hyphens (e.g. "Riley Reid" → riley-reid).
export function pornstarUrl(name: string): string {
  const slug = name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return `${SITE}/pornstar/${slug}/`;
}

// Scrapes an eporner pornstar page (by name) into a partial Performer (enrich
// payload). Returns null when the page is missing or carries no recognizable bio.
export async function fetchPornstar(name: string): Promise<Partial<Performer> | null> {
  return fetchPornstarByUrl(pornstarUrl(name));
}

// Scrapes an eporner pornstar page (by URL) into a partial Performer, including
// the display name from the page heading.
export async function fetchPornstarByUrl(url: string): Promise<Partial<Performer> | null> {
  const response = await fetch(url, {
    headers: { "User-Agent": UA },
    signal: AbortSignal.timeout(15_000),
  });
  if (!response.ok) return null;

  const root = parse(await response.text());
  const stats = extractStats(root);
  const name = heading(root);
  if (Object.keys(stats).length === 0 && !name) return null;

  const performer: Partial<Performer> = { url };
  if (name) performer.name = name;

  const country = stats["country"];
  if (country) performer.country = country;
  if (stats["ethnicity"]) performer.ethnicity = stats["ethnicity"];
  if (stats["eye"]) performer.eye_color = stats["eye"];
  if (stats["hair"]) performer.hair_color = stats["hair"];
  if (stats["measurements"]) performer.measurements = stats["measurements"];

  const heightCm = leadingNumber(stats["height"]);
  if (heightCm) performer.height_cm = heightCm;
  const weightKg = leadingNumber(stats["weight"]);
  if (weightKg) performer.weight_kg = weightKg;

  const aliases = extractAliases(root);
  if (aliases.length > 0) performer.aliases = aliases;

  const bio = extractBiography(root);
  if (bio) performer.details = bio;

  const avatar = extractAvatar(root);
  if (avatar) {
    performer.image_path = avatar;
    const image: Image = { type: "profile", file_path: avatar };
    performer.images = [image];
  }

  return performer;
}

// Stat rows are <li><span>Label:</span><div>Value</div></li>; return them keyed
// by a lowercased label.
function extractStats(root: HTMLElement): Record<string, string> {
  const stats: Record<string, string> = {};
  for (const li of root.querySelectorAll("li")) {
    const label = li.querySelector("span")?.text?.replace(/:\s*$/, "").trim().toLowerCase();
    const value = li.querySelector("div")?.text?.trim();
    if (label && value) stats[label] = value;
  }
  return stats;
}

function heading(root: HTMLElement): string {
  return root.querySelector("h1")?.text?.trim() ?? "";
}

function extractAliases(root: HTMLElement): string[] {
  const container = root.querySelector(".psbioaliases");
  if (!container) return [];
  return container
    .querySelectorAll("a, li")
    .map((node) => node.text.trim())
    .filter((text) => text.length > 0);
}

function extractBiography(root: HTMLElement): string {
  for (const block of root.querySelectorAll(".psbio")) {
    const text = block.text.replace(/\s+/g, " ").trim();
    const marker = text.indexOf("Biography:");
    if (marker >= 0) return text.slice(marker + "Biography:".length).trim();
  }
  return "";
}

// The main portrait is the first CDN image in the profile block.
function extractAvatar(root: HTMLElement): string {
  const block = root.querySelector(".psbio.ps1") ?? root;
  for (const img of block.querySelectorAll("img")) {
    const src = img.getAttribute("data-src") ?? img.getAttribute("src") ?? "";
    if (!/^https?:\/\/.*\.(?:jpe?g|png|webp)/i.test(src)) continue;
    // Skip eporner's "no photo" placeholder.
    if (/def_actress|\/newimg\//i.test(src)) continue;
    return src;
  }
  return "";
}

// "163 cm / 5'4\"" → 163; "52 kg / 114 lbs" → 52.
function leadingNumber(value: string | undefined): number | undefined {
  if (!value) return undefined;
  const match = value.match(/(\d+)/);
  return match ? parseInt(match[1], 10) : undefined;
}
