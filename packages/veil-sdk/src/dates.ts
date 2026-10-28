// Sites often show only relative times ("3 weeks ago", "2 hours ago",
// "yesterday"). relativeDate turns them into an ISO date-time, so a stored
// date doesn't drift as "3 weeks ago" would.

const SECOND_MS = 1000;

const UNIT_MS: Record<string, number> = {
  second: SECOND_MS,
  sec: SECOND_MS,
  minute: 60 * SECOND_MS,
  min: 60 * SECOND_MS,
  hour: 3600 * SECOND_MS,
  hr: 3600 * SECOND_MS,
  day: 86400 * SECOND_MS,
  week: 7 * 86400 * SECOND_MS,
  month: 30 * 86400 * SECOND_MS,
  year: 365 * 86400 * SECOND_MS,
};

/**
 * The moment a relative English time names ("3 weeks ago", "an hour ago",
 * "just now", "yesterday") as an ISO 8601 date-time, approximated from now;
 * undefined when the text isn't one.
 */
export function relativeDate(text: string, now: Date = new Date()): string | undefined {
  const normalized = text.trim().toLowerCase();
  if (normalized === "") return undefined;
  if (normalized === "just now" || normalized === "now" || normalized === "today") return now.toISOString();
  if (normalized === "yesterday") return new Date(now.getTime() - UNIT_MS.day).toISOString();

  const match = normalized.match(/\b(\d+|an?|one)\s*(second|sec|minute|min|hour|hr|day|week|month|year)s?\s+ago\b/);
  if (!match) return undefined;
  let amount = 1;
  if (/^\d+$/.test(match[1])) amount = parseInt(match[1], 10);
  return new Date(now.getTime() - amount * UNIT_MS[match[2]]).toISOString();
}
