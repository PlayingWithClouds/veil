<!--
	Phone drawer contents. Home, Subscriptions, Library, Collections and History
	already sit in the tab bar, so the drawer carries what it doesn't: the
	subscription list as the main content, Random and Downloads, and Plugins,
	Settings and the VPN state in a footer within thumb reach. Rows fade in
	one after another each time the drawer opens.
-->
<script lang="ts">
	import { page } from '$app/stores';
	import FilmReelIcon from 'phosphor-svelte/lib/FilmReelIcon';
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import PlugsIcon from 'phosphor-svelte/lib/PlugsIcon';
	import GearSixIcon from 'phosphor-svelte/lib/GearSixIcon';
	import BellSimpleIcon from 'phosphor-svelte/lib/BellSimpleIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import type { Component } from 'svelte';
	import { subscriptionsUrl } from '$lib/routes';
	import { subscriptionName } from '$lib/searchSubscriptions';
	import { subscriptionList } from '$lib/stores/newSubscriptionScenes';
	import VpnGuard from './VpnGuard.svelte';
	import SubscriptionAvatar from './SubscriptionAvatar.svelte';

	interface Props {
		// Settled open: plays the staggered entrance.
		open: boolean;
		// A finger is dragging the drawer: rows show without animating.
		dragging: boolean;
	}

	let { open, dragging }: Props = $props();

	type NavLink = {
		label: string;
		icon: Component;
		href: string;
	};

	// Subscriptions shown before "Show all".
	const VISIBLE_SUBSCRIPTIONS = 15;
	// Rows past this one enter together, so a long list doesn't trail in.
	const MAX_STAGGER_INDEX = 12;

	const quickLinks: NavLink[] = [
		{ label: 'Random', icon: ShuffleIcon, href: '/random' }
	];

	const footerLinks: NavLink[] = [
		{ label: 'Plugins', icon: PlugsIcon, href: '/plugins' },
		{ label: 'Settings', icon: GearSixIcon, href: '/settings' }
	];

	let visibleSubscriptions = $derived($subscriptionList.slice(0, VISIBLE_SUBSCRIPTIONS));

	// Stagger slots: header, quick links, section title, then the subscriptions.
	const SUBSCRIPTIONS_STAGGER_START = quickLinks.length + 2;

	/** Entrance delay slot for the row at this position. */
	function stagger(index: number): number {
		return Math.min(index, MAX_STAGGER_INDEX);
	}

	/** Whether the link points at the current page. */
	function isActive(href: string): boolean {
		return $page.url.pathname.startsWith(href);
	}

	/** Whether a subscription's feed is the page being shown. */
	function isSubscriptionActive(id: string): boolean {
		return $page.url.pathname === subscriptionsUrl() && $page.url.searchParams.get('id') === id;
	}
</script>

<div class="drawer-menu flex min-h-0 flex-1 flex-col" class:open class:dragging>
	<a href="/" class="drawer-item flex items-center gap-2 px-5 pt-5 pb-3" style:--stagger={0}>
		<span class="text-primary"><FilmReelIcon size={26} weight="fill" /></span>
		<span class="text-lg font-semibold tracking-tight">Veil</span>
	</a>

	<div class="no-scrollbar flex min-h-0 flex-1 flex-col overflow-y-auto px-3 pb-3">
		<nav class="flex flex-col gap-0.5">
			{#each quickLinks as link, index (link.href)}
				<a
					href={link.href}
					class="drawer-item drawer-row"
					class:drawer-row-active={isActive(link.href)}
					style:--stagger={stagger(index + 1)}
				>
					<link.icon size={22} weight={isActive(link.href) ? 'fill' : 'regular'} />
					<span class="truncate">{link.label}</span>
				</a>
			{/each}
		</nav>

		<a
			href={subscriptionsUrl()}
			class="drawer-item text-base-content/50 mt-5 flex items-center gap-1 px-3 pb-2 text-xs font-semibold tracking-wider uppercase"
			style:--stagger={stagger(SUBSCRIPTIONS_STAGGER_START - 1)}
		>
			Subscriptions
			<CaretRightIcon size={12} weight="bold" />
		</a>
		<nav class="flex flex-col gap-0.5">
			{#each visibleSubscriptions as subscription, index (subscription.id)}
				<a
					href={subscriptionsUrl(subscription.id)}
					class="drawer-item drawer-row"
					class:drawer-row-active={isSubscriptionActive(subscription.id)}
					style:--stagger={stagger(SUBSCRIPTIONS_STAGGER_START + index)}
				>
					<SubscriptionAvatar {subscription} size={32} />
					<span class="flex-1 truncate">{subscriptionName(subscription)}</span>
					{#if subscription.newCount > 0}
						<span
							class="bg-primary text-primary-content min-w-5 shrink-0 rounded-full px-1.5 text-center text-[11px] leading-5 font-semibold"
						>
							{subscription.newCount}
						</span>
					{/if}
				</a>
			{:else}
				<div
					class="drawer-item bg-base-content/5 mx-1 flex flex-col items-start gap-2 rounded-2xl p-4"
					style:--stagger={stagger(SUBSCRIPTIONS_STAGGER_START)}
				>
					<BellSimpleIcon size={22} class="text-base-content/60" />
					<p class="text-base-content/60 text-sm">
						Subscribe to a channel, performer, tag or search to follow it here.
					</p>
					<a href="/categories" class="text-primary text-sm font-semibold">Browse categories</a>
				</div>
			{/each}
			{#if $subscriptionList.length > VISIBLE_SUBSCRIPTIONS}
				<a
					href={subscriptionsUrl()}
					class="drawer-item drawer-row text-base-content/60 text-sm"
					style:--stagger={stagger(SUBSCRIPTIONS_STAGGER_START + VISIBLE_SUBSCRIPTIONS)}
				>
					Show all {$subscriptionList.length}
				</a>
			{/if}
		</nav>
	</div>

	<footer
		class="drawer-item border-base-content/10 flex flex-col gap-0.5 border-t px-3 pt-2 pb-3"
		style:--stagger={stagger(MAX_STAGGER_INDEX)}
	>
		{#each footerLinks as link (link.href)}
			<a href={link.href} class="drawer-row" class:drawer-row-active={isActive(link.href)}>
				<link.icon size={22} weight={isActive(link.href) ? 'fill' : 'regular'} />
				<span class="truncate">{link.label}</span>
			</a>
		{/each}
		<div class="flex items-center justify-between px-3 pt-2">
			<VpnGuard showLabel />
			<span class="text-base-content/25 text-[10px] font-medium">v0.1.0</span>
		</div>
	</footer>
</div>

<style>
	.drawer-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		height: 3rem;
		padding-inline: 0.75rem;
		border-radius: 1rem;
		font-size: 15px;
		color: color-mix(in oklab, var(--color-base-content) 85%, transparent);
		transition: background-color 180ms ease-out;
	}

	/* Same pill as the tab bar's active tab. */
	.drawer-row-active {
		background-color: color-mix(in oklab, var(--color-base-content) 10%, transparent);
		color: var(--color-base-content);
		font-weight: 600;
	}

	/* Closed: rows sit slightly left and hidden, and leave without delay. */
	.drawer-item {
		opacity: 0;
		translate: -12px 0;
		transition:
			opacity 160ms ease-out,
			translate 160ms ease-out;
	}

	/* Open: each row follows the previous one by 25 ms once the panel is mostly in. */
	.open .drawer-item {
		opacity: 1;
		translate: 0 0;
		transition:
			opacity 320ms ease-out,
			translate 420ms cubic-bezier(0.22, 1, 0.36, 1);
		transition-delay: calc(90ms + var(--stagger) * 25ms);
	}

	.dragging .drawer-item {
		opacity: 1;
		translate: 0 0;
		transition: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.drawer-item,
		.open .drawer-item {
			opacity: 1;
			translate: 0 0;
			transition: none;
		}
	}
</style>
