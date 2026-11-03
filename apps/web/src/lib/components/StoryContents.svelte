<!--
	A story's chapters as a list of links: numbered in a linear story, as a
	tree of choices in a branching one (each branch under the choice that
	leads to it). The chapter being read is marked.
-->
<script lang="ts">
	import { contentsOf } from '$lib/storyReading';
	import { storyReaderUrl } from '$lib/routes';
	import type { ChapterRow } from '$lib/media';

	interface Props {
		storyId: string;
		chapters: ChapterRow[];
		// The chapter being read, marked in the list.
		currentId?: string | null;
		// Called when a chapter is picked, e.g. to close a drawer.
		onpick?: () => void;
	}

	let { storyId, chapters, currentId = null, onpick }: Props = $props();

	// Deep branches stop indenting here, so long paths stay readable.
	const MAX_INDENT_DEPTH = 10;
	const INDENT_PX = 14;

	let entries = $derived(contentsOf(chapters));
</script>

<ol class="flex flex-col">
	{#each entries as entry, index (entry.chapter.id)}
		{@const chapter = entry.chapter}
		{@const current = chapter.id === currentId}
		<li style:padding-left="{Math.min(entry.depth, MAX_INDENT_DEPTH) * INDENT_PX}px">
			<a
				href={storyReaderUrl(storyId, chapter.position)}
				onclick={() => onpick?.()}
				class="flex items-baseline gap-3 rounded-lg px-3 py-2 hover:bg-current/5"
				class:font-semibold={current}
				style:color={current ? 'var(--reader-accent, currentColor)' : undefined}
				aria-current={current ? 'page' : undefined}
			>
				{#if entry.depth === 0 && chapter.parent === null}
					<span class="w-6 shrink-0 text-right text-xs tabular-nums opacity-50">{index + 1}</span>
				{:else}
					<span class="w-6 shrink-0 text-right text-xs opacity-50">↳</span>
				{/if}
				<span class="min-w-0 flex-1">
					{chapter.choice ?? chapter.title ?? `Chapter ${index + 1}`}
					{#if chapter.choice && chapter.title && chapter.title !== chapter.choice}
						<span class="block text-xs opacity-60">{chapter.title}</span>
					{/if}
				</span>
				{#if chapter.wordCount > 0}
					<span class="shrink-0 text-xs tabular-nums opacity-50">
						{chapter.wordCount.toLocaleString()} words
					</span>
				{/if}
			</a>
		</li>
	{/each}
</ol>
