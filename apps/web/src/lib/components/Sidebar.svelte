<!--
	YouTube-style sidebar: main destinations, the subscription list with a dot for
	unseen videos, the personal section, then app settings. Collapses to a narrow
	icon rail from the top bar's menu button. Phones get DrawerMenu instead.
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
	import { subscriptionsUrl } from '$lib/routes';
	import { subscriptionName } from '$lib/searchSubscriptions';
	import { sidebarCollapsed } from '$lib/stores/sidebar';
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
		{ label: 'Random', icon: ShuffleIcon, href: '/random' }
	];

	// In the icon rail the subscriptions section is hidden, so it gets a link of
	// its own; the full sidebar reaches it through the section header.
	const subscriptionsLink: NavLink = {
		label: 'Subscriptions',
		icon: BellIcon,
		href: subscriptionsUrl(),
		exact: true
	};

	const personalLinks: NavLink[] = [
		{ label: 'Library', icon: FilmStripIcon, href: '/library' },
		{ label: 'History', icon: ClockCounterClockwiseIcon, href: '/history' },
		{ label: 'Collections', icon: FolderSimpleIcon, href: '/collections' }
	];

	const appLinks: NavLink[] = [
		{ label: 'Plugins', icon: PlugsIcon, href: '/plugins' },
		{ label: 'Settings', icon: GearSixIcon, href: '/settings' }
	];

	let visibleSubscriptions = $derived($subscriptionList.slice(0, VISIBLE_SUBSCRIPTIONS));
	let rail = $derived($sidebarCollapsed);

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
		class="text-base-content/80 hover:bg-base-200 hover:text-base-content flex items-center rounded-lg transition-colors"
		class:sidebar-link-active={isActive(link)}
		class:sidebar-link-full={!rail}
		class:sidebar-link-rail={rail}
	>
		<link.icon size={rail ? 22 : 20} weight={isActive(link) ? 'fill' : 'regular'} />
		<span class="truncate">{link.label}</span>
	</a>
{/snippet}

<aside
	class="no-scrollbar flex shrink-0 flex-col overflow-y-auto px-3 pb-3"
	class:w-60={!rail}
	class:w-[76px]={rail}
	class:px-1={rail}
	aria-label="Main menu"
>
	<nav class="flex flex-col gap-0.5">
		{#each mainLinks as link (link.href)}
			{@render navLink(link)}
		{/each}
		{#if rail}
			{@render navLink(subscriptionsLink)}
		{/if}
	</nav>

	{#if !rail}
		<hr class="border-base-300 my-3" />
		<a
			href={subscriptionsUrl()}
			class="hover:bg-base-200 flex items-center gap-1 rounded-lg px-3 py-2 text-[15px] font-semibold transition-colors"
			class:sidebar-link-active={isActive(subscriptionsLink)}
		>
			Subscriptions
			<CaretRightIcon size={14} weight="bold" />
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
			{:else}
				<p class="text-base-content/40 px-3 py-1 text-xs">
					Subscribe to a channel, performer, tag or search to follow it here.
				</p>
			{/each}
			{#if $subscriptionList.length > VISIBLE_SUBSCRIPTIONS}
				<a
					href={subscriptionsUrl()}
					class="text-base-content/60 hover:bg-base-200 sidebar-link-full flex items-center rounded-lg text-sm transition-colors"
				>
					Show all {$subscriptionList.length}
				</a>
			{/if}
		</nav>
		<hr class="border-base-300 my-3" />
		<p class="px-3 pb-1 text-[15px] font-semibold">You</p>
	{/if}

	<nav class="flex flex-col gap-0.5">
		{#each personalLinks as link (link.href)}
			{@render navLink(link)}
		{/each}
	</nav>

	{#if !rail}
		<hr class="border-base-300 my-3" />
	{/if}
	<nav class="flex flex-col gap-0.5">
		{#each appLinks as link (link.href)}
			{@render navLink(link)}
		{/each}
	</nav>

	{#if !rail}
		<div class="mt-auto flex items-center justify-between px-3 pt-4">
			<VpnGuard />
			<span class="text-base-content/20 text-[9px] font-medium">v0.1.0</span>
		</div>
	{/if}
</aside>

<style>
	.sidebar-link-full {
		height: 2.5rem;
		gap: 1.25rem;
		padding-inline: 0.75rem;
		font-size: 14px;
	}

	.sidebar-link-rail {
		flex-direction: column;
		justify-content: center;
		gap: 0.3rem;
		padding-block: 1rem;
		font-size: 10px;
	}

	.sidebar-link-active {
		background-color: var(--color-base-200);
		color: var(--color-base-content);
		font-weight: 600;
	}
</style>
