<!--
	The face search mode of Home and Search: drop, paste or pick a picture, or
	paste its address, and every face search plugin (NameThatPorn, FaceDigger,
	PornstarByFace, …) is asked who is in it. The answer is laid out by
	FaceResults; with no media to credit, names open their page or a search.
-->
<script lang="ts">
	import UserFocusIcon from 'phosphor-svelte/lib/UserFocusIcon';
	import UploadSimpleIcon from 'phosphor-svelte/lib/UploadSimpleIcon';
	import { cacheUrl } from '$lib/img';
	import { faceSearch } from '$lib/faceSearch.svelte';
	import FaceResults from './FaceResults.svelte';

	let fileInput = $state<HTMLInputElement | null>(null);
	let address = $state('');
	let dragging = $state(false);

	// The searched picture as the page can show it: a picked file as is, a web one through the image cache.
	let shownPicture = $derived.by(() => {
		const pictureUrl = faceSearch.pictureUrl;
		if (pictureUrl === null) return null;
		if (pictureUrl.startsWith('data:')) return pictureUrl;
		return cacheUrl(pictureUrl);
	});

	/** Searches the first file of a pick or drop. */
	function searchFiles(files: FileList | null | undefined) {
		if (!files || files.length === 0) return;
		faceSearch.searchFile(files[0]);
	}

	/** Searches the address typed or pasted into the field. */
	function submitAddress(event: SubmitEvent) {
		event.preventDefault();
		faceSearch.searchUrl(address);
	}

	/** A picture pasted anywhere on the page is searched; pasted text is left to the field. */
	function pastePicture(event: ClipboardEvent) {
		const items = event.clipboardData?.items;
		if (!items) return;
		for (const item of items) {
			if (item.kind !== 'file' || !item.type.startsWith('image/')) continue;
			const file = item.getAsFile();
			if (!file) continue;
			event.preventDefault();
			faceSearch.searchFile(file);
			return;
		}
	}

	/** Takes a picture dropped on the zone. */
	function drop(event: DragEvent) {
		event.preventDefault();
		dragging = false;
		searchFiles(event.dataTransfer?.files);
	}
</script>

<svelte:window onpaste={pastePicture} />

<section class="mx-auto flex w-full max-w-3xl flex-col gap-6">
	<div
		class="border-base-300 bg-base-200 flex flex-col items-center gap-4 rounded-2xl border-2 border-dashed p-6 text-center transition-colors"
		class:drop-active={dragging}
		role="region"
		aria-label="Picture to search"
		ondragover={(event) => {
			event.preventDefault();
			dragging = true;
		}}
		ondragleave={() => (dragging = false)}
		ondrop={drop}
	>
		{#if shownPicture}
			<img src={shownPicture} alt="Searched" class="max-h-80 rounded-lg object-contain" />
		{:else}
			<UserFocusIcon size={48} class="text-base-content/30" />
			<div class="flex flex-col gap-1">
				<h2 class="text-lg font-semibold">Who is this?</h2>
				<p class="text-base-content/50 text-sm">
					Drop, paste or pick a picture of a face. NameThatPorn and the other face searches are
					asked who it is.
				</p>
			</div>
		{/if}
		<div class="flex w-full flex-wrap items-center justify-center gap-2">
			<button
				type="button"
				class="flex items-center gap-2 rounded-full bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-white/80"
				onclick={() => fileInput?.click()}
			>
				<UploadSimpleIcon size={16} weight="bold" />
				{#if shownPicture}Another picture{:else}Pick a picture{/if}
			</button>
			<form class="flex min-w-0 flex-1 basis-64 gap-2" onsubmit={submitAddress}>
				<input
					type="url"
					bind:value={address}
					placeholder="…or paste the picture's address"
					class="bg-base-100 border-base-300 placeholder:text-base-content/30 min-w-0 flex-1 rounded-full border px-4 py-2 text-sm outline-none"
				/>
				<button
					type="submit"
					class="bg-base-300 hover:bg-base-content/20 rounded-full px-4 py-2 text-sm font-semibold transition-colors disabled:opacity-50"
					disabled={address.trim() === ''}
				>
					Search
				</button>
			</form>
		</div>
		<input
			bind:this={fileInput}
			type="file"
			accept="image/*"
			class="hidden"
			onchange={(event) => searchFiles(event.currentTarget.files)}
		/>
	</div>

	{#if faceSearch.loading}
		<div class="flex flex-col items-center gap-3 py-10">
			<span class="loading loading-spinner text-base-content/40"></span>
			<p class="text-base-content/40 text-sm">Asking the face searches…</p>
		</div>
	{:else if faceSearch.errorText}
		<p class="text-error py-6 text-center text-sm">{faceSearch.errorText}</p>
	{:else if faceSearch.identification}
		{#key faceSearch.identification}
			<FaceResults identification={faceSearch.identification} />
		{/key}
	{/if}
</section>

<style>
	.drop-active {
		border-color: var(--color-primary);
	}
</style>
