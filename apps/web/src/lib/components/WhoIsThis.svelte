<!--
	"Who is this?": a camera button in the search bar. Pick a picture and the
	identification plugins say who (performers) or what (videos) is in it;
	the answer opens in a panel over the page.
-->
<script lang="ts">
	import { goto, afterNavigate } from '$app/navigation';
	import CameraIcon from 'phosphor-svelte/lib/CameraIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import MemberCard from '$lib/components/cards/MemberCard.svelte';
	import SceneCard from '$lib/components/SceneCard.svelte';
	import { lookUpPicture, type LookupFound } from '$lib/lookup';
	import { cacheUrl } from '$lib/img';
	import { videoUrl } from '$lib/routes';
	import { pushOverlay } from '$lib/stores/overlayStack';

	let picker = $state<HTMLInputElement>();
	let open = $state(false);
	let pictureUrl = $state<string | null>(null);
	let found = $state<LookupFound | null>(null);
	let error = $state<string | null>(null);

	let nothingFound = $derived(
		found !== null &&
			found.performers.length === 0 &&
			found.videos.length === 0 &&
			found.elsewhere.length === 0
	);

	$effect(() => {
		if (open) return pushOverlay(close);
	});

	// Picking a performer or video leaves the panel behind.
	afterNavigate(close);

	/** Looks up the picked picture. */
	async function onPicked() {
		const picture = picker?.files?.[0];
		if (!picture) return;
		if (pictureUrl) URL.revokeObjectURL(pictureUrl);
		pictureUrl = URL.createObjectURL(picture);
		found = null;
		error = null;
		open = true;
		try {
			found = await lookUpPicture(picture);
		} catch (failure) {
			error = String(failure);
		} finally {
			if (picker) picker.value = '';
		}
	}

	/** Closes the panel. */
	function close() {
		open = false;
	}
</script>

<button
	type="button"
	aria-label="Who is this? Search by picture"
	title="Who is this?"
	class="text-base-content/60 hover:text-base-content mr-1.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full"
	onclick={() => picker?.click()}
>
	<CameraIcon size={18} />
</button>
<input bind:this={picker} type="file" accept="image/*" class="hidden" onchange={onPicked} />

{#if open}
	<button
		type="button"
		class="z-modal fixed inset-0 bg-black/60"
		aria-label="Close"
		tabindex="-1"
		onclick={close}
	></button>
	<div
		class="bg-base-100 z-modal fixed inset-x-4 top-16 bottom-4 mx-auto flex max-w-4xl flex-col gap-5 overflow-y-auto rounded-2xl p-6 shadow-2xl ring-1 ring-white/10"
		role="dialog"
		aria-modal="true"
		aria-label="Who is this?"
	>
		<div class="flex items-start gap-4">
			{#if pictureUrl}
				<img src={pictureUrl} alt="" class="h-28 w-28 shrink-0 rounded-xl object-cover" />
			{/if}
			<div class="flex flex-1 flex-col gap-1">
				<h2 class="text-xl font-semibold">Who is this?</h2>
				{#if !found && !error}
					<p class="text-base-content/60 flex items-center gap-2 text-sm">
						<span class="loading loading-spinner loading-sm"></span> Asking the identification sites…
					</p>
				{/if}
				{#if nothingFound}
					<p class="text-base-content/60 text-sm">No site recognised anyone in this picture.</p>
				{/if}
				{#if error}
					<p class="text-error text-sm">{error}</p>
				{/if}
			</div>
			<button
				type="button"
				aria-label="Close"
				class="hover:bg-base-200 flex h-9 w-9 items-center justify-center rounded-full"
				onclick={close}
			>
				<XIcon size={20} />
			</button>
		</div>

		{#if found && found.performers.length > 0}
			<section class="flex flex-col gap-3">
				<h3 class="font-semibold">Performers</h3>
				<div class="grid grid-cols-[repeat(auto-fill,minmax(130px,1fr))] gap-4">
					{#each found.performers as match (match.performer.id)}
						<div class="flex flex-col gap-1">
							<MemberCard kind="performer" member={match.performer} />
							{#if match.evidence[0]?.detail}
								<span class="text-primary text-xs">{match.evidence[0].detail}</span>
							{/if}
						</div>
					{/each}
				</div>
			</section>
		{/if}

		{#if found && found.videos.length > 0}
			<section class="flex flex-col gap-3">
				<h3 class="font-semibold">Videos</h3>
				<div class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-4">
					{#each found.videos as match (match.video.id)}
						<SceneCard
							video={match.video}
							note={match.evidence[0]?.detail}
							onclick={() => goto(videoUrl(match.video.id))}
						/>
					{/each}
				</div>
			</section>
		{/if}

		{#if found && found.elsewhere.length > 0}
			<section class="flex flex-col gap-2">
				<h3 class="font-semibold">Elsewhere</h3>
				{#each found.elsewhere as evidence, index (index)}
					<a
						href={evidence.url ?? undefined}
						target="_blank"
						rel="noreferrer"
						class="hover:bg-base-200 flex items-center gap-3 rounded-lg p-2"
					>
						{#if evidence.image}
							<img
								src={cacheUrl(evidence.image.uri)}
								alt=""
								class="h-12 w-12 shrink-0 rounded-lg object-cover"
							/>
						{/if}
						<span class="flex min-w-0 flex-1 flex-col">
							<span class="truncate text-sm font-medium">{evidence.title ?? evidence.url}</span>
							<span class="text-base-content/50 text-xs">
								{evidence.plugin}{#if evidence.detail}&nbsp;· {evidence.detail}{/if}
							</span>
						</span>
						<ArrowSquareOutIcon size={16} class="text-base-content/40 shrink-0" />
					</a>
				{/each}
			</section>
		{/if}

		{#if found && found.failures.length > 0}
			<p class="text-base-content/40 text-xs">Didn't answer: {found.failures.join(', ')}</p>
		{/if}
	</div>
{/if}
