<!--
	Phone watch page actions: one sideways-scrolling row of pills (like/dislike,
	o-counter, Save, Download, ⋮). Save, Download and ⋮ open glass sheets.
-->
<script lang="ts">
	import ListPlusIcon from 'phosphor-svelte/lib/ListPlusIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import DotsThreeIcon from 'phosphor-svelte/lib/DotsThreeIcon';
	import BookmarkSimpleIcon from 'phosphor-svelte/lib/BookmarkSimpleIcon';
	import FolderSimplePlusIcon from 'phosphor-svelte/lib/FolderSimplePlusIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import StackIcon from 'phosphor-svelte/lib/StackIcon';
	import ShareNetworkIcon from 'phosphor-svelte/lib/ShareNetworkIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import SealCheckIcon from 'phosphor-svelte/lib/SealCheckIcon';
	import ArrowCounterClockwiseIcon from 'phosphor-svelte/lib/ArrowCounterClockwiseIcon';
	import RatingButtons from '$lib/components/RatingButtons.svelte';
	import OCounterButton from '$lib/components/OCounterButton.svelte';
	import ActionSheet, { type SheetAction } from '$lib/components/ActionSheet.svelte';
	import { openMoveDialog } from '$lib/stores/moveDialog';
	import { notifications } from '$lib/stores/notifications';
	import { addToWatchlist, fetchWatchlist, removeFromWatchlist } from '$lib/discovery';
	import { shareLink } from '$lib/share';
	import { formatSpeed } from '$lib/tmdb';
	import { siteDisplayName, sites } from '$lib/stores/sites';
	import { qualityLabel, type SceneDetail, type SceneStream, type StreamGroup } from '$lib/sceneDetail';
	import SceneSources from './SceneSources.svelte';
	import SiteIcon from './SiteIcon.svelte';

	type BlockKind = 'performer' | 'studio';
	type OpenSheet = 'save' | 'download' | 'more' | 'sources' | null;

	interface Props {
		scene: SceneDetail;
		posterUrl: string | null;
		streamGroups: StreamGroup[];
		pluginIcons: Record<string, string | null>;
		// Source id whose download is being queued, for its spinner.
		queueingId: string | null;
		activeSourceId: string | null;
		resolvingId: string | null;
		alikeCount: number;
		ondownload: (stream: SceneStream) => void;
		onplay: (stream: SceneStream) => void;
		onfindsources: () => void;
		onblock: (kind: BlockKind, id: string, label: string) => void;
	}

	let {
		scene,
		posterUrl,
		streamGroups,
		pluginIcons,
		queueingId,
		activeSourceId,
		resolvingId,
		alikeCount,
		ondownload,
		onplay,
		onfindsources,
		onblock
	}: Props = $props();

	const PILL_CLASS =
		'glass-pill flex h-10 shrink-0 items-center gap-1.5 rounded-full px-4 text-sm font-semibold disabled:opacity-50';

	let openSheet = $state<OpenSheet>(null);
	let oCounter = $state<{ undo: () => void; currentCount: () => number } | null>(null);
	let inWatchlist = $state(false);

	// Whether the scene is saved to watch later, for the Save sheet's label.
	$effect(() => {
		const id = scene.id;
		let cancelled = false;
		fetchWatchlist().then((entries) => {
			if (!cancelled) inWatchlist = entries.some((entry) => entry.id === id);
		});
		return () => {
			cancelled = true;
		};
	});

	/** Adds to or removes from watch later. */
	async function toggleWatchlist() {
		const next = !inWatchlist;
		inWatchlist = next;
		try {
			if (next) {
				await addToWatchlist(scene.id);
				notifications.push('Saved to Watch later', 'success');
				return;
			}
			await removeFromWatchlist(scene.id);
			notifications.push('Removed from Watch later', 'success');
		} catch {
			inWatchlist = !next;
			notifications.push('Could not update Watch later', 'error');
		}
	}

	/** The Save sheet: watch later and collections. */
	function saveActions(): SheetAction[] {
		let watchLabel = 'Save to Watch later';
		if (inWatchlist) watchLabel = 'Remove from Watch later';
		return [
			{ label: watchLabel, icon: BookmarkSimpleIcon, run: toggleWatchlist },
			{
				label: 'Save to collection…',
				icon: FolderSimplePlusIcon,
				run: () => openMoveDialog('scene', scene.id, scene.title)
			}
		];
	}

	/** The ⋮ sheet: other sources, sharing and blocking. */
	function moreActions(): SheetAction[] {
		let findLabel = 'Find other sites';
		if (alikeCount > 0) findLabel = `Find other sites (${alikeCount} found)`;
		const actions: SheetAction[] = [];
		const counter = oCounter;
		if (counter && counter.currentCount() > 0) {
			actions.push({ label: 'Undo last 💧', icon: ArrowCounterClockwiseIcon, run: () => counter.undo() });
		}
		if (streamGroups.length > 0) {
			actions.push({ label: 'Sources and qualities…', icon: StackIcon, run: () => (openSheet = 'sources') });
		}
		actions.push({ label: findLabel, icon: MagnifyingGlassIcon, run: onfindsources });
		actions.push({ label: 'Share', icon: ShareNetworkIcon, run: () => shareLink(scene.title, scene.sourceUrl) });
		const studio = scene.studio;
		if (studio) {
			actions.push({
				label: `Block ${studio.name}`,
				icon: ProhibitIcon,
				run: () => onblock('studio', studio.id, studio.name),
				destructive: true
			});
		}
		for (const performer of scene.performers) {
			actions.push({
				label: `Block ${performer.name}`,
				icon: ProhibitIcon,
				run: () => onblock('performer', performer.id, performer.name),
				destructive: true
			});
		}
		return actions;
	}

	/** Closes the Download sheet and queues the chosen source. */
	function download(stream: SceneStream) {
		openSheet = null;
		ondownload(stream);
	}

	/** Closes the sources sheet and plays the chosen source. */
	function play(stream: SceneStream) {
		openSheet = null;
		onplay(stream);
	}
</script>

<div class="no-scrollbar fade-right-edge -mx-3 flex items-center gap-2 overflow-x-auto px-3 pr-8">
	<RatingButtons mediaId={scene.id} joined class="shrink-0" />
	<OCounterButton bind:this={oCounter} mediaId={scene.id} glass class="shrink-0" />
	<button type="button" class={PILL_CLASS} onclick={() => (openSheet = 'save')}>
		<ListPlusIcon size={18} />
		Save
	</button>
	{#if streamGroups.length > 0}
		<button
			type="button"
			class={PILL_CLASS}
			disabled={queueingId !== null}
			onclick={() => (openSheet = 'download')}
		>
			{#if queueingId !== null}
				<span class="loading loading-xs"></span>
			{:else}
				<DownloadSimpleIcon size={18} />
			{/if}
			Download
		</button>
	{/if}
	<button
		type="button"
		class="glass-pill flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
		onclick={() => (openSheet = 'more')}
		aria-label="More actions"
	>
		<DotsThreeIcon size={20} weight="bold" />
	</button>
</div>

{#if openSheet === 'save'}
	<ActionSheet title={scene.title} imageUrl={posterUrl} actions={saveActions()} onclose={() => (openSheet = null)} />
{:else if openSheet === 'more'}
	<ActionSheet title={scene.title} imageUrl={posterUrl} actions={moreActions()} onclose={() => (openSheet = null)} />
{:else if openSheet === 'download'}
	<ActionSheet title="Download" subtitle="Saved to the server" onclose={() => (openSheet = null)}>
		{#each streamGroups as streamGroup (streamGroup.provider)}
			<p class="text-base-content/50 flex items-center gap-1.5 px-3 pt-2 pb-1 text-xs font-semibold">
				<SiteIcon iconUrl={pluginIcons[streamGroup.provider]} class="h-3.5 w-3.5" />
				{siteDisplayName(streamGroup.provider, $sites)}
			</p>
			{#each streamGroup.streams as stream (stream.id)}
				<button
					type="button"
					class="flex h-12 shrink-0 items-center gap-3 rounded-2xl px-3 text-left text-[15px]"
					onclick={() => download(stream)}
				>
					<DownloadSimpleIcon size={20} />
					<span class="font-medium">{qualityLabel(stream)}</span>
					{#if stream.verified}
						<SealCheckIcon size={14} weight="fill" class="text-success" />
					{/if}
					{#if formatSpeed(stream.expectedSpeedBps)}
						<span class="text-success ml-auto text-xs">~{formatSpeed(stream.expectedSpeedBps)}</span>
					{/if}
				</button>
			{/each}
		{/each}
	</ActionSheet>
{:else if openSheet === 'sources'}
	<ActionSheet title="Sources" subtitle="Pick what plays" onclose={() => (openSheet = null)}>
		<div class="px-2 pt-1 pb-2">
			<SceneSources groups={streamGroups} {pluginIcons} {activeSourceId} {resolvingId} onplay={play} />
		</div>
	</ActionSheet>
{/if}
