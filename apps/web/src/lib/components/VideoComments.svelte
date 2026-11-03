<!--
	The comments the video's sites carry, as threads: each comment with an
	initial avatar, its author, age, site and likes, and its replies folded
	under it ("3 replies" opens them); the most liked threads first. Shows
	the first few; "Show all" lists the rest.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
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
	// Threads whose replies are open, by comment id.
	let openThreads = $state<string[]>([]);
	let threads = $derived(threadsOf(comments));
	let shown = $derived(showAll ? threads : threads.slice(0, FIRST_THREADS));

	/** Groups replies under their comment, the most liked threads first. */
	function threadsOf(all: CommentRow[]): Thread[] {
		const byId: Record<string, Thread> = {};
		for (const comment of all) {
			if (!comment.replyTo) byId[comment.id] = { comment, replies: [] };
		}
		for (const comment of all) {
			if (comment.replyTo) byId[comment.replyTo]?.replies.push(comment);
		}
		return Object.values(byId).sort(
			(first, second) => (second.comment.likes ?? 0) - (first.comment.likes ?? 0)
		);
	}

	/** "3 years ago", or nothing without a date. */
	function when(postedAt: string | null): string {
		if (!postedAt) return '';
		return dayjs(postedAt).fromNow();
	}

	/** A steady colour per author, for the avatar. */
	function hueOf(name: string): number {
		let hash = 0;
		for (const character of name) hash = (hash * 31 + character.charCodeAt(0)) % 360;
		return hash;
	}

	/** Opens or folds a thread's replies. */
	function toggleReplies(commentId: string) {
		if (openThreads.includes(commentId)) {
			openThreads = openThreads.filter((id) => id !== commentId);
			return;
		}
		openThreads = [...openThreads, commentId];
	}
</script>

{#snippet entry(comment: CommentRow, small: boolean)}
	{@const author = comment.author ?? 'Anonymous'}
	<div class="flex gap-3">
		<span
			class="flex shrink-0 items-center justify-center rounded-full font-semibold text-black/80"
			class:h-9={!small}
			class:w-9={!small}
			class:text-sm={!small}
			class:h-7={small}
			class:w-7={small}
			class:text-xs={small}
			style:background-color="hsl({hueOf(author)} 55% 70%)"
			aria-hidden="true"
		>
			{author.charAt(0).toUpperCase()}
		</span>
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<span class="text-base-content/45 flex flex-wrap items-center gap-x-2 text-xs">
				<span class="text-base-content text-[13px] font-semibold">{author}</span>
				{#if comment.postedAt}<span>{when(comment.postedAt)}</span>{/if}
				{#if comment.source}<span>· {siteDisplayName(comment.source.plugin, $sites)}</span>{/if}
			</span>
			<p class="text-base-content/90 text-sm leading-relaxed break-words whitespace-pre-line">
				{comment.text}
			</p>
			{#if comment.likes}
				<span
					class="text-base-content/55 flex items-center gap-1 self-start rounded-full bg-white/5 px-2 py-0.5 text-xs tabular-nums"
				>
					<HeartIcon size={12} weight="fill" class="text-error/80" />
					{comment.likes}
				</span>
			{/if}
		</div>
	</div>
{/snippet}

{#if threads.length > 0}
	<section class="flex flex-col gap-5">
		<h2 class="flex items-baseline gap-2 text-lg font-semibold">
			Comments
			<span class="text-base-content/45 text-sm font-normal">
				{comments.length} from the sites
			</span>
		</h2>
		<div class="flex flex-col gap-2">
			{#each shown as thread, position (thread.comment.id)}
				<div
					class="rise-in flex flex-col gap-3 rounded-2xl p-3 transition-colors hover:bg-white/[0.03]"
					style:--rise-delay="{Math.min(position * 30, 300)}ms"
				>
					{@render entry(thread.comment, false)}
					{#if thread.replies.length > 0}
						<button
							type="button"
							class="text-info ml-12 flex items-center gap-1.5 self-start rounded-full px-2 py-1 text-xs font-semibold transition-colors hover:bg-white/5"
							aria-expanded={openThreads.includes(thread.comment.id)}
							onclick={() => toggleReplies(thread.comment.id)}
						>
							<span
								class="flex transition-transform"
								class:rotate-180={openThreads.includes(thread.comment.id)}
							>
								<CaretDownIcon size={12} weight="bold" />
							</span>
							{thread.replies.length}
							{thread.replies.length === 1 ? 'reply' : 'replies'}
						</button>
						{#if openThreads.includes(thread.comment.id)}
							<div class="ml-12 flex flex-col gap-3">
								{#each thread.replies as reply (reply.id)}
									<div class="rise-in">
										{@render entry(reply, true)}
									</div>
								{/each}
							</div>
						{/if}
					{/if}
				</div>
			{/each}
		</div>
		{#if threads.length > FIRST_THREADS && !showAll}
			<button type="button" class="chip self-start" onclick={() => (showAll = true)}>
				Show all {threads.length} threads
			</button>
		{/if}
	</section>
{/if}
