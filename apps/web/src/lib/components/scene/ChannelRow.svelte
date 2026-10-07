<script lang="ts">
	// YouTube-style "channel" block under the title: the studio's avatar, name and
	// scene count, linking to its page, with a subscribe pill. Scenes without a
	// studio show their site.
	import { cacheUrl } from '$lib/img';
	import { studioUrl } from '$lib/routes';
	import type { SceneSite, StudioRef } from '$lib/sceneDetail';
	import SubscribeButton from '$lib/components/SubscribeButton.svelte';
	import { formatCount } from './format';
	import SiteIcon from './SiteIcon.svelte';

	interface Props {
		studio: StudioRef | null;
		site: SceneSite | null;
	}

	let { studio, site }: Props = $props();

	let avatarUrl = $derived(cacheUrl(studio?.imagePath));

	/** Label for a studio's scene total, e.g. "1.2K scenes". */
	function sceneCountLabel(count: number): string {
		if (count === 1) return '1 scene';
		return `${formatCount(count)} scenes`;
	}
</script>

{#if studio}
	<div class="flex min-w-0 items-center gap-6">
		<a href={studioUrl(studio.id)} class="group flex min-w-0 items-center gap-3">
			<span
				class="bg-base-300 flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full text-base font-semibold uppercase"
			>
				{#if avatarUrl}
					<img src={avatarUrl} alt="" loading="lazy" class="h-full w-full object-cover" />
				{:else}
					{studio.name.charAt(0)}
				{/if}
			</span>
			<span class="flex min-w-0 flex-col">
				<span class="truncate font-semibold group-hover:underline">{studio.name}</span>
				{#if studio.sceneCount > 0}
					<span class="text-base-content/50 text-xs">{sceneCountLabel(studio.sceneCount)}</span>
				{/if}
			</span>
		</a>
		<SubscribeButton kind="STUDIO" targetId={studio.id} targetName={studio.name} />
	</div>
{:else if site}
	<div class="flex min-w-0 items-center gap-3">
		<span class="bg-base-300 flex h-10 w-10 shrink-0 items-center justify-center rounded-full">
			<SiteIcon iconUrl={site.iconUrl} class="h-5 w-5" />
		</span>
		<span class="truncate font-semibold">{site.name}</span>
	</div>
{/if}
