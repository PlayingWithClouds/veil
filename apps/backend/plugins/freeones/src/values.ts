const MONTH_NAMES = [
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

// Plausible adult-performer ranges. freeones shows "12 cm" and similar for
// unknown heights, which would otherwise be stored as data.
const HEIGHT_RANGE_CM = { min: 100, max: 250 };
const WEIGHT_RANGE_KG = { min: 30, max: 250 };

/** Collapses whitespace runs and trims, so labels and values compare cleanly. */
export function normalizeText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}

/** Reads a date written like "November 14, 1977" as an ISO date; undefined when it isn't one. */
export function isoDateFromEnglish(text: string): string | undefined {
  const match = normalizeText(text).match(/^([A-Za-z]+) (\d{1,2}), (\d{4})$/);
  if (!match) return undefined;
  const monthIndex = MONTH_NAMES.indexOf(match[1].toLowerCase());
  if (monthIndex < 0) return undefined;
  const month = String(monthIndex + 1).padStart(2, "0");
  const day = match[2].padStart(2, "0");
  return `${match[3]}-${month}-${day}`;
}

/** The leading number of a value such as "71 cm", or undefined when the value starts without one. */
export function leadingNumber(text: string): number | undefined {
  const match = normalizeText(text).match(/^(\d+(?:\.\d+)?)/);
  if (!match) return undefined;
  return Number(match[1]);
}

/** The number when it lies within [min, max], otherwise undefined. */
export function numberWithin(value: number | undefined, range: { min: number; max: number }): number | undefined {
  if (value === undefined) return undefined;
  if (value < range.min || value > range.max) return undefined;
  return value;
}

/** A height in centimetres from a value like "171 cm"; implausible values are dropped. */
export function heightCm(text: string): number | undefined {
  return numberWithin(leadingNumber(text), HEIGHT_RANGE_CM);
}

/** A weight in kilograms from a value like "52 kg"; implausible values are dropped. */
export function weightKg(text: string): number | undefined {
  return numberWithin(leadingNumber(text), WEIGHT_RANGE_KG);
}

/**
 * Body measurements as "81E-71-111 cm" from the bust (with cup), waist and hip
 * values, using whichever of them the site lists; undefined when none is.
 */
export function measurementsFrom(bust: string, cup: string, waist: string, hip: string): string | undefined {
  const parts: string[] = [];
  const bustNumber = leadingNumber(bust);
  if (bustNumber !== undefined) parts.push(`${bustNumber}${cup.trim()}`);
  const waistNumber = leadingNumber(waist);
  if (waistNumber !== undefined) parts.push(String(waistNumber));
  const hipNumber = leadingNumber(hip);
  if (hipNumber !== undefined) parts.push(String(hipNumber));
  if (parts.length === 0) return undefined;
  return `${parts.join("-")} cm`;
}

/**
 * The last comma-separated part of a place such as "Cincinnati , OH - Ohio , United States",
 * which is the country.
 */
export function countryFromPlace(place: string): string | undefined {
  const parts = normalizeText(place).split(",");
  const last = normalizeText(parts[parts.length - 1]);
  if (last === "" || last === "N/A") return undefined;
  return last;
}

/** A value that the site writes as "N/A" or leaves empty means nothing. */
export function knownValue(text: string): string | undefined {
  const normalized = normalizeText(text);
  if (normalized === "" || normalized === "N/A") return undefined;
  return normalized;
}
