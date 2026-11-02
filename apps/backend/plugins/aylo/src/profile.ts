import { parse, HTMLElement } from "node-html-parser";
import type { Performer, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { fetchText, slugFromHref } from "./http.ts";
import { requireSite, type Site } from "./sites.ts";

/** Matches a pornstar or model profile path, capturing [kind, slug]. */
const PROFILE_PATH = /^\/(pornstar|model)\/([^/?#]+)/;

/** Whether a URL is a performer profile page of an Aylo site. */
export function isProfileUrl(url: string): boolean {
  let pathname = "";
  try {
    pathname = new URL(url).pathname;
  } catch {
    return false;
  }
  return PROFILE_PATH.test(pathname);
}

/** The profile page URL itself, without sub-pages such as /videos and without query. */
export function canonicalProfileUrl(site: Site, url: string): string {
  const match = new URL(url).pathname.match(PROFILE_PATH);
  if (!match) throw new Error(`aylo: not a profile URL: ${url}`);
  return `https://${site.host}/${match[1]}/${match[2]}`;
}

/** Profile URL a site would use for a performer slug (words joined by spaces already folded to "-"). */
export function profileUrlForSlug(site: Site, slug: string): string {
  // RedTube spells multi-word slugs with "+" and redirects the hyphen form to it.
  if (site.key === "redtube") return `https://${site.host}/pornstar/${slug.replace(/-/g, "+")}`;
  if (site.key === "pornhub") return `https://${site.host}/pornstar/${slug}`;
  return `https://${site.host}/pornstar/${slug}/`;
}

/** performer:find for a profile URL: one fetch, one Performer. */
export async function profileFind(url: string): Promise<ScrapeResult[]> {
  const site = requireSite(url);
  const profileUrl = canonicalProfileUrl(site, url);
  const performer = parseProfile(site, profileUrl, await fetchText(profileUrl));
  if (!performer) throw new Error(`aylo: no performer profile at ${profileUrl}`);
  return [{ type: "performer", performer }];
}

/**
 * Builds the Performer from a profile page, undefined when the page is not a
 * profile (Pornhub answers an unknown pornstar slug with its pornstar index).
 */
export function parseProfile(site: Site, profileUrl: string, html: string): Performer | undefined {
  const doc = parse(html);
  const name = profileName(site, doc);
  if (!name) return undefined;

  const stats = profileStats(site, doc);
  const performer: Performer = {
    type: "performer",
    external_id: `${site.key}-ps-${slugFromHref(new URL(profileUrl).pathname).replace(/\s+/g, "-")}`,
    source_url: profileUrl,
    name,
  };
  fillBiography(site, doc, stats, performer, name);
  fillBody(performer, stats);
  fillBackground(performer, stats);
  fillLinks(doc, performer);
  return performer;
}

/** The performer's display name from the page heading. */
function profileName(site: Site, doc: HTMLElement): string {
  let heading = doc.querySelector("h1.name-title");
  if (site.key === "pornhub") heading = doc.querySelector(".topProfileHeader h1");
  if (!heading) return "";
  return heading.text.replace(/\s+/g, " ").trim();
}

/** The "label: value" facts of the profile, keyed by lower-cased label. */
function profileStats(site: Site, doc: HTMLElement): Record<string, string> {
  const stats: Record<string, string> = {};
  if (site.key === "pornhub") {
    for (const piece of doc.querySelectorAll(".detailedInfo .infoPiece")) {
      const label = piece.querySelector("span");
      const value = piece.querySelector(".smallInfo");
      addStat(stats, label ? label.text : "", value ? value.text : "");
    }
    return stats;
  }
  for (const item of doc.querySelectorAll(".profile-info li.info-stat")) {
    const label = item.querySelector(".info-stat-label");
    const value = item.querySelector(".info-stat-data");
    addStat(stats, label ? label.text : "", value ? value.text : "");
  }
  return stats;
}

/** Records one fact when both halves are non-empty; labels lose their trailing colon. */
function addStat(stats: Record<string, string>, label: string, value: string): void {
  const key = label.replace(/:\s*$/, "").replace(/\s+/g, " ").trim().toLowerCase();
  const text = value.replace(/\s+/g, " ").trim();
  if (key === "" || text === "") return;
  stats[key] = text;
}

/** Image, bio and aliases. */
function fillBiography(
  site: Site,
  doc: HTMLElement,
  stats: Record<string, string>,
  performer: Performer,
  name: string,
): void {
  const image = profileImage(site, doc);
  if (image) {
    performer.image_path = image;
    performer.images = [{ type: "profile", file_path: image }];
  }
  const bio = profileBio(site, doc);
  if (bio) performer.details = bio;
  const aliases = profileAliases(site, doc, stats, name);
  if (aliases.length > 0) performer.aliases = aliases;
}

/** The avatar URL; Pornhub renders it directly, the others lazy-load it. */
function profileImage(site: Site, doc: HTMLElement): string {
  if (site.key === "pornhub") {
    const avatar = doc.querySelector("#getAvatar");
    if (!avatar) return "";
    return avatar.getAttribute("src") || "";
  }
  const avatar = doc.querySelector(".avatar-wrapper img");
  if (!avatar) return "";
  return avatar.getAttribute("data-src") || "";
}

/** The performer biography, falling back to Pornhub's self-written "About" text. */
function profileBio(site: Site, doc: HTMLElement): string {
  if (site.key !== "pornhub") {
    const bio = doc.querySelector(".profile-bio");
    if (!bio) return "";
    return bio.text.replace(/\s+/g, " ").trim();
  }
  const bio = doc.querySelector('.biographyText [itemprop="description"]');
  if (bio) return bio.text.replace(/\s+/g, " ").trim();
  for (const about of doc.querySelectorAll(".aboutMeSection div")) {
    if (about.classList.contains("title")) continue;
    return about.text.replace(/\s+/g, " ").trim();
  }
  return "";
}

/** Other names shown on the page, without the performer's own name. */
function profileAliases(site: Site, doc: HTMLElement, stats: Record<string, string>, name: string): string[] {
  let element = doc.querySelector(".alias_names");
  if (site.key === "pornhub") element = doc.querySelector(".pornstarAliases .aliases");
  // The AKA fact is the only list on some pages (the alias section sits far below the fold on others).
  let text = stats["performer aka"] || "";
  if (element) text = element.text;
  const aliases: string[] = [];
  for (const part of text.split(",")) {
    const alias = part.replace(/\s+/g, " ").trim();
    if (alias === "" || alias.toLowerCase() === name.toLowerCase()) continue;
    if (aliases.includes(alias)) continue;
    aliases.push(alias);
  }
  return aliases;
}

/** Appearance and measurements. */
function fillBody(performer: Performer, stats: Record<string, string>): void {
  const height = numberBefore(stats["height"], "cm");
  if (height) performer.height_cm = height;
  const weight = numberBefore(stats["weight"], "kg");
  if (weight) performer.weight_kg = weight;
  if (stats["measurements"]) performer.measurements = stats["measurements"];
  if (stats["ethnicity"]) performer.ethnicity = stats["ethnicity"];
  const hair = stats["hair color"] || stats["hair"];
  if (hair) performer.hair_color = hair;
  const eyes = stats["eye color"] || stats["eyes"];
  if (eyes) performer.eye_color = eyes;
  const breasts = breastType(stats);
  if (breasts) performer.fake_tits = breasts;
  const tattoos = presentOrUndefined(stats["tattoos"]);
  if (tattoos) performer.tattoos = tattoos;
  const piercings = presentOrUndefined(stats["piercings"]);
  if (piercings) performer.piercings = piercings;
}

/** Who they are and since when: gender, birthdate, country, career. */
function fillBackground(performer: Performer, stats: Record<string, string>): void {
  const gender = genderValue(stats["gender"]);
  if (gender) performer.gender = gender;
  const birthdate = isoDate(stats["date of birth"] || stats["born"] || stats["birthday"] || stats["birth date"]);
  if (birthdate) performer.birthdate = birthdate;
  const country = countryOfBirthPlace(stats["birth place"]);
  if (country) performer.country = country;
  const career = stats["career start and end"] || stats["years active"];
  if (career) performer.career_length = career.replace(/\s*\(.*\)\s*$/, "").trim();
}

/** Official site, Twitter and Instagram from the profile's social links. */
function fillLinks(doc: HTMLElement, performer: Performer): void {
  for (const anchor of doc.querySelectorAll(".socialList a, .main-stats-wrapper a.social-link")) {
    const href = anchor.getAttribute("href") || "";
    if (!/^https?:\/\//.test(href)) continue;
    if (/(^|\.)(twitter|x)\.com\//.test(href.replace(/^https?:\/\//, ""))) {
      if (!performer.twitter) performer.twitter = href;
    } else if (/instagram\.com\//.test(href)) {
      if (!performer.instagram) performer.instagram = href;
    } else if (anchor.text.trim() === "Official Site") {
      if (!performer.url) performer.url = href;
    }
  }
}

/** The integer right before a unit, "5 ft 7 in (170 cm)" gives 170. */
function numberBefore(text: string | undefined, unit: string): number | undefined {
  if (!text) return undefined;
  const match = text.match(new RegExp(`(\\d+(?:\\.\\d+)?)\\s*${unit}\\b`, "i"));
  if (!match) return undefined;
  return Math.round(parseFloat(match[1]));
}

/** Tattoo/piercing facts; a bare "No" means none, which is nothing to store. */
function presentOrUndefined(text: string | undefined): string | undefined {
  if (!text) return undefined;
  if (/^(no|none)\.?$/i.test(text)) return undefined;
  return text;
}

/** "Natural"/"Fake" from Pornhub's "Fake Boobs: Yes|No" or the others' "Breast Type". */
function breastType(stats: Record<string, string>): string | undefined {
  if (stats["breast type"]) return stats["breast type"];
  const fake = (stats["fake boobs"] || "").toLowerCase();
  if (fake === "yes") return "Fake";
  if (fake === "no") return "Natural";
  return undefined;
}

/** Maps the page's gender word to the Performer gender values. */
function genderValue(text: string | undefined): string | undefined {
  if (!text) return undefined;
  const lower = text.toLowerCase();
  if (lower === "female") return "female";
  if (lower === "male") return "male";
  if (lower.includes("trans") && lower.includes("female")) return "transgender_female";
  if (lower.includes("trans") && lower.includes("male")) return "transgender_male";
  return undefined;
}

const MONTHS = ["jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"];

/** YYYY-MM-DD from "1992-06-01" or "Jun 1, 1992"; undefined for anything else. */
function isoDate(text: string | undefined): string | undefined {
  if (!text) return undefined;
  const iso = text.match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (iso) return `${iso[1]}-${iso[2]}-${iso[3]}`;
  const written = text.match(/^([A-Za-z]{3})[a-z]*\.?\s+(\d{1,2}),?\s+(\d{4})/);
  if (!written) return undefined;
  const month = MONTHS.indexOf(written[1].toLowerCase());
  if (month === -1) return undefined;
  return `${written[3]}-${String(month + 1).padStart(2, "0")}-${written[2].padStart(2, "0")}`;
}

/** The country is the last part of "Palm Springs, California, United States of America". */
function countryOfBirthPlace(text: string | undefined): string | undefined {
  if (!text) return undefined;
  const parts = text.split(",");
  return parts[parts.length - 1].trim() || undefined;
}
