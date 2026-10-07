<script lang="ts">
	import { goto } from '$app/navigation';
	import ArrowLeftIcon from 'phosphor-svelte/lib/ArrowLeftIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import PencilSimpleIcon from 'phosphor-svelte/lib/PencilSimpleIcon';
	import RobotIcon from 'phosphor-svelte/lib/RobotIcon';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import TagChips from '$lib/components/TagChips.svelte';
	import { notifications } from '$lib/stores/notifications';
	import {
		reorderCollection,
		removeFromCollection,
		renameCollection,
		deleteCollection,
		memberUrl,
		type CollectionMember
	} from '$lib/collections';

	let { data } = $props();
	let collectionId = $derived(data.id);
	let name = $state(data.collection?.name ?? 'Collection');
	let members = $state<CollectionMember[]>(data.members);
	// Scraper-created collections render read-only: no rename, reorder, or remove.
	let editable = $derived(data.collection?.origin !== 'scraped');
	let tags = $derived(data.collection?.tags ?? []);

	let editingName = $state(false);
	let dragIndex = $state<number | null>(null);
	let overIndex = $state<number | null>(null);

	function persistOrder() {
		reorderCollection(
			collectionId,
			members.map((member) => member.mediaId)
		);
	}

	function onDrop(targetIndex: number) {
		if (dragIndex === null || dragIndex === targetIndex) {
			dragIndex = null;
			overIndex = null;
			return;
		}
		const next = [...members];
		const [moved] = next.splice(dragIndex, 1);
		next.splice(targetIndex, 0, moved);
		members = next;
		dragIndex = null;
		overIndex = null;
		persistOrder();
	}

	async function remove(member: CollectionMember) {
		members = members.filter((entry) => entry.mediaId !== member.mediaId);
		await removeFromCollection(collectionId, member.mediaId);
	}

	function openMember(member: CollectionMember) {
		const url = memberUrl(member);
		if (url) goto(url);
	}

	async function saveName() {
		editingName = false;
		const trimmed = name.trim();
		if (!trimmed) {
			name = data.collection?.name ?? 'Collection';
			return;
		}
		await renameCollection(collectionId, trimmed);
	}

	async function removeCollection() {
		await deleteCollection(collectionId);
		notifications.push('Collection deleted', 'success');
		goto('/collections');
	}
</script>

<div class="flex flex-col gap-6">
	<div class="flex items-center gap-3">
		<button
			type="button"
			class="btn btn-square btn-ghost btn-sm"
			aria-label="Back to collections"
			onclick={() => goto('/collections')}
		>
			<ArrowLeftIcon size={18} />
		</button>
		{#if editingName && editable}
			<input
				type="text"
				bind:value={name}
				maxlength={60}
				class="bg-base-200 border-base-300 h-10 rounded-lg border px-3 text-2xl font-bold outline-none"
				onblur={saveName}
				onkeydown={(event) => event.key === 'Enter' && saveName()}
			/>
		{:else}
			<h1 class="text-3xl font-bold">{name}</h1>
			{#if editable}
				<button
					type="button"
					class="text-base-content/40 hover:text-base-content"
					aria-label="Rename"
					onclick={() => (editingName = true)}
				>
					<PencilSimpleIcon size={18} />
				</button>
			{:else}
				<span class="text-base-content/40 flex items-center gap-1 text-xs" title="Created by a source plugin">
					<RobotIcon size={14} /> from source
				</span>
			{/if}
		{/if}
		<span class="text-base-content/40 text-sm">{members.length} items</span>
		{#if editable}
			<button
				type="button"
				class="text-base-content/40 hover:text-error ml-auto"
				aria-label="Delete collection"
				onclick={removeCollection}
			>
				<TrashIcon size={18} />
			</button>
		{/if}
	</div>

	{#if tags.length > 0}
		<TagChips {tags} />
	{/if}

	{#if members.length === 0}
		<div class="flex flex-col items-center gap-3 py-20 text-center">
			<span class="text-6xl opacity-10">📁</span>
			<p class="text-base-content/40">This collection is empty.</p>
		</div>
	{:else}
		{#if editable}
			<p class="text-base-content/40 text-xs">Drag cards to reorder.</p>
		{/if}
		<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
			{#each members as member, index (member.mediaId)}
				<div
					role="listitem"
					draggable={editable}
					ondragstart={() => editable && (dragIndex = index)}
					ondragover={(event) => {
						if (!editable) return;
						event.preventDefault();
						overIndex = index;
					}}
					ondragend={() => {
						dragIndex = null;
						overIndex = null;
					}}
					ondrop={() => editable && onDrop(index)}
					class="group/drag relative rounded-xl transition-all {editable
						? 'cursor-grab active:cursor-grabbing'
						: ''} {dragIndex === index ? 'opacity-40' : ''} {overIndex === index &&
					dragIndex !== index
						? 'ring-primary ring-2'
						: ''}"
				>
					<MemberCard {member} onclick={() => openMember(member)} />
					{#if editable}
						<button
							type="button"
							aria-label="Remove from collection"
							onclick={(event) => {
								event.stopPropagation();
								remove(member);
							}}
							class="hover:bg-error absolute right-1.5 top-1.5 z-10 flex h-8 w-8 items-center justify-center rounded-full bg-black/60 text-white opacity-0 transition-opacity group-hover/drag:opacity-100"
						>
							<XIcon size={16} weight="bold" />
						</button>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
