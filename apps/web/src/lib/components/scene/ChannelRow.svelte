<script lang="ts">
	// YouTube-style "channel" block under the title: the studio's avatar, name and
	// scene count, linking to its page, with a subscribe pill. Scenes without a
	// studio use their first performer as the channel, and only with neither the
	// site (which the description names anyway).
	import { cacheUrl } from '$lib/img';
	import { performerUrl, studioUrl } from '$lib/routes';
	import type { PerformerRef, SceneSite, StudioRef } from '$lib/sceneDetail';
	import SubscribeButton from '$lib/components/SubscribeButton.svelte';
	import type { FollowableKind } from '$lib/searchSubscriptions';
	import { formatCount, initials } from './format';
	import SiteIcon from './SiteIcon.svelte';

	interface Props {
		studio: StudioRef | null;
		performers: PerformerRef[];
		site: SceneSite | null;
	}

	let { studio, performers, site }: Props = $props();

	// What the row links to and subscribes to: the studio, else the first performer.
	let channel = $derived(channelFor(studio, performers));

	interface Channel {
		kind: FollowableKind;
		id: string;
		name: string;
		href: string;
		imagePath: string | null;
		sceneCount: number;
	}

	/** The studio as the channel, else the first credited performer, else none. */
	function channelFor(currentStudio: StudioRef | null, cast: PerformerRef[]): Channel | null {
		if (currentStudio) {
			return {
				kind: 'STUDIO',
				id: currentStudio.id,
				name: currentStudio.name,
				href: studioUrl(currentStudio.id),
				imagePath: currentStudio.imagePath,
				sceneCount: currentStudio.sceneCount
			};
		}
		const performer = cast[0];
		if (!performer) return null;
		let sceneCount = 0;
		if (performer.sceneCount) sceneCount = performer.sceneCount;
		return {
			kind: 'PERFORMER',
			id: performer.id,
			name: performer.name,
			href: performerUrl(performer.id),
			imagePath: performer.imagePath,
			sceneCount
		};
	}

	/** Label for a channel's scene total, e.g. "1.2K scenes". */
	function sceneCountLabel(count: number): string {
		if (count === 1) return '1 scene';
		return `${formatCount(count)} scenes`;
	}
</script>

{#if channel}
	<div class="flex min-w-0 items-center gap-6">
		<a href={channel.href} class="group flex min-w-0 items-center gap-3">
			<span
				class="bg-base-300 flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full text-sm font-semibold"
			>
				{#if cacheUrl(channel.imagePath)}
					<img
						src={cacheUrl(channel.imagePath)}
						alt=""
						loading="lazy"
						class="h-full w-full object-cover object-top"
					/>
				{:else}
					{initials(channel.name)}
				{/if}
			</span>
			<span class="flex min-w-0 flex-col">
				<span class="truncate font-semibold capitalize group-hover:underline">{channel.name}</span>
				{#if channel.sceneCount > 0}
					<span class="text-base-content/50 text-xs">{sceneCountLabel(channel.sceneCount)}</span>
				{/if}
			</span>
		</a>
		<SubscribeButton kind={channel.kind} targetId={channel.id} targetName={channel.name} />
	</div>
{:else if site}
	<div class="flex min-w-0 items-center gap-3">
		<span class="bg-base-300 flex h-10 w-10 shrink-0 items-center justify-center rounded-full">
			<SiteIcon iconUrl={site.iconUrl} class="h-5 w-5" />
		</span>
		<span class="truncate font-semibold">{site.name}</span>
	</div>
{/if}
