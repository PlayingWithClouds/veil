<!--
	Finding a performer or studio to pick: type, the matches load below as
	you go, click or Enter picks one. Escape or leaving closes it.
-->
<script lang="ts">
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import { cacheUrl } from '$lib/img';
	import type { MemberRow } from '$lib/library';

	interface Props {
		placeholder: string;
		search: (text: string) => Promise<MemberRow[]>;
		onpick: (member: MemberRow) => void;
		onclose: () => void;
	}

	let { placeholder, search, onpick, onclose }: Props = $props();

	// How long typing rests before the search runs.
	const SEARCH_DELAY_MS = 200;
	const RESULTS = 8;

	let text = $state('');
	let results = $state<MemberRow[]>([]);
	let highlighted = $state(0);
	let timer: ReturnType<typeof setTimeout> | null = null;

	$effect(() => {
		const typed = text.trim();
		if (timer) clearTimeout(timer);
		if (typed === '') {
			results = [];
			return;
		}
		timer = setTimeout(async () => {
			const found = await search(typed);
			if (typed === text.trim()) {
				results = found.slice(0, RESULTS);
				highlighted = 0;
			}
		}, SEARCH_DELAY_MS);
	});

	/** Focuses the input once it's in the page. */
	function focus(input: HTMLInputElement) {
		input.focus();
	}

	/** Arrow keys move, Enter picks, Escape closes. */
	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onclose();
			return;
		}
		if (results.length === 0) return;
		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();
			const step = event.key === 'ArrowDown' ? 1 : -1;
			highlighted = (highlighted + step + results.length) % results.length;
			return;
		}
		if (event.key === 'Enter') {
			event.preventDefault();
			onpick(results[highlighted]);
		}
	}
</script>

<div class="relative w-72">
	<input
		use:focus
		class="input w-full"
		{placeholder}
		bind:value={text}
		onkeydown={onKeydown}
		onblur={() => setTimeout(onclose, 150)}
	/>
	{#if results.length > 0}
		<ul
			class="border-base-content/10 bg-base-200 z-overlay absolute top-full right-0 left-0 mt-1 flex flex-col rounded-xl border p-1 shadow-xl"
			role="listbox"
		>
			{#each results as result, index (result.id)}
				<li role="option" aria-selected={index === highlighted}>
					<button
						type="button"
						class="flex w-full items-center gap-2.5 rounded-lg px-2 py-1.5 text-left text-sm"
						class:bg-base-300={index === highlighted}
						onmouseenter={() => (highlighted = index)}
						onmousedown={(event) => event.preventDefault()}
						onclick={() => onpick(result)}
					>
						<span class="bg-base-300 flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full">
							{#if result.image}
								<img src={cacheUrl(result.image.uri)} alt="" class="h-full w-full object-cover" />
							{:else}
								<UserIcon size={14} class="opacity-40" />
							{/if}
						</span>
						<span class="flex min-w-0 flex-col">
							<span class="truncate">{result.name}</span>
							<span class="text-base-content/50 text-xs">{result.videoCount} videos</span>
						</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
