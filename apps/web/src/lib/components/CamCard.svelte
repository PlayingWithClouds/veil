<!--
	Card for a live cam room: thumbnail with a LIVE badge and viewer count, the
	room's name with the site icon, and its first tags. Offline rooms (the
	Following row) are dimmed and say when they were last seen.
-->
<script lang="ts">
	import EyeIcon from 'phosphor-svelte/lib/EyeIcon';
	import RecordIcon from 'phosphor-svelte/lib/RecordIcon';
	import SplitAddButton from './SplitAddButton.svelte';
	import { activeCamRecordings } from '$lib/stores/camRecordings';
	import { cacheUrl } from '$lib/img';
	import { formatViewers, type CamModel, type CamSource } from '$lib/cams';
	import { formatRelativeTime } from '$lib/time';

	interface Props {
		camModel: CamModel;
		// The room's site, for the icon next to its name.
		source?: CamSource | null;
		onclick: () => void;
		// Called with a tag name when one of the room's tag chips is pressed.
		ontag?: (tag: string) => void;
	}

	let { camModel, source = null, onclick, ontag }: Props = $props();

	const VISIBLE_TAGS = 3;

	let thumbnail = $derived(cacheUrl(camModel.thumbnailUrl));
	let iconUrl = $derived.by(() => {
		if (!source?.iconUrl) return null;
		return cacheUrl(source.iconUrl);
	});
	let isRecording = $derived(
		$activeCamRecordings.some((entry) => entry.camModelId === camModel.id)
	);
	let visibleTags = $derived(camModel.tags.slice(0, VISIBLE_TAGS));
	let lastSeenLabel = $derived.by(() => {
		if (camModel.online || !camModel.lastSeenOnlineAt) return null;
		return `Seen ${formatRelativeTime(camModel.lastSeenOnlineAt)}`;
	});
</script>

<div class="group flex w-full flex-col gap-2">
	<div class="relative">
		<button
			type="button"
			{onclick}
			class="tv-card bg-base-200 relative block aspect-video w-full overflow-hidden rounded-xl"
			aria-label="Watch {camModel.displayName}"
		>
			{#if thumbnail}
				<img
					src={thumbnail}
					alt=""
					loading="lazy"
					class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
					class:opacity-40={!camModel.online}
				/>
			{/if}
			{#if camModel.online}
				<span
					class="absolute top-2 left-2 flex items-center gap-1 rounded bg-red-600 px-1.5 py-0.5 text-[11px] font-bold tracking-wide text-white uppercase"
				>
					<span class="h-1.5 w-1.5 rounded-full bg-white"></span>
					Live
				</span>
				<span
					class="absolute right-2 bottom-2 flex items-center gap-1 rounded bg-black/70 px-1.5 py-0.5 text-xs font-medium text-white"
				>
					<EyeIcon size={12} weight="fill" />
					{formatViewers(camModel.viewers)}
				</span>
			{:else}
				<span
					class="absolute top-2 left-2 rounded bg-black/70 px-1.5 py-0.5 text-[11px] font-semibold tracking-wide text-white uppercase"
				>
					Offline
				</span>
			{/if}
			{#if isRecording}
				<span
					class="absolute top-2 right-2 flex items-center gap-1 rounded bg-black/70 px-1.5 py-0.5 text-[11px] font-bold tracking-wide text-white uppercase"
				>
					<RecordIcon size={12} weight="fill" class="text-red-500" />
					Rec
				</span>
			{:else if camModel.autoRecord}
				<span
					class="absolute top-2 right-2 flex items-center gap-1 rounded bg-black/70 px-1.5 py-0.5 text-[11px] font-semibold text-white/80"
					title="Recorded automatically when live"
				>
					<RecordIcon size={12} class="text-red-400" />
					Auto
				</span>
			{/if}
		</button>
		{#if camModel.online}
			<SplitAddButton
				kind="cam"
				mediaId={camModel.id}
				title={camModel.displayName}
				class="absolute bottom-2 left-2"
			/>
		{/if}
	</div>

	<div class="flex min-w-0 flex-col gap-1.5">
		<button type="button" {onclick} class="flex min-w-0 items-center gap-2 text-left">
			{#if iconUrl}
				<img src={iconUrl} alt={source?.displayName ?? ''} class="h-4 w-4 shrink-0 rounded-sm" />
			{/if}
			<span class="line-clamp-1 text-sm font-semibold">{camModel.displayName}</span>
			{#if lastSeenLabel}
				<span class="text-base-content/40 shrink-0 text-xs">{lastSeenLabel}</span>
			{/if}
		</button>
		{#if visibleTags.length > 0}
			<div class="flex flex-wrap gap-1.5">
				{#each visibleTags as tag (tag)}
					<button
						type="button"
						class="border-base-300 bg-base-200 text-base-content/70 hover:text-base-content rounded-full border px-2.5 py-0.5 text-xs transition-colors"
						onclick={() => ontag?.(tag)}
					>
						{tag}
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>
