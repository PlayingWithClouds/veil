<script lang="ts">
	// The watch page's one "keep this" pill: a menu toggling Favorites and the
	// watchlist and opening the collection picker. The pill shows filled while the
	// scene is in Favorites or the watchlist.
	import BookmarkSimpleIcon from 'phosphor-svelte/lib/BookmarkSimpleIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import ListPlusIcon from 'phosphor-svelte/lib/ListPlusIcon';
	import { notifications } from '$lib/stores/notifications';
	import { openMoveDialog } from '$lib/stores/moveDialog';
	import {
		FAVORITES_COLLECTION_ID,
		addToCollection,
		collectionIdsForEntity,
		removeFromCollection
	} from '$lib/collections';
	import { addToWatchlist, fetchWatchlist, removeFromWatchlist } from '$lib/discovery';
	import PopoverMenu from './PopoverMenu.svelte';

	interface Props {
		sceneId: string;
		sceneTitle: string;
		triggerClass: string;
		menuItemClass: string;
	}

	let { sceneId, sceneTitle, triggerClass, menuItemClass }: Props = $props();

	let favorite = $state(false);
	let inWatchlist = $state(false);
	let busy = $state(false);
	let saved = $derived(favorite || inWatchlist);

	// Reflect current membership; re-checks when the scene changes.
	$effect(() => {
		const id = sceneId;
		favorite = false;
		inWatchlist = false;
		let cancelled = false;
		collectionIdsForEntity(id).then((ids) => {
			if (!cancelled) favorite = ids.includes(FAVORITES_COLLECTION_ID);
		});
		fetchWatchlist().then((entries) => {
			if (!cancelled) inWatchlist = entries.some((entry) => entry.id === id);
		});
		return () => {
			cancelled = true;
		};
	});

	/** Adds to or removes from Favorites, rolling back when that fails. */
	async function toggleFavorite(close: () => void) {
		close();
		if (busy) return;
		busy = true;
		const next = !favorite;
		favorite = next;
		try {
			if (next) {
				await addToCollection(FAVORITES_COLLECTION_ID, sceneId);
				notifications.push('Added to favorites', 'success');
			} else {
				await removeFromCollection(FAVORITES_COLLECTION_ID, sceneId);
				notifications.push('Removed from favorites', 'success');
			}
		} catch {
			favorite = !next;
			notifications.push('Could not update favorites', 'error');
		} finally {
			busy = false;
		}
	}

	/** Adds to or removes from the watchlist, rolling back when that fails. */
	async function toggleWatchlist(close: () => void) {
		close();
		if (busy) return;
		busy = true;
		const next = !inWatchlist;
		inWatchlist = next;
		try {
			if (next) {
				await addToWatchlist(sceneId);
				notifications.push('Added to watchlist', 'success');
			} else {
				await removeFromWatchlist(sceneId);
				notifications.push('Removed from watchlist', 'success');
			}
		} catch {
			inWatchlist = !next;
			notifications.push('Could not update watchlist', 'error');
		} finally {
			busy = false;
		}
	}

	/** Opens the collection picker. */
	function addToCollectionDialog(close: () => void) {
		close();
		openMoveDialog('scene', sceneId, sceneTitle);
	}
</script>

<PopoverMenu label="Save" {triggerClass} disabled={busy}>
	{#snippet trigger()}
		<BookmarkSimpleIcon size={18} weight={saved ? 'fill' : 'regular'} />
		{#if saved}Saved{:else}Save{/if}
		<CaretDownIcon size={12} weight="bold" />
	{/snippet}
	{#snippet children(close)}
		<button
			type="button"
			role="menuitemcheckbox"
			aria-checked={favorite}
			class={menuItemClass}
			onclick={() => toggleFavorite(close)}
		>
			<HeartIcon size={16} weight={favorite ? 'fill' : 'regular'} class={favorite ? 'text-error' : ''} />
			Favorite
			{#if favorite}<CheckIcon size={14} weight="bold" class="ml-auto" />{/if}
		</button>
		<button
			type="button"
			role="menuitemcheckbox"
			aria-checked={inWatchlist}
			class={menuItemClass}
			onclick={() => toggleWatchlist(close)}
		>
			<BookmarkSimpleIcon size={16} weight={inWatchlist ? 'fill' : 'regular'} />
			Watch later
			{#if inWatchlist}<CheckIcon size={14} weight="bold" class="ml-auto" />{/if}
		</button>
		<button
			type="button"
			role="menuitem"
			class={menuItemClass}
			onclick={() => addToCollectionDialog(close)}
		>
			<ListPlusIcon size={16} />
			Add to collection…
		</button>
	{/snippet}
</PopoverMenu>
