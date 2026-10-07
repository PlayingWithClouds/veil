export const BASE_URL = "https://hqporner.com";

export function resolveUrl(path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  return BASE_URL + (path.startsWith("/") ? path : "/" + path);
}

// Video pages look like /hdporn/126872-the_little_hole_that_can.html.
// The leading numeric id is the stable per-video identifier.
export function videoIdFromUrl(url: string): string {
  const match = resolveUrl(url).match(/\/hdporn\/(\d+)-/);
  return match ? match[1] : "";
}

export function isVideoUrl(url: string): boolean {
  return /\/hdporn\/\d+-.*\.html/.test(resolveUrl(url));
}

// Actress pages look like /actress/little-caprice (optionally /<page>).
export function actressSlugFromUrl(url: string): string {
  const match = resolveUrl(url).match(/\/actress\/([a-z0-9-]+)/i);
  return match ? match[1] : "";
}

export function isActressUrl(url: string): boolean {
  return actressSlugFromUrl(url) !== "";
}

// Category pages look like /category/anal-sex-hd (optionally /<page>).
export function categorySlugFromUrl(url: string): string {
  const match = resolveUrl(url).match(/\/category\/([a-z0-9-]+)/i);
  return match ? match[1] : "";
}

export function isCategoryUrl(url: string): boolean {
  return categorySlugFromUrl(url) !== "";
}

// Durations render as "33m 15s", "1h 33m 16s" or "59m 1s".
export function parseDurationSeconds(text: string): number {
  const hours = text.match(/(\d+)\s*h/i);
  const minutes = text.match(/(\d+)\s*m/i);
  const seconds = text.match(/(\d+)\s*s/i);
  let total = 0;
  if (hours) total += parseInt(hours[1], 10) * 3600;
  if (minutes) total += parseInt(minutes[1], 10) * 60;
  if (seconds) total += parseInt(seconds[1], 10);
  return total;
}

const DAY_MS = 24 * 60 * 60 * 1000;

const RELATIVE_UNIT_DAYS: Record<string, number> = {
  day: 1,
  week: 7,
  month: 30,
  year: 365,
};

// The site only shows relative publish dates ("today", "3 days ago",
// "2 months ago"). Approximate them as an ISO date; empty when unparseable.
export function relativeDateToISO(text: string, now = new Date()): string {
  const normalized = text.trim().toLowerCase();
  if (!normalized) return "";
  if (normalized === "today") return isoDate(now);
  if (normalized === "yesterday") return isoDate(new Date(now.getTime() - DAY_MS));

  const match = normalized.match(/(\d+)\s*(day|week|month|year)s?\s*ago/);
  if (!match) return "";
  const amount = parseInt(match[1], 10);
  const days = amount * RELATIVE_UNIT_DAYS[match[2]];
  return isoDate(new Date(now.getTime() - days * DAY_MS));
}

function isoDate(date: Date): string {
  return date.toISOString().slice(0, 10);
}

export function providerFromUrl(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "").split(".")[0];
  } catch {
    return "unknown";
  }
}
