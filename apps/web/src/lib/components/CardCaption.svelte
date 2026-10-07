<!--
	YouTube-style caption under an Explore card: avatar (channel, performer or
	site) on the left; title, who is in it, and site · age on the right. Lines
	that have nothing to show collapse, so the text sits right under its media.
	Phones fold everything under the title into one line and may add a ⋮ button.
-->
<script module lang="ts">
	/**
	 * Typical caption height including the gap above it, for the masonry's first
	 * placement estimate; it re-measures rendered columns before placing more.
	 */
	export const CARD_CAPTION_HEIGHT = 64;
</script>

<script lang="ts">
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import type { SearchPlugin } from '$lib/search';
	import { isCompact } from '$lib/stores/viewport';

	interface Props {
		title: string;
		// Channel/studio name, else credited performers. Null leaves the line empty.
		byline?: string | null;
		// Channel or performer image for the avatar; falls back to the site icon.
		avatarUrl?: string | null;
		site?: SearchPlugin | null;
		// Trailing detail on the last line (age, image count, …).
		detail?: string | null;
		onclick: () => void;
		// Opens the card's action sheet; shows a ⋮ button when set.
		onmore?: () => void;
	}

	let {
		title,
		byline = null,
		avatarUrl = null,
		site = null,
		detail = null,
		onclick,
		onmore
	}: Props = $props();

	// Phones fold byline, site and age into one grey line under the title.
	let compactMeta = $derived.by(() => {
		const parts: string[] = [];
		if (byline) parts.push(byline);
		if (site) parts.push(siteLabel(site));
		if (detail) parts.push(detail);
		return parts.join(' · ');
	});

	let avatarSource = $derived(avatarUrl || site?.iconUrl || null);
	let avatarInitial = $derived((byline || title || '?').trim().charAt(0).toUpperCase());

	/** The site's display label. */
	function siteLabel(plugin: SearchPlugin): string {
		if (plugin.displayName) return plugin.displayName;
		return plugin.name;
	}
</script>

<div class="flex items-start gap-1">
<button type="button" {onclick} class="flex min-w-0 flex-1 gap-3 text-left" {title}>
	<span class="bg-base-300 flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-full">
		{#if avatarSource}
			<img src={avatarSource} alt="" loading="lazy" class="h-full w-full object-cover" />
		{:else}
			<span class="text-base-content/60 text-sm font-semibold">{avatarInitial}</span>
		{/if}
	</span>
	<span class="flex min-w-0 flex-1 flex-col gap-0.5">
		<span class="line-clamp-2 text-[15px] leading-5 font-medium">{title}</span>
		{#if $isCompact}
			{#if compactMeta}
				<span class="text-base-content/60 truncate text-[13px] leading-[18px]">{compactMeta}</span>
			{/if}
		{:else if byline}
			<span class="text-base-content/60 truncate text-[13px] leading-[18px]">{byline}</span>
		{/if}
		{#if !$isCompact && (site || detail)}
			<span class="text-base-content/60 flex items-center gap-1 text-[13px] leading-[18px]">
				{#if site}
					<span class="truncate">{siteLabel(site)}</span>
				{/if}
				{#if site && detail}
					<span aria-hidden="true">·</span>
				{/if}
				{#if detail}
					<span class="shrink-0">{detail}</span>
				{/if}
			</span>
		{/if}
	</span>
</button>
{#if onmore}
	<button
		type="button"
		onclick={onmore}
		aria-label="More actions"
		class="text-base-content/70 -mr-2 flex h-9 w-9 shrink-0 items-center justify-center rounded-full"
	>
		<DotsThreeVerticalIcon size={20} weight="bold" />
	</button>
{/if}
</div>
