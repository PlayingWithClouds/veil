<script lang="ts">
	// Pulsing placeholders for the channel row and description panel while the
	// scene's first-visit detail fetch runs.
	interface Props {
		variant: 'channel' | 'description';
	}

	let { variant }: Props = $props();
</script>

{#if variant === 'channel'}
	<div class="flex animate-pulse items-center gap-3" aria-hidden="true">
		<span class="bg-base-300 h-10 w-10 rounded-full"></span>
		<span class="flex flex-col gap-1.5">
			<span class="bg-base-300 h-4 w-36 rounded"></span>
			<span class="bg-base-300 h-3 w-20 rounded"></span>
		</span>
	</div>
{:else}
	<div
		class="bg-base-200 flex animate-pulse flex-col gap-4 rounded-xl p-4"
		role="status"
		aria-busy="true"
		aria-label="Loading scene details"
	>
		<span class="bg-base-300 h-4 w-64 max-w-full rounded"></span>
		<span class="flex gap-2">
			<span class="bg-base-300 h-9 w-32 rounded-full"></span>
			<span class="bg-base-300 h-9 w-28 rounded-full"></span>
		</span>
		<span class="flex gap-1.5">
			<span class="bg-base-300 h-6 w-16 rounded-full"></span>
			<span class="bg-base-300 h-6 w-20 rounded-full"></span>
			<span class="bg-base-300 h-6 w-14 rounded-full"></span>
		</span>
		<span class="flex flex-col gap-2">
			<span class="bg-base-300 h-3 w-full rounded"></span>
			<span class="bg-base-300 h-3 w-5/6 rounded"></span>
			<span class="bg-base-300 h-3 w-2/3 rounded"></span>
		</span>
	</div>
{/if}
