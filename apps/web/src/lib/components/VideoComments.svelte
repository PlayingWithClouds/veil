<!--
	The comments the video's sites carry, as threads: each comment with its
	replies under it, the most liked threads first. Shows the first few;
	"Show all" lists the rest.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import type { VideoDetail } from '$lib/library';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';

	type CommentRow = VideoDetail['comments'][number];

	/** A comment and the replies to it. */
	interface Thread {
		comment: CommentRow;
		replies: CommentRow[];
	}

	let { comments }: { comments: CommentRow[] } = $props();

	const FIRST_THREADS = 6;

	let showAll = $state(false);
	let threads = $derived(threadsOf(comments));
	let shown = $derived(showAll ? threads : threads.slice(0, FIRST_THREADS));

	/** Groups replies under their comment, the most liked threads first. */
	function threadsOf(all: CommentRow[]): Thread[] {
		const byId = new Map<string, Thread>();
		for (const comment of all) {
			if (!comment.replyTo) byId.set(comment.id, { comment, replies: [] });
		}
		for (const comment of all) {
			if (comment.replyTo) byId.get(comment.replyTo)?.replies.push(comment);
		}
		return [...byId.values()].sort((first, second) => (second.comment.likes ?? 0) - (first.comment.likes ?? 0));
	}

	/** "3 years ago", or nothing without a date. */
	function when(postedAt: string | null): string {
		if (!postedAt) return '';
		return dayjs(postedAt).fromNow();
	}
</script>

{#snippet line(comment: CommentRow)}
	<div class="flex flex-col gap-0.5">
		<span class="text-base-content/50 flex flex-wrap items-center gap-x-2 text-xs">
			<span class="text-base-content/80 font-medium">{comment.author ?? 'Anonymous'}</span>
			{#if comment.postedAt}<span>{when(comment.postedAt)}</span>{/if}
			{#if comment.source}<span>on {siteDisplayName(comment.source.plugin, $sites)}</span>{/if}
			{#if comment.likes}
				<span class="flex items-center gap-0.5"><HeartIcon size={11} /> {comment.likes}</span>
			{/if}
		</span>
		<p class="text-sm leading-relaxed whitespace-pre-line">{comment.text}</p>
	</div>
{/snippet}

{#if threads.length > 0}
	<section class="flex flex-col gap-4">
		<h2 class="text-lg font-semibold">Comments <span class="text-base-content/40 font-normal">{comments.length}</span></h2>
		{#each shown as thread (thread.comment.id)}
			<div class="flex flex-col gap-3">
				{@render line(thread.comment)}
				{#if thread.replies.length > 0}
					<div class="border-base-content/10 ml-4 flex flex-col gap-3 border-l pl-4">
						{#each thread.replies as reply (reply.id)}
							{@render line(reply)}
						{/each}
					</div>
				{/if}
			</div>
		{/each}
		{#if threads.length > FIRST_THREADS && !showAll}
			<button type="button" class="btn btn-sm self-start" onclick={() => (showAll = true)}>
				Show all {threads.length} threads
			</button>
		{/if}
	</section>
{/if}
