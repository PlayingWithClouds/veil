<!--
	Adding or editing a category: its name, the group it's in (or the top),
	and the words that put media in it, as a site would tag it ("big tits,
	busty"). Words are compared as tag names are (case and punctuation
	don't matter). Hands the saved category to onsaved.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import {
		createCategory,
		editCategory,
		fetchParentOptions,
		parseTerms,
		type CategoryDetail,
		type ParentOption
	} from '$lib/categories';

	interface Props {
		// The category to edit; a new one when null.
		category?: CategoryDetail | null;
		// The group a new category starts in.
		parentId?: string | null;
		onsaved: (category: CategoryDetail) => void;
		onclose: () => void;
	}

	let { category = null, parentId = null, onsaved, onclose }: Props = $props();

	// The form starts from the category (or the group given); it isn't reset
	// when they change while it's open.
	let name = $state(untrack(() => category?.name ?? ''));
	let parent = $state(untrack(() => category?.parent?.id ?? parentId ?? ''));
	let termsText = $state(untrack(() => (category?.matchTerms ?? []).join(', ')));
	let parents = $state<ParentOption[]>([]);
	let saving = $state(false);
	let error = $state<string | null>(null);

	// A category can't go inside itself.
	let parentChoices = $derived(parents.filter((option) => option.id !== category?.id));

	$effect(() => {
		fetchParentOptions().then((found) => (parents = found));
	});

	/** Opens the dialog as a modal once it's in the page. */
	function showModal(dialog: HTMLDialogElement) {
		dialog.showModal();
	}

	/** Saves the category and hands it over. */
	async function save(event: SubmitEvent) {
		event.preventDefault();
		if (name.trim() === '') {
			error = 'A category needs a name.';
			return;
		}
		saving = true;
		error = null;
		const input = { name: name.trim(), parentId: parent, matchTerms: parseTerms(termsText) };
		try {
			if (category) {
				onsaved(await editCategory(category.id, input));
				return;
			}
			onsaved(await createCategory(input));
		} catch (failure) {
			error = String(failure);
		} finally {
			saving = false;
		}
	}
</script>

<dialog
	use:showModal
	class="modal"
	{onclose}
	aria-label={category ? 'Edit category' : 'New category'}
>
	<form class="modal-box flex flex-col gap-4" onsubmit={save}>
		<h2 class="text-lg font-semibold">{category ? 'Edit category' : 'New category'}</h2>

		<label class="form-control flex flex-col gap-1">
			<span class="label-text text-sm">Name</span>
			<input class="input w-full" bind:value={name} placeholder="Latex" required />
		</label>

		<label class="form-control flex flex-col gap-1">
			<span class="label-text text-sm">Group</span>
			<select class="select w-full" bind:value={parent}>
				<option value="">None (a group of its own)</option>
				{#each parentChoices as option (option.id)}
					<option value={option.id}>{option.name}</option>
				{/each}
			</select>
		</label>

		<label class="form-control flex flex-col gap-1">
			<span class="label-text text-sm">Tags it matches</span>
			<textarea
				class="textarea w-full"
				rows="3"
				bind:value={termsText}
				placeholder="latex, pvc, rubber, catsuit"
			></textarea>
			<span class="text-base-content/50 text-xs">
				Separated by commas. Media with a tag of one of these names is in the category; case and
				punctuation don't matter.
			</span>
		</label>

		{#if category && category.matchTags.length > 0}
			<p class="text-base-content/60 text-xs">
				Also matches the tags {category.matchTags.map((tag) => tag.name).join(', ')}.
			</p>
		{/if}

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}

		<div class="modal-action mt-0">
			<button type="button" class="btn btn-ghost" onclick={onclose}>Cancel</button>
			<button type="submit" class="btn btn-primary" disabled={saving}>
				{saving ? 'Saving…' : 'Save'}
			</button>
		</div>
	</form>
</dialog>
