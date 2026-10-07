<script lang="ts">
	// The watch page's action pill group: rating, o-counter, save, watchlist,
	// download (quality menu), find sources, and an overflow menu for blocking.
	// Left/Right arrows move between the pills for D-pad navigation.
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import ListPlusIcon from 'phosphor-svelte/lib/ListPlusIcon';
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import SealCheckIcon from 'phosphor-svelte/lib/SealCheckIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import RatingButtons from '$lib/components/RatingButtons.svelte';
	import OCounterButton from '$lib/components/OCounterButton.svelte';
	import WatchlistButton from '$lib/components/WatchlistButton.svelte';
	import { openMoveDialog } from '$lib/stores/moveDialog';
	import { formatSpeed } from '$lib/tmdb';
	import {
		qualityLabel,
		type PerformerRef,
		type SceneStream,
		type StreamGroup,
		type StudioRef
	} from '$lib/sceneDetail';
	import PopoverMenu from './PopoverMenu.svelte';
	import SiteIcon from './SiteIcon.svelte';

	type BlockKind = 'performer' | 'studio';
	type Direction = 'left' | 'right';

	interface Props {
		sceneId: string;
		sceneTitle: string;
		streamGroups: StreamGroup[];
		pluginIcons: Record<string, string | null>;
		// Source id whose download is being queued, for its spinner.
		queueingId: string | null;
		alikeCount: number;
		studio: StudioRef | null;
		performers: PerformerRef[];
		ondownload: (stream: SceneStream) => void;
		onfindsources: () => void;
		onblock: (kind: BlockKind, id: string, label: string) => void;
		// Arrow key past the first/last pill, for handing focus to a neighbour zone.
		onedge?: (direction: Direction) => void;
	}

	let {
		sceneId,
		sceneTitle,
		streamGroups,
		pluginIcons,
		queueingId,
		alikeCount,
		studio,
		performers,
		ondownload,
		onfindsources,
		onblock,
		onedge
	}: Props = $props();

	const PILL_BASE_CLASS =
		'flex items-center gap-1.5 rounded-full bg-white/10 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-white/20 disabled:opacity-50';
	const PILL_CLASS = `${PILL_BASE_CLASS} px-4`;
	const ICON_PILL_CLASS = `${PILL_BASE_CLASS} px-2.5`;
	const MENU_ITEM_CLASS =
		'flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-white/10';

	let group = $state<HTMLElement | null>(null);

	/** Enabled buttons in the pill row, in DOM order. */
	function pillButtons(): HTMLButtonElement[] {
		if (!group) return [];
		return [...group.querySelectorAll<HTMLButtonElement>('button:not(:disabled)')];
	}

	/** Focuses the first pill — the entry point when D-pad focus arrives here. */
	export function focusFirst() {
		pillButtons()[0]?.focus();
	}

	/** Moves DOM focus one pill left/right; reports the edge when there is none. */
	function moveFocus(delta: number, edge: Direction) {
		const buttons = pillButtons();
		const index = buttons.findIndex((button) => button === document.activeElement);
		const next = buttons[index + delta];
		if (index === -1 || !next) {
			onedge?.(edge);
			return;
		}
		next.focus();
	}

	/** Arrow-key navigation across the pills. */
	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'ArrowRight') {
			event.preventDefault();
			moveFocus(1, 'right');
		} else if (event.key === 'ArrowLeft') {
			event.preventDefault();
			moveFocus(-1, 'left');
		}
	}

	/** Blocks an entity and closes the overflow menu. */
	function block(kind: BlockKind, id: string, label: string, close: () => void) {
		close();
		onblock(kind, id, label);
	}

	/** Queues a download and closes the quality menu. */
	function download(stream: SceneStream, close: () => void) {
		close();
		ondownload(stream);
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="flex flex-wrap items-center gap-2" bind:this={group} onkeydown={onKeydown}>
	<RatingButtons mediaId={sceneId} />
	<OCounterButton mediaId={sceneId} />
	<button
		type="button"
		class={PILL_CLASS}
		title="Add to collection"
		onclick={() => openMoveDialog('scene', sceneId, sceneTitle)}
	>
		<ListPlusIcon size={18} />
		Save
	</button>
	<WatchlistButton mediaId={sceneId} />

	{#if streamGroups.length > 0}
		<PopoverMenu label="Download to NAS" triggerClass={PILL_CLASS} disabled={queueingId !== null}>
			{#snippet trigger()}
				{#if queueingId !== null}
					<span class="loading loading-xs"></span>
				{:else}
					<DownloadSimpleIcon size={18} />
				{/if}
				Download
				<CaretDownIcon size={12} weight="bold" />
			{/snippet}
			{#snippet children(close)}
				{#each streamGroups as streamGroup (streamGroup.provider)}
					<p class="text-base-content/50 flex items-center gap-1.5 px-3 pt-2 pb-1 text-xs font-semibold">
						<SiteIcon iconUrl={pluginIcons[streamGroup.provider]} class="h-3.5 w-3.5" />
						{streamGroup.provider}
					</p>
					{#each streamGroup.streams as stream (stream.id)}
						<button
							type="button"
							role="menuitem"
							class={MENU_ITEM_CLASS}
							onclick={() => download(stream, close)}
						>
							<span class="font-medium">{qualityLabel(stream)}</span>
							{#if stream.verified}
								<SealCheckIcon size={12} weight="fill" class="text-success" />
							{/if}
							{#if formatSpeed(stream.expectedSpeedBps)}
								<span class="text-success ml-auto text-xs">~{formatSpeed(stream.expectedSpeedBps)}</span>
							{/if}
						</button>
					{/each}
				{/each}
			{/snippet}
		</PopoverMenu>
	{/if}

	<button type="button" class={PILL_CLASS} title="Find alternate sources" onclick={onfindsources}>
		<MagnifyingGlassIcon size={18} />
		Find sources
		{#if alikeCount > 0}
			<span
				class="bg-primary text-primary-content flex h-5 min-w-5 items-center justify-center rounded-full px-1.5 text-xs font-bold"
			>
				{alikeCount}
			</span>
		{/if}
	</button>

	<PopoverMenu label="More actions" align="right" triggerClass={ICON_PILL_CLASS}>
		{#snippet trigger()}
			<DotsThreeVerticalIcon size={18} weight="bold" />
		{/snippet}
		{#snippet children(close)}
			{#if studio}
				<button
					type="button"
					role="menuitem"
					class="{MENU_ITEM_CLASS} hover:text-error"
					onclick={() => block('studio', studio.id, studio.name, close)}
				>
					<ProhibitIcon size={14} weight="bold" />
					Block studio {studio.name}
				</button>
			{/if}
			{#each performers as performer (performer.id)}
				<button
					type="button"
					role="menuitem"
					class="{MENU_ITEM_CLASS} hover:text-error"
					onclick={() => block('performer', performer.id, performer.name, close)}
				>
					<ProhibitIcon size={14} weight="bold" />
					Block {performer.name}
				</button>
			{/each}
			{#if !studio && performers.length === 0}
				<p class="text-base-content/40 px-3 py-2 text-sm">Nothing to block</p>
			{/if}
		{/snippet}
	</PopoverMenu>
</div>
