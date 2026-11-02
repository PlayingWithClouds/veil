<!--
	A searchable grid of performers or studios, loading more on scroll.
-->
<script lang="ts">
	import PageHeader from '$lib/components/PageHeader.svelte';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { fetchPerformers, fetchStudios, type MemberRow } from '$lib/library';

	let { kind }: { kind: 'performer' | 'studio' } = $props();

	const PAGE_SIZE = 60;

	let members = $state<MemberRow[]>([]);
	let text = $state('');
	let loading = $state(false);
	let exhausted = $state(false);

	let title = $derived(kind === 'performer' ? 'Performers' : 'Studios');

	$effect(() => {
		void loadFirst(text);
	});

	/** Fetches a page of members whose name matches. */
	function fetchPage(search: string, offset: number): Promise<MemberRow[]> {
		if (kind === 'performer') return fetchPerformers(search, PAGE_SIZE, offset);
		return fetchStudios(search, PAGE_SIZE, offset);
	}

	/** Loads the first page for the current search. */
	async function loadFirst(search: string) {
		loading = true;
		try {
			members = await fetchPage(search, 0);
			exhausted = members.length < PAGE_SIZE;
		} finally {
			loading = false;
		}
	}

	/** Appends the next page. */
	async function loadMore() {
		if (loading || exhausted) return;
		loading = true;
		try {
			const nextPage = await fetchPage(text, members.length);
			members = [...members, ...nextPage];
			exhausted = nextPage.length < PAGE_SIZE;
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>{title} · Veil</title></svelte:head>

<div class="flex flex-col gap-5">
	<PageHeader {title}>
		{#snippet actions()}
			<input
				type="search"
				bind:value={text}
				placeholder="Filter by name"
				class="bg-base-200 h-9 w-56 rounded-lg px-3 text-sm outline-none"
			/>
		{/snippet}
	</PageHeader>

	<div
		class="grid gap-x-4 gap-y-6"
		class:grid-cols-[repeat(auto-fill,minmax(140px,1fr))]={kind === 'performer'}
		class:grid-cols-[repeat(auto-fill,minmax(220px,1fr))]={kind === 'studio'}
	>
		{#each members as member (member.id)}
			<MemberCard {kind} {member} />
		{/each}
	</div>

	{#if !loading && members.length === 0}
		<p class="text-base-content/50 text-sm">
			None yet. They arrive with the videos and galleries you browse.
		</p>
	{/if}

	<ScrollSentinel onvisible={loadMore} loading={loading && members.length > 0} done={exhausted} />
</div>
