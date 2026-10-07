<script lang="ts">
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import { moveDialogTarget, closeMoveDialog } from '$lib/stores/moveDialog';
	import { setLastCollection } from '$lib/stores/lastCollection';
	import {
		fetchCollections,
		collectionIdsForEntity,
		addToCollection,
		removeFromCollection,
		createCollection,
		type Collection
	} from '$lib/collections';

	let target = $derived($moveDialogTarget);
	let collections = $state<Collection[]>([]);
	let memberIds = $state<Set<string>>(new Set());
	let loading = $state(false);
	let newName = $state('');
	let creating = $state(false);

	// Load the user-created collections + current membership whenever the
	// dialog opens for an entity.
	$effect(() => {
		const current = target;
		if (!current) {
			collections = [];
			memberIds = new Set();
			return;
		}
		loading = true;
		let cancelled = false;
		Promise.all([
			fetchCollections('user'),
			collectionIdsForEntity(current.entityId)
		]).then(([lists, ids]) => {
			if (cancelled) return;
			collections = lists;
			memberIds = new Set(ids);
			loading = false;
		});
		return () => {
			cancelled = true;
		};
	});

	async function toggle(collection: Collection) {
		if (!target) return;
		const entityId = target.entityId;
		const next = new Set(memberIds);
		if (memberIds.has(collection.id)) {
			next.delete(collection.id);
			memberIds = next;
			await removeFromCollection(collection.id, entityId);
		} else {
			next.add(collection.id);
			memberIds = next;
			await addToCollection(collection.id, entityId);
			setLastCollection(collection.id);
		}
	}

	async function create() {
		const name = newName.trim();
		if (!name || creating) return;
		creating = true;
		const collection = await createCollection(name);
		creating = false;
		newName = '';
		if (collection) {
			collections = [collection, ...collections];
			await toggle(collection);
		}
	}
</script>

{#if target}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4"
		role="button"
		tabindex="-1"
		onclick={closeMoveDialog}
		onkeydown={(event) => event.key === 'Escape' && closeMoveDialog()}
	>
		<!-- Dialog -->
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
					<h2 class="text-lg font-bold">Add to collection</h2>
					<p class="text-base-content/50 line-clamp-1 text-sm">{target.title}</p>
				</div>
				<button
					type="button"
					class="text-base-content/40 hover:text-base-content"
					aria-label="Close"
					onclick={closeMoveDialog}
				>
					<XIcon size={20} />
				</button>
			</div>

			{#if loading}
				<div class="flex justify-center py-8">
					<span class="loading loading-spinner text-base-content/40"></span>
				</div>
			{:else}
				<div class="no-scrollbar flex max-h-72 flex-col gap-1 overflow-y-auto">
					{#each collections as collection (collection.id)}
						{@const member = memberIds.has(collection.id)}
						<button
							type="button"
							onclick={() => toggle(collection)}
							class="hover:bg-base-300 flex items-center justify-between gap-3 rounded-lg px-3 py-2.5 text-left transition-colors"
						>
							<span class="flex flex-col">
								<span class="text-sm font-medium">{collection.name}</span>
								<span class="text-base-content/40 text-xs">{collection.itemCount} items</span>
							</span>
							<span
								class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 {member
									? 'border-primary bg-primary text-black'
									: 'border-base-content/25'}"
							>
								{#if member}<CheckIcon size={12} weight="bold" />{/if}
							</span>
						</button>
					{/each}
					{#if collections.length === 0}
						<p class="text-base-content/40 px-3 py-4 text-sm">
							No collections yet. Create one below.
						</p>
					{/if}
				</div>
			{/if}

			<!-- Create new collection -->
			<form
				class="border-base-300 flex items-center gap-2 border-t pt-3"
				onsubmit={(event) => {
					event.preventDefault();
					create();
				}}
			>
				<input
					type="text"
					bind:value={newName}
					placeholder="New collection name"
					maxlength={60}
					class="bg-base-100 border-base-300 focus:border-base-content/30 h-10 flex-1 rounded-lg border px-3 text-sm outline-none"
				/>
				<button
					type="submit"
					disabled={!newName.trim() || creating}
					class="bg-primary text-primary-content flex h-10 items-center gap-1 rounded-lg px-3 text-sm font-semibold disabled:opacity-40"
				>
					<PlusIcon size={16} weight="bold" /> Create
				</button>
			</form>
		</div>
	</div>
{/if}
