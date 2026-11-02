<!--
	Editing a performer in a dialog: name, aliases, picture (by URL, with a
	preview), description, links and every attribute. Only fields that changed
	are saved, since a saved field is locked against the sites.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { cacheUrl } from '$lib/img';
	import {
		editPerformer,
		type PerformerAttributes,
		type PerformerDetail,
		type PerformerEdit
	} from '$lib/performers';

	interface Props {
		performer: PerformerDetail;
		onsaved: (performer: PerformerDetail) => void;
		onclose: () => void;
	}

	let { performer, onsaved, onclose }: Props = $props();

	type AttributeKey = keyof PerformerAttributes;

	// The attribute inputs, in form order.
	const ATTRIBUTE_FIELDS: { key: AttributeKey; label: string; type: 'text' | 'number' | 'date' }[] =
		[
			{ key: 'gender', label: 'Gender', type: 'text' },
			{ key: 'birthdate', label: 'Born', type: 'date' },
			{ key: 'deathDate', label: 'Died', type: 'date' },
			{ key: 'country', label: 'Country', type: 'text' },
			{ key: 'ethnicity', label: 'Ethnicity', type: 'text' },
			{ key: 'hairColor', label: 'Hair', type: 'text' },
			{ key: 'eyeColor', label: 'Eyes', type: 'text' },
			{ key: 'heightCm', label: 'Height (cm)', type: 'number' },
			{ key: 'weightKg', label: 'Weight (kg)', type: 'number' },
			{ key: 'measurements', label: 'Measurements', type: 'text' },
			{ key: 'fakeTits', label: 'Enhanced', type: 'text' },
			{ key: 'tattoos', label: 'Tattoos', type: 'text' },
			{ key: 'piercings', label: 'Piercings', type: 'text' },
			{ key: 'careerStart', label: 'Active since', type: 'number' },
			{ key: 'careerEnd', label: 'Active until', type: 'number' }
		];

	// The performer as the dialog opened: the form starts from it and saves what differs.
	const original = untrack(() => performer);

	let name = $state(original.name);
	let aliasesText = $state(original.aliases.join(', '));
	let imageUrl = $state(original.image?.uri ?? '');
	let details = $state(original.details ?? '');
	let linksText = $state(original.links.join('\n'));
	let attributes = $state<PerformerAttributes>({ ...original.attributes });
	let saving = $state(false);
	let error = $state<string | null>(null);

	let preview = $derived(cacheUrl(imageUrl.trim()));

	/** Opens the dialog as a modal once it is in the page. */
	function showModal(dialog: HTMLDialogElement) {
		dialog.showModal();
	}

	/** Splits a list typed into one field, dropping blanks. */
	function splitList(text: string, separator: RegExp): string[] {
		return text
			.split(separator)
			.map((item) => item.trim())
			.filter((item) => item !== '');
	}

	/** Blank inputs as null, so cleared attributes are cleared. */
	function cleanAttributes(values: PerformerAttributes): PerformerAttributes {
		const cleaned = { ...values };
		for (const field of ATTRIBUTE_FIELDS) {
			const value = cleaned[field.key];
			if (value === '' || value === undefined) {
				Object.assign(cleaned, { [field.key]: null });
			}
		}
		return cleaned;
	}

	/** Whether two lists hold the same items in the same order. */
	function sameList(first: string[], second: string[]): boolean {
		return first.length === second.length && first.every((item, index) => item === second[index]);
	}

	/** The fields that differ from the stored performer. */
	function changes(): PerformerEdit {
		const edit: PerformerEdit = {};
		if (name.trim() !== original.name) edit.name = name.trim();
		const aliases = splitList(aliasesText, /,/);
		if (!sameList(aliases, original.aliases)) edit.aliases = aliases;
		if (imageUrl.trim() !== (original.image?.uri ?? '')) edit.imageUrl = imageUrl.trim();
		if (details.trim() !== (original.details ?? '')) edit.details = details.trim();
		const links = splitList(linksText, /\n/);
		if (!sameList(links, original.links)) edit.links = links;
		const cleaned = cleanAttributes(attributes);
		if (JSON.stringify(cleaned) !== JSON.stringify(cleanAttributes(original.attributes))) {
			edit.attributes = cleaned;
		}
		return edit;
	}

	/** Saves what changed and closes. */
	async function save(event: SubmitEvent) {
		event.preventDefault();
		const edit = changes();
		if (Object.keys(edit).length === 0) {
			onclose();
			return;
		}
		saving = true;
		error = null;
		try {
			const saved = await editPerformer(performer.id, edit);
			if (saved) onsaved(saved);
			onclose();
		} catch (failure) {
			error = String(failure);
		} finally {
			saving = false;
		}
	}
</script>

<dialog use:showModal class="modal" {onclose} aria-label="Edit {performer.name}">
	<form
		class="modal-box flex max-h-[90vh] !max-w-3xl flex-col gap-5 overflow-y-auto"
		onsubmit={save}
	>
		<h2 class="text-lg font-semibold">Edit performer</h2>

		<div class="grid grid-cols-[8rem_1fr] gap-5 max-md:grid-cols-1">
			<div class="bg-base-200 aspect-[2/3] overflow-hidden rounded-xl">
				{#if preview}
					<img src={preview} alt="" class="h-full w-full object-cover" />
				{/if}
			</div>
			<div class="flex flex-col gap-3">
				<label class="form-control">
					<span class="label-text">Name</span>
					<input class="input" bind:value={name} required />
				</label>
				<label class="form-control">
					<span class="label-text">Also known as (comma-separated)</span>
					<input class="input" bind:value={aliasesText} />
				</label>
				<label class="form-control">
					<span class="label-text">Picture URL (empty removes it)</span>
					<input class="input" type="url" bind:value={imageUrl} />
				</label>
			</div>
		</div>

		<label class="form-control">
			<span class="label-text">About</span>
			<textarea class="input min-h-24 py-2" bind:value={details}></textarea>
		</label>

		<div class="grid grid-cols-3 gap-3 max-md:grid-cols-2">
			{#each ATTRIBUTE_FIELDS as field (field.key)}
				<label class="form-control">
					<span class="label-text">{field.label}</span>
					<input class="input" type={field.type} bind:value={attributes[field.key]} />
				</label>
			{/each}
		</div>

		<label class="form-control">
			<span class="label-text">Links (one per line)</span>
			<textarea class="input min-h-20 py-2" bind:value={linksText}></textarea>
		</label>

		<p class="text-base-content/50 text-xs">
			Fields you change here stay as you set them; the sites no longer update them.
		</p>
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
