// Endless feeds: a ranked list that runs out, continued by the rest of the
// library, without repeats.

/** The next items of an endless feed, about limit of them; none once it has nothing more. */
export type Feed<Item> = (limit: number) => Promise<Item[]>;

/** A page of a ranked list, and how many entries it held (some may not load). */
export interface RankedPage<Item> {
	items: Item[];
	entryCount: number;
}

/** Where an endless feed takes its items from. */
export interface FeedSources<Item> {
	// The ranked list, read until a page comes back short.
	ranked: (limit: number, offset: number) => Promise<RankedPage<Item>>;
	// Everything else, read until it's through.
	rest: (limit: number, offset: number) => Promise<Item[]>;
}

/**
 * An endless feed: the ranked items, then, once those run out, the rest.
 * Nothing shows twice, and the ids in leaveOut (shown elsewhere on the page)
 * not at all.
 */
export function endlessFeed<Item extends { id: string }>(
	sources: FeedSources<Item>,
	leaveOut: string[] = []
): Feed<Item> {
	const seen = new Set(leaveOut);
	let rankedOffset = 0;
	// Where the rest continues; null while the ranked list lasts.
	let restOffset: number | null = null;

	/** The next batch from wherever the feed is; null once the rest is through. */
	async function nextBatch(limit: number): Promise<Item[] | null> {
		if (restOffset === null) {
			const page = await sources.ranked(limit, rankedOffset);
			rankedOffset += page.entryCount;
			if (page.entryCount < limit) restOffset = 0;
			return page.items;
		}
		const page = await sources.rest(limit, restOffset);
		if (page.length === 0) return null;
		restOffset += page.length;
		return page;
	}

	return async (limit) => {
		const fresh: Item[] = [];
		while (fresh.length < limit) {
			const batch = await nextBatch(limit);
			if (batch === null) break;
			for (const item of batch) {
				if (seen.has(item.id)) continue;
				seen.add(item.id);
				fresh.push(item);
			}
		}
		return fresh;
	};
}
