<script lang="ts">
	// Heart pill that adds a scene, gallery or other media to the built-in
	// Favorites collection, or takes it out again.
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import { notifications } from '$lib/stores/notifications';
	import {
		FAVORITES_COLLECTION_ID,
		addToCollection,
		collectionIdsForEntity,
		removeFromCollection
	} from '$lib/collections';

	interface Props {
		mediaId: string;
		class?: string;
	}

	let { mediaId, class: className = '' }: Props = $props();

	let favorite = $state(false);
	let busy = $state(false);

	// Reflect current membership; re-checks when the media changes.
	$effect(() => {
		if (!mediaId) {
			favorite = false;
			return;
		}
		let cancelled = false;
		collectionIdsForEntity(mediaId).then((ids) => {
			if (!cancelled) favorite = ids.includes(FAVORITES_COLLECTION_ID);
		});
		return () => {
			cancelled = true;
		};
	});

	/** Adds to or removes from Favorites, rolling back the heart when that fails. */
	async function toggle() {
		if (!mediaId || busy) return;
		busy = true;
		const next = !favorite;
		favorite = next;
		try {
			if (next) await addToCollection(FAVORITES_COLLECTION_ID, mediaId);
			else await removeFromCollection(FAVORITES_COLLECTION_ID, mediaId);
			notifications.push(next ? 'Added to favorites' : 'Removed from favorites', 'success');
		} catch {
			favorite = !next;
			notifications.push('Could not update favorites', 'error');
		} finally {
			busy = false;
		}
	}
</script>

<button
	type="button"
	onclick={toggle}
	disabled={busy}
	aria-pressed={favorite}
	aria-label={favorite ? 'Remove from favorites' : 'Add to favorites'}
	title={favorite ? 'Remove from favorites' : 'Add to favorites'}
	class="flex items-center gap-1.5 rounded-full bg-white/10 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-white/20 disabled:opacity-50 {className}"
	class:text-error={favorite}
>
	<HeartIcon size={18} weight={favorite ? 'fill' : 'regular'} />
	Favorite
</button>
