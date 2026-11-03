<!--
	The way through a branching story so far, as a timeline: the start, each
	choice made, and the chapter reading stopped in, with the choices it
	offers next. Each step opens the reader there; a step with other choices
	lists them on request. A long path shows its start and its end, the
	steps between on request.
-->
<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import { choicesOf, followingChapters, pathTo } from '$lib/storyReading';
	import { storyReaderUrl } from '$lib/routes';
	import type { ChapterRow } from '$lib/media';

	interface Props {
		storyId: string;
		chapters: ChapterRow[];
		// Where reading stopped; null before reading starts.
		current: ChapterRow | null;
	}

	let { storyId, chapters, current }: Props = $props();

	// A longer path shows its first FIRST_STEPS and last LAST_STEPS steps.
	const FIRST_STEPS = 2;
	const LAST_STEPS = 5;

	let showAll = $state(false);
	// Steps whose other choices are listed.
	let opened = new SvelteSet<string>();

	let steps = $derived.by(() => {
		if (!current) return chapters.filter((chapter) => !chapter.parent).slice(0, 1);
		const path = pathTo(chapters, current);
		const start = chapters.find((chapter) => chapter.id === path[0]?.parent);
		if (path.length === 0) return [current];
		if (!start) return path;
		return [start, ...path];
	});
	let hiddenSteps = $derived(showAll ? 0 : Math.max(steps.length - FIRST_STEPS - LAST_STEPS, 0));
	let shownSteps = $derived.by(() => {
		if (hiddenSteps === 0) return steps;
		return [...steps.slice(0, FIRST_STEPS), ...steps.slice(steps.length - LAST_STEPS)];
	});
	let last = $derived(steps[steps.length - 1] ?? null);
	let nextChoices = $derived(last ? followingChapters(chapters, last) : []);

	/** The other choices that were open at a step. */
	function alternativesOf(step: ChapterRow): ChapterRow[] {
		if (!step.parent) return [];
		return choicesOf(chapters, step.parent).filter((choice) => choice.id !== step.id);
	}

	/** Lists or hides a step's other choices. */
	function toggleAlternatives(stepId: string) {
		if (opened.has(stepId)) {
			opened.delete(stepId);
			return;
		}
		opened.add(stepId);
	}
</script>

<ol class="flex flex-col">
	{#each shownSteps as step, index (step.id)}
		{@const alternatives = alternativesOf(step)}
		{@const isLast = step.id === last?.id}
		{#if hiddenSteps > 0 && index === FIRST_STEPS}
			<li class="flex gap-4">
				<span class="flex w-3 justify-center"><span class="bg-base-content/15 w-px"></span></span>
				<button
					type="button"
					class="text-base-content/50 hover:text-base-content py-2 text-sm"
					onclick={() => (showAll = true)}
				>
					{hiddenSteps} more {hiddenSteps === 1 ? 'choice' : 'choices'} along the way
				</button>
			</li>
		{/if}
		<li class="flex gap-4">
			<span class="flex w-3 flex-col items-center">
				<span
					class="mt-1.5 h-3 w-3 shrink-0 rounded-full border-2"
					class:bg-accent={isLast}
					class:border-accent={isLast}
					class:border-base-content={!isLast}
					class:opacity-40={!isLast}
				></span>
				{#if !isLast || nextChoices.length > 0}
					<span class="bg-base-content/15 w-px flex-1"></span>
				{/if}
			</span>
			<div class="flex min-w-0 flex-1 flex-col gap-1 pb-5">
				{#if !step.parent}
					<span class="text-base-content/50 text-xs tracking-wide uppercase">Start</span>
				{/if}
				<a
					href={storyReaderUrl(storyId, step.position)}
					class="hover:text-accent leading-snug font-medium"
					class:text-lg={isLast}
				>
					{step.choice ?? step.title ?? 'The beginning'}
				</a>
				{#if step.choice && step.title && step.title !== step.choice}
					<span class="text-base-content/60 text-sm">{step.title}</span>
				{/if}
				{#if isLast && current}
					<span class="text-accent text-xs font-medium">You stopped here</span>
				{/if}
				{#if alternatives.length > 0}
					<button
						type="button"
						class="text-base-content/50 hover:text-base-content self-start text-xs"
						onclick={() => toggleAlternatives(step.id)}
					>
						{alternatives.length}
						other {alternatives.length === 1 ? 'choice' : 'choices'} here
					</button>
					{#if opened.has(step.id)}
						<div class="flex flex-wrap gap-2 pt-1">
							{#each alternatives as alternative (alternative.id)}
								<a
									href={storyReaderUrl(storyId, alternative.position)}
									class="bg-base-200 hover:bg-base-300 rounded-full px-3 py-1 text-sm"
									>{alternative.choice ?? alternative.title}</a
								>
							{/each}
						</div>
					{/if}
				{/if}
			</div>
		</li>
	{/each}
	{#if nextChoices.length > 0}
		<li class="flex gap-4">
			<span class="w-3"></span>
			<div class="flex flex-col gap-2">
				<span class="text-base-content/50 text-xs tracking-wide uppercase">What happens next?</span>
				<div class="flex flex-wrap gap-2">
					{#each nextChoices as choice (choice.id)}
						<a
							href={storyReaderUrl(storyId, choice.position)}
							class="bg-base-200 hover:bg-base-300 rounded-xl px-3 py-2 text-sm"
							>{choice.choice ?? choice.title}</a
						>
					{/each}
				</div>
			</div>
		</li>
	{/if}
</ol>
