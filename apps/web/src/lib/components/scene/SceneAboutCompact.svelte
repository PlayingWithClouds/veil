<!--
	Phone watch page body under the actions: performers (inline chips for one
	or two, with Subscribe for a lone one; a row of avatars for more), tags on
	one sideways line ending in "+N", and a details card (description,
	chapters) that springs open on tap, shown only when there is something in it.
-->
<script lang="ts">
	import { slide } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import TagChips from '$lib/components/TagChips.svelte';
	import SubscribeButton from '$lib/components/SubscribeButton.svelte';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, tagUrl } from '$lib/routes';
	import { tagLabel, type TagRef } from '$lib/tags';
	import type { SceneMarker } from '$lib/markers';
	import type { PerformerRef, SceneDetail } from '$lib/sceneDetail';
	import SceneChapters from './SceneChapters.svelte';

	interface Props {
		scene: SceneDetail;
		markers: SceneMarker[];
		onjump: (seconds: number) => void;
		onremovemarker: (markerId: string) => void;
		onblocktag: (tag: TagRef) => void;
	}

	let { scene, markers, onjump, onremovemarker, onblocktag }: Props = $props();

	// Tags on the folded line before the "+N" chip.
	const FOLDED_TAG_COUNT = 6;
	// Up to this many performers show as inline chips instead of the avatar row.
	const INLINE_PERFORMERS = 2;

	let tagsExpanded = $state(false);
	let detailsExpanded = $state(false);

	// A new scene starts folded again.
	$effect(() => {
		scene.id;
		tagsExpanded = false;
		detailsExpanded = false;
	});

	let hiddenTagCount = $derived(Math.max(0, scene.tags.length - FOLDED_TAG_COUNT));
	let hasDetails = $derived(Boolean(scene.details) || markers.length > 0);

	/** The card's collapsed hint for what opening it shows. */
	function chapterHint(): string | null {
		if (markers.length === 0) return null;
		if (markers.length === 1) return '1 chapter';
		return `${markers.length} chapters`;
	}
</script>

{#if scene.performers.length > 0 && scene.performers.length <= INLINE_PERFORMERS}
	<div class="flex items-center gap-2">
		<div class="no-scrollbar flex min-w-0 flex-1 gap-2 overflow-x-auto">
			{#each scene.performers as performer (performer.id)}
				{@render performerChip(performer)}
			{/each}
		</div>
		{#if scene.performers.length === 1}
			<SubscribeButton
				kind="PERFORMER"
				targetId={scene.performers[0].id}
				targetName={scene.performers[0].name}
			/>
		{/if}
	</div>
{:else if scene.performers.length > INLINE_PERFORMERS}
	<ul class="no-scrollbar fade-right-edge -mx-3 flex gap-4 overflow-x-auto px-3 pr-8">
		{#each scene.performers as performer (performer.id)}
			<li class="shrink-0">
				<a href={performerUrl(performer.id)} class="flex w-16 flex-col items-center gap-1.5">
					<span
						class="bg-base-200 flex h-14 w-14 items-center justify-center overflow-hidden rounded-full ring-1 ring-white/10"
					>
						{@render performerPhoto(performer, 22)}
					</span>
					<span class="line-clamp-2 text-center text-xs leading-tight font-medium">{performer.name}</span>
				</a>
			</li>
		{/each}
	</ul>
{/if}

{#if scene.tags.length > 0}
	{#if tagsExpanded}
		<div transition:slide={{ duration: 220 }}>
			<TagChips tags={scene.tags} onblock={onblocktag} />
		</div>
	{:else}
		<div class="no-scrollbar fade-right-edge -mx-3 flex items-center gap-1.5 overflow-x-auto px-3 pr-8">
			{#each scene.tags.slice(0, FOLDED_TAG_COUNT) as tag (tag.id)}
				<a
					href={tagUrl(tag.id)}
					class="border-base-300 bg-base-200 text-base-content/70 shrink-0 rounded-full border px-2.5 py-1 text-xs whitespace-nowrap"
				>
					{tagLabel(tag.name)}
				</a>
			{/each}
			{#if hiddenTagCount > 0}
				<button
					type="button"
					class="bg-base-content/10 shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold"
					onclick={() => (tagsExpanded = true)}
				>
					+{hiddenTagCount}
				</button>
			{/if}
		</div>
	{/if}
{/if}

{#if hasDetails}
	<button
		type="button"
		class="bg-base-200 flex flex-col gap-2 rounded-2xl p-4 text-left text-sm"
		onclick={() => (detailsExpanded = !detailsExpanded)}
		aria-expanded={detailsExpanded}
	>
		<span class="flex items-center justify-between gap-3 font-semibold">
			<span>Details</span>
			<span class="shrink-0 transition-transform duration-300" class:rotate-180={detailsExpanded}>
				<CaretDownIcon size={16} weight="bold" />
			</span>
		</span>
		{#if scene.details}
			<span
				class="text-base-content/75 leading-relaxed whitespace-pre-line"
				class:line-clamp-2={!detailsExpanded}
			>
				{scene.details}
			</span>
		{/if}
		{#if !detailsExpanded && chapterHint()}
			<span class="text-base-content/50 text-xs">{chapterHint()}</span>
		{/if}
	</button>
	{#if detailsExpanded && markers.length > 0}
		<div class="flex flex-col gap-2" transition:slide={{ duration: 420, easing: backOut }}>
			<h2 class="font-semibold">Chapters</h2>
			<SceneChapters {markers} {onjump} onremove={onremovemarker} />
		</div>
	{/if}
{/if}

{#snippet performerChip(performer: PerformerRef)}
	<a
		href={performerUrl(performer.id)}
		class="glass-pill flex h-10 shrink-0 items-center gap-2 rounded-full py-1 pr-4 pl-1"
	>
		<span class="bg-base-200 flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full">
			{@render performerPhoto(performer, 14)}
		</span>
		<span class="truncate text-sm font-semibold">{performer.name}</span>
	</a>
{/snippet}

{#snippet performerPhoto(performer: PerformerRef, iconSize: number)}
	{#if cacheUrl(performer.imagePath)}
		<img src={cacheUrl(performer.imagePath)} alt="" loading="lazy" class="h-full w-full object-cover" />
	{:else}
		<UserIcon size={iconSize} class="opacity-40" />
	{/if}
{/snippet}
