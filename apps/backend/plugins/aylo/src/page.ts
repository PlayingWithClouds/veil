/** Helpers for reading the JSON the sites embed in their scene pages. */

export interface VideoObject {
  name?: string;
  description?: string;
  thumbnailUrl?: string;
  uploadDate?: string;
  duration?: string;
}

/** The schema.org VideoObject from the page's ld+json blocks (flat or inside an @graph). */
export function extractVideoObject(html: string): VideoObject | undefined {
  const blockPattern = /<script type="application\/ld\+json">([\s\S]*?)<\/script>/g;
  let block: RegExpExecArray | null;
  while ((block = blockPattern.exec(html)) !== null) {
    const found = findVideoObject(parseJson(block[1]));
    if (found) return found;
  }
  return undefined;
}

/** Searches a parsed ld+json value for the VideoObject node. */
function findVideoObject(value: any): VideoObject | undefined {
  if (!value || typeof value !== "object") return undefined;
  if (Array.isArray(value)) {
    for (const entry of value) {
      const found = findVideoObject(entry);
      if (found) return found;
    }
    return undefined;
  }
  if (value["@type"] === "VideoObject") return value as VideoObject;
  return findVideoObject(value["@graph"]);
}

/** Pornhub's `flashvars_<id>` player config: title, duration, poster and mediaDefinitions. */
export function extractFlashvars(html: string): any | undefined {
  const line = lineAfter(html, "var flashvars_");
  if (line === undefined) return undefined;
  return parseJson(sliceBetween(line, "{", "}"));
}

/** The `mediaDefinition: [...]` player array that RedTube, YouPorn and Tube8 inline. */
export function extractMediaDefinitions(html: string): any[] {
  const line = lineAfter(html, "mediaDefinition:");
  if (line === undefined) return [];
  const parsed = parseJson(sliceBetween(line, "[", "]"));
  if (Array.isArray(parsed)) return parsed;
  return [];
}

/** The rest of the line on which `marker` first occurs (the sites emit each config on one line). */
function lineAfter(html: string, marker: string): string | undefined {
  const start = html.indexOf(marker);
  if (start === -1) return undefined;
  let end = html.indexOf("\n", start);
  if (end === -1) end = html.length;
  return html.slice(start, end);
}

/** Substring from the first `open` to the last `close` of a line. */
function sliceBetween(line: string, open: string, close: string): string {
  const start = line.indexOf(open);
  const end = line.lastIndexOf(close);
  if (start === -1 || end <= start) return "";
  return line.slice(start, end + 1);
}

function parseJson(text: string): any {
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}
