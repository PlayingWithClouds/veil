// How the story reader shows text: theme, typeface, size, line spacing and
// column width. Persisted to localStorage, so every story opens the same way.
import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export type ReaderTheme = 'dark' | 'black' | 'sepia' | 'light';
export type ReaderFont = 'serif' | 'classic' | 'sans' | 'readable' | 'mono';
export type ReaderWidth = 'narrow' | 'medium' | 'wide';

export interface ReaderSettings {
	theme: ReaderTheme;
	font: ReaderFont;
	// Text size in px.
	fontSize: number;
	lineHeight: number;
	width: ReaderWidth;
}

/** The themes in the order the settings offer them, with their names. */
export const READER_THEMES: { value: ReaderTheme; label: string }[] = [
	{ value: 'dark', label: 'Dark' },
	{ value: 'black', label: 'Black' },
	{ value: 'sepia', label: 'Sepia' },
	{ value: 'light', label: 'Light' }
];

/** The typefaces, each a font stack of fonts commonly installed. */
export const READER_FONTS: { value: ReaderFont; label: string; stack: string }[] = [
	{
		value: 'serif',
		label: 'Serif',
		stack: "'Literata', 'Iowan Old Style', 'Charter', 'Source Serif Pro', Georgia, serif"
	},
	{
		value: 'classic',
		label: 'Classic',
		stack: "'Palatino Linotype', Palatino, 'Book Antiqua', 'URW Palladio L', serif"
	},
	{ value: 'sans', label: 'Sans', stack: 'var(--font-sans)' },
	{
		value: 'readable',
		label: 'Legible',
		stack: "'Atkinson Hyperlegible', Verdana, 'DejaVu Sans', sans-serif"
	},
	{ value: 'mono', label: 'Mono', stack: 'var(--font-mono)' }
];

/** Column widths in em of the text size: about 60, 70 and 90 characters a line. */
export const READER_WIDTHS: { value: ReaderWidth; label: string; ems: number }[] = [
	{ value: 'narrow', label: 'Narrow', ems: 32 },
	{ value: 'medium', label: 'Medium', ems: 38 },
	{ value: 'wide', label: 'Wide', ems: 48 }
];

export const FONT_SIZE_RANGE = { min: 14, max: 32 };
export const LINE_HEIGHT_RANGE = { min: 1.3, max: 2.2 };

const DEFAULTS: ReaderSettings = {
	theme: 'dark',
	font: 'serif',
	fontSize: 19,
	lineHeight: 1.7,
	width: 'medium'
};

const READER_STORAGE_KEY = 'veil:reader';

/** The settings saved last, with defaults for anything missing. */
function loadSettings(): ReaderSettings {
	if (!browser) return DEFAULTS;
	const stored = localStorage.getItem(READER_STORAGE_KEY);
	if (!stored) return DEFAULTS;
	try {
		return { ...DEFAULTS, ...(JSON.parse(stored) as Partial<ReaderSettings>) };
	} catch {
		return DEFAULTS;
	}
}

export const readerSettings = writable<ReaderSettings>(loadSettings());

readerSettings.subscribe((settings) => {
	if (!browser) return;
	localStorage.setItem(READER_STORAGE_KEY, JSON.stringify(settings));
});

/** The font stack of a typeface. */
export function fontStack(font: ReaderFont): string {
	const found = READER_FONTS.find((entry) => entry.value === font);
	if (!found) return READER_FONTS[0].stack;
	return found.stack;
}

/** The column width of a width setting, in em. */
export function columnEms(width: ReaderWidth): number {
	const found = READER_WIDTHS.find((entry) => entry.value === width);
	if (!found) return READER_WIDTHS[1].ems;
	return found.ems;
}
