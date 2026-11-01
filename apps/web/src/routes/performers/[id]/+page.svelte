<!-- One performer: picture, facts, links, tags, and their videos and galleries. -->
<script lang="ts">
	import { page } from '$app/state';
	import ArrowSquareOutIcon from 'phosphor-svelte/lib/ArrowSquareOutIcon';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import TagChips from '$lib/components/TagChips.svelte';
	import RecordMedia from '$lib/components/RecordMedia.svelte';
	import { fetchPerformer, type PerformerDetail } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { cacheUrl } from '$lib/img';

	// Labels for the facts a performer page shows, in this order.
	const ATTRIBUTE_LABELS: [string, string][] = [
		['gender', 'Gender'],
		['birthdate', 'Born'],
		['country', 'Country'],
		['ethnicity', 'Ethnicity'],
		['hairColor', 'Hair'],
		['eyeColor', 'Eyes'],
		['heightCm', 'Height (cm)'],
		['weightKg', 'Weight (kg)'],
		['measurements', 'Measurements'],
		['tattoos', 'Tattoos'],
		['piercings', 'Piercings'],
		['careerStart', 'Active since'],
		['careerEnd', 'Active until']
	];

	let performer = $state<PerformerDetail | null>(null);
	let loading = $state(true);

	let id = $derived(page.params.id ?? '');
	let facts = $derived(
		ATTRIBUTE_LABELS.flatMap(([key, label]) => {
			const value = performer?.attributes[key];
			if (value === null || value === undefined || value === '') return [];
			return [{ label, value: String(value) }];
		})
	);

	$effect(() => {
		loading = true;
		fetchPerformer(id)
			.then((found) => (performer = found))
			.finally(() => (loading = false));
	});
</script>

<svelte:head><title>{performer?.name ?? 'Performer'} · Veil</title></svelte:head>

<div class="flex flex-col gap-6 pt-2">
	{#if performer}
		<div class="flex gap-5 max-md:flex-col">
			<div
				class="bg-base-200 flex aspect-[2/3] w-48 shrink-0 items-center justify-center overflow-hidden rounded-xl"
			>
				{#if performer.image}
					<img
						src={cacheUrl(performer.image.uri)}
						alt={performer.name}
						class="h-full w-full object-cover"
					/>
				{:else}
					<UserIcon size={48} class="opacity-15" />
				{/if}
			</div>
			<div class="flex min-w-0 flex-col gap-3">
				<div>
					<h1 class="text-2xl font-bold tracking-tight">{performer.name}</h1>
					{#if performer.aliases.length > 0}
						<p class="text-base-content/50 text-sm">Also: {performer.aliases.join(', ')}</p>
					{/if}
				</div>
				{#if facts.length > 0}
					<dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
						{#each facts as fact (fact.label)}
							<dt class="text-base-content/50">{fact.label}</dt>
							<dd>{fact.value}</dd>
						{/each}
					</dl>
				{/if}
				{#if performer.details}
					<p class="text-base-content/80 max-w-3xl text-sm whitespace-pre-line">
						{performer.details}
					</p>
				{/if}
				<TagChips tags={performer.tags} />
				<div class="flex flex-wrap gap-x-4 gap-y-1 text-sm">
					{#each performer.links as link (link)}
						<a
							href={link}
							target="_blank"
							rel="noreferrer"
							class="text-base-content/70 hover:text-base-content flex items-center gap-1"
						>
							<ArrowSquareOutIcon size={13} />
							{new URL(link).hostname.replace(/^www\./, '')}
						</a>
					{/each}
					{#each performer.sources as source (source.url)}
						<a
							href={source.url}
							target="_blank"
							rel="noreferrer"
							class="text-base-content/70 hover:text-base-content flex items-center gap-1"
						>
							<ArrowSquareOutIcon size={13} />
							{siteDisplayName(source.plugin, $sites)}
						</a>
					{/each}
				</div>
			</div>
		</div>

		<RecordMedia
			filter={{ performerId: performer.id }}
			allVideosHref={`/?${new URLSearchParams({ performer: performer.id })}`}
		/>
	{:else if !loading}
		<p class="text-base-content/50 text-sm">This performer isn't in the library.</p>
	{/if}
</div>
