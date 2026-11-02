import { parse, type HTMLElement } from "node-html-parser";
import type { Performer, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { fetchPage } from "./http.ts";
import { siteForUrl, type SiteConfig } from "./sites.ts";
import { cleanText } from "./text.ts";
import { canonicalUrl, entityExternalId, entitySlug } from "./urls.ts";

// A KVS model page. Themes differ in markup, so every fact is read as a
// "label: value" pair from wherever the page lists one (info lists, stat
// boxes, definition lists, JSON-LD) and mapped by label.

/** What a profile fact is called on this plugin's Performer. */
type FactKey =
  | "birthdate"
  | "country"
  | "ethnicity"
  | "hair"
  | "eyes"
  | "height"
  | "weight"
  | "measurements"
  | "aliases"
  | "tattoos"
  | "piercings"
  | "gender"
  | "website";

/** Page labels (lower case) and the fact each one names. */
const FACT_LABELS: Record<string, FactKey> = {
  born: "birthdate",
  "born on": "birthdate",
  birthdate: "birthdate",
  "birth date": "birthdate",
  "date of birth": "birthdate",
  birthday: "birthdate",
  country: "country",
  nationality: "country",
  ethnicity: "ethnicity",
  hair: "hair",
  "hair color": "hair",
  "hair colour": "hair",
  eyes: "eyes",
  "eye color": "eyes",
  "eye colour": "eyes",
  height: "height",
  weight: "weight",
  measurements: "measurements",
  alias: "aliases",
  aliases: "aliases",
  "also known as": "aliases",
  aka: "aliases",
  tattoos: "tattoos",
  piercings: "piercings",
  gender: "gender",
  "official site": "website",
  website: "website",
};

/** Values themes print for a fact they do not have. */
const EMPTY_VALUES = new Set(["", "n/a", "na", "none", "-", "--", "0", "unknown", "not specified"]);

const MONTHS = ["jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"];

/** Elements whose text holds a short "label: value" fact. */
const FACT_ELEMENT_SELECTOR = "li, .channel-stat-item, dl div, .website, .ent-facts";

/** Containers of the model's biography, most specific first. */
const BIO_SELECTORS = [".model-bio-text", ".ent-bio", ".model-description", ".asside-desc", ".channel-cover-short span", ".desc"];

const FEET_TO_CM = 30.48;
const INCH_TO_CM = 2.54;
const POUND_TO_KG = 0.45359237;

/** Whether a model page is under one of the site's model prefixes; returns its slug or "". */
export function profileSlug(site: SiteConfig, url: string): string {
  return entitySlug(site, url, site.modelPrefixes);
}

/** performer:find for a model page URL: fetches the page and parses it. */
export async function profileFind(url: string): Promise<ScrapeResult[]> {
  const site = siteForUrl(url);
  if (!site) throw new Error(`kvs: unsupported site ${url}`);
  if (profileSlug(site, url) === "") throw new Error(`kvs: not a model page: ${url}`);
  const pageUrl = canonicalUrl(site, url);
  const html = await fetchPage(site, pageUrl);
  return [{ type: "performer", performer: parseProfile(site, pageUrl, html) }];
}

/** The Performer a model page describes. */
export function parseProfile(site: SiteConfig, pageUrl: string, html: string): Performer {
  const root = parse(html);
  const slug = profileSlug(site, pageUrl);
  const person = jsonLdPerson(root);
  const facts = readFacts(root);

  const performer: Performer = {
    type: "performer",
    external_id: entityExternalId(site, "model", slug),
    source_url: pageUrl,
    name: profileName(root, person, slug),
  };
  setAliases(performer, facts, person);
  setBiography(performer, root, person);
  setDetails(performer, facts, person);
  setLinks(performer, facts.get("website"), root);
  setImage(performer, root, person, pageUrl);
  return performer;
}

/** The model's name: JSON-LD, else the heading without "'s New Videos"-style tails, else the slug. */
function profileName(root: HTMLElement, person: Record<string, unknown>, slug: string): string {
  if (typeof person.name === "string" && cleanText(person.name) !== "") return cleanText(person.name);
  for (const heading of root.querySelectorAll("h1")) {
    const name = nameFromHeading(cleanText(heading.text), slug);
    if (name !== "") return name;
  }
  return slug.split("-").map(capitalize).join(" ");
}

/**
 * The name in a heading such as "Riley Reid xxx videos": the heading's own
 * spelling of the slug's words when it starts with them, else the heading
 * without its tail.
 */
function nameFromHeading(heading: string, slug: string): string {
  const slugWords = slug.split("-").join(" ");
  if (heading.toLowerCase().startsWith(slugWords.toLowerCase())) return heading.slice(0, slugWords.length);
  return withoutHeadingTail(heading);
}

/** A heading such as "Ryan Smiles's New Videos" or "Riley Reid xxx videos" reduced to the name. */
export function withoutHeadingTail(heading: string): string {
  return heading
    .replace(/\s+(?:porn(?:o)?|xxx|sex|hd|free|hardcore|new|latest|best|top)?\s*(?:videos?|movies?|porn(?:o)?|collection|scenes?|[- ]videos?)\b.*$/i, "")
    .replace(/(?:'s|’s|s')$/i, (tail) => (tail === "s'" ? "s" : ""))
    .trim();
}

/** `word` with a capital first letter. */
function capitalize(word: string): string {
  return word.charAt(0).toUpperCase() + word.slice(1);
}

/** The page's JSON-LD Person, or an empty object. */
function jsonLdPerson(root: HTMLElement): Record<string, unknown> {
  for (const script of root.querySelectorAll('script[type="application/ld+json"]')) {
    const entry = personIn(parseLenientJson(script.rawText));
    if (entry) return entry;
  }
  return {};
}

/** The Person among a parsed JSON-LD value (a list, an object or a @graph). */
function personIn(value: unknown): Record<string, unknown> | undefined {
  if (Array.isArray(value)) {
    for (const entry of value) {
      const found = personIn(entry);
      if (found) return found;
    }
    return undefined;
  }
  if (!value || typeof value !== "object") return undefined;
  const record = value as Record<string, unknown>;
  if (record["@type"] === "Person") return record;
  return personIn(record["@graph"]);
}

/** JSON.parse tolerating the raw control characters some themes leave in strings; null on failure. */
function parseLenientJson(text: string): unknown {
  try {
    return JSON.parse(text.replace(/[\u0000-\u001f]+/g, " "));
  } catch {
    return null;
  }
}

/** Every fact the page lists, by key; the first listing of a fact wins. */
function readFacts(root: HTMLElement): Map<FactKey, string> {
  const facts = new Map<FactKey, string>();
  for (const element of root.querySelectorAll(FACT_ELEMENT_SELECTOR)) {
    const text = cleanText(element.text);
    if (text.length > 240) continue;
    addFactsFromText(facts, text, element);
  }
  return facts;
}

/** Adds the fact(s) one element's text gives: "Label: value", "Born On value", or "Age 28 · Height 172 cm". */
function addFactsFromText(facts: Map<FactKey, string>, text: string, element: HTMLElement): void {
  if (element.classList.contains("ent-facts")) {
    for (const part of text.split("·")) addFact(facts, ...labelAndValue(part.trim()));
    return;
  }
  const stat = statPair(element);
  if (stat) {
    addFact(facts, stat[0], stat[1]);
    return;
  }
  addFact(facts, ...labelAndValue(text));
}

/** The title/value pair of a stat box (`.channel-stat-title` + `.channel-stat-value`), if it is one. */
function statPair(element: HTMLElement): [string, string] | undefined {
  const title = element.querySelector(".channel-stat-title");
  const value = element.querySelector(".channel-stat-value");
  const termTitle = element.querySelector("dt");
  const termValue = element.querySelector("dd");
  if (title && value) return [cleanText(title.text), cleanText(value.text)];
  if (termTitle && termValue) return [cleanText(termTitle.text), cleanText(termValue.text)];
  return undefined;
}

/** Splits "Height: 172 cm", "Born On June 1, 1990" or "Height 172 cm" into label and value. */
function labelAndValue(text: string): [string, string] {
  const colon = text.match(/^([A-Za-z][A-Za-z ]{1,20}?)\s*:\s*(.+)$/);
  if (colon) return [colon[1], colon[2]];
  const bornOn = text.match(/^(born on)\s+(.+)$/i);
  if (bornOn) return [bornOn[1], bornOn[2]];
  const spaced = text.match(/^([A-Za-z][A-Za-z ]{1,20}?)\s+(\d.*)$/);
  if (spaced) return [spaced[1], spaced[2]];
  return ["", ""];
}

/** Records a labelled value under its fact when the label is known and the value says something. */
function addFact(facts: Map<FactKey, string>, label: string, value: string): void {
  const key = FACT_LABELS[label.trim().toLowerCase()];
  if (key === undefined) return;
  const cleaned = value.trim();
  if (EMPTY_VALUES.has(cleaned.toLowerCase())) return;
  if (!facts.has(key)) facts.set(key, cleaned);
}

/** Alternative names from the "Also known as" fact or JSON-LD alternateName. */
function setAliases(performer: Performer, facts: Map<FactKey, string>, person: Record<string, unknown>): void {
  const names: string[] = [];
  const listed = facts.get("aliases");
  if (listed) names.push(...listed.split(/[,;/]/));
  if (typeof person.alternateName === "string") names.push(...person.alternateName.split(/[,;/]/));
  const aliases: string[] = [];
  for (const entry of names) {
    const alias = cleanText(entry);
    if (alias === "" || alias === performer.name || aliases.includes(alias)) continue;
    aliases.push(alias);
  }
  if (aliases.length > 0) performer.aliases = aliases;
}

/** The biography: JSON-LD description, else the first bio container with real text. */
function setBiography(performer: Performer, root: HTMLElement, person: Record<string, unknown>): void {
  const candidates: string[] = [];
  if (typeof person.description === "string") candidates.push(person.description);
  for (const selector of BIO_SELECTORS) {
    const element = root.querySelector(selector);
    if (!element) continue;
    candidates.push(element.getAttribute("title") ?? "", element.text);
  }
  const bio = longestReal(candidates);
  if (bio !== "") performer.details = bio;
}

/** The longest candidate that is real text, not a "No description" placeholder. */
function longestReal(candidates: string[]): string {
  let best = "";
  for (const candidate of candidates) {
    const text = cleanText(candidate);
    if (/^(no description|no bio|n\/a)\b/i.test(text)) continue;
    if (text.length > best.length) best = text;
  }
  // The first source is the structured one; later containers only win when it is missing or cut short.
  if (best.length < 40) return "";
  return best;
}

/** Birthdate, country, gender and body facts. */
function setDetails(performer: Performer, facts: Map<FactKey, string>, person: Record<string, unknown>): void {
  const birthdate = parseBirthdate(facts.get("birthdate") ?? stringOf(person.birthDate));
  if (birthdate) performer.birthdate = birthdate;
  const country = parseCountry(facts.get("country") ?? stringOf(person.nationality));
  if (country) performer.country = country;
  const gender = parseGender(facts.get("gender") ?? "");
  if (gender) performer.gender = gender;
  const height = parseHeightCm(facts.get("height") ?? quantity(person.height));
  if (height) performer.height_cm = height;
  const weight = parseWeightKg(facts.get("weight") ?? "");
  if (weight) performer.weight_kg = weight;
  const measurements = parseMeasurements(facts.get("measurements") ?? "");
  if (measurements) performer.measurements = measurements;
  setText(performer, "ethnicity", facts.get("ethnicity"));
  setText(performer, "hair_color", facts.get("hair"));
  setText(performer, "eye_color", facts.get("eyes"));
  setText(performer, "tattoos", facts.get("tattoos"));
  setText(performer, "piercings", facts.get("piercings"));
}

/** Sets a plain text field when the page gave a value. */
function setText(
  performer: Performer,
  field: "ethnicity" | "hair_color" | "eye_color" | "tattoos" | "piercings",
  value: string | undefined,
): void {
  if (value) performer[field] = value;
}

/** A JSON-LD value as a string, "" when it is not one. */
function stringOf(value: unknown): string {
  if (typeof value !== "string") return "";
  return value;
}

/** A schema.org QuantitativeValue (or plain number) as text, "" when it is neither. */
function quantity(value: unknown): string {
  if (typeof value === "number") return `${value}cm`;
  if (!value || typeof value !== "object") return "";
  const amount = (value as { value?: unknown }).value;
  if (typeof amount === "number" || typeof amount === "string") return `${amount}cm`;
  return "";
}

/** A birthdate as YYYY-MM-DD from "1997-06-16", "June 16, 1997" or "16 January 1984"; undefined without a full date. */
export function parseBirthdate(text: string): string | undefined {
  const iso = text.match(/\b(\d{4})-(\d{2})-(\d{2})\b/);
  if (iso) return `${iso[1]}-${iso[2]}-${iso[3]}`;
  const monthFirst = text.match(/\b([A-Za-z]{3,9})\.?\s+(\d{1,2})(?:st|nd|rd|th)?,?\s+(\d{4})\b/);
  if (monthFirst) return datedParts(monthFirst[3], monthFirst[1], monthFirst[2]);
  const dayFirst = text.match(/\b(\d{1,2})(?:st|nd|rd|th)?\s+([A-Za-z]{3,9})\.?,?\s+(\d{4})\b/);
  if (dayFirst) return datedParts(dayFirst[3], dayFirst[2], dayFirst[1]);
  return undefined;
}

/** YYYY-MM-DD from a year, month name and day; undefined for an unknown month. */
function datedParts(year: string, monthName: string, day: string): string | undefined {
  const month = MONTHS.indexOf(monthName.slice(0, 3).toLowerCase());
  if (month < 0) return undefined;
  return `${year}-${String(month + 1).padStart(2, "0")}-${day.padStart(2, "0")}`;
}

/** The country from "Hungary, Budapest" (country first) or "United States". */
function parseCountry(text: string): string | undefined {
  const country = text.split(",")[0].trim();
  if (country === "" || EMPTY_VALUES.has(country.toLowerCase())) return undefined;
  return country;
}

/** The Performer gender value for the page's wording, or undefined. */
function parseGender(text: string): string | undefined {
  const lower = text.toLowerCase();
  if (lower === "female" || lower === "woman") return "female";
  if (lower === "male" || lower === "man") return "male";
  if (lower.includes("trans")) return "transgender_female";
  return undefined;
}

/** Height in cm from "162cm", "5'2\"", "5 feet 4 inches (163 cm)" or a bare number of cm; plausible heights only. */
export function parseHeightCm(text: string): number | undefined {
  const centimetres = text.match(/(\d+(?:\.\d+)?)\s*cm/i);
  if (centimetres) return plausible(Math.round(parseFloat(centimetres[1])), 120, 230);
  const feet = text.match(/(\d)\s*(?:'|ft|feet)\s*(\d{1,2})?/i);
  if (feet) return plausible(Math.round(parseInt(feet[1], 10) * FEET_TO_CM + parseInt(feet[2] ?? "0", 10) * INCH_TO_CM), 120, 230);
  const bare = text.match(/^(\d{3})$/);
  if (bare) return plausible(parseInt(bare[1], 10), 120, 230);
  return undefined;
}

/** Weight in kg from "73kg", "105lb", "128 pounds (58 kg)" or a bare number of kg; plausible weights only. */
export function parseWeightKg(text: string): number | undefined {
  const kilograms = text.match(/(\d+(?:\.\d+)?)\s*kg/i);
  if (kilograms) return plausible(Math.round(parseFloat(kilograms[1])), 30, 200);
  const pounds = text.match(/(\d+(?:\.\d+)?)\s*(?:lbs?|pounds?)/i);
  if (pounds) return plausible(Math.round(parseFloat(pounds[1]) * POUND_TO_KG), 30, 200);
  const bare = text.match(/^(\d{2,3})$/);
  if (bare) return plausible(parseInt(bare[1], 10), 30, 200);
  return undefined;
}

/** `value` when within the range, else undefined (themes print placeholders such as "62cm"). */
function plausible(value: number, min: number, max: number): number | undefined {
  if (value < min || value > max) return undefined;
  return value;
}

/** "34D-24-34" style measurements, from the first such triple in the text. */
function parseMeasurements(text: string): string | undefined {
  const match = text.match(/\b\d{2}[A-Za-z]{0,4}-\d{2}-\d{2}\b/);
  if (!match) return undefined;
  return match[0].toUpperCase();
}

/** Twitter/Instagram/personal site from the "Official Site" fact or the page's profile links. */
function setLinks(performer: Performer, website: string | undefined, root: HTMLElement): void {
  const links: string[] = [];
  if (website) links.push(website);
  for (const anchor of root.querySelectorAll(".website a, .model-links a, .jpm-social a")) {
    links.push(anchor.getAttribute("href") ?? "");
  }
  for (const link of links) {
    const url = link.trim();
    if (!/^https?:\/\//i.test(url)) continue;
    classifyLink(performer, url);
  }
}

/** Files a profile link as twitter, instagram or the model's own site; the first of each wins. */
function classifyLink(performer: Performer, url: string): void {
  if (/^https?:\/\/(?:www\.)?(?:twitter|x)\.com\//i.test(url)) {
    if (!performer.twitter) performer.twitter = url;
    return;
  }
  if (/^https?:\/\/(?:www\.)?instagram\.com\//i.test(url)) {
    if (!performer.instagram) performer.instagram = url;
    return;
  }
  if (!performer.url) performer.url = url;
}

/** The model's picture: JSON-LD image, else the model photo next to the page heading. */
function setImage(performer: Performer, root: HTMLElement, person: Record<string, unknown>, pageUrl: string): void {
  let image = stringOf(person.image);
  if (image === "") image = headerPhoto(root);
  if (image === "") return;
  try {
    performer.image_path = new URL(image, pageUrl).toString();
  } catch {
    // an unusable URL is no picture
  }
}

/** The first model photo in the closest ancestor of the heading that holds one (related-model sidebars come later). */
function headerPhoto(root: HTMLElement): string {
  const heading = root.querySelector("h1");
  let container: HTMLElement | null = heading;
  for (let level = 0; level < 5 && container !== null; level++) {
    container = container.parentNode as HTMLElement | null;
    const photo = container?.querySelector("img[src*='/contents/models/'], img[data-src*='/contents/models/'], img[data-original*='/contents/models/']");
    if (!photo) continue;
    return photoSource(photo);
  }
  return "";
}

/** An image's source URL, protocol-relative ones completed. */
function photoSource(photo: HTMLElement): string {
  const candidates = [photo.getAttribute("data-src"), photo.getAttribute("data-original"), photo.getAttribute("src")];
  for (const source of candidates) {
    if (!source || source.startsWith("data:")) continue;
    if (source.startsWith("//")) return "https:" + source;
    return source;
  }
  return "";
}
