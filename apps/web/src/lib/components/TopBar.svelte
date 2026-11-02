<!--
	Global top bar: logo on the left, the search (with "Who is this?") in the
	middle, notifications on the right. A search opens home with ?q=, which
	asks every site and stores what they return. Phones get
	a slim bar floating over the page that slides away while scrolling down
	(chromeHidden), with a back button on detail pages.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import FilmReelIcon from 'phosphor-svelte/lib/FilmReelIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import { chromeHidden, isCompact } from '$lib/stores/viewport';
	import { activeNavIndex } from '$lib/navigation';
	import WhoIsThis from '$lib/components/WhoIsThis.svelte';
	import NotificationBell from '$lib/components/NotificationBell.svelte';

	let query = $state(page.url.searchParams.get('q') ?? '');

	// Top-level pages show the logo; detail pages a back button.
	let onTopLevelPage = $derived(activeNavIndex(page.url.pathname) !== -1);

	// Measured, not assumed: sticky rows below (top-chrome) must sit flush under it.
	let compactHeaderHeight = $state(0);
	$effect(() => {
		if (compactHeaderHeight === 0) return;
		document.documentElement.style.setProperty(
			'--compact-header-height',
			`${compactHeaderHeight}px`
		);
	});

	/** Opens home searching the sites for the typed query. */
	function submit(event: SubmitEvent) {
		event.preventDefault();
		const trimmed = query.trim();
		if (trimmed === '') {
			goto('/');
			return;
		}
		goto(`/?${new URLSearchParams({ q: trimmed })}`);
	}
</script>

{#snippet searchBox()}
	<form
		onsubmit={submit}
		class="bg-base-content/8 flex h-10 min-w-0 flex-1 items-center rounded-full"
		role="search"
	>
		<MagnifyingGlassIcon size={18} class="text-base-content/50 ml-3.5 shrink-0" />
		<input
			id="global-search"
			type="search"
			bind:value={query}
			placeholder="Search the sites"
			class="h-full min-w-0 flex-1 bg-transparent px-2.5 text-[15px] outline-none"
		/>
		<WhoIsThis />
	</form>
{/snippet}

{#if $isCompact}
	<header
		bind:offsetHeight={compactHeaderHeight}
		class="compact-header pt-safe fixed inset-x-0 top-0 z-30 flex items-center gap-3 px-3"
		class:compact-header-hidden={$chromeHidden}
	>
		<div class="flex h-14 w-full items-center gap-3">
			{#if onTopLevelPage}
				<a href="/" class="flex shrink-0 items-center" aria-label="Veil home">
					<span class="text-primary"><FilmReelIcon size={24} weight="fill" /></span>
				</a>
			{:else}
				<button
					type="button"
					onclick={() => history.back()}
					aria-label="Back"
					class="-ml-1 flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
				>
					<ArrowLeftIcon size={22} />
				</button>
			{/if}
			{@render searchBox()}
			<NotificationBell />
		</div>
	</header>
{:else}
	<header class="flex h-14 shrink-0 items-center gap-4 px-4">
		<div class="flex w-56 shrink-0 items-center gap-3">
			<a href="/" class="flex items-center gap-1.5">
				<span class="text-primary"><FilmReelIcon size={24} weight="fill" /></span>
				<span class="text-lg font-semibold tracking-tight">Veil</span>
				<span class="text-error -mt-3 text-[10px] font-bold">18+</span>
			</a>
		</div>
		<div class="mx-auto flex w-full max-w-2xl items-center">
			{@render searchBox()}
		</div>
		<div class="flex w-56 shrink-0 justify-end">
			<NotificationBell />
		</div>
	</header>
{/if}

<style>
	.compact-header {
		background-color: var(--color-base-100);
		transition: translate 280ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	.compact-header-hidden {
		translate: 0 -100%;
	}

	@media (prefers-reduced-motion: reduce) {
		.compact-header {
			transition: none;
		}
	}
</style>
