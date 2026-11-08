// Endless feeds: a ranked list that runs out, continued by the rest of the
// library, then by pages fetched from the sites, without repeats.

/** The next items of an endless feed, about limit of them; none once it has nothing more. */
export type Feed<Item> = (limit: number) => Promise<Item[]>;

/** A page of a ranked list, and how many entries it held (some may not load). */
export interface RankedPage<Item> {
	items: Item[];
	entryCount: number;
}

/** The next page from wherever a source is; null once it has nothing more. */
export type Source<Item> = (limit: number) => Promise<Item[] | null>;

/** Where an endless feed takes its items from, in this order. */
export interface FeedSources<Item> {
	// The ranked list, read until a page comes back short.
	ranked?: (limit: number, offset: number) => Promise<RankedPage<Item>>;
	// Everything stored, read until it's through.
	rest: (limit: number, offset: number) => Promise<Item[]>;
	// New items from the sites once the library is through (siteSource).
	sites?: Source<Item>;
}

/**
 * An endless feed: the ranked items, then the rest of the library, then what
 * the sites add. Nothing shows twice, and the ids in leaveOut (shown
 * elsewhere on the page) not at all.
 */
export function endlessFeed<Item extends { id: string }>(
	sources: FeedSources<Item>,
	leaveOut: string[] = []
): Feed<Item> {
	const phases: Source<Item>[] = [];
	if (sources.ranked) phases.push(rankedSource(sources.ranked));
	phases.push(offsetSource(sources.rest));
	if (sources.sites) phases.push(sources.sites);
	return chainFeed(phases, leaveOut);
}

/**
 * A feed reading each source until it's through, then the next. Nothing
 * shows twice, and the ids in leaveOut not at all.
 */
export function chainFeed<Item extends { id: string }>(
	phases: Source<Item>[],
	leaveOut: string[] = []
): Feed<Item> {
	const seen = new Set(leaveOut);

	/** The next batch from the first phase that still has items; null once all are through. */
	async function nextBatch(limit: number): Promise<Item[] | null> {
		while (phases.length > 0) {
			const batch = await phases[0](limit);
			if (batch !== null) return batch;
			phases.shift();
		}
		return null;
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

/** A ranked list as a source: through once a page comes back short. */
function rankedSource<Item>(
	ranked: (limit: number, offset: number) => Promise<RankedPage<Item>>
): Source<Item> {
	let offset = 0;
	let through = false;
	return async (limit) => {
		if (through) return null;
		const page = await ranked(limit, offset);
		offset += page.entryCount;
		if (page.entryCount < limit) through = true;
		return page.items;
	};
}

/** A list fetched once as a source: all of it the first time, through after. */
export function onceSource<Item>(list: () => Promise<Item[]>): Source<Item> {
	let read = false;
	return async () => {
		if (read) return null;
		read = true;
		return list();
	};
}

/** An offset-paged list as a source: through once a page comes back empty. */
export function offsetSource<Item>(
	list: (limit: number, offset: number) => Promise<Item[]>
): Source<Item> {
	let offset = 0;
	return async (limit) => {
		const page = await list(limit, offset);
		if (page.length === 0) return null;
		offset += page.length;
		return page;
	};
}

/**
 * Pages from the sites, the sites taking turns, each paged by its own
 * offset. The sites are asked for when first needed (the installed plugins
 * load after the page). A site that returns nothing or fails drops out (a
 * failure is logged); the source is through once every site has.
 */
export function siteSource<Item>(
	plugins: () => string[],
	fetchPage: (plugin: string, limit: number, offset: number) => Promise<Item[]>
): Source<Item> {
	let remaining: string[] | null = null;
	const offsets = new Map<string, number>();
	let turn = 0;
	return async (limit) => {
		if (remaining === null) remaining = [...plugins()];
		while (remaining.length > 0) {
			turn = turn % remaining.length;
			const plugin = remaining[turn];
			const offset = offsets.get(plugin) ?? 0;
			const page = await fetchPage(plugin, limit, offset).catch((failure) => {
				console.warn(`${plugin} gave no more`, failure);
				return [];
			});
			if (page.length === 0) {
				remaining.splice(turn, 1);
				continue;
			}
			offsets.set(plugin, offset + page.length);
			turn++;
			return page;
		}
		return null;
	};
}
