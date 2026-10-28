<script lang="ts">
	// The watch page's action pill group: rating, o-counter, the Save menu
	// (favorite, watchlist, collection) and an overflow menu with "Who is this?",
	// find on other sites, download and blocking.
	// Left/Right arrows move between the pills for D-pad navigation.
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import UserFocusIcon from 'phosphor-svelte/lib/UserFocusIcon';
	import DotsThreeVerticalIcon from 'phosphor-svelte/lib/DotsThreeVerticalIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import SealCheckIcon from 'phosphor-svelte/lib/SealCheckIcon';
	import RatingButtons from '$lib/components/RatingButtons.svelte';
	import OCounterButton from '$lib/components/OCounterButton.svelte';
	import { formatSpeed } from '$lib/tmdb';
	import {
		qualityLabel,
		type PerformerRef,
		type SceneStream,
		type StreamGroup,
		type StudioRef
	} from '$lib/sceneDetail';
	import PopoverMenu from './PopoverMenu.svelte';
	import SaveMenu from './SaveMenu.svelte';
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
		onidentify: () => void;
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
		onidentify,
		onblock,
		onedge
	}: Props = $props();

	const PILL_BASE_CLASS =
		'flex items-center gap-1.5 rounded-full bg-white/10 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-white/20 disabled:opacity-50';
	const PILL_CLASS = `${PILL_BASE_CLASS} px-4`;
	const ICON_PILL_CLASS = `${PILL_BASE_CLASS} px-2.5`;
	const MENU_ITEM_CLASS =
		'flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-white/10 disabled:opacity-50';
	const MENU_HEADING_CLASS =
		'text-base-content/50 mt-1 flex items-center gap-1.5 border-t border-white/5 px-3 pt-2.5 pb-1 text-xs font-semibold';

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

	/** Queues a download and closes the menu. */
	function download(stream: SceneStream, close: () => void) {
		close();
		ondownload(stream);
	}

	/** Closes the menu, then runs the picked action. */
	function run(action: () => void, close: () => void) {
		close();
		action();
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="flex flex-wrap items-center gap-2" bind:this={group} onkeydown={onKeydown}>
	<RatingButtons mediaId={sceneId} />
	<OCounterButton mediaId={sceneId} />
	<SaveMenu {sceneId} {sceneTitle} triggerClass={PILL_CLASS} menuItemClass={MENU_ITEM_CLASS} />

	<PopoverMenu label="More actions" align="right" triggerClass={ICON_PILL_CLASS}>
		{#snippet trigger()}
			{#if queueingId !== null}
				<span class="loading loading-xs"></span>
			{:else}
				<DotsThreeVerticalIcon size={18} weight="bold" />
			{/if}
		{/snippet}
		{#snippet children(close)}
			<button type="button" role="menuitem" class={MENU_ITEM_CLASS} onclick={() => run(onidentify, close)}>
				<UserFocusIcon size={16} />
				Who is this?
			</button>
			<button
				type="button"
				role="menuitem"
				class={MENU_ITEM_CLASS}
				onclick={() => run(onfindsources, close)}
			>
				<MagnifyingGlassIcon size={16} />
				Find on other sites
				{#if alikeCount > 0}
					<span class="text-base-content/50 ml-auto text-xs">{alikeCount} found</span>
				{/if}
			</button>

			{#if streamGroups.length > 0}
				<p class={MENU_HEADING_CLASS}>
					<DownloadSimpleIcon size={14} />
					Download to NAS
				</p>
				{#each streamGroups as streamGroup (streamGroup.provider)}
					{#each streamGroup.streams as stream (stream.id)}
						<button
							type="button"
							role="menuitem"
							class={MENU_ITEM_CLASS}
							disabled={queueingId !== null}
							onclick={() => download(stream, close)}
						>
							<SiteIcon iconUrl={pluginIcons[streamGroup.provider]} class="h-3.5 w-3.5" />
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
			{/if}

			{#if studio || performers.length > 0}
				<p class={MENU_HEADING_CLASS}>
					<ProhibitIcon size={14} />
					Don't recommend
				</p>
			{/if}
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
		{/snippet}
	</PopoverMenu>
</div>
