<!--
	The cam rooms of the person in a video, one line each: the room's picture,
	its name and site, and whether it's live right now (with its audience and
	a Watch live button) or when it was last seen.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import { cacheUrl } from '$lib/img';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { camUrl } from '$lib/routes';
	import type { CamRow } from '$lib/cams';

	interface Props {
		rooms: CamRow[];
	}

	let { rooms }: Props = $props();

	/** When the offline room was last seen, or that it's offline. */
	function offlineLabel(room: CamRow): string {
		if (room.lastSeenOnlineAt) return `Offline · seen ${dayjs(room.lastSeenOnlineAt).fromNow()}`;
		return 'Offline';
	}
</script>

<div class="flex flex-col gap-2">
	{#each rooms as room (room.id)}
		<a
			href={camUrl(room.id)}
			class="bg-base-200 flex items-center gap-3 rounded-xl p-2 pr-3 hover:brightness-125"
		>
			<span class="bg-base-300 relative aspect-video w-24 shrink-0 overflow-hidden rounded-lg">
				{#if cacheUrl(room.live?.thumbnail?.uri ?? room.thumbnail?.uri)}
					<img
						src={cacheUrl(room.live?.thumbnail?.uri ?? room.thumbnail?.uri)}
						alt=""
						loading="lazy"
						class="h-full w-full object-cover"
						class:opacity-50={!room.live}
					/>
				{:else}
					<span class="flex h-full w-full items-center justify-center">
						<VideoCameraIcon size={20} class="opacity-30" />
					</span>
				{/if}
			</span>
			<span class="flex min-w-0 flex-1 flex-col">
				<span class="truncate text-sm font-medium">
					{room.displayName}
					<span class="text-base-content/50 font-normal"
						>· {siteDisplayName(room.plugin, $sites)}</span
					>
				</span>
				{#if room.live}
					<span class="text-error flex items-center gap-1.5 text-xs font-medium">
						<span class="bg-error h-1.5 w-1.5 animate-pulse rounded-full"></span>
						Live now
						{#if room.live.viewers > 0}
							<span class="text-base-content/60 font-normal">
								· {room.live.viewers.toLocaleString()} watching
							</span>
						{/if}
					</span>
				{:else}
					<span class="text-base-content/50 text-xs">{offlineLabel(room)}</span>
				{/if}
			</span>
			{#if room.live}
				<span class="btn btn-sm btn-primary shrink-0">Watch live</span>
			{/if}
		</a>
	{/each}
</div>
