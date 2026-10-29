<!--
	A row of identification-forum threads (NameThatPorn) matching a search,
	shown above the search results: people asking who is in a video or where
	it is from, often answered with names and links. Each opens its thread in
	the app (/forum). Collapses when the forums find nothing or fail.
-->
<script lang="ts">
	import { cacheUrl } from '$lib/img';
	import { forumThreadUrl } from '$lib/routes';
	import { fetchForumSearch, type ForumPostHit } from '$lib/forumThread';

	interface Props {
		/** The search; empty shows nothing. */
		query: string;
	}

	let { query }: Props = $props();

	const SKELETON_CARDS = 6;

	let hits = $state<ForumPostHit[]>([]);
	let loading = $state(false);
	// Tells searches apart, so a slow answer never replaces a newer one.
	let requestToken = 0;

	// The sites the threads come from, for the heading ("namethatporn.com").
	let siteNames = $derived([...new Set(hits.map((hit) => siteOf(hit.url)))].join(', '));

	$effect(() => {
		load(query.trim());
	});

	/** Asks the forums about the query. */
	async function load(text: string) {
		const token = ++requestToken;
		hits = [];
		if (text === '') {
			loading = false;
			return;
		}
		loading = true;
		try {
			const found = await fetchForumSearch(text);
			if (token !== requestToken) return;
			hits = found;
		} catch (error) {
			console.error('forumSearch failed', text, error);
		} finally {
			if (token === requestToken) loading = false;
		}
	}

	/** The host of a thread, without "www.". */
	function siteOf(url: string): string {
		try {
			return new URL(url).hostname.replace(/^www\./, '');
		} catch {
			return url;
		}
	}
</script>

{#if loading || hits.length > 0}
	<section class="flex flex-col gap-3" aria-label="Forum threads">
		<h2 class="flex items-baseline gap-2 text-lg font-semibold">
			Threads
			{#if hits.length > 0}
				<span class="text-base-content/40 text-sm font-normal">{hits.length} · {siteNames}</span>
			{/if}
		</h2>
		<ul class="no-scrollbar -mx-1 flex gap-3 overflow-x-auto px-1 pb-1">
			{#if loading}
				{#each { length: SKELETON_CARDS } as _, index (index)}
					<li class="flex w-56 shrink-0 flex-col gap-2">
						<div class="skeleton aspect-square w-full rounded-xl"></div>
						<div class="skeleton h-3.5 w-11/12 rounded"></div>
						<div class="skeleton h-3 w-1/2 rounded"></div>
					</li>
				{/each}
			{:else}
				{#each hits as hit (hit.url)}
					<li class="w-56 shrink-0">
						<a href={forumThreadUrl(hit.url)} class="group flex flex-col gap-2">
							<div class="relative">
								{#if hit.thumbnailUrl}
									<img
										src={cacheUrl(hit.thumbnailUrl)}
										alt=""
										loading="lazy"
										class="aspect-square w-full rounded-xl bg-white/5 object-cover transition-opacity group-hover:opacity-85"
									/>
								{:else}
									<div class="aspect-square w-full rounded-xl bg-white/5"></div>
								{/if}
								{#if hit.solved}
									<span
										class="bg-primary absolute top-2 left-2 rounded px-1.5 py-0.5 text-xs font-bold text-black"
									>
										Answered
									</span>
								{/if}
							</div>
							<span class="line-clamp-2 text-sm font-semibold">{hit.title}</span>
							{#if hit.snippet}
								<span class="text-base-content/50 line-clamp-2 text-xs">{hit.snippet}</span>
							{/if}
						</a>
					</li>
				{/each}
			{/if}
		</ul>
	</section>
{/if}
