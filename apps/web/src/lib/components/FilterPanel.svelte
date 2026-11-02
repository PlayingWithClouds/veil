<!--
	The library grid's advanced filters, opened from the chip row: length,
	quality, the user's rating, unwatched and favourites, sites, sort, and tags
	to hide. Every choice is a link that changes the URL (libraryFilter), so the
	grid reloads from it.
-->
<script lang="ts">
	import EyeSlashIcon from 'phosphor-svelte/lib/EyeSlashIcon';
	import type { SitePlugin } from '$lib/plugins';
	import {
		LENGTH_OPTIONS,
		RESOLUTION_OPTIONS,
		RATING_OPTIONS,
		SORT_OPTIONS,
		toggledParam,
		toggledListParam,
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
</script>

{#snippet pill(href: string, label: string, active: boolean)}
	{#if active}
		<a
			{href}
			data-sveltekit-noscroll
			data-sveltekit-keepfocus
			aria-current="true"
			class="border-base-content bg-base-content text-base-100 rounded-full border px-3 py-1 text-sm whitespace-nowrap"
		>
			{label}
		</a>
	{:else}
		<a
			{href}
			data-sveltekit-noscroll
			data-sveltekit-keepfocus
			class="border-base-content/10 hover:bg-base-300 rounded-full border px-3 py-1 text-sm whitespace-nowrap"
		>
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

	{#if sites.length > 0}
		<div class="flex flex-col gap-2">
			<h3 class="text-base-content/60 text-xs font-semibold tracking-wide uppercase">Sites</h3>
			<div class="flex flex-wrap gap-2">
				{#each sites as site (site.name)}
					{@render pill(
						toggledListParam(url, 'site', site.name),
						site.displayName ?? site.name,
						chosenSites.includes(site.name)
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
