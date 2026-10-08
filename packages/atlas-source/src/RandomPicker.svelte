<script lang="ts">
	// Opens a session of randomly sampled veil images.
	import type { Project } from '@atlas/contracts';

	let { open }: { project: Project; open: (params: Record<string, unknown>) => Promise<void> } = $props();

	let count = $state(20);
	let opening = $state(false);
	let error = $state('');

	async function openRandom() {
		opening = true;
		error = '';
		try {
			await open({ count });
		} catch (failure) {
			error = failure instanceof Error ? failure.message : String(failure);
			opening = false;
		}
	}
</script>

{#if error}<div class="alert alert-error mb-4 text-sm">{error}</div>{/if}

<div class="flex items-center gap-2">
	<label class="text-dim text-sm" for="veil-random-count">Images</label>
	<input id="veil-random-count" class="input input-sm w-24" type="number" min="1" max="500" bind:value={count} />
	<button type="button" class="btn btn-sm btn-primary" disabled={opening || count < 1} onclick={openRandom}>
		{opening ? 'Sampling…' : 'Open random sample'}
	</button>
</div>
