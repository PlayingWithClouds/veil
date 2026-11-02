const NAMED_ENTITIES: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
};

/** Decodes named and numeric HTML entities in one pass, so "&amp;lt;" becomes "&lt;". */
function decodeEntities(text: string): string {
  return text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (match, name: string) => decodeEntity(name, match));
}

/** The character of one entity name (the text between "&" and ";"); the whole match when unknown. */
function decodeEntity(name: string, match: string): string {
  if (/^#x/i.test(name)) {
    return String.fromCodePoint(parseInt(name.slice(2), 16));
  }
  if (name.startsWith("#")) {
    return String.fromCodePoint(parseInt(name.slice(1), 10));
  }
  const decoded = NAMED_ENTITIES[name.toLowerCase()];
  if (decoded === undefined) {
    return match;
  }
  return decoded;
}

/**
 * Turns one page of a work's `pageText` into plain paragraphs. The API sends
 * CRLF text with a little inline markup (em, strong, sup) and entities.
 */
export function plainPageText(pageText: string): string {
  const text = pageText
    .replace(/\r\n?/g, "\n")
    .replace(/<br\s*\/?>/gi, "\n")
    .replace(/<\/p>/gi, "\n\n")
    .replace(/<[^>]+>/g, "");
  return decodeEntities(text)
    .replace(/[ \t]+\n/g, "\n")
    .replace(/\n{3,}/g, "\n\n")
    .trim();
}

/** Joins the plain text of a work's pages into one text. */
export function joinPages(pages: string[]): string {
  return pages
    .map(plainPageText)
    .filter((page) => page !== "")
    .join("\n\n");
}

/** An API date such as "11/02/2026" as "2026-11-02"; undefined for anything else. */
export function isoDateOf(apiDate: string | undefined): string | undefined {
  if (apiDate === undefined) {
    return undefined;
  }
  const match = apiDate.match(/^(\d{2})\/(\d{2})\/(\d{4})$/);
  if (match === null) {
    return undefined;
  }
  return `${match[3]}-${match[1]}-${match[2]}`;
}
