// Display formatting shared by the scene watch page components.

const RELATIVE_UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
	['year', 365 * 24 * 3600],
	['month', 30 * 24 * 3600],
	['week', 7 * 24 * 3600],
	['day', 24 * 3600],
	['hour', 3600],
	['minute', 60]
];

/** Parses a scene/listing date ("2024-08-04" or ISO timestamp), null when invalid. */
function parseDate(value: string | null | undefined): Date | null {
	if (!value) return null;
	const parsed = new Date(value);
	if (Number.isNaN(parsed.getTime())) return null;
	return parsed;
}

/** "3 months ago"-style label for a date, falling back to the raw value. */
export function formatRelativeDate(value: string | null | undefined): string | null {
	const parsed = parseDate(value);
	if (!parsed) return value ?? null;
	const secondsAgo = (Date.now() - parsed.getTime()) / 1000;
	const formatter = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });
	for (const [unit, unitSeconds] of RELATIVE_UNITS) {
		if (Math.abs(secondsAgo) >= unitSeconds) {
			return formatter.format(-Math.round(secondsAgo / unitSeconds), unit);
		}
	}
	return 'just now';
}

/** Absolute, locale-formatted date for hover titles. */
export function formatAbsoluteDate(value: string | null | undefined): string | undefined {
	const parsed = parseDate(value);
	if (!parsed) return undefined;
	return parsed.toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' });
}

/** Runtime like "37 min" or "1 h 12 min" ("12m" read like months next to a date). */
export function formatDuration(seconds: number | null | undefined): string | null {
	if (!seconds || seconds <= 0) return null;
	const minutes = Math.max(1, Math.round(seconds / 60));
	if (minutes < 60) return `${minutes} min`;
	const hours = Math.floor(minutes / 60);
	return `${hours} h ${minutes % 60} min`;
}

/** Whole years from a birthdate to a date (now when absent), null when either is unknown. */
export function ageAt(
	birthdate: string | null | undefined,
	at: string | null | undefined
): number | null {
	const born = parseDate(birthdate);
	if (!born) return null;
	let reference = new Date();
	if (at) {
		const parsed = parseDate(at);
		if (!parsed) return null;
		reference = parsed;
	}
	let years = reference.getFullYear() - born.getFullYear();
	const birthdayPending =
		reference.getMonth() < born.getMonth() ||
		(reference.getMonth() === born.getMonth() && reference.getDate() < born.getDate());
	if (birthdayPending) years -= 1;
	if (years < 18 || years > 99) return null;
	return years;
}

/** Up to two initials of a name, for avatars without a photo. */
export function initials(name: string): string {
	const words = name.trim().split(/\s+/).filter(Boolean);
	return words
		.slice(0, 2)
		.map((word) => word.charAt(0).toUpperCase())
		.join('');
}

/** Compact count like "1.2K" for view and scene counts. */
export function formatCount(count: number): string {
	return new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 1 }).format(
		count
	);
}

/** Whether a plugin icon value is a loadable image URL (some plugins use an emoji). */
export function isImageUrl(value: string | null | undefined): value is string {
	if (!value) return false;
	return value.startsWith('http://') || value.startsWith('https://') || value.startsWith('/');
}
