<script lang="ts">
	import { goto } from '$app/navigation';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import { tagUrl } from '$lib/routes';
	import { tagLabel, type TagRef } from '$lib/tags';

	let {
		tags,
		inherited = false,
		onblock
	}: {
		tags: TagRef[];
		// Muted styling for tags that came from a studio/performer, not the record.
		inherited?: boolean;
		// When set, a block button appears on each chip.
		onblock?: (tag: TagRef) => void;
	} = $props();
</script>

{#if tags.length > 0}
	<div class="flex flex-wrap items-center gap-1.5">
		{#each tags as tag (tag.id)}
			<span
				class="group/chip flex items-center overflow-hidden rounded-full border text-xs {inherited
					? 'border-base-300 text-base-content/40'
					: 'border-base-300 bg-base-200 text-base-content/70'}"
			>
				<button
					type="button"
					class="hover:text-base-content px-2.5 py-1 transition-colors"
					title={inherited ? `${tag.name} (via studio/performer)` : tag.name}
					onclick={() => goto(tagUrl(tag.id))}
				>
					{tagLabel(tag.name)}
				</button>
				{#if onblock}
					<button
						type="button"
						class="hover:text-error hidden pr-2 text-current/40 group-hover/chip:block"
						aria-label={`Block ${tag.name}`}
						onclick={() => onblock(tag)}
					>
						<ProhibitIcon size={12} />
					</button>
				{/if}
			</span>
		{/each}
	</div>
{/if}
