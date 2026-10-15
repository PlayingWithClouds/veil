<script lang="ts">
	import { page } from '$app/state';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import EyeIcon from 'phosphor-svelte/lib/EyeIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import CamPlayer from '$lib/components/CamPlayer.svelte';
	import {
		fetchCamModel,
		fetchCamSources,
		fetchCamStream,
		formatViewers,
		setCamModelFollowed,
		type CamModel,
		type CamSource
	} from '$lib/cams';
	import { cacheUrl } from '$lib/img';
	import { camsUrl } from '$lib/routes';
	import { pluginIconUrl } from '$lib/server';
	import { notifications } from '$lib/stores/notifications';
	import { formatRelativeTime } from '$lib/time';

	let camModel = $state<CamModel | null>(null);
	let source = $state<CamSource | null>(null);
	let streamUrl = $state<string | null>(null);
	let loading = $state(true);
	let notFound = $state(false);
	let offline = $state(false);
	let followBusy = $state(false);
	// Bumped to start playback again (retry after the room went offline).
	let attempt = $state(0);

	let roomId = $derived(page.params.id ?? '');
	let thumbnail = $derived(cacheUrl(camModel?.thumbnailUrl));
	let sourceIcon = $derived.by(() => {
		if (!source?.iconUrl) return null;
		return cacheUrl(source.iconUrl);
	});

	// Load the model and, when it is live, its stream.
	$effect(() => {
		void attempt;
		const id = roomId;
		let cancelled = false;
		loading = true;
		offline = false;
		notFound = false;
		streamUrl = null;
		start(id, () => cancelled);
		return () => {
			cancelled = true;
		};
	});

	/** Loads the model, then asks for its live stream. */
	async function start(id: string, isCancelled: () => boolean) {
		const loaded = await fetchCamModel(id);
		if (isCancelled()) return;
		if (!loaded) {
			notFound = true;
			loading = false;
			return;
		}
		camModel = loaded;
		const sources = await fetchCamSources();
		if (isCancelled()) return;
		source = sources.find((entry) => entry.name === loaded.pluginName) ?? null;
		if (source) source = { ...source, iconUrl: pluginIconUrl(source.iconUrl) };
		if (!loaded.online) {
			offline = true;
			loading = false;
			return;
		}
		try {
			const stream = await fetchCamStream(id);
			if (isCancelled()) return;
			streamUrl = stream.url;
		} catch {
			if (isCancelled()) return;
			offline = true;
		}
		loading = false;
	}

	/** Shows the offline state once the stream stops. */
	function onStreamEnded() {
		streamUrl = null;
		offline = true;
	}

	/** Follows or unfollows the model. */
	async function toggleFollow() {
		if (!camModel || followBusy) return;
		followBusy = true;
		const wanted = !camModel.followed;
		try {
			const updated = await setCamModelFollowed(camModel.id, wanted);
			camModel = { ...camModel, followed: updated.followed };
		} catch {
			notifications.push('Could not update follow', 'error');
		} finally {
			followBusy = false;
		}
	}
</script>

<div class="flex flex-col gap-4">
	<a href={camsUrl()} class="text-base-content/50 hover:text-base-content w-fit text-sm">
		&larr; All cams
	</a>

	{#if notFound}
		<p class="text-base-content/40 py-24 text-center">This room is not known.</p>
	{:else}
		<div class="bg-base-200 relative aspect-video w-full overflow-hidden rounded-xl">
			{#if streamUrl}
				{#key streamUrl}
					<CamPlayer {streamUrl} onended={onStreamEnded} />
				{/key}
			{:else if offline}
				{#if thumbnail}
					<img src={thumbnail} alt="" class="absolute inset-0 h-full w-full object-cover opacity-20" />
				{/if}
				<div class="absolute inset-0 flex flex-col items-center justify-center gap-3 text-center">
					<p class="text-lg font-semibold">{camModel?.displayName} is offline</p>
					{#if camModel?.lastSeenOnlineAt}
						<p class="text-base-content/50 text-sm">
							Last seen {formatRelativeTime(camModel.lastSeenOnlineAt)}
						</p>
					{/if}
					<button type="button" class="btn btn-sm" onclick={() => (attempt += 1)}>
						<ArrowClockwiseIcon size={16} />
						Check again
					</button>
				</div>
			{:else if loading}
				<div class="absolute inset-0 flex items-center justify-center">
					<span class="loading loading-spinner text-base-content/40"></span>
				</div>
			{/if}
		</div>

		{#if camModel}
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex min-w-0 flex-col gap-2">
					<div class="flex items-center gap-2">
						{#if sourceIcon}
							<img src={sourceIcon} alt={source?.displayName ?? ''} class="h-5 w-5 rounded-sm" />
						{/if}
						<h1 class="truncate text-2xl font-bold">{camModel.displayName}</h1>
						{#if streamUrl}
							<span
								class="flex items-center gap-1 rounded bg-red-600 px-1.5 py-0.5 text-[11px] font-bold tracking-wide text-white uppercase"
							>
								<span class="h-1.5 w-1.5 rounded-full bg-white"></span>
								Live
							</span>
						{/if}
					</div>
					{#if camModel.subject}
						<p class="text-base-content/70 text-sm">{camModel.subject}</p>
					{/if}
					<div class="text-base-content/50 flex items-center gap-3 text-sm">
						{#if streamUrl}
							<span class="flex items-center gap-1">
								<EyeIcon size={14} weight="fill" />
								{formatViewers(camModel.viewers)} watching
							</span>
						{/if}
						<a
							href={camModel.sourceUrl}
							target="_blank"
							rel="noreferrer noopener"
							class="hover:text-base-content flex items-center gap-1"
						>
							<ArrowSquareOutIcon size={14} />
							Open on {source?.displayName ?? 'site'}
						</a>
					</div>
					{#if camModel.tags.length > 0}
						<div class="flex flex-wrap gap-1.5">
							{#each camModel.tags as tag (tag)}
								<a
									href="{camsUrl()}?tag={encodeURIComponent(tag)}"
									class="border-base-300 bg-base-200 text-base-content/70 hover:text-base-content rounded-full border px-2.5 py-1 text-xs transition-colors"
								>
									{tag}
								</a>
							{/each}
						</div>
					{/if}
				</div>
				<button
					type="button"
					class="btn"
					class:btn-primary={!camModel.followed}
					disabled={followBusy}
					onclick={toggleFollow}
				>
					<HeartIcon size={18} weight={camModel.followed ? 'fill' : 'regular'} />
					{camModel.followed ? 'Following' : 'Follow'}
				</button>
			</div>
		{/if}
	{/if}
</div>
