<!--
	A story's chapters as links. A linear story is a numbered list, the
	chapters before the current one ticked as read. A branching story is its
	map: each choice under the chapter that offers it, joined by guide lines,
	with the path to the current chapter drawn in the accent colour.
-->
<script lang="ts">
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import { isBranching, pathTo } from '$lib/storyReading';
	import { storyReaderUrl } from '$lib/routes';
	import type { ChapterRow } from '$lib/media';

	interface Props {
		storyId: string;
		chapters: ChapterRow[];
		// The chapter being read (or where reading stopped), marked in the list.
		currentId?: string | null;
		// Called when a chapter is picked, e.g. to close a drawer.
		onpick?: () => void;
	}

	let { storyId, chapters, currentId = null, onpick }: Props = $props();

	// Deep branches stop indenting here, so long paths stay readable.
	const MAX_INDENT_DEPTH = 8;

	let branching = $derived(isBranching(chapters));
	let current = $derived(chapters.find((chapter) => chapter.id === currentId) ?? null);
	let currentIndex = $derived(current ? chapters.indexOf(current) : -1);
	// The chapters from the start to the current one: the path read.
	let pathIds = $derived.by(() => {
		if (!current) return new Set<string>();
		const path = pathTo(chapters, current);
		const ids = new Set(path.map((chapter) => chapter.id));
		ids.add(current.id);
		// The path starts at the first choice; the chapter offering it is the start.
		const start = path[0]?.parent;
		if (start) ids.add(start);
		return ids;
	});
	let childrenByParent = $derived.by(() => {
		const known = new Set(chapters.map((chapter) => chapter.id));
		const children = new Map<string | null, ChapterRow[]>();
		for (const chapter of chapters) {
			const parent = chapter.parent && known.has(chapter.parent) ? chapter.parent : null;
			children.set(parent, [...(children.get(parent) ?? []), chapter]);
		}
		return children;
	});

	/** A chapter's line in the map: the choice that leads to it, else its title. */
	function labelOf(chapter: ChapterRow, index: number): string {
		return chapter.choice ?? chapter.title ?? `Chapter ${index + 1}`;
	}
</script>

{#snippet branch(parentId: string | null, depth: number)}
	<ul
		class="flex flex-col"
		class:ml-2.5={depth > 0 && depth <= MAX_INDENT_DEPTH}
		class:pl-3={depth > 0 && depth <= MAX_INDENT_DEPTH}
		class:border-l={depth > 0 && depth <= MAX_INDENT_DEPTH}
		style:border-color="color-mix(in oklab, currentColor 14%, transparent)"
	>
		{#each childrenByParent.get(parentId) ?? [] as chapter (chapter.id)}
			{@const onPath = pathIds.has(chapter.id)}
			{@const isCurrent = chapter.id === currentId}
			<li class="flex flex-col">
				<a
					href={storyReaderUrl(storyId, chapter.position)}
					onclick={() => onpick?.()}
					class="flex items-start gap-2 rounded-lg px-2 py-1.5 hover:bg-current/5"
					class:opacity-60={pathIds.size > 0 && !onPath}
					aria-current={isCurrent ? 'page' : undefined}
				>
					<span
						class="mt-[0.45em] h-2 w-2 shrink-0 rounded-full"
						style:background-color={onPath
							? 'var(--reader-accent, var(--color-accent))'
							: 'color-mix(in oklab, currentColor 30%, transparent)'}
					></span>
					<span class="min-w-0 flex-1 leading-snug" class:font-semibold={isCurrent}>
						{labelOf(chapter, chapters.indexOf(chapter))}
						{#if chapter.choice && chapter.title && chapter.title !== chapter.choice}
							<span class="block text-xs opacity-60">{chapter.title}</span>
						{/if}
					</span>
					{#if isCurrent}
						<span class="shrink-0 text-xs opacity-60">You're here</span>
					{/if}
				</a>
				{#if childrenByParent.has(chapter.id)}
					{@render branch(chapter.id, depth + 1)}
				{/if}
			</li>
		{/each}
	</ul>
{/snippet}

{#if branching}
	{@render branch(null, 0)}
{:else}
	<ol class="flex flex-col">
		{#each chapters as chapter, index (chapter.id)}
			{@const isCurrent = chapter.id === currentId}
			{@const read = currentIndex !== -1 && index < currentIndex}
			<li>
				<a
					href={storyReaderUrl(storyId, chapter.position)}
					onclick={() => onpick?.()}
					class="flex items-baseline gap-3 rounded-lg px-3 py-2 hover:bg-current/5"
					class:opacity-60={read}
					aria-current={isCurrent ? 'page' : undefined}
				>
					<span class="w-6 shrink-0 text-right text-xs tabular-nums opacity-60">{index + 1}</span>
					<span class="min-w-0 flex-1" class:font-semibold={isCurrent}>
						{chapter.title ?? `Chapter ${index + 1}`}
					</span>
					{#if isCurrent}
						<span class="shrink-0 text-xs opacity-60">You're here</span>
					{:else if read}
						<CheckIcon size={14} class="shrink-0 self-center" />
					{/if}
				</a>
			</li>
		{/each}
	</ol>
{/if}
