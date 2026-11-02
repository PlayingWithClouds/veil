import { describe, expect, test } from 'bun:test';
import {
	countryNames,
	joinMeasurements,
	mergeSuggestions,
	normalizeCountry,
	splitMeasurements,
	yearsBetween
} from '../src/lib/performerPresets';

describe('performer presets', () => {
	test('measurements split into parts and join back', () => {
		expect(splitMeasurements('34DD-26-38')).toEqual({ bust: '34DD', waist: '26', hips: '38' });
		expect(splitMeasurements('34DD')).toBeNull();
		expect(splitMeasurements(null)).toEqual({ bust: '', waist: '', hips: '' });
		expect(joinMeasurements({ bust: '34C ', waist: '24', hips: '34' })).toBe('34C-24-34');
		expect(joinMeasurements({ bust: '', waist: '', hips: '' })).toBe('');
	});

	test('countries are named and short spellings normalised', () => {
		expect(countryNames()).toContain('United States');
		expect(normalizeCountry(' USA')).toBe('United States');
		expect(normalizeCountry('Czechia')).toBe('Czechia');
	});

	test('library values follow the presets without repeating them', () => {
		expect(mergeSuggestions(['Blonde', 'Red'], ['blonde', 'Platinum'])).toEqual(['Blonde', 'Red', 'Platinum']);
	});

	test('years count whole years', () => {
		expect(yearsBetween('1992-07-01', '2026-06-30')).toBe(33);
		expect(yearsBetween('1992-07-01', '2026-07-01')).toBe(34);
		expect(yearsBetween('not a date')).toBeNull();
	});
});
