<!--
	Merging two performers in a dialog: find the other one by name (or start
	from a suggested duplicate), see both side by side, pick which one stays,
	confirm. The page offers Undo afterwards.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import ArrowsLeftRightIcon from 'phosphor-svelte/lib/ArrowsLeftRightIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import { cacheUrl } from '$lib/img';
	import { fetchPerformers, type MemberRow } from '$lib/library';
	import { mergePerformers } from '$lib/performers';

	interface Props {
		// The performer whose page this is.
		performer: MemberRow;
		// A suggested duplicate to start from.
		other?: MemberRow | null;
		// Called with who stayed, who was merged in, and the merge to undo.
		onmerged: (kept: MemberRow, merged: MemberRow, mergeId: string | null) => void;
		onclose: () => void;
	}

	let { performer, other = null, onmerged, onclose }: Props = $props();

	const SEARCH_LIMIT = 8;

	let query = $state('');
	let results = $state<MemberRow[]>([]);
	// Starts from the suggested duplicate; picking someone else replaces it.
	let chosen = $state<MemberRow | null>(untrack(() => other));
	let keepThisOne = $state(true);
	let merging = $state(false);
	let error = $state<string | null>(null);

	let kept = $derived(keepThisOne ? performer : chosen);
	let folded = $derived(keepThisOne ? chosen : performer);

	$effect(() => {
		const text = query.trim();
		if (text === '') {
			results = [];
			return;
		}
		fetchPerformers(text, SEARCH_LIMIT + 1).then((found) => {
			if (query.trim() !== text) return;
			results = found.filter((candidate) => candidate.id !== performer.id).slice(0, SEARCH_LIMIT);
		});
	});

	/** Opens the dialog as a modal once it is in the page. */
	function showModal(dialog: HTMLDialogElement) {
		dialog.showModal();
	}

	/** Merges the folded performer into the kept one. */
	async function merge() {
		if (!kept || !folded) return;
		merging = true;
		error = null;
		try {
			const mergeId = await mergePerformers(kept.id, folded.id);
			onmerged(kept, folded, mergeId);
		} catch (failure) {
			error = String(failure);
			merging = false;
		}
	}
</script>

{#snippet side(member: MemberRow, role: string)}
	<div class="bg-base-200 flex flex-1 flex-col items-center gap-2 rounded-xl p-4 text-center">
		<span class="text-base-content/50 text-xs font-semibold tracking-wide uppercase">{role}</span>
		<div
			class="bg-base-300 flex aspect-[2/3] w-24 items-center justify-center overflow-hidden rounded-lg"
		>
			{#if member.image}
				<img src={cacheUrl(member.image.uri)} alt="" class="h-full w-full object-cover" />
			{:else}
				<UserIcon size={32} class="opacity-20" />
			{/if}
		</div>
		<span class="font-medium">{member.name}</span>
		<span class="text-base-content/50 text-xs">
			{member.videoCount} videos · {member.galleryCount} galleries
		</span>
	</div>
{/snippet}

<dialog use:showModal class="modal" {onclose} aria-label="Merge {performer.name}">
	<div class="modal-box flex !max-w-xl flex-col gap-4">
		<h2 class="text-lg font-semibold">Merge performers</h2>

		{#if chosen && kept && folded}
			<div class="flex items-stretch gap-3">
				{@render side(kept, 'Stays')}
				<button
					type="button"
					class="btn btn-ghost btn-square self-center"
					onclick={() => (keepThisOne = !keepThisOne)}
					aria-label="Swap which one stays"
				>
					<ArrowsLeftRightIcon size={18} />
				</button>
				{@render side(folded, 'Merged in')}
			</div>
			<p class="text-base-content/70 text-sm">
				{folded.name}'s videos, galleries, sources and your ratings move to {kept.name}; the name
				{folded.name} becomes an alias. You can undo it afterwards.
			</p>
			<button
				type="button"
				class="text-base-content/60 self-start text-sm underline"
				onclick={() => (chosen = null)}
			>
				Pick someone else
			</button>
		{:else}
			<input class="input" placeholder="Find the other performer by name" bind:value={query} />
			<div class="flex max-h-72 flex-col overflow-y-auto">
				{#each results as candidate (candidate.id)}
					<button
						type="button"
						class="hover:bg-base-200 flex items-center gap-3 rounded-lg p-2 text-left"
						onclick={() => (chosen = candidate)}
					>
						<span class="bg-base-300 h-10 w-10 shrink-0 overflow-hidden rounded-full">
							{#if candidate.image}
								<img
									src={cacheUrl(candidate.image.uri)}
									alt=""
									class="h-full w-full object-cover"
								/>
							{/if}
						</span>
						<span class="flex min-w-0 flex-col">
							<span class="truncate text-sm font-medium">{candidate.name}</span>
							<span class="text-base-content/50 text-xs">{candidate.videoCount} videos</span>
						</span>
					</button>
				{:else}
					{#if query.trim() !== ''}
						<p class="text-base-content/50 p-2 text-sm">Nobody else by that name.</p>
					{/if}
				{/each}
			</div>
		{/if}

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}

		<div class="modal-action mt-0">
			<button type="button" class="btn btn-ghost" onclick={onclose}>Cancel</button>
			<button type="button" class="btn btn-error" disabled={!chosen || merging} onclick={merge}>
				{merging ? 'Merging…' : 'Merge'}
			</button>
		</div>
	</div>
</dialog>
