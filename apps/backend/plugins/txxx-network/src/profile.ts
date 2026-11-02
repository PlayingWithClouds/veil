import { fetchHtml, type Performer, type ScrapeResult } from "@playingwithclouds/veil-sdk";
import { baseUrl, fetchJson, hostFromUrl } from "./http.ts";
import { hostScopedId } from "./listing.ts";

// A model page. The site is a single-page app; its page asks
// /api/json/model/<lifetime>/<first two letters>/<dir>.json for the record,
// and /langs/json/en.json names the codes it uses for country, hair and eyes.

const MODEL_LIFETIME = 86400;
const CM_PER_INCH = 2.54;

/** The `model` record of the model API; everything arrives as strings. */
export interface ApiModel {
  model_id?: string;
  title?: string;
  alias?: string;
  description?: string;
  screenshot1?: string;
  img?: string;
  height?: string;
  weight?: string;
  measurements?: string;
  birth_date?: string;
  dir?: string;
  career_start?: string;
  career_end?: string;
  tattoo_description?: string;
  piercing_description?: string;
  website?: string;
  country?: string;
  hair?: string;
  eye?: string;
  gender?: string;
  race?: string;
}

/** The slice of the site's English language file the model record's codes refer to. */
export interface SiteLanguage {
  countries?: Record<string, string>;
  component?: { filter?: { hair_color?: Record<string, string>; race?: Record<string, string> } };
  page?: { model?: { eye_color?: Record<string, string> } };
}

/** The model record API's answer: the record, or an error code such as model_not_found. */
export interface ModelResponse {
  model?: ApiModel;
  error?: number;
  code?: string;
}

/** The model slug of a /models/<dir>/ page URL (any language prefix allowed), or "". */
export function modelDir(url: string): string {
  let pathname: string;
  try {
    pathname = new URL(url).pathname;
  } catch {
    return "";
  }
  const match = pathname.match(/\/models\/([^/]+)\/?$/);
  if (!match) return "";
  return decodeURIComponent(match[1]);
}

/** The API path of a model record. */
export function modelApiPath(dir: string): string {
  return `/api/json/model/${MODEL_LIFETIME}/${dir.slice(0, 2)}/${dir}.json`;
}

/** The model page URL on a network site. */
export function modelPageUrl(host: string, dir: string): string {
  return `${baseUrl(host)}/models/${dir}/`;
}

/** The record of one model, undefined when the site does not know the slug. */
export async function fetchModel(host: string, dir: string): Promise<ApiModel | undefined> {
  const response = await fetchJson<ModelResponse>(host, modelApiPath(dir));
  return response.model;
}

/** The site's English language file, for its country/hair/eye code names. */
export async function fetchLanguage(host: string): Promise<SiteLanguage> {
  return JSON.parse(await fetchHtml(`${baseUrl(host)}/langs/json/en.json`, {})) as SiteLanguage;
}

/** performer:find for a /models/<dir>/ page URL. */
export async function profileFind(url: string): Promise<ScrapeResult[]> {
  const host = hostFromUrl(url);
  const dir = modelDir(url);
  if (dir === "") throw new Error(`txxx-network: not a model page: ${url}`);
  const model = await fetchModel(host, dir);
  if (!model) throw new Error(`txxx-network: no model ${dir} on ${host}`);
  const performer = parseModel(host, model, await fetchLanguage(host));
  return [{ type: "performer", performer }];
}

/** The Performer a model record describes. */
export function parseModel(host: string, model: ApiModel, language: SiteLanguage): Performer {
  const dir = text(model.dir);
  const performer: Performer = {
    type: "performer",
    external_id: hostScopedId(host, "model", dir),
    source_url: modelPageUrl(host, dir),
    name: text(model.title),
  };
  setAliases(performer, model);
  setDetails(performer, model);
  setBody(performer, model, language);
  setLinks(performer, text(model.website));
  setImage(performer, model);
  return performer;
}

/** A record field as trimmed text; the API sends "null" or nothing for missing ones. */
function text(value: string | number | null | undefined): string {
  if (typeof value === "number") return String(value);
  if (typeof value !== "string") return "";
  const trimmed = decodeEntities(value).trim();
  if (trimmed.toLowerCase() === "null") return "";
  return trimmed;
}

/** The entities the API leaves in its text fields. */
function decodeEntities(value: string): string {
  return value
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#0?39;/g, "'");
}

/** The comma-separated "alias" field as aliases. */
function setAliases(performer: Performer, model: ApiModel): void {
  const aliases: string[] = [];
  for (const entry of text(model.alias).split(",")) {
    const alias = entry.trim();
    if (alias === "" || alias === performer.name || aliases.includes(alias)) continue;
    aliases.push(alias);
  }
  if (aliases.length > 0) performer.aliases = aliases;
}

/** Biography, birthdate, career and the tattoo/piercing notes. */
function setDetails(performer: Performer, model: ApiModel): void {
  const details = text(model.description);
  if (details !== "") performer.details = details;
  const birthdate = parseBirthDate(text(model.birth_date));
  if (birthdate) performer.birthdate = birthdate;
  const career = careerLength(text(model.career_start), text(model.career_end));
  if (career) performer.career_length = career;
  const tattoos = text(model.tattoo_description);
  if (tattoos !== "") performer.tattoos = tattoos;
  const piercings = text(model.piercing_description);
  if (piercings !== "") performer.piercings = piercings;
}

/** "2010-2023", "2010-" for a model still working, undefined without a start. */
function careerLength(start: string, end: string): string | undefined {
  if (!/^\d{4}$/.test(start)) return undefined;
  if (/^\d{4}$/.test(end)) return `${start}-${end}`;
  return `${start}-`;
}

const MONTHS = ["jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"];

/** YYYY-MM-DD from the site's "9 July, 1991"; undefined for anything else. */
export function parseBirthDate(value: string): string | undefined {
  const match = value.match(/^(\d{1,2})\s+([A-Za-z]{3,9}),?\s+(\d{4})$/);
  if (!match) return undefined;
  const month = MONTHS.indexOf(match[2].slice(0, 3).toLowerCase());
  if (month < 0) return undefined;
  return `${match[3]}-${String(month + 1).padStart(2, "0")}-${match[1].padStart(2, "0")}`;
}

/** Country, gender, ethnicity, hair, eyes, height, weight and measurements. */
function setBody(performer: Performer, model: ApiModel, language: SiteLanguage): void {
  const country = language.countries?.[text(model.country).toLowerCase()];
  if (country) performer.country = country;
  const gender = text(model.gender).toLowerCase();
  if (gender === "female" || gender === "male") performer.gender = gender;
  const race = language.component?.filter?.race?.[text(model.race)];
  if (race) performer.ethnicity = race;
  const hair = language.component?.filter?.hair_color?.[text(model.hair)];
  if (hair) performer.hair_color = hair;
  const eyes = language.page?.model?.eye_color?.[text(model.eye)];
  if (eyes) performer.eye_color = eyes;
  const height = positiveNumber(text(model.height));
  if (height) performer.height_cm = height;
  const weight = positiveNumber(text(model.weight));
  if (weight) performer.weight_kg = weight;
  const measurements = parseMeasurements(text(model.measurements));
  if (measurements) performer.measurements = measurements;
}

/** A positive whole number from text, 0 for "0", empty or non-numeric. */
function positiveNumber(value: string): number {
  const number = parseInt(value, 10);
  if (!Number.isFinite(number) || number <= 0) return 0;
  return number;
}

/** "32A-58 cm-86 cm" or "34D-24-36" as bust-waist-hips in inches; undefined when not three parts. */
export function parseMeasurements(value: string): string | undefined {
  const match = value.match(/^(\d{2}[A-Za-z]{0,4})\s*-\s*(\d{2,3})\s*(cm)?\s*-\s*(\d{2,3})\s*(cm)?$/);
  if (!match) return undefined;
  const waist = inches(match[2], match[3] !== undefined);
  const hips = inches(match[4], match[5] !== undefined);
  return `${match[1].toUpperCase()}-${waist}-${hips}`;
}

/** A measurement in inches, converted from centimetres when the site gave it in cm. */
function inches(amount: string, centimetres: boolean): number {
  const number = parseInt(amount, 10);
  if (!centimetres) return number;
  return Math.round(number / CM_PER_INCH);
}

/** The model's own site, filed as twitter, instagram or url. */
function setLinks(performer: Performer, website: string): void {
  if (!/^https?:\/\//i.test(website)) return;
  if (/^https?:\/\/(?:www\.)?(?:twitter|x)\.com\//i.test(website)) {
    performer.twitter = website;
    return;
  }
  if (/^https?:\/\/(?:www\.)?instagram\.com\//i.test(website)) {
    performer.instagram = website;
    return;
  }
  performer.url = website;
}

/** The model's picture. */
function setImage(performer: Performer, model: ApiModel): void {
  let image = text(model.img);
  if (image === "") image = text(model.screenshot1);
  if (image === "") return;
  if (image.startsWith("//")) image = "https:" + image;
  performer.image_path = image;
}
