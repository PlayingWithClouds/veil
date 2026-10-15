<!--
	Split view: up to four cams and videos playing side by side. Tabs above the
	panes name them (drag to reorder, × to close); clicking a pane focuses it
	and gives it the sound. The panes live in lib/stores/split.ts, so the view
	survives reload. The layout keeps this component mounted on every page: on
	/split it fills the page, elsewhere it shrinks to a muted dock that keeps
	playing while the user browses for more to add, and opens /split on click.
-->
<script lang="ts">
	import ArrowsInIcon from 'phosphor-svelte/lib/ArrowsInIcon';
	import ArrowsOutIcon from 'phosphor-svelte/lib/ArrowsOutIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { goto } from '$app/navigation';
	import SplitPaneView from './SplitPane.svelte';
	import { splitUrl } from '$lib/routes';
	import {
		clearSplitPanes,
		closeSplitPane,
		focusSplitPane,
		moveSplitPane,
		splitPaneKey,
		splitState
	} from '$lib/stores/split';

	interface Props {
		/** Shown as the small dock while another page is open. */
		docked: boolean;
	}

	let { docked }: Props = $props();

	let gridElement = $state<HTMLElement>();
	let isFullscreen = $state(false);
	// Index of the tab being dragged, and the one it hovers over.
	let draggedIndex = $state<number | null>(null);
	let dropIndex = $state<number | null>(null);

	let paneCount = $derived($splitState.panes.length);

	// Panes keep one DOM position for as long as they are open and are placed by
	// CSS order: moving a player's element within the page would stall its stream.
	let renderKeys = $state<string[]>([]);

	// Follow the store: new panes join at the end, closed ones leave.
	$effect(() => {
		const currentKeys = $splitState.panes.map(splitPaneKey);
		const kept = renderKeys.filter((key) => currentKeys.includes(key));
		const added = currentKeys.filter((key) => !kept.includes(key));
		if (added.length === 0 && kept.length === renderKeys.length) return;
		renderKeys = [...kept, ...added];
	});

	/** Where a pane sits in the tab order, or -1 once it is closed. */
	function paneIndex(key: string): number {
		return $splitState.panes.findIndex((pane) => splitPaneKey(pane) === key);
	}

	/** Puts the grid of panes, without the app around it, in fullscreen, or leaves it. */
	async function toggleFullscreen() {
		if (document.fullscreenElement) {
			await document.exitFullscreen();
			return;
		}
		await gridElement?.requestFullscreen();
	}

	/** Tracks fullscreen so the button shows the right icon. */
	function onFullscreenChange() {
		isFullscreen = document.fullscreenElement === gridElement;
	}

	/** Starts dragging a tab. */
	function onTabDragStart(event: DragEvent, index: number) {
		draggedIndex = index;
		event.dataTransfer?.setData('text/plain', String(index));
		if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
	}

	/** Marks the tab under the pointer as the drop target. */
	function onTabDragOver(event: DragEvent, index: number) {
		if (draggedIndex === null) return;
		event.preventDefault();
		dropIndex = index;
	}

	/** Moves the dragged tab (and its pane) to the dropped position. */
	function onTabDrop(event: DragEvent, index: number) {
		event.preventDefault();
		if (draggedIndex !== null) moveSplitPane(draggedIndex, index);
		onTabDragEnd();
	}

	/** Clears the drag state. */
	function onTabDragEnd() {
		draggedIndex = null;
		dropIndex = null;
	}
</script>

<svelte:document onfullscreenchange={onFullscreenChange} />

<div class="flex flex-col gap-2" class:pt-2={!docked} class:split-dock={docked}>
	{#if docked}
		<div class="flex items-center gap-2 px-2 pt-1.5 text-xs text-white/80">
			<span class="flex-1 font-semibold">Split view · {paneCount}</span>
			<button
				type="button"
				class="hover:text-white"
				onclick={() => goto(splitUrl())}
				aria-label="Open split view"
				title="Open split view"
			>
				<ArrowsOutIcon size={14} />
			</button>
			<button
				type="button"
				class="hover:text-white"
				onclick={clearSplitPanes}
				aria-label="Close split view"
				title="Close split view"
			>
				<XIcon size={14} weight="bold" />
			</button>
		</div>
	{/if}
	<div class="flex items-center gap-2" class:hidden={docked}>
		<div class="no-scrollbar flex min-w-0 flex-1 gap-1 overflow-x-auto" role="tablist">
			{#each $splitState.panes as pane, index (splitPaneKey(pane))}
				<div
					role="tab"
					tabindex="0"
					aria-selected={index === $splitState.focusedIndex}
					draggable="true"
					data-split-tab
					class="bg-base-200 text-base-content/70 hover:text-base-content flex max-w-56 min-w-0 cursor-grab items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm"
					class:tab-focused={index === $splitState.focusedIndex}
					class:opacity-40={draggedIndex === index}
					class:tab-drop={dropIndex === index && draggedIndex !== index}
					ondragstart={(event) => onTabDragStart(event, index)}
					ondragover={(event) => onTabDragOver(event, index)}
					ondrop={(event) => onTabDrop(event, index)}
					ondragend={onTabDragEnd}
					onclick={() => focusSplitPane(splitPaneKey(pane))}
					onkeydown={(event) => event.key === 'Enter' && focusSplitPane(splitPaneKey(pane))}
				>
					{#if index === $splitState.focusedIndex}
						<SpeakerHighIcon size={14} weight="fill" class="shrink-0" />
					{/if}
					<span class="truncate">{pane.title}</span>
					<button
						type="button"
						class="hover:bg-base-300 -mr-1 flex h-5 w-5 shrink-0 items-center justify-center rounded-full"
						aria-label="Close {pane.title}"
						onclick={(event) => {
							event.stopPropagation();
							closeSplitPane(splitPaneKey(pane));
						}}
					>
						<XIcon size={12} weight="bold" />
					</button>
				</div>
			{/each}
		</div>
		{#if paneCount > 0}
			<button type="button" class="btn btn-sm" onclick={clearSplitPanes}>Close all</button>
			<button
				type="button"
				class="btn btn-sm"
				onclick={toggleFullscreen}
				aria-label="Fullscreen"
				title="Fullscreen"
			>
				<ArrowsOutIcon size={16} />
				Fullscreen
			</button>
		{/if}
	</div>

	{#if paneCount === 0}
		<p class="text-base-content/40 py-24 text-center">
			Nothing in the split view. Use the split button on a cam, a card or a followed cam in the
			sidebar to add it.
		</p>
	{:else}
		<div
			bind:this={gridElement}
			data-split-grid
			class="split-grid relative grid gap-1 bg-black"
			class:split-grid-fullscreen={isFullscreen}
			class:split-grid-docked={docked}
			style:grid-template-columns={paneCount === 1 ? '1fr' : 'repeat(2, minmax(0, 1fr))'}
			style:grid-template-rows={paneCount > 2 ? 'repeat(2, minmax(0, 1fr))' : '1fr'}
		>
			{#each renderKeys as key (key)}
				{@const index = paneIndex(key)}
				{#if index >= 0}
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						data-split-pane={key}
						class="focus-ring relative min-h-0 min-w-0 overflow-hidden"
						class:focus-ring-on={index === $splitState.focusedIndex && paneCount > 1}
						style:order={index}
						onpointerdowncapture={() => focusSplitPane(key)}
					>
						<SplitPaneView
							pane={$splitState.panes[index]}
							focused={!docked && index === $splitState.focusedIndex}
						/>
					</div>
				{/if}
			{/each}
			{#if docked}
				<!-- The dock is a preview: a click anywhere on it opens the full view. -->
				<button
					type="button"
					class="absolute inset-0 z-30 cursor-pointer"
					onclick={() => goto(splitUrl())}
					aria-label="Open split view"
				></button>
			{/if}
			{#if isFullscreen}
				<button
					type="button"
					class="absolute top-2 right-2 z-30 flex h-8 w-8 items-center justify-center rounded-full bg-black/60 text-white opacity-30 transition-opacity hover:opacity-100"
					onclick={toggleFullscreen}
					aria-label="Exit fullscreen"
				>
					<ArrowsInIcon size={16} />
				</button>
			{/if}
		</div>
	{/if}
</div>

<style>
	.split-grid {
		height: calc(100dvh - 8.5rem);
		min-height: 20rem;
	}

	/* Bottom left, clear of the sidebar and of the scene mini player (bottom right). */
	.split-dock {
		position: fixed;
		z-index: 60;
		bottom: 1rem;
		left: 16rem;
		width: 26rem;
		gap: 0.25rem;
		overflow: hidden;
		border-radius: 0.75rem;
		background-color: black;
		box-shadow: 0 10px 30px rgb(0 0 0 / 0.5);
	}

	@media (width < 720px) {
		.split-dock {
			left: 0.75rem;
			bottom: calc(var(--bottom-nav-clearance) + 0.5rem);
			width: 16rem;
		}
	}

	.split-grid-docked {
		height: auto;
		min-height: 0;
		aspect-ratio: 16 / 9;
		gap: 1px;
	}

	.split-grid-docked :global(*) {
		pointer-events: none;
	}

	.split-grid-docked > button {
		pointer-events: auto;
	}

	.split-grid-fullscreen {
		height: 100dvh;
		gap: 2px;
	}

	.focus-ring::after {
		content: '';
		position: absolute;
		inset: 0;
		pointer-events: none;
		z-index: 25;
		border: 2px solid transparent;
		transition: border-color 0.15s;
	}

	.focus-ring-on::after {
		border-color: var(--color-primary);
	}

	.tab-focused {
		background-color: var(--color-base-300);
		color: var(--color-base-content);
		font-weight: 600;
	}

	.tab-drop {
		outline: 2px solid var(--color-primary);
	}
</style>
