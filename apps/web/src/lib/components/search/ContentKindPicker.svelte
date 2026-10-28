<!--
	"What are you looking for?": the kinds of content the installed sites carry
	(videos, cam shows, JAV, photos), so a search goes to the sites that have it
	without the reader knowing which plugin does. Each chip names its sites on
	hover. Hidden when the sites carry only one kind.
-->
<script lang="ts">
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { CONTENT_KINDS, carriesContent, type ContentKind, type SearchPlugin } from '$lib/search';

	interface Props {
		// The sites a search can ask (enabled, offering the media type shown).
		sites: SearchPlugin[];
		// The picked kind, null for anything.
		selected: ContentKind | null;
		onselect: (kind: ContentKind | null) => void;
	}

	let { sites, selected, onselect }: Props = $props();

	// The kinds some site carries, each with the names of those sites.
	let offered = $derived(
		CONTENT_KINDS.map((entry) => ({
			...entry,
			siteNames: sites
				.filter((site) => carriesContent(site, entry.kind))
				.map((site) => site.displayName || site.name)
		})).filter((entry) => entry.siteNames.length > 0)
	);

	/** Picks a kind, or goes back to anything when it is the picked one. */
	function toggle(kind: ContentKind) {
		if (selected === kind) {
			onselect(null);
			return;
		}
		onselect(kind);
	}
</script>

{#if offered.length > 1}
	<div class="flex flex-wrap items-center gap-2" role="group" aria-label="Kind of content">
		<span class="text-base-content/60 mr-1 text-sm">
			{#if selected === null}
				What are you looking for?
			{:else}
				Looking for
			{/if}
		</span>
		{#each offered as entry (entry.kind)}
			<button
				type="button"
				class="kind-chip"
				class:kind-chip-active={selected === entry.kind}
				aria-pressed={selected === entry.kind}
				title="Searches {entry.siteNames.join(', ')}"
				onclick={() => toggle(entry.kind)}
			>
				{entry.label}
				{#if selected === entry.kind}
					<XIcon size={11} weight="bold" />
				{/if}
			</button>
		{/each}
	</div>
{/if}

<style>
	.kind-chip {
		display: flex;
		align-items: center;
		gap: 0.3rem;
		height: 1.75rem;
		padding-inline: 0.75rem;
		border-radius: 9999px;
		border: 1px solid var(--color-base-300);
		font-size: 13px;
		font-weight: 500;
		transition:
			background-color 150ms,
			border-color 150ms;
	}

	.kind-chip:hover {
		background-color: var(--color-base-200);
	}

	.kind-chip-active,
	.kind-chip-active:hover {
		border-color: var(--color-base-content);
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}
</style>
