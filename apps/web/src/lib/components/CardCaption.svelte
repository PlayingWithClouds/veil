<!--
	YouTube-style caption under a card: avatar (channel, performer or site) on
	the left; title, who is in it, and site · age on the right. Lines that have
	nothing to show collapse, so the text sits right under its media. The title
	is the card's button and its click area covers the caption; performer names
	sit above it as their own links. Phones fold everything under the title into
	one line. A ⋮ button shows when onmore is set.
-->
<script lang="ts">
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import type { SitePlugin } from '$lib/plugins';
	import type { NamedRef } from '$lib/library';
	import { performerUrl } from '$lib/routes';
	import { isCompact } from '$lib/stores/viewport';

	interface Props {
		title: string;
		// Channel/studio name, else credited performers. Null leaves the line empty.
		byline?: string | null;
		// Credited performers, linked, on their own line above the byline.
		performers?: NamedRef[];
		// Channel or performer image for the avatar; falls back to the site icon.
		avatarUrl?: string | null;
		site?: SitePlugin | null;
		// Shown instead of the site's name, e.g. the host for a plugin covering many sites.
		siteName?: string | null;
		// Trailing detail on the last line (age, image count, …).
		detail?: string | null;
		// A last, quieter line, e.g. why the card was recommended.
		note?: string | null;
		onclick: () => void;
		// Opens the card's menu; shows a ⋮ button when set. Gets the button to anchor to.
		onmore?: (button: HTMLElement) => void;
	}

	let {
		title,
		byline = null,
		performers = [],
		avatarUrl = null,
		site = null,
		siteName = null,
		detail = null,
		note = null,
		onclick,
		onmore
	}: Props = $props();

	// Performers named on the card; the rest show as "+n".
	const SHOWN_PERFORMERS = 3;

	let moreButton = $state<HTMLElement | null>(null);

	let shownSite = $derived.by(() => {
		if (siteName) return siteName;
		if (site) return siteLabel(site);
		return null;
	});
	let shownPerformers = $derived(performers.slice(0, SHOWN_PERFORMERS));
	let hiddenPerformerCount = $derived(performers.length - shownPerformers.length);

	// Phones fold byline, site and age into one grey line under the title.
	let compactMeta = $derived.by(() => {
		const parts: string[] = [];
		if (byline) parts.push(byline);
		if (shownSite) parts.push(shownSite);
		if (detail) parts.push(detail);
		return parts.join(' · ');
	});

	let avatarSource = $derived(avatarUrl || site?.iconUrl || null);
	let avatarInitial = $derived((byline || title || '?').trim().charAt(0).toUpperCase());

	/** The site's display label. */
	function siteLabel(plugin: SitePlugin): string {
		if (plugin.displayName) return plugin.displayName;
		return plugin.name;
	}

	/** Opens the menu from the ⋮ button. */
	function openMore() {
		if (onmore && moreButton) onmore(moreButton);
	}
</script>

<div class="relative flex items-start gap-1">
	<div class="flex min-w-0 flex-1 gap-3">
		<span
			class="bg-base-300 flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-full"
		>
			{#if avatarSource}
				<img src={avatarSource} alt="" loading="lazy" class="h-full w-full object-cover" />
			{:else}
				<span class="text-base-content/60 text-sm font-semibold">{avatarInitial}</span>
			{/if}
		</span>
		<span class="flex min-w-0 flex-1 flex-col gap-0.5">
			<button
				type="button"
				{onclick}
				{title}
				class="text-left after:absolute after:inset-0 focus:outline-none"
			>
				<span class="line-clamp-2 text-[15px] leading-5 font-medium">{title}</span>
			</button>
			{#if shownPerformers.length > 0}
				<span class="text-base-content/80 relative truncate text-[13px] leading-[18px]">
					{#each shownPerformers as performer, position (performer.id)}
						{#if position > 0},
						{/if}<a href={performerUrl(performer.id)} class="hover:underline">{performer.name}</a>
					{/each}
					{#if hiddenPerformerCount > 0}
						<span class="text-base-content/50">+{hiddenPerformerCount}</span>
					{/if}
				</span>
			{/if}
			{#if $isCompact}
				{#if compactMeta}
					<span class="text-base-content/60 truncate text-[13px] leading-[18px]">{compactMeta}</span
					>
				{/if}
			{:else if byline}
				<span class="text-base-content/60 truncate text-[13px] leading-[18px]">{byline}</span>
			{/if}
			{#if !$isCompact && (shownSite || detail)}
				<span class="text-base-content/60 flex items-center gap-1 text-[13px] leading-[18px]">
					{#if shownSite}
						<span class="truncate">{shownSite}</span>
					{/if}
					{#if shownSite && detail}
						<span aria-hidden="true">·</span>
					{/if}
					{#if detail}
						<span class="shrink-0">{detail}</span>
					{/if}
				</span>
			{/if}
			{#if note}
				<span class="text-base-content/40 truncate text-xs leading-4">{note}</span>
			{/if}
		</span>
	</div>
	{#if onmore}
		<button
			type="button"
			bind:this={moreButton}
			onclick={openMore}
			aria-label="More actions"
			class="text-base-content/70 hover:bg-base-content/10 relative -mr-2 flex h-9 w-9 shrink-0 items-center justify-center rounded-full"
		>
			<DotsThreeVerticalIcon size={20} weight="bold" />
		</button>
	{/if}
</div>
