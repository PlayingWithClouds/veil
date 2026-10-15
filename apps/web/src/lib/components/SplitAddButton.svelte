<!--
	Small round button that adds a cam or scene to the split view instead of
	opening it. Meant to sit inside a hover-revealing group; the parent
	positions it with class. Safe inside links (it stops the click).
-->
<script lang="ts">
	import SquareSplitHorizontalIcon from 'phosphor-svelte/lib/SquareSplitHorizontalIcon';
	import { openInSplit } from '$lib/splitActions';
	import type { SplitPaneKind } from '$lib/stores/split';

	interface Props {
		kind: SplitPaneKind;
		mediaId: string;
		title: string;
		class?: string;
	}

	let { kind, mediaId, title, class: className = '' }: Props = $props();

	/** Adds the stream to the split view without triggering the surrounding link or card. */
	function add(event: MouseEvent) {
		event.preventDefault();
		event.stopPropagation();
		openInSplit(kind, mediaId, title);
	}
</script>

<button
	type="button"
	onclick={add}
	tabindex="-1"
	aria-label="Add {title} to split view"
	title="Add to split view"
	data-split-add
	class="z-20 flex h-8 w-8 items-center justify-center rounded-md bg-black/70 text-white opacity-0 backdrop-blur-sm transition-opacity group-focus-within:opacity-100 group-hover:opacity-100 hover:bg-black/90 {className}"
>
	<SquareSplitHorizontalIcon size={16} weight="bold" />
</button>
