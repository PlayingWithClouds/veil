<script lang="ts">
	// Playback sources grouped per site: one row per site, its qualities as chips
	// (best first). Picking a chip plays that source.
	import SealCheckIcon from 'phosphor-svelte/lib/SealCheckIcon';
	import { formatSpeed } from '$lib/tmdb';
	import { qualityLabel, type SceneStream, type StreamGroup } from '$lib/sceneDetail';
	import SiteIcon from './SiteIcon.svelte';
	import { sites, siteDisplayName } from '$lib/stores/sites';

	interface Props {
		groups: StreamGroup[];
		pluginIcons: Record<string, string | null>;
		// Source currently playing, highlighted.
		activeSourceId: string | null;
		// Source being resolved, shown as loading.
		resolvingId: string | null;
		onplay: (stream: SceneStream) => void;
	}

	let { groups, pluginIcons, activeSourceId, resolvingId, onplay }: Props = $props();
</script>

<ul class="flex flex-col gap-2">
	{#each groups as group (group.provider)}
		<li class="flex flex-wrap items-center gap-2">
			<span class="flex w-32 shrink-0 items-center gap-1.5 truncate text-sm font-medium">
				<SiteIcon iconUrl={pluginIcons[group.provider]} />
				{siteDisplayName(group.provider, $sites)}
			</span>
			{#each group.streams as stream (stream.id)}
				<button
					type="button"
					class="border-base-content/10 hover:bg-white/10 flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs transition-colors"
					class:ring-1={activeSourceId === stream.id}
					class:ring-white={activeSourceId === stream.id}
					aria-pressed={activeSourceId === stream.id}
					onclick={() => onplay(stream)}
				>
					{#if resolvingId === stream.id}
						<span class="loading loading-xs"></span>
					{/if}
					<span class="font-medium">{qualityLabel(stream)}</span>
					{#if stream.verified}
						<SealCheckIcon size={12} weight="fill" class="text-success" />
					{/if}
					{#if formatSpeed(stream.expectedSpeedBps)}
						<span class="text-success">~{formatSpeed(stream.expectedSpeedBps)}</span>
					{/if}
				</button>
			{/each}
		</li>
	{/each}
</ul>
