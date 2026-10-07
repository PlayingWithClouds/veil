<script lang="ts">
	import BookmarkSimpleIcon from 'phosphor-svelte/lib/BookmarkSimpleIcon';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import { notifications } from '$lib/stores/notifications';
	import { addToWatchlist, removeFromWatchlist, fetchWatchlist } from '$lib/discovery';

	interface Props {
		mediaId: string;
		class?: string;
	}

	let { mediaId, class: className = '' }: Props = $props();

	let inWatchlist = $state(false);
	let busy = $state(false);

	// Reflect current membership; re-checks when the media changes.
	$effect(() => {
		if (!mediaId) {
			inWatchlist = false;
			return;
		}
		let cancelled = false;
		fetchWatchlist().then((entries) => {
			if (!cancelled) inWatchlist = entries.some((entry) => entry.id === mediaId);
		});
		return () => {
			cancelled = true;
		};
	});

	async function toggle() {
		if (!mediaId || busy) return;
		busy = true;
		const next = !inWatchlist;
		inWatchlist = next;
		try {
			if (next) await addToWatchlist(mediaId);
			else await removeFromWatchlist(mediaId);
			notifications.push(next ? 'Added to watchlist' : 'Removed from watchlist', 'success');
		} catch {
			inWatchlist = !next;
			notifications.push('Could not update watchlist', 'error');
		} finally {
			busy = false;
		}
	}
</script>

<button
	type="button"
	onclick={toggle}
	disabled={busy}
	aria-pressed={inWatchlist}
	class="flex items-center gap-1.5 rounded-full bg-white/10 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-white/20 disabled:opacity-50 {className}"
>
	{#if inWatchlist}
		<CheckIcon size={18} weight="bold" />
		In watchlist
	{:else}
		<BookmarkSimpleIcon size={18} />
		Watchlist
	{/if}
</button>
