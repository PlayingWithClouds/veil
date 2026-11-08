<script lang="ts">
	import './layout.css';
	import '$lib/tv.css';
	import AmbientGlow from '$lib/components/AmbientGlow.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import TopBar from '$lib/components/TopBar.svelte';
	import BottomNav from '$lib/components/BottomNav.svelte';
	import UndoToast from '$lib/components/UndoToast.svelte';
	import CollectionPickerDialog from '$lib/components/CollectionPickerDialog.svelte';
	import { collectionPick } from '$lib/stores/collectionPicker';
	import { afterNavigate } from '$app/navigation';
	import { chromeHidden, isCompact } from '$lib/stores/viewport';
	import { isNativeApp } from '$lib/server';
	import { listenToBackButton } from '$lib/native';
	import dayjs from 'dayjs';
	import duration from 'dayjs/plugin/duration';
	import relativeTime from 'dayjs/plugin/relativeTime';

	let { children } = $props();

	// Registered once on the shared instance; every import of dayjs has them.
	dayjs.extend(duration);
	dayjs.extend(relativeTime);

	afterNavigate(() => {
		chromeHidden.set(false);
	});

	// Scroll in one direction before the phone top bar hides or returns, so
	// jitter doesn't flicker it; it never hides within its own height of the top.
	const CHROME_SCROLL_THRESHOLD = 8;
	const CHROME_HIDE_AFTER = 56;
	// Scrolled past this, the desktop top bar frosts over the page under it.
	const HEADER_FROST_AFTER = 4;
	let lastScrollTop = 0;
	// How far the page has scrolled, for the glow and the top bar.
	let mainScrollTop = $state(0);

	/**
	 * Follows the page's scroll: the glow fades and the desktop top bar
	 * frosts; the phone top bar slides away while scrolling down and back on
	 * any scroll up.
	 */
	function onMainScroll(event: Event) {
		if (!(event.currentTarget instanceof HTMLElement)) return;
		const scrollTop = event.currentTarget.scrollTop;
		mainScrollTop = scrollTop;
		if (!$isCompact) return;
		const delta = scrollTop - lastScrollTop;
		if (Math.abs(delta) < CHROME_SCROLL_THRESHOLD) return;
		lastScrollTop = scrollTop;
		chromeHidden.set(delta > 0 && scrollTop > CHROME_HIDE_AFTER);
	}

	$effect(() => {
		if (!isNativeApp()) return;
		return listenToBackButton();
	});

	/** Ctrl/Cmd+K focuses the search. */
	function onWindowKeydown(event: KeyboardEvent) {
		if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
			event.preventDefault();
			document.getElementById('global-search')?.focus();
		}
	}
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div
	class="app-shell bg-base-100 relative isolate flex h-dvh flex-col overflow-hidden"
	class:chrome-hidden={$chromeHidden}
	class:scrolled={mainScrollTop > HEADER_FROST_AFTER}
>
	<AmbientGlow scrollTop={mainScrollTop} />
	<TopBar />
	<div class="flex min-h-0 flex-1">
		{#if !$isCompact}
			<Sidebar />
		{/if}
		<main
			onscroll={onMainScroll}
			class="no-scrollbar max-md:pb-bottom-nav min-w-0 flex-1 overflow-x-hidden overflow-y-auto overscroll-y-contain px-4 pb-4 max-md:px-3"
		>
			{#if $isCompact}
				<!-- Room for the floating top bar. A spacer, not padding: sticky
				     offsets count from inside the scroll container's padding. -->
				<div class="shrink-0" style:height="var(--compact-header-height)" aria-hidden="true"></div>
			{:else}
				<!-- Room for the see-through desktop top bar the page scrolls under. -->
				<div class="shrink-0" style:height="var(--desktop-header-height)" aria-hidden="true"></div>
			{/if}
			{@render children()}
		</main>
	</div>
	{#if $isCompact}
		<BottomNav />
	{/if}
	<UndoToast />
	{#if $collectionPick}
		{#key $collectionPick}
			<CollectionPickerDialog pick={$collectionPick} />
		{/key}
	{/if}
</div>

<style>
	/* Sticky rows stick under the desktop top bar, which floats over the page. */
	.app-shell {
		--chrome-offset: var(--desktop-header-height, 0px);
		transition: --chrome-offset 280ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	/* The phone layout; same rule as isCompact. */
	@media (width < 720px), ((height < 540px) and (pointer: coarse)) {
		.app-shell {
			--chrome-offset: var(--compact-header-height);
		}
		.app-shell.chrome-hidden {
			--chrome-offset: 0px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.app-shell {
			transition: none;
		}
	}
</style>
