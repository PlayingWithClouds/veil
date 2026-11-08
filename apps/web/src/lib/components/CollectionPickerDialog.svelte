<!--
	"Add to collection": the user's collections as a list, the ones the record
	is in checked; a click adds it or takes it out again. A name typed in the
	field creates a collection with the record in it. Shown by the layout
	whenever collectionPick is set.
-->
<script lang="ts">
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import StackIcon from 'phosphor-svelte/lib/StackIcon';
	import {
		addToCollection,
		createCollection,
		fetchCollectionIdsWith,
		fetchCollections,
		removeRecordFromCollection,
		type CollectionCardRow
	} from '$lib/curation';
	import { cacheUrl } from '$lib/img';
	import { collectionPick, type CollectionPick } from '$lib/stores/collectionPicker';
	import { showUndo } from '$lib/stores/undoToast';

	interface Props {
		pick: CollectionPick;
	}

	let { pick }: Props = $props();

	let collections = $state<CollectionCardRow[] | null>(null);
	let containing = $state<Set<string>>(new Set());
	let name = $state('');
	let busy = $state(false);
	let error = $state<string | null>(null);

	$effect(() => {
		Promise.all([fetchCollections(), fetchCollectionIdsWith(pick.subject)])
			.then(([found, ids]) => {
				collections = found;
				containing = ids;
			})
			.catch((failure) => (error = String(failure)));
	});

	/** Adds the record to a collection, or takes it out when it's in it. */
	async function toggle(collection: CollectionCardRow) {
		const wasIn = containing.has(collection.id);
		containing = toggled(containing, collection.id);
		try {
			if (wasIn) {
				await removeRecordFromCollection(collection.id, pick.subject);
				return;
			}
			await addToCollection(collection.id, pick.subject);
			showUndo({ message: `Added to ${collection.name}` });
		} catch (failure) {
			containing = toggled(containing, collection.id);
			error = String(failure);
		}
	}

	/** The set with id added when it's missing, removed when it's there. */
	function toggled(ids: Set<string>, id: string): Set<string> {
		const next = new Set(ids);
		if (next.has(id)) {
			next.delete(id);
			return next;
		}
		next.add(id);
		return next;
	}

	/** Creates a collection by the typed name with the record in it, and closes. */
	async function create(event: SubmitEvent) {
		event.preventDefault();
		const trimmed = name.trim();
		if (trimmed === '') return;
		busy = true;
		try {
			const id = await createCollection(trimmed);
			await addToCollection(id, pick.subject);
			showUndo({ message: `Added to ${trimmed}` });
			close();
		} catch (failure) {
			error = String(failure);
			busy = false;
		}
	}

	/** Closes the dialog. */
	function close() {
		collectionPick.set(null);
	}

	/** Opens the dialog as a modal once it is in the page. */
	function showModal(dialog: HTMLDialogElement) {
		dialog.showModal();
	}

	/** Closes on a click beside the box (the dialog itself fills the screen). */
	function closeOnBackdrop(event: MouseEvent) {
		if (event.target === event.currentTarget) close();
	}

	/** The first cover a collection's tile shows, if any. */
	function coverOf(collection: CollectionCardRow): string | null {
		if (collection.cover) return cacheUrl(collection.cover.uri);
		const first = collection.covers[0];
		if (!first) return null;
		return cacheUrl(first);
	}
</script>

<dialog
	use:showModal
	class="modal picker-backdrop"
	onclose={close}
	onclick={closeOnBackdrop}
	aria-label="Add to collection"
>
	<div class="menu-panel flex max-h-[80vh] w-[min(28rem,calc(100vw-2rem))] flex-col gap-4 p-5">
		<div class="flex min-w-0 flex-col">
			<h2 class="text-lg font-semibold">Add to collection</h2>
			<p class="text-base-content/50 truncate text-sm">{pick.title}</p>
		</div>

		<form class="flex gap-2" onsubmit={create}>
			<input class="input w-full" bind:value={name} placeholder="New collection" />
			<button type="submit" class="btn" disabled={busy || name.trim() === ''}>
				<PlusIcon size={16} />
				Create
			</button>
		</form>

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}

		<div class="-mx-1 flex min-h-24 flex-col gap-1 overflow-y-auto px-1">
			{#if collections === null}
				<div class="flex justify-center py-6">
					<span class="loading loading-spinner"></span>
				</div>
			{:else if collections.length === 0}
				<p class="text-base-content/50 py-6 text-center text-sm">
					No collections yet. Name one above.
				</p>
			{:else}
				{#each collections as collection (collection.id)}
					{@const cover = coverOf(collection)}
					{@const added = containing.has(collection.id)}
					<button
						type="button"
						class="picker-card flex items-center gap-3 rounded-xl p-2 text-left"
						class:picker-card-on={added}
						aria-pressed={added}
						onclick={() => toggle(collection)}
					>
						<span
							class="bg-base-300 flex h-10 w-16 shrink-0 items-center justify-center overflow-hidden rounded-lg"
						>
							{#if cover}
								<img src={cover} alt="" loading="lazy" class="h-full w-full object-cover" />
							{:else}
								<StackIcon size={18} class="opacity-30" />
							{/if}
						</span>
						<span class="flex min-w-0 grow flex-col">
							<span class="truncate text-sm font-medium">{collection.name}</span>
							<span class="text-base-content/45 text-xs">{collection.itemCount} items</span>
						</span>
						{#if added}
							<span
								class="bg-success flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-black"
							>
								<CheckIcon size={12} weight="bold" />
							</span>
						{/if}
					</button>
				{/each}
			{/if}
		</div>

		<div class="flex justify-end">
			<button type="button" class="btn btn-ghost btn-sm" onclick={close}>Done</button>
		</div>
	</div>
</dialog>
