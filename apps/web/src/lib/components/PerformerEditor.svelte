<!--
	Editing a performer in a dialog: name, aliases, picture (uploaded by click
	or drop, or by URL), description, links and every attribute. Text
	attributes suggest common values and the library's own, but take anything
	typed; measurements are bust, waist and hips; born and active since show
	the age and years active. Only fields that changed are saved, since a
	saved field is locked against the sites.
-->
<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import UploadSimpleIcon from 'phosphor-svelte/lib/UploadSimpleIcon';
	import Combobox from '$lib/components/Combobox.svelte';
	import { cacheUrl } from '$lib/img';
	import {
		editPerformer,
		fetchAttributeValues,
		uploadPicture,
		type PerformerAttributes,
		type PerformerDetail,
		type PerformerEdit
	} from '$lib/performers';
	import {
		BODY_ART,
		ENHANCEMENTS,
		ETHNICITIES,
		EYE_COLORS,
		GENDERS,
		HAIR_COLORS,
		capitalize,
		countryNames,
		joinMeasurements,
		mergeSuggestions,
		normalizeCountry,
		splitMeasurements,
		yearsBetween,
		type Measurements
	} from '$lib/performerPresets';

	interface Props {
		performer: PerformerDetail;
		onsaved: (performer: PerformerDetail) => void;
		onclose: () => void;
	}

	let { performer, onsaved, onclose }: Props = $props();

	type AttributeKey = keyof PerformerAttributes;

	/** One attribute input: suggested text, a date, a number, or the measurement parts. */
	interface AttributeField {
		key: AttributeKey;
		label: string;
		type: 'suggest' | 'date' | 'number' | 'measurements';
		presets?: string[];
	}

	// The attribute inputs, in form order.
	const ATTRIBUTE_FIELDS: AttributeField[] = [
		{ key: 'gender', label: 'Gender', type: 'suggest', presets: GENDERS },
		{ key: 'birthdate', label: 'Born', type: 'date' },
		{ key: 'deathDate', label: 'Died', type: 'date' },
		{ key: 'country', label: 'Country', type: 'suggest', presets: countryNames() },
		{ key: 'ethnicity', label: 'Ethnicity', type: 'suggest', presets: ETHNICITIES },
		{ key: 'hairColor', label: 'Hair', type: 'suggest', presets: HAIR_COLORS },
		{ key: 'eyeColor', label: 'Eyes', type: 'suggest', presets: EYE_COLORS },
		{ key: 'heightCm', label: 'Height (cm)', type: 'number' },
		{ key: 'weightKg', label: 'Weight (kg)', type: 'number' },
		{ key: 'measurements', label: 'Measurements', type: 'measurements' },
		{ key: 'fakeTits', label: 'Enhanced', type: 'suggest', presets: ENHANCEMENTS },
		{ key: 'tattoos', label: 'Tattoos', type: 'suggest', presets: BODY_ART },
		{ key: 'piercings', label: 'Piercings', type: 'suggest', presets: BODY_ART },
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
	let attributes = $state<PerformerAttributes>(capitalizedText(original.attributes));
	// Null when the stored measurements aren't "bust-waist-hips"; they're then edited as text.
	let measurementParts = $state(splitMeasurements(original.attributes.measurements));
	let libraryValues = $state<Partial<Record<AttributeKey, string[]>>>({});
	let saving = $state(false);
	let uploading = $state(false);
	let dragging = $state(false);
	let error = $state<string | null>(null);
	let fileInput = $state<HTMLInputElement>();

	let preview = $derived(cacheUrl(imageUrl.trim()));

	onMount(() => {
		for (const field of ATTRIBUTE_FIELDS) {
			if (field.type !== 'suggest') continue;
			void fetchAttributeValues(field.key).then((values) => (libraryValues[field.key] = values));
		}
	});

	/** A quiet note under an input: the age by Born, years active by Active since. */
	function hint(key: AttributeKey): string | null {
		if (key === 'birthdate' && attributes.birthdate) {
			const age = yearsBetween(attributes.birthdate, attributes.deathDate);
			if (age === null) return null;
			return attributes.deathDate ? `died aged ${age}` : `${age} years old`;
		}
		if (key === 'careerStart' && attributes.careerStart) {
			let until = attributes.careerEnd;
			if (!until) until = new Date().getFullYear();
			return `${until - attributes.careerStart} years active`;
		}
		return null;
	}

	/** Uploads a picked or dropped picture and shows it as the new one. */
	async function upload(file: File | undefined) {
		if (!file) return;
		if (!file.type.startsWith('image/')) {
			error = 'That file is not a picture.';
			return;
		}
		uploading = true;
		error = null;
		try {
			imageUrl = await uploadPicture(file);
		} catch (failure) {
			error = String(failure);
		} finally {
			uploading = false;
		}
	}

	/** Takes a picture dropped on the preview. */
	function drop(event: DragEvent) {
		event.preventDefault();
		dragging = false;
		void upload(event.dataTransfer?.files[0]);
	}

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

	/**
	 * The attributes as the form holds them: measurements joined from their
	 * parts, and a country the user changed in the list's spelling ("USA" as
	 * "United States"). An untouched country stays as the site spelled it, so
	 * opening and saving locks nothing.
	 */
	function formAttributes(): PerformerAttributes {
		const values = { ...attributes };
		if (measurementParts) values.measurements = joinedMeasurements(measurementParts);
		if (values.country && values.country !== original.attributes.country) {
			values.country = normalizeCountry(values.country);
		}
		for (const field of ATTRIBUTE_FIELDS) {
			if (field.type !== 'suggest') continue;
			Object.assign(values, { [field.key]: savedText(field.key, values[field.key] as string | null) });
		}
		return cleanAttributes(values);
	}

	/** The text attributes capitalised, as the form shows them. */
	function capitalizedText(values: PerformerAttributes): PerformerAttributes {
		const shown = { ...values };
		for (const field of ATTRIBUTE_FIELDS) {
			const value = shown[field.key];
			if (field.type === 'suggest' && typeof value === 'string') {
				Object.assign(shown, { [field.key]: capitalize(value) });
			}
		}
		return shown;
	}

	/**
	 * A text attribute as it's saved: capitalised, but the stored spelling
	 * when only the case differs, so showing it capitalised locks nothing.
	 */
	function savedText(key: AttributeKey, value: string | null): string | null {
		if (!value) return value;
		const stored = original.attributes[key];
		if (typeof stored === 'string' && stored.toLowerCase() === value.trim().toLowerCase()) return stored;
		return capitalize(value.trim());
	}

	/** The measurements from their parts; the stored spelling when the parts didn't change. */
	function joinedMeasurements(parts: Measurements): string {
		const joined = joinMeasurements(parts);
		const stored = splitMeasurements(original.attributes.measurements);
		if (stored && joinMeasurements(stored) === joined) return original.attributes.measurements ?? '';
		return joined;
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
		const cleaned = formAttributes();
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
			<button
				type="button"
				class="bg-base-200 group relative aspect-[2/3] overflow-hidden rounded-xl"
				class:ring-2={dragging}
				class:ring-primary={dragging}
				aria-label="Upload a picture"
				title="Click or drop a picture"
				disabled={uploading}
				onclick={() => fileInput?.click()}
				ondragover={(event) => {
					event.preventDefault();
					dragging = true;
				}}
				ondragleave={() => (dragging = false)}
				ondrop={drop}
			>
				{#if preview}
					<img src={preview} alt="" class="h-full w-full object-cover" />
				{/if}
				<span
					class="bg-base-300/80 absolute inset-0 flex flex-col items-center justify-center gap-1 text-xs opacity-0 transition-opacity group-hover:opacity-100"
					class:opacity-100={uploading || dragging || !preview}
				>
					<UploadSimpleIcon size={22} />
					{uploading ? 'Uploading…' : 'Click or drop a picture'}
				</span>
			</button>
			<input
				bind:this={fileInput}
				type="file"
				accept="image/*"
				class="hidden"
				onchange={(event) => upload(event.currentTarget.files?.[0])}
			/>
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
					<input class="input" bind:value={imageUrl} />
				</label>
			</div>
		</div>

		<label class="form-control">
			<span class="label-text">About</span>
			<textarea class="input min-h-24 py-2" bind:value={details}></textarea>
		</label>

		<div class="grid grid-cols-3 gap-3 max-md:grid-cols-2">
			{#each ATTRIBUTE_FIELDS as field (field.key)}
				{@const note = hint(field.key)}
				<label class="form-control">
					<span class="label-text">{field.label}</span>
					{#if field.type === 'measurements' && measurementParts}
						<span class="flex gap-1.5">
							<input
								class="input min-w-0"
								placeholder="Bust"
								aria-label="Bust with cup"
								bind:value={measurementParts.bust}
							/>
							<input
								class="input min-w-0"
								placeholder="Waist"
								aria-label="Waist"
								inputmode="numeric"
								bind:value={measurementParts.waist}
							/>
							<input
								class="input min-w-0"
								placeholder="Hips"
								aria-label="Hips"
								inputmode="numeric"
								bind:value={measurementParts.hips}
							/>
						</span>
					{:else if field.type === 'measurements'}
						<input class="input" bind:value={attributes.measurements} />
					{:else if field.type === 'suggest'}
						<Combobox
							label={field.label}
							options={mergeSuggestions(field.presets ?? [], libraryValues[field.key] ?? [])}
							bind:value={
								() => attributes[field.key] as string | null,
								(text) => Object.assign(attributes, { [field.key]: text })
							}
						/>
					{:else}
						<input class="input" type={field.type} bind:value={attributes[field.key]} />
					{/if}
					{#if note}
						<span class="text-base-content/50 text-xs">{note}</span>
					{/if}
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
