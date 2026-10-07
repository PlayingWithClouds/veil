<!--
	Sticky tab bar of a performer/studio page (Videos · Galleries · About). It
	sticks under the phone top bar (top-chrome) and a pill slides under the
	active tab, like the bottom tab bar.
-->
<script module lang="ts">
	export type EntityTab = { key: string; label: string };
</script>

<script lang="ts">
	interface Props {
		tabs: EntityTab[];
		active: string;
		onchange: (key: string) => void;
	}

	let { tabs, active, onchange }: Props = $props();

	let activeIndex = $derived(
		Math.max(
			0,
			tabs.findIndex((tab) => tab.key === active)
		)
	);
</script>

<div class="entity-tabs bg-base-100 top-chrome sticky z-20 -mx-4 px-4 py-2 max-md:-mx-3 max-md:px-3">
	<div class="bg-base-200 relative flex rounded-full p-1" role="tablist">
		<span
			class="entity-tabs-pill bg-base-content absolute inset-y-1 left-1 rounded-full"
			style:width="calc((100% - 0.5rem) / {tabs.length})"
			style:transform="translateX({activeIndex * 100}%)"
		></span>
		{#each tabs as tab (tab.key)}
			<button
				type="button"
				role="tab"
				aria-selected={tab.key === active}
				onclick={() => onchange(tab.key)}
				class="text-base-content/70 relative z-10 flex h-9 flex-1 items-center justify-center text-sm font-semibold transition-colors"
				class:entity-tab-active={tab.key === active}
			>
				{tab.label}
			</button>
		{/each}
	</div>
</div>

<style>
	.entity-tabs-pill {
		transition: transform 380ms cubic-bezier(0.34, 1.3, 0.64, 1);
	}

	.entity-tab-active {
		color: var(--color-base-100);
	}

	@media (prefers-reduced-motion: reduce) {
		.entity-tabs-pill {
			transition: none;
		}
	}
</style>
