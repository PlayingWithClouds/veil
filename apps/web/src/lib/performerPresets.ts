// The common values the performer editor suggests, before what the library
// already holds. Suggestions only: any value can be typed.

export const GENDERS = ['Female', 'Male', 'Trans woman', 'Trans man', 'Non-binary'];
export const ETHNICITIES = ['Caucasian', 'Latina', 'Asian', 'Black', 'Middle Eastern', 'Indian', 'Mixed'];
export const HAIR_COLORS = ['Blonde', 'Brunette', 'Black', 'Red', 'Auburn', 'Brown', 'Gray', 'Dyed'];
export const EYE_COLORS = ['Blue', 'Green', 'Brown', 'Hazel', 'Gray', 'Amber'];
export const ENHANCEMENTS = ['Natural', 'Enhanced'];
export const BODY_ART = ['None', 'Some'];

// ISO 3166 region codes, named by the browser in English.
const REGION_CODES =
	'AD AE AF AG AL AM AO AR AT AU AZ BA BB BD BE BF BG BH BI BJ BN BO BR BS BT BW BY BZ CA CD CF CG CH CI CL CM CN CO CR CU CV CY CZ DE DJ DK DM DO DZ EC EE EG ER ES ET FI FJ FR GA GB GD GE GH GM GN GQ GR GT GW GY HK HN HR HT HU ID IE IL IN IQ IR IS IT JM JO JP KE KG KH KM KN KP KR KW KZ LA LB LC LI LK LR LS LT LU LV LY MA MC MD ME MG MK ML MM MN MO MR MT MU MV MW MX MY MZ NA NE NG NI NL NO NP NZ OM PA PE PG PH PK PL PR PS PT PY QA RO RS RU RW SA SB SC SD SE SG SI SK SL SM SN SO SR SS ST SV SY SZ TD TG TH TJ TL TM TN TO TR TT TW TZ UA UG US UY UZ VA VC VE VN VU WS YE ZA ZM ZW'.split(
		' '
	);

/** Every country by its English name, alphabetically. */
export function countryNames(): string[] {
	const names = new Intl.DisplayNames(['en'], { type: 'region' });
	return REGION_CODES.map((code) => names.of(code) ?? code).sort((first, second) =>
		first.localeCompare(second)
	);
}

// Short spellings sites use for a country, by the name the list gives it.
const COUNTRY_SPELLINGS: Record<string, string> = {
	usa: 'United States',
	us: 'United States',
	'u.s.a.': 'United States',
	'united states of america': 'United States',
	uk: 'United Kingdom',
	'great britain': 'United Kingdom',
	england: 'United Kingdom'
};

/** A typed country in the list's spelling: "USA" becomes "United States". */
export function normalizeCountry(country: string): string {
	const trimmed = country.trim();
	return COUNTRY_SPELLINGS[trimmed.toLowerCase()] ?? trimmed;
}

/** A value with its first letter capitalised: "female" as "Female". */
export function capitalize(value: string): string {
	return value.charAt(0).toUpperCase() + value.slice(1);
}

/**
 * The presets first, then the library's values they don't already cover,
 * ignoring case; all capitalised, since sites spell them either way.
 */
export function mergeSuggestions(presets: string[], libraryValues: string[]): string[] {
	const suggestions: string[] = [];
	const known = new Set<string>();
	for (const value of [...presets, ...libraryValues]) {
		const key = value.trim().toLowerCase();
		if (key === '' || known.has(key)) continue;
		known.add(key);
		suggestions.push(capitalize(value.trim()));
	}
	return suggestions;
}

/** Bust (with cup), waist and hips, as the editor's three inputs hold them. */
export interface Measurements {
	bust: string;
	waist: string;
	hips: string;
}

/** Splits "34C-24-34" into its parts; null when it isn't in that form. */
export function splitMeasurements(measurements: string | null): Measurements | null {
	if (!measurements) return { bust: '', waist: '', hips: '' };
	const parts = measurements.split('-').map((part) => part.trim());
	if (parts.length !== 3) return null;
	return { bust: parts[0], waist: parts[1], hips: parts[2] };
}

/** "34C-24-34" from its parts; empty when every part is. */
export function joinMeasurements(measurements: Measurements): string {
	const parts = [measurements.bust, measurements.waist, measurements.hips].map((part) =>
		part.trim()
	);
	if (parts.every((part) => part === '')) return '';
	return parts.join('-');
}

/** Whole years between two dates, or from a date until now. */
export function yearsBetween(from: string, until: string | null = null): number | null {
	const start = new Date(from);
	const end = until ? new Date(until) : new Date();
	if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return null;
	let years = end.getFullYear() - start.getFullYear();
	const beforeAnniversary =
		end.getMonth() < start.getMonth() ||
		(end.getMonth() === start.getMonth() && end.getDate() < start.getDate());
	if (beforeAnniversary) years--;
	return years;
}
