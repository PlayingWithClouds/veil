import type { HTMLElement } from "node-html-parser";
import { cleanText, decodeEntities } from "./text.ts";

// Page-level metadata KVS themes expose in machine-readable form: OpenGraph
// and video:* meta tags, and a schema.org VideoObject in JSON-LD.

/** The schema.org VideoObject fields used here. */
export interface VideoObject {
  name?: string;
  description?: string;
  duration?: string;
  uploadDate?: string;
  thumbnailUrl?: string | string[];
  actor?: unknown;
  author?: unknown;
  productionCompany?: unknown;
  keywords?: unknown;
  genre?: unknown;
}

/** The content of the first meta tag with this property/name/itemprop, or "". */
export function metaContent(root: HTMLElement, name: string): string {
  const selector = `meta[property="${name}"], meta[name="${name}"], meta[itemprop="${name}"]`;
  return cleanText(root.querySelector(selector)?.getAttribute("content") ?? "");
}

/** Every content of meta tags with this property (e.g. repeated video:tag). */
export function metaContents(root: HTMLElement, name: string): string[] {
  const values: string[] = [];
  for (const meta of root.querySelectorAll(`meta[property="${name}"]`)) {
    const value = cleanText(meta.getAttribute("content") ?? "");
    if (value) values.push(value);
  }
  return values;
}

/** The page's JSON-LD VideoObject, or an empty object. */
export function videoObject(root: HTMLElement): VideoObject {
  for (const script of root.querySelectorAll('script[type="application/ld+json"]')) {
    // rawText: script bodies are not entity-encoded.
    for (const candidate of asList(parseLenientJson(script.rawText))) {
      if (isVideoObject(candidate)) return candidate as VideoObject;
    }
  }
  return {};
}

/** Whether a parsed JSON-LD entry is a schema.org VideoObject. */
function isVideoObject(entry: unknown): boolean {
  if (!entry || typeof entry !== "object") return false;
  return (entry as { "@type"?: unknown })["@type"] === "VideoObject";
}

/** A JSON-LD value as a list: arrays as they are, anything else as one entry. */
function asList(value: unknown): unknown[] {
  if (Array.isArray(value)) return value;
  return [value];
}

/** JSON.parse tolerating the raw control characters some themes leave in strings; null on failure. */
function parseLenientJson(text: string): unknown {
  try {
    return JSON.parse(text.replace(/[\u0000-\u001f]+/g, " "));
  } catch {
    return null;
  }
}

/**
 * The names in a JSON-LD person/organization field: a string, an object with
 * a name, or a list of either. Objects of `excludedType` (e.g. uploader
 * "Person"s in author) are skipped.
 */
export function jsonLdNames(value: unknown, excludedType = ""): string[] {
  const names: string[] = [];
  for (const entry of asList(value)) {
    const name = jsonLdName(entry, excludedType);
    if (name) names.push(name);
  }
  return names;
}

/** The name of one JSON-LD entry, "" when it has none or is of the excluded type. */
function jsonLdName(entry: unknown, excludedType: string): string {
  if (typeof entry === "string") return cleanText(decodeEntities(entry));
  if (!entry || typeof entry !== "object") return "";
  const record = entry as { name?: unknown; "@type"?: unknown };
  if (excludedType && record["@type"] === excludedType) return "";
  if (typeof record.name !== "string") return "";
  return cleanText(decodeEntities(record.name));
}

/** JSON-LD keywords/genre as a list: comma-separated strings are split. */
export function jsonLdList(value: unknown): string[] {
  const items: string[] = [];
  for (const entry of asList(value)) {
    if (typeof entry !== "string") continue;
    for (const part of entry.split(",")) {
      const item = cleanText(decodeEntities(part));
      if (item) items.push(item);
    }
  }
  return items;
}

/** A JSON-LD string field, "" when absent or not a string. */
export function jsonLdString(value: unknown): string {
  if (Array.isArray(value)) return jsonLdString(value[0]);
  if (typeof value !== "string") return "";
  return cleanText(decodeEntities(value));
}
