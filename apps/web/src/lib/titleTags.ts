// Audio and story sites put tags into titles in brackets: "[F4M] Your
// coworker begs you [Whispers][Office]". Cards show the title without them
// and the audience tag ("F4M": who's speaking to whom) as a badge.

/** A title split into its readable text and the tags bracketed around it. */
export interface TitleParts {
	title: string;
	// The audience tags (F4M, M4F, F4A, FF4M, …), as written.
	audiences: string[];
	// The other bracketed tags.
	tags: string[];
}

const BRACKETED = /\[([^\]]{1,40})\]/g;
const AUDIENCE = /^[mfatn]+\s?4\s?[mfatn]+$/i;

/** Splits a title into its text and its bracketed tags; the text stays whole when it's nothing else. */
export function splitTitle(raw: string): TitleParts {
	const audiences: string[] = [];
	const tags: string[] = [];
	for (const match of raw.matchAll(BRACKETED)) {
		const tag = match[1].trim();
		if (AUDIENCE.test(tag)) {
			audiences.push(tag.replaceAll(' ', '').toUpperCase());
			continue;
		}
		tags.push(tag);
	}
	const title = raw
		.replace(BRACKETED, ' ')
		.replace(/\s+/g, ' ')
		.replace(/^[\s\-–|:]+|[\s\-–|:]+$/g, '');
	if (title === '') return { title: raw.trim(), audiences: [], tags: [] };
	return { title, audiences, tags };
}
