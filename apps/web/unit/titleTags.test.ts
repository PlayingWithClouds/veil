import { expect, test } from 'bun:test';
import { splitTitle } from '../src/lib/titleTags';

test('audience tags become badges and other bracketed tags leave the title', () => {
	expect(splitTitle('[F4M] Your Sexy Coworker Begs You [Whispers][Office]')).toEqual({
		title: 'Your Sexy Coworker Begs You',
		audiences: ['F4M'],
		tags: ['Whispers', 'Office']
	});
	expect(splitTitle('[f4a] [ASMR] - Live Karaoke').title).toBe('Live Karaoke');
	expect(splitTitle('[FF 4 M] Twins').audiences).toEqual(['FF4M']);
});

test('a title without brackets stays, and so does one made only of tags', () => {
	expect(splitTitle('Morning Facesitting')).toEqual({
		title: 'Morning Facesitting',
		audiences: [],
		tags: []
	});
	expect(splitTitle('[F4M]').title).toBe('[F4M]');
});
