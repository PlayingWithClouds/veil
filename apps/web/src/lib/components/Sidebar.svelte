<!--
	YouTube-style sidebar: main destinations, the subscription list with a dot for
	unseen videos and the followed cam models (a dot while live), the personal section (library, downloads, history), then app settings. Phones get DrawerMenu instead.
-->
<script lang="ts">
	import { page } from '$app/stores';
	import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
	import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
	import FolderSimpleIcon from 'phosphor-svelte/lib/FolderSimpleIcon';
	import PlugsIcon from 'phosphor-svelte/lib/PlugsIcon';
	import GearSixIcon from 'phosphor-svelte/lib/GearSixIcon';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import { camsUrl, camUrl, subscriptionsUrl } from '$lib/routes';
	import { followedCams, refreshFollowedCams } from '$lib/stores/followedCams';
	import CamAvatar from './CamAvatar.svelte';
	import { subscriptionName } from '$lib/searchSubscriptions';
	import { subscriptionList } from '$lib/stores/newSubscriptionScenes';
	import type { Component } from 'svelte';
	import VpnGuard from './VpnGuard.svelte';
	import SubscriptionAvatar from './SubscriptionAvatar.svelte';

	type NavLink = {
		label: string;
		icon: Component;
		href: string;
		// Match the route exactly rather than by prefix.
		exact?: boolean;
	};

	// Subscriptions shown before "Show more".
	const VISIBLE_SUBSCRIPTIONS = 7;

	const mainLinks: NavLink[] = [
		{ label: 'Home', icon: HouseIcon, href: '/', exact: true },
		{ label: 'Cams', icon: VideoCameraIcon, href: camsUrl() },
		{ label: 'Random', icon: ShuffleIcon, href: '/random' }
	];

	// The "Subscriptions" row links here.
	const subscriptionsLink: NavLink = {
		label: 'Subscriptions',
		icon: BellIcon,
		href: subscriptionsUrl(),
		exact: true
	};

	const personalLinks: NavLink[] = [
		{ label: 'Library', icon: FilmStripIcon, href: '/library' },
		// The download queue sits at the top of the library page, so only Library
		// reads as the current page there.
		{ label: 'Downloads', icon: DownloadSimpleIcon, href: '/library#downloads' },
		{ label: 'History', icon: ClockCounterClockwiseIcon, href: '/history' },
		{ label: 'Collections', icon: FolderSimpleIcon, href: '/collections' }
	];

	const appLinks: NavLink[] = [
		{ label: 'Plugins', icon: PlugsIcon, href: '/plugins' },
		{ label: 'Settings', icon: GearSixIcon, href: '/settings' }
	];

	let visibleSubscriptions = $derived($subscriptionList.slice(0, VISIBLE_SUBSCRIPTIONS));
	// Live followed cams always show; offline ones fill the rows left over.
	let visibleCams = $derived.by(() => {
		const live = $followedCams.filter((camModel) => camModel.online);
		const offline = $followedCams.filter((camModel) => !camModel.online);
		const leftOver = Math.max(0, VISIBLE_SUBSCRIPTIONS - visibleSubscriptions.length - live.length);
		return [...live, ...offline.slice(0, leftOver)];
	});
	let followingCount = $derived($subscriptionList.length + $followedCams.length);

	// Reload on every navigation, so a follow from a room page shows up and the
	// live dots stay current.
	$effect(() => {
		void $page.url.pathname;
		refreshFollowedCams();
	});

	/** Whether the nav link points at the current page. */
	function isActive(link: NavLink): boolean {
		if (link.exact) return $page.url.pathname === link.href;
		return $page.url.pathname.startsWith(link.href);
	}

	/** Whether a subscription's feed is the page being shown. */
	function isSubscriptionActive(id: string): boolean {
		return $page.url.pathname === subscriptionsUrl() && $page.url.searchParams.get('id') === id;
	}
</script>

{#snippet navLink(link: NavLink)}
	<a
		href={link.href}
		title={link.label}
		class="text-base-content/80 hover:bg-base-200 hover:text-base-content sidebar-link-full flex items-center rounded-lg transition-colors"
		class:sidebar-link-active={isActive(link)}
	>
		<link.icon size={20} weight={isActive(link) ? 'fill' : 'regular'} />
		<span class="truncate">{link.label}</span>
	</a>
{/snippet}

<aside
	class="no-scrollbar flex w-60 shrink-0 flex-col overflow-y-auto px-3 pb-3"
	aria-label="Main menu"
>
	<nav class="flex flex-col gap-0.5">
		{#each mainLinks as link (link.href)}
			{@render navLink(link)}
		{/each}
	</nav>

	<hr class="border-base-300 my-3" />
	<a
		href={subscriptionsUrl()}
		class="text-base-content hover:bg-base-200 sidebar-link-full flex items-center rounded-lg font-semibold transition-colors"
		class:sidebar-link-active={isActive(subscriptionsLink)}
	>
		<BellIcon size={20} weight={isActive(subscriptionsLink) ? 'fill' : 'regular'} />
		<span class="flex-1">Subscriptions</span>
		<CaretRightIcon size={14} weight="bold" class="text-base-content/50" />
	</a>
	<nav class="flex flex-col gap-0.5">
		{#each visibleSubscriptions as subscription (subscription.id)}
			<a
				href={subscriptionsUrl(subscription.id)}
				title={subscriptionName(subscription)}
				class="text-base-content/80 hover:bg-base-200 hover:text-base-content sidebar-link-full flex items-center rounded-lg transition-colors"
				class:sidebar-link-active={isSubscriptionActive(subscription.id)}
			>
				<SubscriptionAvatar {subscription} />
				<span class="flex-1 truncate">{subscriptionName(subscription)}</span>
				{#if subscription.newCount > 0}
					<span
						class="bg-primary h-1.5 w-1.5 shrink-0 rounded-full"
						title="{subscription.newCount} new"
					></span>
				{/if}
			</a>
		{/each}
		{#each visibleCams as camModel (camModel.id)}
			<a
				href={camUrl(camModel.id)}
				title={camModel.displayName}
				class="text-base-content/80 hover:bg-base-200 hover:text-base-content sidebar-link-full flex items-center rounded-lg transition-colors"
				class:sidebar-link-active={$page.url.pathname === camUrl(camModel.id)}
			>
				<CamAvatar {camModel} />
				<span class="flex-1 truncate">{camModel.displayName}</span>
			</a>
		{/each}
		{#if followingCount === 0}
			<p class="text-base-content/40 px-3 py-1 text-xs">
				Subscribe to a channel, performer, tag or search, or follow a cam, to see it here.
			</p>
		{/if}
		{#if followingCount > visibleSubscriptions.length + visibleCams.length}
			<a
				href={subscriptionsUrl()}
				class="text-base-content/60 hover:bg-base-200 sidebar-link-full flex items-center rounded-lg text-sm transition-colors"
			>
				Show all {followingCount}
			</a>
		{/if}
	</nav>
	<hr class="border-base-300 my-3" />
	<p class="text-base-content/40 px-3 pb-1 text-xs font-semibold tracking-wide uppercase">You</p>

	<nav class="flex flex-col gap-0.5">
		{#each personalLinks as link (link.href)}
			{@render navLink(link)}
		{/each}
	</nav>

	<hr class="border-base-300 my-3" />
	<nav class="flex flex-col gap-0.5">
		{#each appLinks as link (link.href)}
			{@render navLink(link)}
		{/each}
	</nav>

	<div class="mt-auto flex items-center justify-between px-3 pt-4">
		<VpnGuard showLabel />
		<span class="text-base-content/20 text-[9px] font-medium">v0.1.0</span>
	</div>
</aside>

<style>
	.sidebar-link-full {
		height: 2.5rem;
		gap: 1.25rem;
		padding-inline: 0.75rem;
		font-size: 14px;
	}

	.sidebar-link-active {
		background-color: var(--color-base-200);
		color: var(--color-base-content);
		font-weight: 600;
	}
</style>
