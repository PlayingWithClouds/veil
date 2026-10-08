import type { HTMLElement } from "node-html-parser";

/** Elements inside an entity link that hold a count, not part of the name. */
const COUNTER_SELECTOR =
  '[class*="count"], [class*="num"], [class*="total"], [class*="views"], [class*="amount"], [class*="qty"]';

/** Text with whitespace runs collapsed and the ends trimmed. */
export function cleanText(text: string): string {
  return decodeEntities(text).replace(/\s+/g, " ").trim();
}

/**
 * The name an entity link (model, tag, channel) shows: its title attribute
 * when set, else a name-classed element inside it, else its text minus nested
 * counters ("Adult Time<span>1348</span>").
 */
export function entityName(anchor: HTMLElement): string {
  const title = cleanText(anchor.getAttribute("title") ?? "");
  if (title) return title;
  const named = cleanText(anchor.querySelector('[class*="name"]')?.text ?? "");
  if (named) return named;

  const counters = [
    ...anchor.querySelectorAll(COUNTER_SELECTOR),
    ...anchor.querySelectorAll("*").filter((element) => /^\s*[\d.,]+\s*[kKmM]?\s*$/.test(element.text)),
  ];
  let text = anchor.text;
  for (const counter of counters) {
    const counterText = counter.text;
    if (counterText) text = text.replace(counterText, " ");
  }
  const name = cleanText(text).replace(/^[@#]\s*/, "");
  // A name needs a letter or digit; separators like "-" are no tag.
  if (!/[A-Za-z0-9\u00C0-\uFFFF]/.test(name)) return "";
  return name;
}

/**
 * Seconds of a runtime badge: "14:16", "1:02:03", "5m:59s" or
 * "1h 2min 55sec". 0 when the text holds none.
 */
export function parseDurationText(text: string): number {
  const clock = text.match(/\b(?:(\d{1,2}):)?(\d{1,3}):(\d{2})\b/);
  if (clock) {
    let hours = 0;
    if (clock[1]) hours = parseInt(clock[1], 10);
    return hours * 3600 + parseInt(clock[2], 10) * 60 + parseInt(clock[3], 10);
  }
  const units = text.match(/\b(?:(\d{1,2})\s*h(?:ours?|rs?)?\s*:?\s*)?(\d{1,3})\s*m(?:in|ins|inutes?)?\s*:?\s*(?:(\d{1,2})\s*s(?:ec|ecs|econds?)?)?\b/i);
  if (units && (units[1] || units[3])) {
    let seconds = parseInt(units[2], 10) * 60;
    if (units[1]) seconds += parseInt(units[1], 10) * 3600;
    if (units[3]) seconds += parseInt(units[3], 10);
    return seconds;
  }
  return 0;
}

/** Seconds of an ISO 8601 duration ("PT12M01S", "PT0H7M17S", "T8M01S"); 0 when unparseable. */
export function parseIsoDuration(text: string): number {
  const match = text.trim().match(/^P?(?:\d+D)?T?(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$/i);
  if (!match || !(match[1] || match[2] || match[3])) return 0;
  let seconds = 0;
  if (match[1]) seconds += parseInt(match[1], 10) * 3600;
  if (match[2]) seconds += parseInt(match[2], 10) * 60;
  if (match[3]) seconds += parseInt(match[3], 10);
  return seconds;
}

/** The YYYY-MM-DD date at the start of an ISO timestamp, or undefined. */
export function isoDate(text: string | undefined): string | undefined {
  const match = (text ?? "").match(/^(\d{4}-\d{2}-\d{2})/);
  if (!match) return undefined;
  return match[1];
}

const NAMED_ENTITIES: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  apos: "'",
  nbsp: " ",
  ndash: "–",
  mdash: "—",
  hellip: "…",
};

/**
 * Decodes HTML entities left in attribute values and script-embedded strings
 * (the HTML parsers already decode text nodes).
 */
export function decodeEntities(text: string): string {
  return text.replace(/&(#x[0-9a-f]+|#\d+|[a-z]+);/gi, (entity, body: string) => {
    if (body.startsWith("#x") || body.startsWith("#X")) return String.fromCodePoint(parseInt(body.slice(2), 16));
    if (body.startsWith("#")) return String.fromCodePoint(parseInt(body.slice(1), 10));
    const named = NAMED_ENTITIES[body.toLowerCase()];
    if (named === undefined) return entity;
    return named;
  });
}
