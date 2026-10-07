import { apiBase } from '$lib/server';

export type MediaAddedEvent = { id: string; type: string; title: string };
export type MediaRemovedEvent = { id: string };
export type EpisodeAddedEvent = {
	mediaId: string;
	title: string;
	season: number;
	episode: number;
	airDate?: string;
};
export type JobUpdatedEvent = {
	id: string;
	kind: string;
	status: string;
	error?: string;
	updatedAt: string;
	downloadTitle?: string;
	downloadUrl?: string;
	progress?: number;
	bytesReceived?: number;
	bytesTotal?: number;
};

type EventMap = {
	'movie:added': MediaAddedEvent;
	'movie:updated': MediaAddedEvent;
	'movie:removed': MediaRemovedEvent;
	'show:added': MediaAddedEvent;
	'show:updated': MediaAddedEvent;
	'show:removed': MediaRemovedEvent;
	'gallery:added': MediaAddedEvent;
	'gallery:removed': MediaRemovedEvent;
	'episode:added': EpisodeAddedEvent;
	'job:updated': JobUpdatedEvent;
	[key: string]: unknown;
};

type EventHandler<T> = (data: T) => void;

export interface EventStream {
	on<K extends keyof EventMap>(topic: K, handler: EventHandler<EventMap[K]>): () => void;
	close(): void;
}

/**
 * Open an SSE stream to /api/events.
 *
 * @param topics  Array of topic patterns to subscribe to.
 *                Supports exact names ("movie:added"), wildcards ("show:*"), or "*" for all.
 *                Defaults to ["*"].
 *
 * Usage in Svelte:
 *   onMount(() => {
 *     const stream = createEventStream(['movie:added', 'movie:removed']);
 *     const off = stream.on('movie:added', (e) => items = [...items, e]);
 *     return () => { off(); stream.close(); };
 *   });
 */
export function createEventStream(topics: string[] = ['*']): EventStream {
	const url = new URL(`${apiBase()}/api/events`);
	url.searchParams.set('topics', topics.join(','));
	const source = new EventSource(url.toString());

	return {
		on<K extends keyof EventMap>(topic: K, handler: EventHandler<EventMap[K]>): () => void {
			const listener = (e: MessageEvent) => {
				try {
					handler(JSON.parse(e.data) as EventMap[K]);
				} catch {
					// malformed event data — ignore
				}
			};
			source.addEventListener(topic as string, listener);
			return () => source.removeEventListener(topic as string, listener);
		},
		close(): void {
			source.close();
		}
	};
}
