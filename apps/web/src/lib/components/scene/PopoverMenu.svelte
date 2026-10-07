<script lang="ts">
	// A pill button that toggles a floating menu panel. Closes on outside click,
	// Escape, or when an item calls the `close` callback passed to the content.
	import type { Snippet } from 'svelte';

	interface Props {
		label: string;
		trigger: Snippet;
		children: Snippet<[() => void]>;
		// Which edge of the trigger the panel lines up with.
		align?: 'left' | 'right';
		disabled?: boolean;
		triggerClass?: string;
	}

	let {
		label,
		trigger,
		children,
		align = 'left',
		disabled = false,
		triggerClass = ''
	}: Props = $props();

	let open = $state(false);
	let container = $state<HTMLElement | null>(null);

	/** Closes the panel. */
	function close() {
		open = false;
	}

	/** Toggles the panel. */
	function toggle() {
		open = !open;
	}

	/** Closes the panel when a click lands outside of it. */
	function onWindowClick(event: MouseEvent) {
		if (!open || !container) return;
		if (event.target instanceof Node && container.contains(event.target)) return;
		close();
	}

	/** Closes the panel on Escape. */
	function onWindowKeydown(event: KeyboardEvent) {
		if (open && event.key === 'Escape') close();
	}
</script>

<svelte:window onclick={onWindowClick} onkeydown={onWindowKeydown} />

<div class="relative" bind:this={container}>
	<button
		type="button"
		class={triggerClass}
		aria-label={label}
		title={label}
		aria-haspopup="menu"
		aria-expanded={open}
		{disabled}
		onclick={toggle}
	>
		{@render trigger()}
	</button>
	{#if open}
		<div
			role="menu"
			class="bg-base-300 border-base-content/10 absolute top-full z-30 mt-2 flex max-h-96 min-w-56 flex-col gap-0.5 overflow-y-auto rounded-xl border p-1.5 shadow-xl"
			class:left-0={align === 'left'}
			class:right-0={align === 'right'}
		>
			{@render children(close)}
		</div>
	{/if}
</div>
