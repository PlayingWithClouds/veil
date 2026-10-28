<script lang="ts">
	// The performers and studios a search names, as channel cards (photo, name,
	// scenes in the library, Subscribe), and the tags it names as chips. A name
	// only spelled close to the query reads "Did you mean …".
	import TagIcon from 'phosphor-svelte/lib/TagIcon';
	import SubscribeButton from '$lib/components/SubscribeButton.svelte';
	import { initials, formatCount } from '$lib/components/scene/format';
	import { cacheUrl } from '$lib/img';
	import { performerUrl, studioUrl, tagUrl } from '$lib/routes';
	import type { SearchMatchedEntity } from '$lib/search';
	import type { FollowableKind } from '$lib/searchSubscriptions';
	import { fetchMatchedEntityCards, type MatchedEntityCard } from '$lib/searchEntities';

	interface Props {
		entities: SearchMatchedEntity[];
	}

	let { entities }: Props = $props();

	let cards = $state<MatchedEntityCard[]>([]);
	let tags = $derived(entities.filter((entity) => entity.kind === 'tag'));

	$effect(() => {
		const current = entities;
		cards = [];
		fetchMatchedEntityCards(current).then((loaded) => {
			if (entities === current) cards = loaded;
		});
	});

	/** The page of a card's performer or studio. */
	function cardUrl(card: MatchedEntityCard): string {
		if (card.kind === 'performer') return performerUrl(card.id);
		return studioUrl(card.id);
	}

	/** What subscribing to the card follows. */
	function subscriptionKind(card: MatchedEntityCard): FollowableKind {
		if (card.kind === 'performer') return 'PERFORMER';
		return 'STUDIO';
	}

	/** "11 scenes in your library", or what kind of entity it is when there are none. */
	function cardDetail(card: MatchedEntityCard): string {
		if (card.sceneCount === 1) return '1 scene in your library';
		if (card.sceneCount > 1) return `${formatCount(card.sceneCount)} scenes in your library`;
		if (card.kind === 'performer') return 'Performer';
		return 'Studio';
	}
</script>

{#if cards.length > 0 || tags.length > 0}
	<div class="flex flex-col gap-3">
		{#each cards as card (card.id)}
			<div class="bg-base-200 flex items-center gap-4 rounded-2xl p-3 pr-4">
				<a href={cardUrl(card)} class="group flex min-w-0 flex-1 items-center gap-4">
					<span
						class="bg-base-300 text-base-content/60 flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-full text-lg font-semibold"
						class:rounded-xl={card.kind === 'studio'}
					>
						{#if cacheUrl(card.imagePath)}
							<img
								src={cacheUrl(card.imagePath)}
								alt=""
								loading="lazy"
								class="h-full w-full object-cover object-top"
								class:object-contain={card.kind === 'studio'}
							/>
						{:else}
							{initials(card.name)}
						{/if}
					</span>
					<span class="flex min-w-0 flex-col gap-0.5">
						{#if card.fuzzy}
							<span class="text-base-content/50 text-xs">Did you mean</span>
						{/if}
						<span class="truncate text-lg font-semibold group-hover:underline">{card.name}</span>
						<span class="text-base-content/50 text-sm">{cardDetail(card)}</span>
					</span>
				</a>
				<SubscribeButton
					kind={subscriptionKind(card)}
					targetId={card.id}
					targetName={card.name}
				/>
			</div>
		{/each}
		{#if tags.length > 0}
			<div class="flex flex-wrap gap-2">
				{#each tags as tag (tag.id)}
					<a
						href={tagUrl(tag.id)}
						class="bg-base-200 hover:bg-base-300 flex h-8 items-center gap-1.5 rounded-lg px-3 text-sm font-medium transition-colors"
					>
						<TagIcon size={14} class="opacity-60" />
						{#if tag.fuzzy}<span class="text-base-content/50">Did you mean</span>{/if}
						{tag.name}
					</a>
				{/each}
			</div>
		{/if}
	</div>
{/if}
