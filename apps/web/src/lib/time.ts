import dayjs from 'dayjs';

/** Formats a duration as a clock: "9:05", "1:02:03"; null for none. */
export function formatClock(seconds: number | null | undefined): string | null {
	if (!seconds || seconds <= 0) return null;
	const length = dayjs.duration(Math.round(seconds), 'seconds');
	if (length.asHours() >= 1) return length.format('H:mm:ss');
	return length.format('m:ss');
}
