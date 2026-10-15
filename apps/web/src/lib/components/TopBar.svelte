<!--
	Global top bar, YouTube style: logo on the left, the search and the
	videos/images toggle in the middle, the library and downloads link on the
	right. Phones get a slim bar floating over the page: logo plus a search button that opens the full-screen SearchOverlay. It slides
	away while scrolling down (chromeHidden); the drawer opens by swiping, and
	the media toggle moves into Home's chip row.
-->
<script lang="ts">
	import { page } from '$app/stores';
	import FilmReelIcon from 'phosphor-svelte/lib/FilmReelIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { searchQuery } from '$lib/stores/searchQuery';
	import { chromeHidden, isCompact } from '$lib/stores/viewport';
	import { searchOverlayOpen } from '$lib/stores/searchOverlay';
	import { pageHeroVisible, pageTitle } from '$lib/stores/pageTitle';
	import { activeTabIndex } from '$lib/tabs';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import { fly } from 'svelte/transition';
	import DesktopSearchBox from './DesktopSearchBox.svelte';
	import MediaFilterToggle from './MediaFilterToggle.svelte';

	// Phones show the running search in the bar while Home shows its results.
	let showingSearch = $derived($page.url.pathname === '/' && $searchQuery.trim().length > 0);

	// Tab pages show the logo; everything else (detail pages) a back button.
	let onTabPage = $derived(activeTabIndex($page.url.pathname) !== -1);

	// Measured, not assumed: sticky rows below (top-chrome) must sit flush under it.
	let compactHeaderHeight = $state(0);
	$effect(() => {
		if (compactHeaderHeight === 0) return;
		document.documentElement.style.setProperty(
			'--compact-header-height',
			`${compactHeaderHeight}px`
		);
	});
</script>

{#if $isCompact}
	<header
		bind:offsetHeight={compactHeaderHeight}
		class="compact-header pt-safe fixed inset-x-0 top-0 z-30 flex items-center gap-3 px-3"
		class:compact-header-hidden={$chromeHidden}
		class:compact-header-clear={$pageHeroVisible}
	>
		<div class="flex h-14 w-full items-center gap-3">
			{#if onTabPage}
				<a href="/" class="flex shrink-0 items-center gap-1.5" aria-label="Veil home">
					<span class="text-primary"><FilmReelIcon size={24} weight="fill" /></span>
					{#if !showingSearch}
						<span class="text-base font-semibold tracking-tight">Veil</span>
					{/if}
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
				{#if $pageTitle && !$pageHeroVisible}
					<!-- The detail page's name, once its hero has scrolled away. -->
					<div class="flex min-w-0 items-center gap-2.5" in:fly={{ y: 8, duration: 220 }}>
						{#if $pageTitle.imageUrl}
							<img
								src={$pageTitle.imageUrl}
								alt=""
								class="h-8 w-8 shrink-0 object-cover"
								class:rounded-full={$pageTitle.imageShape === 'avatar'}
								class:rounded-lg={$pageTitle.imageShape === 'logo'}
							/>
						{/if}
						<span class="truncate text-base font-semibold">{$pageTitle.title}</span>
					</div>
				{/if}
			{/if}
			{#if showingSearch}
				<!-- The running search: tap to edit it in the overlay, ✕ to go back to the feed. -->
				<div class="bg-base-content/8 flex h-10 min-w-0 flex-1 items-center rounded-full">
					<button
						type="button"
						onclick={() => searchOverlayOpen.set(true)}
						class="flex h-full min-w-0 flex-1 items-center gap-2 pl-3.5 text-left text-[15px]"
					>
						<MagnifyingGlassIcon size={18} class="text-base-content/50 shrink-0" />
						<span class="truncate">{$searchQuery}</span>
					</button>
					<button
						type="button"
						onclick={() => searchQuery.set('')}
						aria-label="Clear search"
						class="text-base-content/60 flex h-full w-10 shrink-0 items-center justify-center"
					>
						<XIcon size={18} />
					</button>
				</div>
			{:else}
				<button
					type="button"
					onclick={() => searchOverlayOpen.set(true)}
					aria-label="Search"
					class="-mr-1 ml-auto flex h-10 w-10 items-center justify-center rounded-full"
				>
					<MagnifyingGlassIcon size={22} />
				</button>
			{/if}
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
			<DesktopSearchBox />
			<div class="ml-3">
				<MediaFilterToggle />
			</div>
		</div>

		<div class="flex w-56 shrink-0 items-center justify-end gap-1">
			<a
				href="/settings"
				title="Account"
				aria-label="Account"
				class="bg-base-300 hover:bg-base-content/20 text-base-content/70 flex h-9 w-9 items-center justify-center rounded-full transition-colors"
			>
				<UserIcon size={20} weight="fill" />
			</a>
		</div>
	</header>
{/if}

<style>
	/* Solid like the chip row sticking under it: two glass layers blur different
	   content and never quite match, so the pair wouldn't read as one bar. */
	.compact-header {
		background-color: var(--color-base-100);
		transition:
			translate 280ms cubic-bezier(0.22, 1, 0.36, 1),
			background-color 220ms ease-out;
	}

	.compact-header-hidden {
		translate: 0 -100%;
	}

	/* Over a detail page's hero the bar lets its blurred backdrop show through. */
	.compact-header-clear {
		background-color: transparent;
	}

	@media (prefers-reduced-motion: reduce) {
		.compact-header {
			transition: none;
		}
	}
</style>
