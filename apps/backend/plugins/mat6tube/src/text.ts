const COUNT_MULTIPLIERS: Record<string, number> = { "": 1, k: 1_000, m: 1_000_000 };

/** Seconds of a "40:51" or "1:02:03" clock; 0 when the text is none. */
export function clockSeconds(text: string): number {
  const match = text.trim().match(/^(?:(\d+):)?(\d{1,2}):(\d{2})$/);
  if (!match) return 0;
  let hours = 0;
  if (match[1] !== undefined) hours = parseInt(match[1], 10);
  return hours * 3600 + parseInt(match[2], 10) * 60 + parseInt(match[3], 10);
}

/** A view count like "14.81M", "8.22K" or "1,234" as a whole number; 0 when unreadable. */
export function countText(text: string): number {
  const match = text.trim().match(/^(\d[\d.,]*)\s*([kKmM]?)$/);
  if (!match) return 0;
  if (match[2] === "") return parseInt(match[1].replace(/[.,]/g, ""), 10);
  const multiplier = COUNT_MULTIPLIERS[match[2].toLowerCase()];
  return Math.round(parseFloat(match[1].replace(",", ".")) * multiplier);
}

/** Decodes the HTML entities left in attribute values (the parser keeps them raw). */
export function decodeEntities(text: string): string {
  return text
    .replace(/&#(\d+);/g, (match, code: string) => String.fromCodePoint(parseInt(code, 10)))
    .replace(/&quot;/g, '"')
    .replace(/&#39;|&apos;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&amp;/g, "&");
}

/** Whitespace runs collapsed and the ends trimmed. */
export function cleanText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}

/** An element's attribute value, or "" when it has none. */
export function attributeOf(element: { getAttribute(name: string): string | undefined }, name: string): string {
  const value = element.getAttribute(name);
  if (value === undefined) return "";
  return value;
}
