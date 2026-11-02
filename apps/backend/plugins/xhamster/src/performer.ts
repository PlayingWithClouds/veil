import { fetchHtml } from "@playingwithclouds/veil-sdk";
import type { Image, Performer, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { BASE_URL, decodeEntities } from "./http.ts";

const PROFILE_MARKER = "window.initials=";

/** One "label: value" row of the profile's personal or appearance lists. */
interface InfoRow {
  label?: string;
  value?: string;
}

/** The subset of a pornstar page's `window.initials` state the scraper reads. */
interface PornstarInitials {
  infoComponent?: {
    aliases?: string;
    pageTitle?: string;
    pornstarTop?: {
      name?: string;
      translatedCountryName?: string;
      thumbUrl?: string;
    };
  };
  aboutMeComponent?: {
    text?: string;
    personalInfoList?: InfoRow[];
    appearanceInfoList?: InfoRow[];
  };
  channelInfo?: { slug?: string };
}

/** Profile URL of a pornstar slug. */
export function pornstarUrl(slug: string): string {
  return `${BASE_URL}/pornstars/${slug}`;
}

/** The slug of a /pornstars/<slug> URL, or "" when the URL is another kind of page. */
export function pornstarSlugFromUrl(url: string): string {
  const match = url.match(/^https?:\/\/(?:www\.)?xhamster\d*\.[a-z]+\/pornstars\/([^/?#]+)\/?(?:[?#].*)?$/i);
  if (!match) return "";
  return match[1].toLowerCase();
}

/** Whether a URL is a pornstar profile page (not the /pornstars listing or a sub-listing). */
export function isPornstarUrl(url: string): boolean {
  const slug = pornstarSlugFromUrl(url);
  if (slug === "" || slug === "all" || slug === "best") return false;
  return true;
}

/** The slug xhamster derives from a performer name: lower case, runs of other characters become one hyphen. */
export function slugFromName(name: string): string {
  return name
    .normalize("NFKD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

/** Reads the `window.initials=…;` JSON of a pornstar page, undefined when absent or broken. */
function parsePornstarInitials(html: string): PornstarInitials | undefined {
  const markerIndex = html.indexOf(PROFILE_MARKER);
  if (markerIndex === -1) return undefined;
  const start = markerIndex + PROFILE_MARKER.length;
  const end = html.indexOf("</script>", start);
  if (end === -1) return undefined;
  try {
    return JSON.parse(html.slice(start, end).trim().replace(/;$/, "")) as PornstarInitials;
  } catch {
    return undefined;
  }
}

/** Rows of an info list keyed by lower-cased label; rows without a value are dropped. */
function rowsByLabel(rows: InfoRow[] | undefined): Record<string, string> {
  const byLabel: Record<string, string> = {};
  if (!rows) return byLabel;
  for (const row of rows) {
    if (!row.label || !row.value) continue;
    byLabel[row.label.trim().toLowerCase()] = row.value.trim();
  }
  return byLabel;
}

/** The "I am" row ("Female", "Female, bisexual") as a schema gender, undefined when not one of them. */
function genderFromRow(value: string | undefined): string | undefined {
  if (!value) return undefined;
  const first = value.split(",")[0].trim().toLowerCase();
  if (first === "female") return "female";
  if (first === "male") return "male";
  if (first === "transgender female" || first === "trans woman") return "transgender_female";
  if (first === "transgender male" || first === "trans man") return "transgender_male";
  return undefined;
}

/** The first number of a value like "160 cm (5 ft 62 in)" or "48 kg / 105 lbs". */
function leadingNumber(value: string | undefined): number | undefined {
  if (!value) return undefined;
  const match = value.match(/(\d+(?:\.\d+)?)/);
  if (!match) return undefined;
  return parseFloat(match[1]);
}

/** The aliases line ("A, B, C") as a list, without the performer's own name. */
function aliasesFromLine(line: string | undefined, name: string): string[] {
  if (!line) return [];
  return line
    .split(",")
    .map((alias) => decodeEntities(alias).trim())
    .filter((alias) => alias !== "" && alias.toLowerCase() !== name.toLowerCase());
}

/** The appearance rows the schema has a field for, copied onto the performer. */
function applyAppearance(performer: Performer, rows: Record<string, string>): void {
  if (rows["ethnicity"]) performer.ethnicity = rows["ethnicity"];
  if (rows["hair color"]) performer.hair_color = rows["hair color"];
  if (rows["eye color"]) performer.eye_color = rows["eye color"];
  if (rows["tattoos"]) performer.tattoos = rows["tattoos"];
  if (rows["piercings"]) performer.piercings = rows["piercings"];
  if (rows["measurements"]) performer.measurements = rows["measurements"];
  const heightCm = leadingNumber(rows["height"]);
  if (heightCm !== undefined) performer.height_cm = heightCm;
  const weightKg = leadingNumber(rows["weight"]);
  if (weightKg !== undefined) performer.weight_kg = weightKg;
}

/** The avatar as image_path and a profile image. */
function applyAvatar(performer: Performer, thumbUrl: string | undefined): void {
  if (!thumbUrl) return;
  performer.image_path = thumbUrl;
  const image: Image = { type: "profile", file_path: thumbUrl };
  performer.images = [image];
}

/**
 * Parses a pornstar page into a Performer, undefined when the page carries no
 * profile. The page states no birthdate (only generated prose mentions one,
 * which is not trusted), so the field stays empty.
 */
export function parsePornstarPage(html: string, sourceUrl: string): Performer | undefined {
  const initials = parsePornstarInitials(html);
  if (!initials || !initials.infoComponent) return undefined;
  const top = initials.infoComponent.pornstarTop;
  const name = (top && top.name) || initials.infoComponent.pageTitle;
  if (!name) return undefined;

  const slug = (initials.channelInfo && initials.channelInfo.slug) || pornstarSlugFromUrl(sourceUrl);
  const performer: Performer = {
    type: "performer",
    external_id: `xhamster-ps-${slug}`,
    source_url: pornstarUrl(slug),
    name: decodeEntities(name).trim(),
  };

  const aliases = aliasesFromLine(initials.infoComponent.aliases, performer.name);
  if (aliases.length > 0) performer.aliases = aliases;

  const about = initials.aboutMeComponent;
  if (about) {
    const details = about.text ? about.text.trim() : "";
    if (details !== "") performer.details = details;
    const personal = rowsByLabel(about.personalInfoList);
    const gender = genderFromRow(personal["i am"]);
    if (gender) performer.gender = gender;
    applyAppearance(performer, rowsByLabel(about.appearanceInfoList));
  }

  if (top && top.translatedCountryName) performer.country = top.translatedCountryName;
  applyAvatar(performer, top && top.thumbUrl);
  return performer;
}

/** Fetches a pornstar page; undefined when xhamster answers 404 (no such profile). */
export async function fetchPornstarPage(url: string): Promise<string | undefined> {
  try {
    return await fetchHtml(url);
  } catch (error) {
    if (String(error).includes("HTTP 404")) return undefined;
    throw error;
  }
}

/** performer:find: one pornstar profile by URL. */
export async function performerFind(url: string): Promise<ScrapeResult[]> {
  const html = await fetchHtml(url);
  const performer = parsePornstarPage(html, url);
  if (!performer) throw new Error(`no performer profile at ${url}`);
  return [{ type: "performer", performer }];
}
