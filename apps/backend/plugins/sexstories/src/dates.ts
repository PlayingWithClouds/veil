import { relativeDate } from "@playingwithclouds/veil-sdk";

const MONTHS = [
  "january",
  "february",
  "march",
  "april",
  "may",
  "june",
  "july",
  "august",
  "september",
  "october",
  "november",
  "december",
];

// What follows "Posted" is a date or a relative time, both short; the text
// after it (genres, report links) must not be searched.
const POSTED_WINDOW = 32;

/** "Thu 22nd of October 2026" as "2026-10-22"; undefined for any other text. */
function absoluteDate(text: string): string | undefined {
  const match = text.match(/(\d{1,2})(?:st|nd|rd|th)\s+of\s+([a-z]+)\s+(\d{4})/i);
  if (!match) return undefined;
  const month = MONTHS.indexOf(match[2].toLowerCase());
  if (month < 0) return undefined;
  const day = match[1].padStart(2, "0");
  const monthNumber = String(month + 1).padStart(2, "0");
  return `${match[3]}-${monthNumber}-${day}`;
}

/**
 * The date a card or story page gives after "Posted", as "YYYY-MM-DD". Recent
 * stories read "3 days ago" instead of the date; those are approximated.
 */
export function postedDate(text: string): string | undefined {
  const index = text.indexOf("Posted");
  if (index < 0) return undefined;
  const posted = text.slice(index + "Posted".length, index + "Posted".length + POSTED_WINDOW);
  const absolute = absoluteDate(posted);
  if (absolute !== undefined) return absolute;
  const relative = relativeDate(posted);
  if (relative === undefined) return undefined;
  return relative.slice(0, 10);
}
