import { expect, test } from 'bun:test';
import { endlessFeed, siteSource, type RankedPage } from '../src/lib/feed';

/** Items with ids from names. */
function items(...names: string[]) {
	return names.map((id) => ({ id }));
}

/** A ranked source over a fixed list. */
function rankedOver(list: { id: string }[]) {
	return async (limit: number, offset: number): Promise<RankedPage<{ id: string }>> => {
		const page = list.slice(offset, offset + limit);
		return { items: page, entryCount: page.length };
	};
}

/** A rest source over a fixed list. */
function restOver(list: { id: string }[]) {
	return async (limit: number, offset: number) => list.slice(offset, offset + limit);
}

/** Every id the feed gives, page by page, until it gives nothing. */
async function drain(feed: (limit: number) => Promise<{ id: string }[]>, limit: number) {
	const out: string[] = [];
	for (let page = await feed(limit); page.length > 0; page = await feed(limit)) {
		out.push(...page.map((item) => item.id));
	}
	return out;
}

test('the ranked items come first, then the rest, without repeats or left-out ones', async () => {
	const feed = endlessFeed(
		{ ranked: rankedOver(items('a', 'b', 'c')), rest: restOver(items('d', 'b', 'e', 'x', 'f')) },
		['x']
	);
	expect(await drain(feed, 2)).toEqual(['a', 'b', 'c', 'd', 'e', 'f']);
});

test('a ranked page with entries that did not load still advances by its entries', async () => {
	const offsets: number[] = [];
	const feed = endlessFeed(
		{
			ranked: async (limit, offset) => {
				offsets.push(offset);
				if (offset === 0) return { items: items('a'), entryCount: limit };
				return { items: [], entryCount: 0 };
			},
			rest: restOver([])
		},
		[]
	);
	expect(await drain(feed, 3)).toEqual(['a']);
	expect(offsets).toEqual([0, 3]);
});

test('after the library the sites take turns until each runs dry or fails', async () => {
	const catalogs: Record<string, { id: string }[]> = {
		one: items('o1', 'o2', 'o3'),
		two: items('t1', 'a'),
		broken: []
	};
	const sites = siteSource(
		() => ['one', 'two', 'broken'],
		async (plugin, limit, offset) => {
			if (plugin === 'broken') throw new Error('login needed');
			return catalogs[plugin].slice(offset, offset + limit);
		}
	);
	const feed = endlessFeed({ rest: restOver(items('a')), sites }, []);
	expect(await drain(feed, 2)).toEqual(['a', 'o1', 'o2', 't1', 'o3']);
});
