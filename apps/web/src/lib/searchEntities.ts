// What the search page shows about the performers and studios a query names:
// picture, name and how many scenes the library has with them.
import { gqlClient } from '$lib/veil';
import type { SearchMatchedEntity } from '$lib/search';

/** A recognised performer or studio, ready for the entity strip. */
export interface MatchedEntityCard {
	kind: 'performer' | 'studio';
	id: string;
	name: string;
	imagePath: string | null;
	sceneCount: number;
	// The query only came close to the name (a likely typo).
	fuzzy: boolean;
}

/** How many recognised performers and studios get a card. */
const MAX_ENTITY_CARDS = 2;

/** Loads the cards of the first recognised performers and studios; tags get none. */
export async function fetchMatchedEntityCards(
	entities: SearchMatchedEntity[]
): Promise<MatchedEntityCard[]> {
	const wanted = entities.filter((entity) => entity.kind !== 'tag').slice(0, MAX_ENTITY_CARDS);
	const cards = await Promise.all(wanted.map(fetchEntityCard));
	return cards.filter((card): card is MatchedEntityCard => card !== null);
}

/** Loads one performer's or studio's card, null when it can't be read. */
async function fetchEntityCard(entity: SearchMatchedEntity): Promise<MatchedEntityCard | null> {
	try {
		if (entity.kind === 'performer') {
			const data = await gqlClient.query({
				performer: {
					__args: { id: entity.id },
					id: true,
					name: true,
					imagePath: true,
					sceneCount: true
				}
			});
			if (!data.performer) return null;
			return { kind: 'performer', fuzzy: entity.fuzzy, ...data.performer };
		}
		const data = await gqlClient.query({
			studio: { __args: { id: entity.id }, id: true, name: true, imagePath: true, sceneCount: true }
		});
		if (!data.studio) return null;
		return { kind: 'studio', fuzzy: entity.fuzzy, ...data.studio };
	} catch {
		return null;
	}
}
