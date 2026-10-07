<script lang="ts">
	import './layout.css';
	import '$lib/tv.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import MobileDrawer from '$lib/components/MobileDrawer.svelte';
	import SceneActionSheet from '$lib/components/SceneActionSheet.svelte';
	import SearchOverlay from '$lib/components/SearchOverlay.svelte';
	import { searchOverlayOpen } from '$lib/stores/searchOverlay';
	import TopBar from '$lib/components/TopBar.svelte';
	import BottomNav from '$lib/components/BottomNav.svelte';
	import NotificationManager from '$lib/components/NotificationManager.svelte';
	import CollectionDialog from '$lib/components/CollectionDialog.svelte';
	import MiniPlayer from '$lib/components/MiniPlayer.svelte';
	import { afterNavigate } from '$app/navigation';
	import { loadWatchProgress } from '$lib/stores/watchProgress';
	import { refreshNewSubscriptionScenes } from '$lib/stores/newSubscriptionScenes';
	import {
		chromeHidden,
		immersiveTop,
		drawerDrag,
		drawerOpen,
		drawerProgress,
		isCompact
	} from '$lib/stores/viewport';
	import { isNativeApp } from '$lib/server';
	import { listenToBackButton } from '$lib/native';
	import { listenForPressFeedback } from '$lib/pressFeedback';
	import { listenForSwipes } from '$lib/swipeGestures';

	let { children } = $props();

	/**
	 * Pushes the page back as the phone drawer opens: it shrinks a little, slides
	 * right and rounds its corners, like a card behind the drawer. Unset while
	 * closed, so fixed children (players, dialogs) keep the viewport as their box.
	 */
	function pushedBackStyle(progress: number): string | undefined {
		if (!$isCompact || progress === 0) return undefined;
		return `transform: translateX(${progress * 1.5}rem) scale(${1 - progress * 0.04}); border-radius: ${progress * 1.5}rem;`;
	}

	// Keep the shared resume-progress map fresh: reload after every navigation, so
	// a scene watched on the detail page shows an updated bar when the user
	// returns to a list. Following a link also closes the phone drawer.
	/** Room above the page: the top bar, or just the status bar on immersive pages. */
	function topSpacerHeight(immersive: boolean): string {
		if (immersive) return 'var(--safe-area-inset-top, env(safe-area-inset-top, 0px))';
		return 'var(--compact-header-height)';
	}

	// An immersive page that mounts after the navigation hides the bar at once.
	$effect(() => {
		if ($immersiveTop) chromeHidden.set(true);
	});

	afterNavigate(() => {
		loadWatchProgress();
		drawerOpen.set(false);
		chromeHidden.set($immersiveTop);
	});

	// Scroll in one direction before the phone top bar hides or returns, so
	// jitter doesn't flicker it; it never hides within its own height of the top.
	const CHROME_SCROLL_THRESHOLD = 8;
	const CHROME_HIDE_AFTER = 56;
	let lastScrollTop = 0;

	/** Slides the phone top bar away while scrolling down and back on any scroll up. */
	function onMainScroll(event: Event) {
		if (!$isCompact || !(event.currentTarget instanceof HTMLElement)) return;
		const scrollTop = event.currentTarget.scrollTop;
		const delta = scrollTop - lastScrollTop;
		if (Math.abs(delta) < CHROME_SCROLL_THRESHOLD) return;
		lastScrollTop = scrollTop;
		const scrollingDown = delta > 0 && scrollTop > CHROME_HIDE_AFTER;
		const atImmersiveTop = $immersiveTop && scrollTop <= CHROME_HIDE_AFTER;
		chromeHidden.set(scrollingDown || atImmersiveTop);
	}

	// The subscription list and unseen badge feed both the sidebar and the phone tab bar.
	$effect(() => {
		refreshNewSubscriptionScenes();
	});

	$effect(() => {
		if (!isNativeApp()) return;
		return listenToBackButton();
	});

	$effect(() => listenForPressFeedback());

	$effect(() => {
		if (!$isCompact) return;
		return listenForSwipes();
	});

	// Opening the sidebar via the left edge is handled by the FocusList that
	// currently owns keyboard navigation (JS-tracked position); the layout only
	// handles global shortcuts.
	function onWindowKeydown(e: KeyboardEvent) {
		if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
			e.preventDefault();
			document.getElementById('global-search')?.focus();
		}
	}
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div
	class="app-shell bg-base-100 flex h-dvh flex-col overflow-hidden"
	class:app-shell-dragging={$drawerDrag !== null}
	class:chrome-hidden={$chromeHidden}
	style={pushedBackStyle($drawerProgress)}
>
	<TopBar />
	<NotificationManager />
	<CollectionDialog />
	<MiniPlayer />
	<div class="flex min-h-0 flex-1">
		{#if !$isCompact}
			<Sidebar />
		{/if}
		<main
			data-swipe-page
			onscroll={onMainScroll}
			class="no-scrollbar min-w-0 flex-1 overflow-x-hidden overflow-y-auto overscroll-y-contain px-4 pb-4 max-md:px-3 max-md:pb-bottom-nav"
		>
			{#if $isCompact}
				<!-- Room for the floating top bar. A spacer, not padding: sticky
				     offsets count from inside the scroll container's padding, so
				     padding would push sticky rows (top-chrome) down twice. -->
				<div class="shrink-0" style:height={topSpacerHeight($immersiveTop)} aria-hidden="true"></div>
			{/if}
			{@render children()}
		</main>
	</div>
	{#if $isCompact}
		<BottomNav />
	{/if}
</div>

{#if $isCompact}
	<MobileDrawer />
	<SceneActionSheet />
	{#if $searchOverlayOpen}
		<SearchOverlay />
	{/if}
{/if}

<style>
	/* Transform timed like the drawer panel so page and drawer move as one;
	   --chrome-offset like the top bar sliding away. */
	.app-shell {
		--chrome-offset: 0px;
		transition:
			transform 340ms cubic-bezier(0.22, 1, 0.36, 1),
			border-radius 340ms cubic-bezier(0.22, 1, 0.36, 1),
			--chrome-offset 280ms cubic-bezier(0.22, 1, 0.36, 1);
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

	.app-shell-dragging {
		transition: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.app-shell {
			transition: none;
		}
	}
</style>
