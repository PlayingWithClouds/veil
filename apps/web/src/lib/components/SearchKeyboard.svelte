<script lang="ts">
	import BackspaceIcon from 'phosphor-svelte/lib/BackspaceIcon';
	import ArrowFatUpIcon from 'phosphor-svelte/lib/ArrowFatUpIcon';
	import EraserIcon from 'phosphor-svelte/lib/EraserIcon';
	import FocusGrid from './FocusGrid.svelte';

	type KeyType = 'char' | 'space' | 'backspace' | 'clear' | 'shift';

	interface Key {
		id: string;
		type: KeyType;
		// Base character for letter keys (uppercase); ignored for action keys.
		char?: string;
	}

	interface Props {
		// Fired with the character to append (already cased for the shift state).
		oninput: (char: string) => void;
		onbackspace: () => void;
		onclear: () => void;
		// Fired when navigation leaves the keyboard at its right edge, so the page
		// can hand focus to the results grid.
		onedge?: () => void;
	}

	let { oninput, onbackspace, onclear, onedge }: Props = $props();

	const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('');

	// 26 letters followed by the action row keeps a clean 6-column layout.
	const keys: Key[] = [
		...letters.map((char) => ({ id: char, type: 'char' as KeyType, char })),
		{ id: 'shift', type: 'shift' },
		{ id: 'space', type: 'space' },
		{ id: 'backspace', type: 'backspace' },
		{ id: 'clear', type: 'clear' }
	];

	let shift = $state(false);
	let grid = $state<{ focusFirst: () => void; focus: () => void }>();

	export function focus() {
		grid?.focus();
	}
	export function focusFirst() {
		grid?.focusFirst();
	}

	function press(key: Key) {
		if (key.type === 'char' && key.char) {
			oninput(shift ? key.char : key.char.toLowerCase());
			return;
		}
		if (key.type === 'space') {
			oninput(' ');
			return;
		}
		if (key.type === 'backspace') {
			onbackspace();
			return;
		}
		if (key.type === 'clear') {
			onclear();
			return;
		}
		if (key.type === 'shift') {
			shift = !shift;
		}
	}

	function onGridEdge(direction: 'left' | 'right' | 'up' | 'down') {
		if (direction === 'right') onedge?.();
	}

	function keyLabel(key: Key): string {
		if (key.type === 'char' && key.char) {
			return shift ? key.char : key.char.toLowerCase();
		}
		return '';
	}
</script>

<FocusGrid
	bind:this={grid}
	items={keys}
	ariaLabel="On-screen keyboard"
	gridClass="grid-cols-6 gap-2"
	onselect={(key) => press(key)}
	onedge={onGridEdge}
>
	{#snippet item(key)}
		<div
			class="flex h-14 items-center justify-center rounded-lg text-lg font-semibold select-none
				{key.type === 'char'
				? 'bg-base-200'
				: 'bg-base-300 text-base-content/70'}
				{key.type === 'shift' && shift ? 'ring-primary text-primary ring-2' : ''}"
		>
			{#if key.type === 'char'}
				{keyLabel(key)}
			{:else if key.type === 'space'}
				<span class="text-base-content/50 text-xs tracking-wide uppercase">Space</span>
			{:else if key.type === 'backspace'}
				<BackspaceIcon size={22} />
			{:else if key.type === 'clear'}
				<EraserIcon size={20} />
			{:else if key.type === 'shift'}
				<ArrowFatUpIcon size={20} weight={shift ? 'fill' : 'regular'} />
			{/if}
		</div>
	{/snippet}
</FocusGrid>
