<!--
	The library grid's advanced filters, opened from the chip row: length,
	quality, the user's rating, unwatched and favourites, sites (all of a
	style at once, or one by one, with their logos), sort, and tags to hide. Every choice is a link that changes the URL (libraryFilter), so the
	grid reloads from it.
-->
<script lang="ts">
	import EyeSlashIcon from 'phosphor-svelte/lib/EyeSlashIcon';
	import SiteIcon from '$lib/components/scene/SiteIcon.svelte';
	import { siteChoices, type SitePlugin } from '$lib/plugins';
	import { siteStyles, type SiteStyle } from '$lib/siteStyles';
	import {
		LENGTH_OPTIONS,
		RESOLUTION_OPTIONS,
		RATING_OPTIONS,
		SORT_OPTIONS,
		toggledParam,
		toggledListParam,
		toggledListGroup,
		excludedTagUrl,
		type FilterOption,
		type TagCountRow
	} from '$lib/libraryFilter';

	interface Props {
		url: URL;
		sites: SitePlugin[];
		// The grid's most common tags, offered for hiding.
		tagCounts: TagCountRow[];
	}

	let { url, sites, tagCounts }: Props = $props();

	let params = $derived(url.searchParams);
	let chosenSites = $derived(params.getAll('site'));
	// Every KVS tube on its own, beside the other plugins.
	let choices = $derived(siteChoices(sites));
	let styles = $derived(siteStyles(choices));

	/** Whether every site of a style is picked. */
	function styleOn(style: SiteStyle): boolean {
		return style.sites.every((site) => chosenSites.includes(site));
	}
</script>

{#snippet pill(href: string, label: string, active: boolean, iconUrl: string | null = null)}
	{#if active}
		<a
			{href}
			data-sveltekit-noscroll
			data-sveltekit-keepfocus
			aria-current="true"
			class="border-base-content bg-base-content text-base-100 flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm whitespace-nowrap"
		>
			{#if iconUrl}<SiteIcon {iconUrl} />{/if}
			{label}
		</a>
	{:else}
		<a
			{href}
			data-sveltekit-noscroll
			data-sveltekit-keepfocus
			class="border-base-content/10 hover:bg-base-300 flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm whitespace-nowrap"
		>
			{#if iconUrl}<SiteIcon {iconUrl} />{/if}
			{label}
		</a>
	{/if}
{/snippet}

{#snippet group(title: string, name: string, options: FilterOption[])}
	<div class="flex flex-col gap-2">
		<h3 class="text-base-content/60 text-xs font-semibold tracking-wide uppercase">{title}</h3>
		<div class="flex flex-wrap gap-2">
			{#each options as option (option.value)}
				{@render pill(
					toggledParam(url, name, option.value),
					option.label,
					params.get(name) === option.value
				)}
			{/each}
		</div>
	</div>
{/snippet}

<section
	class="bg-base-200/60 border-base-content/10 grid gap-x-8 gap-y-5 rounded-2xl border p-5 md:grid-cols-2 xl:grid-cols-3"
	aria-label="Filters"
>
	{@render group('Length', 'len', LENGTH_OPTIONS)}
	{@render group('Quality', 'res', RESOLUTION_OPTIONS)}
	{@render group('Your rating', 'rating', RATING_OPTIONS)}

	<div class="flex flex-col gap-2">
		<h3 class="text-base-content/60 text-xs font-semibold tracking-wide uppercase">Show only</h3>
		<div class="flex flex-wrap gap-2">
			{@render pill(toggledParam(url, 'unwatched', '1'), 'Unwatched', params.has('unwatched'))}
			{@render pill(toggledParam(url, 'fav', '1'), 'Favourites', params.has('fav'))}
		</div>
	</div>

	{@render group('Sort by', 'sort', SORT_OPTIONS)}

	{#if choices.length > 0}
		<div class="flex flex-col gap-2 md:col-span-2 xl:col-span-3">
			<h3 class="text-base-content/60 text-xs font-semibold tracking-wide uppercase">Sites</h3>
			<div class="flex flex-wrap gap-2" aria-label="Styles">
				{#each styles as style (style.key)}
					{@render pill(
						toggledListGroup(url, 'site', style.sites),
						`${style.label} · ${style.sites.length}`,
						styleOn(style)
					)}
				{/each}
			</div>
			<div class="flex flex-wrap gap-2">
				{#each choices as choice (choice.value)}
					{@render pill(
						toggledListParam(url, 'site', choice.value),
						choice.label,
						chosenSites.includes(choice.value),
						choice.iconUrl
					)}
				{/each}
			</div>
		</div>
	{/if}

	{#if tagCounts.length > 0}
		<div class="flex flex-col gap-2 md:col-span-2 xl:col-span-3">
			<h3 class="text-base-content/60 text-xs font-semibold tracking-wide uppercase">Hide tags</h3>
			<div class="flex flex-wrap gap-2">
				{#each tagCounts as entry (entry.tag.id)}
					<a
						href={excludedTagUrl(url, entry.tag.id)}
						data-sveltekit-noscroll
						data-sveltekit-keepfocus
						class="border-base-content/10 hover:border-error hover:text-error flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm whitespace-nowrap"
					>
						<EyeSlashIcon size={14} />
						{entry.tag.name}
						<span class="text-base-content/40 tabular-nums">{entry.count}</span>
					</a>
				{/each}
			</div>
		</div>
	{/if}
</section>
