// Where search results go as the sites answer, so the screen never
// reshuffles: results are only ever appended. The top places go to each
// site's best few, in the order the sites answer; the rest waits below the
// fold and is let in sites taking turns. Sites the user picks from more
// often get more of the top and an earlier turn.

/** How many places the top of the results has. */
export const TOP_PLACES = 20;

/** How many of the top places one site may take, without a preference. */
export const TOP_PER_SITE = 4;

/**
 * Top places by preference score (the site's pick rate over the overall
 * one): the first bound a score reaches decides.
 */
const TOP_PLACES_BY_SCORE: [number, number][] = [
	[1.5, 6],
	[1.15, 5],
	[0.85, TOP_PER_SITE],
	[0.6, 3],
	[0, 2]
];

/** How many top places a site with a preference score may take. */
export function topPlacesFor(score: number | undefined): number {
	if (score === undefined) return TOP_PER_SITE;
	for (const [bound, places] of TOP_PLACES_BY_SCORE) {
		if (score >= bound) return places;
	}
	return TOP_PER_SITE;
}

/** Something with an id: a result. */
export interface Identified {
	id: string;
}

/** The order results appear in, as the sites answer. */
export class SearchLayout<Item extends Identified> {
	private seen = new Set<string>();
	private waiting = new Map<string, Item[]>();
	private topBySite = new Map<string, number>();
	private topCount = 0;
	private topOpen = true;

	/** Takes each site's preference score; sites without one count as average. */
	constructor(private preference: ReadonlyMap<string, number> = new Map()) {}

	/** Whether an item is already on screen or waiting for its place. */
	has(id: string): boolean {
		return this.seen.has(id);
	}

	/** Counts an item shown elsewhere on the page, so results don't repeat it. */
	claim(id: string) {
		this.seen.add(id);
	}

	/**
	 * Takes part of a site's answer and returns what goes on screen now: while
	 * the top is open, up to TOP_PER_SITE of the site's results. The rest
	 * waits for release; filling the top releases it at once.
	 */
	add(site: string, items: Item[]): Item[] {
		const shown: Item[] = [];
		for (const item of items) {
			if (this.seen.has(item.id)) continue;
			this.seen.add(item.id);
			if (this.takesTopPlace(site)) {
				shown.push(item);
				continue;
			}
			this.wait(site, item);
		}
		if (this.topOpen && this.topCount >= TOP_PLACES) {
			return [...shown, ...this.release()];
		}
		return shown;
	}

	/**
	 * Closes the top and returns every waiting result, sites taking turns,
	 * the most preferred first in each turn.
	 */
	release(): Item[] {
		this.topOpen = false;
		const released: Item[] = [];
		const queues = [...this.waiting.entries()]
			.sort(([first], [second]) => this.scoreOf(second) - this.scoreOf(first))
			.map(([, queue]) => queue);
		this.waiting.clear();
		for (let position = 0; queues.some((queue) => position < queue.length); position++) {
			for (const queue of queues) {
				if (position < queue.length) released.push(queue[position]);
			}
		}
		return released;
	}

	/** Whether results are waiting for release. */
	get hasWaiting(): boolean {
		return this.waiting.size > 0;
	}

	/** Whether the site's next result gets a top place, and takes it. */
	private takesTopPlace(site: string): boolean {
		if (!this.topOpen || this.topCount >= TOP_PLACES) return false;
		const taken = this.topBySite.get(site) ?? 0;
		if (taken >= topPlacesFor(this.preference.get(site))) return false;
		this.topBySite.set(site, taken + 1);
		this.topCount++;
		return true;
	}

	/** A site's preference score, 1 (average) without one. */
	private scoreOf(site: string): number {
		const score = this.preference.get(site);
		if (score === undefined) return 1;
		return score;
	}

	/** Queues a result behind the site's others. */
	private wait(site: string, item: Item) {
		const queue = this.waiting.get(site);
		if (queue) {
			queue.push(item);
			return;
		}
		this.waiting.set(site, [item]);
	}
}
