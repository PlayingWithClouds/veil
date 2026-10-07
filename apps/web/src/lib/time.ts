// Human-readable relative times ("3 hours ago", "2 years ago", "in 20 minutes").

const RELATIVE_TIME_UNITS: { unit: Intl.RelativeTimeFormatUnit; seconds: number }[] = [
	{ unit: 'year', seconds: 365 * 86400 },
	{ unit: 'month', seconds: 30 * 86400 },
	{ unit: 'week', seconds: 7 * 86400 },
	{ unit: 'day', seconds: 86400 },
	{ unit: 'hour', seconds: 3600 },
	{ unit: 'minute', seconds: 60 }
];

const relativeTimeFormat = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });

/** Formats an ISO timestamp or date relative to now, using the largest fitting unit. */
export function formatRelativeTime(timestamp: string): string {
	const differenceSeconds = (Date.parse(timestamp) - Date.now()) / 1000;
	for (const { unit, seconds } of RELATIVE_TIME_UNITS) {
		if (Math.abs(differenceSeconds) >= seconds) {
			return relativeTimeFormat.format(Math.round(differenceSeconds / seconds), unit);
		}
	}
	if (differenceSeconds < 0) return 'just now';
	return 'in under a minute';
}
