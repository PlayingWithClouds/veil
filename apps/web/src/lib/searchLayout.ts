// Where search results go as the sites answer, so the screen never
// reshuffles: results are only ever appended. The top places go to each
// site's best few, in the order the sites answer; the rest waits below the
// fold and is let in sites taking turns.

/** How many places the top of the results has. */
export const TOP_PLACES = 20;

/** How many of the top places one site may take. */
export const TOP_PER_SITE = 4;

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

	/** Closes the top and returns every waiting result, sites taking turns. */
	release(): Item[] {
		this.topOpen = false;
		const released: Item[] = [];
		const queues = [...this.waiting.values()];
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
		if (taken >= TOP_PER_SITE) return false;
		this.topBySite.set(site, taken + 1);
		this.topCount++;
		return true;
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
