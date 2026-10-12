/** Seconds in a runtime such as "47:18" or "1:02:03"; zero when the text is not a clock. */
export function clockSeconds(text: string): number {
  const parts = text.trim().split(":");
  if (parts.length < 2 || parts.length > 3) return 0;
  let seconds = 0;
  for (const part of parts) {
    if (!/^\d+$/.test(part)) return 0;
    seconds = seconds * 60 + Number(part);
  }
  return seconds;
}

/** Seconds in an ISO 8601 duration such as "PT47M18S"; zero when it is not one. */
export function isoDurationSeconds(text: string): number {
  const match = text.match(/^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$/);
  if (!match) return 0;
  return unitValue(match[1], 3600) + unitValue(match[2], 60) + unitValue(match[3], 1);
}

/** A captured number of one duration unit, scaled to seconds; zero when the unit is absent. */
function unitValue(captured: string | undefined, secondsPerUnit: number): number {
  if (captured === undefined) return 0;
  return Number(captured) * secondsPerUnit;
}

/** A count such as "428", "12.5K" or "1.2M" as a whole number; undefined when it isn't one. */
export function countFromText(text: string): number | undefined {
  const match = text.trim().match(/^(\d+(?:\.\d+)?)\s*([KMB]?)$/i);
  if (!match) return undefined;
  const multipliers: Record<string, number> = { "": 1, K: 1_000, M: 1_000_000, B: 1_000_000_000 };
  const multiplier = multipliers[match[2].toUpperCase()];
  return Math.round(Number(match[1]) * multiplier);
}

/** The view count of a post's "2 days ago · 428 views" line; undefined when the line has none. */
export function viewsFromText(text: string): number | undefined {
  const match = text.match(/([\d][\d.,]*\s*[KMB]?)\s*views/i);
  if (!match) return undefined;
  return countFromText(match[1].replace(/,/g, ""));
}
