<!--
	Phone tab bar: the most used destinations within thumb reach. Everything else
	stays in the drawer opened from the top bar.
-->
<script lang="ts">
	import { page } from '$app/stores';
	import { subscriptionsUrl } from '$lib/routes';
	import { newSubscriptionScenes } from '$lib/stores/newSubscriptionScenes';
	import { activeTabIndex, isTabActive, tabs, type Tab } from '$lib/tabs';

	/** Whether the tab points at the current page. */
	function isActive(tab: Tab): boolean {
		return isTabActive(tab, $page.url.pathname);
	}

	// -1 on pages without a tab (settings, detail pages): the pill fades out.
	let activeIndex = $derived(activeTabIndex($page.url.pathname));
</script>

<nav class="bottom-nav z-overlay fixed inset-x-3 flex rounded-[1.75rem] p-1.5" aria-label="Main">
	<span
		class="bottom-nav-pill bg-base-content/10 absolute inset-y-1.5 left-1.5 rounded-[1.375rem]"
		class:opacity-0={activeIndex === -1}
		style:width="calc((100% - 0.75rem) / {tabs.length})"
		style:transform="translateX({Math.max(activeIndex, 0) * 100}%)"
	></span>
	{#each tabs as tab (tab.href)}
		<a
			href={tab.href}
			class="text-base-content/60 relative flex h-13 flex-1 flex-col items-center justify-center gap-0.5 text-[10px] transition-colors"
			class:bottom-tab-active={isActive(tab)}
		>
			<span class="bottom-tab-icon flex">
				<tab.icon size={22} weight={isActive(tab) ? 'fill' : 'regular'} />
			</span>
			<span class="max-w-full truncate px-1">{tab.label}</span>
			{#if tab.href === subscriptionsUrl() && $newSubscriptionScenes > 0}
				<span class="bg-primary absolute top-2 left-1/2 ml-2 h-2 w-2 rounded-full"></span>
			{/if}
		</a>
	{/each}
</nav>

<style>
	/* Frosted glass: the feed scrolls under the bar and shows through blurred. */
	.bottom-nav {
		bottom: calc(0.75rem + var(--safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
		background-color: color-mix(in oklab, var(--color-base-200) 65%, transparent);
		backdrop-filter: blur(20px) saturate(1.8);
		border: 1px solid color-mix(in oklab, var(--color-base-content) 10%, transparent);
		box-shadow:
			0 10px 30px rgba(0, 0, 0, 0.45),
			inset 0 1px 0 color-mix(in oklab, var(--color-base-content) 8%, transparent);
	}

	.bottom-nav-pill {
		transition:
			transform 420ms cubic-bezier(0.34, 1.3, 0.64, 1),
			opacity 200ms ease-out;
	}

	.bottom-tab-active {
		color: var(--color-base-content);
		font-weight: 600;
	}

	/* Pops once when a tab becomes active. */
	.bottom-tab-active .bottom-tab-icon {
		animation: tab-pop 380ms cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	@keyframes tab-pop {
		0% {
			scale: 0.8;
		}
		100% {
			scale: 1;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.bottom-nav-pill {
			transition: none;
		}
		.bottom-tab-active .bottom-tab-icon {
			animation: none;
		}
	}
</style>
