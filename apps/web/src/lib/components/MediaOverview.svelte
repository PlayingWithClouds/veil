<!--
	The rows at the top of the galleries, audio and stories pages: what you
	stopped partway (or viewed last), your favourites when the page has them,
	the fullest categories, the newest releases, the performers, voices or
	authors (favourites first),
	then a row for each of the most common tags with "See all". Each row loads
	on its own and stays hidden while empty.
-->
<script lang="ts" generics="Item extends { id: string }">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import Shelf from '$lib/components/Shelf.svelte';
	import CategoryShelf from '$lib/components/CategoryShelf.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import { tagLabel } from '$lib/components/TagChips.svelte';
	import {
		fetchCreditedPerformers,
		fetchTopTags,
		type MediaListing,
		type OverviewKind,
		type TagCountRow
	} from '$lib/media';
	import type { MemberRow } from '$lib/library';

	// Items per row, and tag rows shown.
	const ROW_ITEMS = 16;
	const TAG_ROWS = 3;
	const PERFORMER_WIDTH = 140;

	interface Props {
		kind: OverviewKind;
		// "Continue listening", "Continue reading", "Recently viewed".
		continueTitle: string;
		// "Voices", "Authors", "Performers".
		peopleTitle: string;
		fetchContinue: () => Promise<Item[]>;
		// The favourites row, on pages that have one.
		fetchFavorites?: () => Promise<Item[]>;
		load: (limit: number, offset: number, listing: MediaListing) => Promise<Item[]>;
		itemWidth: number;
		mediaCard: Snippet<[Item]>;
		// The card in Continue, e.g. one that opens the reader.
		continueCard?: Snippet<[Item]>;
	}

	let {
		kind,
		continueTitle,
		peopleTitle,
		fetchContinue,
		fetchFavorites,
		load,
		itemWidth,
		mediaCard,
		continueCard
	}: Props = $props();

	/** One tag's row. */
	interface TagRow {
		tag: TagCountRow['tag'];
		items: Item[];
	}

	let started = $state<Item[]>([]);
	let favorites = $state<Item[]>([]);
	let newest = $state<Item[]>([]);
	let people = $state<MemberRow[]>([]);
	let tagRows = $state<TagRow[]>([]);

	$effect(() => {
		fill(fetchContinue(), (found) => (started = found));
		if (fetchFavorites) fill(fetchFavorites(), (found) => (favorites = found));
		fill(load(ROW_ITEMS, 0, { byRelease: true }), (found) => (newest = found));
		fill(fetchCreditedPerformers(kind), (found) => (people = found));
		fill(loadTagRows(), (found) => (tagRows = found));
	});

	/** The rows of the most common tags. */
	async function loadTagRows(): Promise<TagRow[]> {
		const counts = await fetchTopTags(kind, TAG_ROWS);
		return Promise.all(
			counts.map(async (count) => ({
				tag: count.tag,
				items: await load(ROW_ITEMS, 0, { tagId: count.tag.id, byRelease: true })
			}))
		);
	}

	/** Fills a row when its data arrives; a row that fails stays hidden and says why in the console. */
	function fill<T>(request: Promise<T>, apply: (found: T) => void) {
		request.then(apply, (failure) => console.warn('overview row failed to load', failure));
	}

	/** The page narrowed to a tag. */
	function tagHref(tagId: string): string {
		return `${page.url.pathname}?${new URLSearchParams({ tag: tagId })}`;
	}
</script>

<div class="flex flex-col gap-8">
	<Shelf title={continueTitle} items={started} key={(item) => item.id} {itemWidth}>
		{#snippet card(item)}
			{#if continueCard}
				{@render continueCard(item)}
			{:else}
				{@render mediaCard(item)}
			{/if}
		{/snippet}
	</Shelf>

	<Shelf title="Favourites" items={favorites} key={(item) => item.id} {itemWidth}>
		{#snippet card(item)}
			{@render mediaCard(item)}
		{/snippet}
	</Shelf>

	<CategoryShelf {kind} />

	<Shelf title="Newest releases" items={newest} key={(item) => item.id} {itemWidth}>
		{#snippet card(item)}
			{@render mediaCard(item)}
		{/snippet}
	</Shelf>

	<Shelf
		title={peopleTitle}
		items={people}
		key={(performer) => performer.id}
		itemWidth={PERFORMER_WIDTH}
	>
		{#snippet card(performer)}
			<MemberCard kind="performer" member={performer} />
		{/snippet}
	</Shelf>

	{#each tagRows as row (row.tag.id)}
		<Shelf
			title={tagLabel(row.tag.name)}
			items={row.items}
			key={(item) => item.id}
			{itemWidth}
			href={tagHref(row.tag.id)}
		>
			{#snippet card(item)}
				{@render mediaCard(item)}
			{/snippet}
		</Shelf>
	{/each}
</div>
