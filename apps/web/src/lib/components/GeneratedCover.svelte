<!--
	A cover for media without artwork, so a grid of them isn't a wall of grey
	tiles: a gradient picked from the title (the same title always gets the
	same colours) with the title set on it. Audio looks like podcast art
	(headphones, title and voice at the bottom); a story like a book cover
	(title in serif in the middle, the author under a rule).
-->
<script lang="ts">
	import HeadphonesIcon from 'phosphor-svelte/lib/HeadphonesIcon';

	interface Props {
		title: string;
		// The voice or author, when known.
		byline?: string | null;
		kind: 'audio' | 'story';
	}

	let { title, byline = null, kind }: Props = $props();

	let hue = $derived(hueOf(title));
	let background = $derived(
		`linear-gradient(150deg, hsl(${hue} 55% 34%), hsl(${(hue + 55) % 360} 60% 16%))`
	);

	/** A hue from 0 to 359 that a title always gets. */
	function hueOf(text: string): number {
		let hash = 0;
		for (const character of text) {
			hash = (hash * 31 + (character.codePointAt(0) ?? 0)) % 360;
		}
		return hash;
	}
</script>

<span
	class="@container flex h-full w-full flex-col p-[9%] text-left text-white"
	style:background-image={background}
>
	{#if kind === 'audio'}
		<HeadphonesIcon size={22} class="opacity-60" />
		<span class="mt-auto flex flex-col gap-1">
			<span class="line-clamp-4 text-[clamp(0.8rem,6cqw,1.15rem)] leading-tight font-bold">
				{title}
			</span>
			{#if byline}<span class="truncate text-xs opacity-70">{byline}</span>{/if}
		</span>
	{:else}
		<span class="my-auto flex flex-col items-center gap-3 text-center">
			<span
				class="line-clamp-5 text-[clamp(0.85rem,8cqw,1.25rem)] leading-snug font-semibold"
				style:font-family="Georgia, 'Iowan Old Style', serif"
			>
				{title}
			</span>
			<span class="h-px w-8 bg-white/50"></span>
			{#if byline}<span class="truncate text-xs tracking-wide uppercase opacity-75">{byline}</span
				>{/if}
		</span>
	{/if}
</span>
