<script lang="ts">
	// Stand-in for the player before playback starts: the preview frame with a
	// round glass play button and the length, or the source-resolution status
	// when nothing is playable yet.
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';

	interface Props {
		imageUrl: string | null;
		title: string;
		// "25:41", shown bottom-right when known.
		durationLabel?: string | null;
		canPlay: boolean;
		// True while the best source is being resolved.
		loading: boolean;
		// True while the first-visit fetch is still looking for sources.
		ensuring: boolean;
		onplay: () => void;
		// A back button over the poster, for pages that hide the top bar.
		onback?: () => void;
	}

	let { imageUrl, title, durationLabel = null, canPlay, loading, ensuring, onplay, onback }: Props = $props();
</script>

<div class="relative aspect-video w-full">
	{#if imageUrl}
		<img src={imageUrl} alt={title} class="h-full w-full object-cover" />
	{/if}
	<div class="absolute inset-0 flex items-center justify-center bg-black/25">
		{#if canPlay}
			<button
				type="button"
				class="glass-round flex h-16 w-16 items-center justify-center rounded-full text-white transition-transform hover:scale-105"
				onclick={onplay}
				aria-label="Play"
			>
				{#if loading}
					<span class="loading loading-spinner loading-md"></span>
				{:else}
					<PlayIcon size={28} weight="fill" />
				{/if}
			</button>
		{:else if ensuring}
			<span class="flex items-center gap-2 text-sm text-white/60">
				<span class="loading loading-sm"></span>
				Resolving sources…
			</span>
		{:else}
			<span class="text-sm text-white/50">No playable source</span>
		{/if}
	</div>
	{#if onback}
		<button
			type="button"
			class="glass-round absolute top-2 left-2 flex h-9 w-9 items-center justify-center rounded-full text-white"
			onclick={onback}
			aria-label="Back"
		>
			<ArrowLeftIcon size={18} weight="bold" />
		</button>
	{/if}
	{#if durationLabel}
		<span
			class="absolute right-2 bottom-2 rounded bg-black/75 px-1.5 py-0.5 text-xs font-medium text-white tabular-nums"
		>
			{durationLabel}
		</span>
	{/if}
</div>
