import { parse, type HTMLElement } from "node-html-parser";
import type { Performer, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { attributeOf, fetchPage, BASE_URL, externalIdFor, profileUrl, slugFromUrl } from "./http.ts";
import {
  countryFromPlace,
  heightCm,
  isoDateFromEnglish,
  knownValue,
  measurementsFrom,
  normalizeText,
  weightKg,
} from "./values.ts";

/** Fetches a performer's bio page and returns the performer as the scrape stream's root record. */
export async function scrapePerformer(url: string): Promise<ScrapeResult[]> {
  const slug = slugFromUrl(url);
  if (slug === "") throw new Error(`freeones: no performer in ${url}`);
  const html = await fetchPage(`${BASE_URL}/${slug}/bio`);
  return [{ type: "performer", performer: performerFromBio(html, slug) }];
}

/**
 * The "Label: value" rows of the personal information list, keyed by label
 * without its colon. Each value is the element holding the value text.
 */
export function profileRows(root: HTMLElement): Map<string, HTMLElement> {
  const rows = new Map<string, HTMLElement>();
  for (const item of root.querySelectorAll("li.hide-on-edit")) {
    const spans = item.querySelectorAll("span");
    if (spans.length < 2) continue;
    const label = normalizeText(spans[0].text).replace(/:$/, "");
    rows.set(label, spans[1]);
  }
  return rows;
}

/** The normalized text of the labelled row, or an empty string when the page has no such row. */
function rowText(rows: Map<string, HTMLElement>, label: string): string {
  const row = rows.get(label);
  if (row === undefined) return "";
  return normalizeText(row.text);
}

/** Builds the performer record from a bio page's HTML. */
export function performerFromBio(html: string, slug: string): Performer {
  const root = parse(html);
  const rows = profileRows(root);
  const name = rowText(rows, "Name");
  if (name === "") throw new Error(`freeones: no name on the bio page of ${slug}`);

  return {
    type: "performer",
    external_id: externalIdFor(slug),
    source_url: profileUrl(slug),
    name,
    aliases: aliasesFrom(rows),
    details: detailsFrom(root),
    birthdate: isoDateFromEnglish(rowText(rows, "Date of birth")),
    death_date: knownValue(rowText(rows, "Date of death")),
    country: countryOf(rows),
    ethnicity: knownValue(rowText(rows, "Ethnicity")),
    eye_color: knownValue(rowText(rows, "Eye Color")),
    hair_color: knownValue(rowText(rows, "Hair Color")),
    height_cm: heightCm(rowText(rows, "Height")),
    weight_kg: weightKg(rowText(rows, "Weight")),
    measurements: measurementsOf(rows),
    fake_tits: knownValue(rowText(rows, "Boobs")),
    tattoos: featureOf(rows, "Tattoos", "Tattoo locations"),
    piercings: featureOf(rows, "Piercings", "Piercing locations"),
    career_length: careerOf(rows),
    url: officialWebsiteOf(root),
    twitter: followLinkOf(root, "twitter.com"),
    instagram: followLinkOf(root, "instagram.com"),
    image_path: portraitOf(root),
  };
}

/** The alternative names of the performer, from the aliases row. */
function aliasesFrom(rows: Map<string, HTMLElement>): string[] | undefined {
  const row = rows.get("Aliases");
  if (row === undefined) return undefined;
  const aliases = row.querySelectorAll('[data-test="link_span_aliases"]').map((alias) => normalizeText(alias.text));
  if (aliases.length === 0) return undefined;
  return aliases;
}

/** The biography: the paragraphs of the "Additional Information" block. */
function detailsFrom(root: HTMLElement): string | undefined {
  const block = findProfileBlock(root, "Additional Information");
  if (block === undefined) return undefined;
  const paragraphs = block.querySelectorAll(".common-text p").map((paragraph) => normalizeText(paragraph.text));
  const text = paragraphs.filter((paragraph) => paragraph !== "").join("\n\n");
  if (text === "") return undefined;
  return text;
}

/** The profile block whose heading is the given text, if the page has one. */
function findProfileBlock(root: HTMLElement, heading: string): HTMLElement | undefined {
  for (const block of root.querySelectorAll(".profile-meta-item")) {
    const title = block.querySelector("p.heading");
    if (title && normalizeText(title.text) === heading) return block;
  }
  return undefined;
}

/** The country: the last part of the place of birth, falling back to nationality. */
function countryOf(rows: Map<string, HTMLElement>): string | undefined {
  const fromPlace = countryFromPlace(rowText(rows, "Place of birth"));
  if (fromPlace !== undefined) return fromPlace;
  return knownValue(rowText(rows, "Nationality"));
}

/** Body measurements from the bust, cup, waist and hip rows. */
function measurementsOf(rows: Map<string, HTMLElement>): string | undefined {
  return measurementsFrom(rowText(rows, "Bust"), rowText(rows, "Cup"), rowText(rows, "Waist"), rowText(rows, "Hip"));
}

/**
 * A yes/no body feature such as tattoos, with its locations when the site
 * lists them ("Yes: back of neck"). Undefined for "No" or an unlisted feature.
 */
function featureOf(rows: Map<string, HTMLElement>, yesNoLabel: string, locationLabel: string): string | undefined {
  const answer = knownValue(rowText(rows, yesNoLabel));
  if (answer === undefined || answer.toLowerCase() !== "yes") return undefined;
  const locations = knownValue(rowText(rows, locationLabel));
  if (locations === undefined) return answer;
  return `${answer}: ${locations}`;
}

/** The career span from the start and end years, "2000-2024" or "2016-" while it runs on. */
function careerOf(rows: Map<string, HTMLElement>): string | undefined {
  const start = knownValue(rowText(rows, "Career start"));
  if (start === undefined) return undefined;
  const end = knownValue(rowText(rows, "Career end"));
  if (end === undefined) return `${start}-`;
  return `${start}-${end}`;
}

/** The performer's own official website, unwrapped from the site's affiliate redirect when it has one. */
function officialWebsiteOf(root: HTMLElement): string | undefined {
  const link = root.querySelector('[data-test="link_official_website"]');
  if (link === null) return undefined;
  const href = attributeOf(link, "href");
  const target = redirectTarget(href);
  if (target === "") return undefined;
  return target;
}

/** The URL behind a tracking redirect (its HTML= parameter), or the link itself. */
export function redirectTarget(href: string): string {
  const decoded = href.replace(/&amp;/g, "&");
  const match = decoded.match(/[?&]HTML=([^&]+)/);
  if (match) return decodeURIComponent(match[1]);
  return decoded;
}

/** The profile link of the "Follow On" block that points at the given host, not the site's own account. */
function followLinkOf(root: HTMLElement, host: string): string | undefined {
  const block = findProfileBlock(root, "Follow On");
  if (block === undefined) return undefined;
  for (const link of block.querySelectorAll("a")) {
    const href = attributeOf(link, "href");
    if (!href.includes(host)) continue;
    if (href.toLowerCase().includes("/freeones")) continue;
    return href.replace(/#$/, "");
  }
  return undefined;
}

/** The portrait in the page header, when present. */
function portraitOf(root: HTMLElement): string | undefined {
  const image = root.querySelector(".dashboard-image-container img");
  if (image === null) return undefined;
  const source = attributeOf(image, "src");
  if (source === "") return undefined;
  return source;
}
