/**
 * Seconds in a runtime such as "13:40" or "1:02:03"; zero when the text is not a clock.
 */
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

/** A view count such as "685.92K" or "1.2M" as a whole number; undefined when it isn't one. */
export function countFromText(text: string): number | undefined {
  const match = text.trim().match(/^(\d+(?:\.\d+)?)\s*([KMB]?)$/i);
  if (!match) return undefined;
  const multipliers: Record<string, number> = { "": 1, K: 1_000, M: 1_000_000, B: 1_000_000_000 };
  const multiplier = multipliers[match[2].toUpperCase()];
  return Math.round(Number(match[1]) * multiplier);
}
