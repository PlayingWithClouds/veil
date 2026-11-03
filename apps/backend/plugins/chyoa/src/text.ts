const NAMED_ENTITIES: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
  lsquo: "‘",
  rsquo: "’",
  ldquo: "“",
  rdquo: "”",
  ndash: "–",
  mdash: "—",
  hellip: "…",
  laquo: "«",
  raquo: "»",
};

/** The character of one entity name (the text between "&" and ";"); the whole match when unknown. */
function decodeEntity(name: string, match: string): string {
  if (/^#x/i.test(name)) return String.fromCodePoint(parseInt(name.slice(2), 16));
  if (name.startsWith("#")) return String.fromCodePoint(parseInt(name.slice(1), 10));
  const decoded = NAMED_ENTITIES[name.toLowerCase()];
  if (decoded === undefined) return match;
  return decoded;
}

/** Decodes named and numeric HTML entities in one pass, so "&amp;lt;" becomes "&lt;". */
export function decodeEntities(text: string): string {
  return text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (match, name: string) => decodeEntity(name, match));
}

/**
 * Turns a chapter's markup into plain text. Paragraphs, headings and rules
 * become blank lines, `<br>` a line break; pictures are dropped. Newlines in
 * the markup are plain whitespace.
 */
export function plainText(markup: string): string {
  const text = markup
    .replace(/<!--[\s\S]*?-->/g, "")
    .replace(/<img[^>]*>/gi, "")
    .replace(/\s+/g, " ")
    .replace(/<br\s*\/?>/gi, "\n")
    .replace(/<hr\s*\/?>|<\/(?:p|h[1-6]|div|li|blockquote)>/gi, "\n\n")
    .replace(/<[^>]+>/g, "");
  return decodeEntities(text)
    .replace(/ *\n */g, "\n")
    .replace(/\n{3,}/g, "\n\n")
    .trim();
}

/** The number of whitespace-separated words in a text. */
export function countWords(text: string): number {
  const words = text.split(/\s+/).filter((word) => word !== "");
  return words.length;
}
