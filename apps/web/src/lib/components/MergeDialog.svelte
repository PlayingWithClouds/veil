<!--
	Merges the record whose page is open into another one the user picks: search
	for the record to keep, confirm, and everything of the open record (credits,
	scenes, history, site addresses) moves over. Irreversible, so the confirm step
	says so. Mounted once in the layout; open it with openMergeDialog.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { mergeDialogTarget, closeMergeDialog } from '$lib/stores/mergeDialog';
	import { notifications } from '$lib/stores/notifications';
	import { mergeRecords, searchMergeCandidates, type MergeCandidate } from '$lib/merge';
	import { performerUrl, studioUrl, tagUrl } from '$lib/routes';

	/** Pause after typing before the candidates are searched. */
	const SEARCH_DEBOUNCE_MS = 200;

	let target = $derived($mergeDialogTarget);
	let search = $state('');
	let candidates = $state<MergeCandidate[]>([]);
	let chosen = $state<MergeCandidate | null>(null);
	let merging = $state(false);

	// Reset when the dialog opens for another record.
	$effect(() => {
		void target;
		search = '';
		candidates = [];
		chosen = null;
	});

	// Search as the user types; the open record itself is not a candidate.
	$effect(() => {
		const current = target;
		const text = search.trim();
		if (!current || text.length < 2) {
			candidates = [];
			return;
		}
		let cancelled = false;
		const timer = setTimeout(async () => {
			const found = await searchMergeCandidates(current.kind, text).catch(() => []);
			if (!cancelled) candidates = found.filter((candidate) => candidate.id !== current.id);
		}, SEARCH_DEBOUNCE_MS);
		return () => {
			cancelled = true;
			clearTimeout(timer);
		};
	});

	/** Page of the kept record after the merge. */
	function pageOf(kind: string, id: string): string {
		if (kind === 'performer') return performerUrl(id);
		if (kind === 'studio') return studioUrl(id);
		return tagUrl(id);
	}

	/** Merges the open record into the chosen one and opens the survivor's page. */
	async function confirm() {
		if (!target || !chosen || merging) return;
		merging = true;
		const { kind, id, name } = target;
		try {
			const survivorId = await mergeRecords(chosen.id, id);
			notifications.push(`Merged ${name} into ${chosen.name}`, 'success');
			closeMergeDialog();
			await goto(pageOf(kind, survivorId), { invalidateAll: true });
		} catch {
			notifications.push('Could not merge', 'error');
		} finally {
			merging = false;
		}
	}
</script>

{#if target}
	<div
		class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4"
		role="button"
		tabindex="-1"
		onclick={closeMergeDialog}
		onkeydown={(event) => event.key === 'Escape' && closeMergeDialog()}
	>
		<div
			class="border-base-300 bg-base-200 flex w-full max-w-md flex-col gap-4 rounded-2xl border p-6 shadow-2xl"
			role="dialog"
			aria-modal="true"
			tabindex="-1"
			onclick={(event) => event.stopPropagation()}
			onkeydown={() => {}}
		>
			<div class="flex items-start justify-between gap-3">
				<div class="flex flex-col">
					<h2 class="text-lg font-bold">Merge into another {target.kind}</h2>
					<p class="text-base-content/50 line-clamp-1 text-sm">{target.name}</p>
				</div>
				<button
					type="button"
					class="text-base-content/40 hover:text-base-content"
					aria-label="Close"
					onclick={closeMergeDialog}
				>
					<XIcon size={20} />
				</button>
			</div>

			{#if chosen}
				<p class="text-sm">
					<strong>{target.name}</strong> will be merged into <strong>{chosen.name}</strong>: its
					scenes, history and site addresses move over and {target.name} is deleted. This can't be
					undone.
				</p>
				<div class="flex justify-end gap-2">
					<button
						type="button"
						class="hover:bg-base-300 h-10 rounded-lg px-4 text-sm"
						onclick={() => (chosen = null)}
					>
						Back
					</button>
					<button
						type="button"
						disabled={merging}
						class="bg-error text-error-content h-10 rounded-lg px-4 text-sm font-semibold disabled:opacity-40"
						onclick={confirm}
					>
						Merge
					</button>
				</div>
			{:else}
				<input
					type="text"
					bind:value={search}
					placeholder="Search for the {target.kind} to keep"
					class="bg-base-100 border-base-300 focus:border-base-content/30 h-10 rounded-lg border px-3 text-sm outline-none"
				/>
				<div class="no-scrollbar flex max-h-72 flex-col gap-1 overflow-y-auto">
					{#each candidates as candidate (candidate.id)}
						<button
							type="button"
							onclick={() => (chosen = candidate)}
							class="hover:bg-base-300 rounded-lg px-3 py-2.5 text-left text-sm font-medium transition-colors"
						>
							{candidate.name}
						</button>
					{/each}
					{#if search.trim().length >= 2 && candidates.length === 0}
						<p class="text-base-content/40 px-3 py-4 text-sm">Nothing found.</p>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}
