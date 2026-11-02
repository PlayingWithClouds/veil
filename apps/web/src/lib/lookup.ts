// "Who is this?": asks every identification plugin about a picture. What they
// name (performers, videos) is stored, so lookups are mutations.
import { gqlClient } from '$lib/veil';
import { VIDEO_CARD, MEMBER, type VideoCardRow, type MemberRow } from '$lib/library';

/** Why a plugin named something: a page, a picture, a detail like "94% face match". */
export interface EvidenceRow {
	plugin: string;
	url: string | null;
	title: string | null;
	image: { uri: string } | null;
	detail: string | null;
}

/** What a lookup found, sorted by kind. */
export interface LookupFound {
	performers: { performer: MemberRow; evidence: EvidenceRow[] }[];
	videos: { video: VideoCardRow; evidence: EvidenceRow[] }[];
	// Matches the library has no record for: just the plugins' evidence.
	elsewhere: EvidenceRow[];
	failures: string[];
}

const EVIDENCE = { plugin: true, url: true, title: true, image: { uri: true }, detail: true } as const;

/** Asks the identification plugins who or what is in a picture. */
export async function lookUpPicture(picture: File): Promise<LookupFound> {
	const data = await gqlClient.mutation({
		lookup: {
			__args: { input: { imageBase64: await base64Of(picture), mimeType: picture.type } },
			matches: {
				record: {
					on_Performer: { __typename: true, ...MEMBER },
					on_Video: { __typename: true, ...VIDEO_CARD }
				},
				evidence: EVIDENCE
			},
			failures: true
		}
	});
	const found: LookupFound = { performers: [], videos: [], elsewhere: [], failures: data.lookup.failures };
	for (const match of data.lookup.matches) {
		if (match.record?.__typename === 'Performer') {
			found.performers.push({ performer: match.record, evidence: match.evidence });
			continue;
		}
		if (match.record?.__typename === 'Video') {
			found.videos.push({ video: match.record, evidence: match.evidence });
			continue;
		}
		found.elsewhere.push(...match.evidence);
	}
	return found;
}

/** A file's contents as base64, without the data-URL prefix. */
function base64Of(file: File): Promise<string> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onload = () => resolve(String(reader.result).replace(/^data:[^,]*,/, ''));
		reader.onerror = () => reject(reader.error);
		reader.readAsDataURL(file);
	});
}
