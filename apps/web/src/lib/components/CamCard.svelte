<!--
	A cam room in a grid or row: its picture, which turns into the muted live
	stream while hovered, the viewer count, marks for a room followed,
	auto-recorded or recording now, and the caption with the room's name, its
	performer, how long it's been on (or when it was last seen) and its topic.
	The Live badge shows only where offline rooms are listed too; the site
	only when more than one cam site is installed. The ⋮ menu follows,
	auto-records, records and blocks the room; a blocked room's card hides.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import BellSlashIcon from 'phosphor-svelte/lib/BellSlashIcon';
	import BellRingingIcon from 'phosphor-svelte/lib/BellRingingIcon';
	import EyeIcon from 'phosphor-svelte/lib/EyeIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import RecordIcon from 'phosphor-svelte/lib/RecordIcon';
	import StopIcon from 'phosphor-svelte/lib/StopIcon';
	import VideoCameraIcon from 'phosphor-svelte/lib/VideoCameraIcon';
	import type { SheetAction } from '$lib/components/ActionSheet.svelte';
	import CardBadge from '$lib/components/CardBadge.svelte';
	import CardCaption from '$lib/components/CardCaption.svelte';
	import CardMenu from '$lib/components/CardMenu.svelte';
	import CoverFrame from '$lib/components/CoverFrame.svelte';
	import LiveRoomPreview from '$lib/components/LiveRoomPreview.svelte';
	import {
		camSites,
		camSubject,
		followedCams,
		loadFollowedCams,
		previewStream,
		roomTopic,
		setCamAutoRecord,
		startCamRecording,
		stopCamRecording,
		toggleCamFollow,
		type CamRow
	} from '$lib/cams';
	import { logChosen, type Shown } from '$lib/events';
	import { impression } from '$lib/impression';
	import { cacheUrl } from '$lib/img';
	import { block } from '$lib/recordActions';
	import { sites } from '$lib/stores/sites';
	import { showUndo } from '$lib/stores/undoToast';
	import { hiddenRecords, hiddenKey } from '$lib/userState';

	interface Props {
		cam: CamRow;
		onclick: () => void;
		// Draws the Live badge, for rows that mix live and offline rooms.
		markLive?: boolean;
		// Whether a recording of the room runs now.
		recording?: boolean;
		// Called after a menu action started or stopped a recording.
		onrecordingchange?: () => void;
		// Why the room is recommended; the row's own reason when not given.
		reason?: string | null;
		// Where the card is shown; logs it as shown there and, on click, chosen.
		shown?: Shown | null;
	}

	let {
		cam,
		onclick,
		markLive = false,
		recording = false,
		onrecordingchange,
		reason = null,
		shown = null
	}: Props = $props();

	// Hovering this long starts the live preview, so passing over cards doesn't.
	const PREVIEW_DELAY_MS = 400;

	const compactNumber = new Intl.NumberFormat(undefined, { notation: 'compact' });

	// The room as the menu last changed it (auto-record); the row is the parent's.
	let changed = $state<CamRow | null>(null);
	let room = $derived(changed?.id === cam.id ? changed : cam);
	let menuAnchor = $state<HTMLElement | null>(null);
	let previewing = $state(false);
	let previewTimer: ReturnType<typeof setTimeout> | null = null;

	let followed = $derived($followedCams.has(cam.id));
	let hidden = $derived($hiddenRecords.has(hiddenKey('camModel', cam.id)));
	let cover = $derived(cacheUrl(room.live?.thumbnail?.uri ?? room.thumbnail?.uri));
	let site = $derived.by(() => {
		const cameraSites = camSites($sites);
		if (cameraSites.length < 2) return null;
		return cameraSites.find((candidate) => candidate.name === room.plugin) ?? null;
	});
	let performers = $derived(room.performer ? [room.performer] : []);
	let detail = $derived(onlineDetail(room));
	let menuActions = $derived(actionsFor(room, followed, recording));
	let shownReason = $derived(reason ?? cam.reason ?? null);
	let subject = $derived(camSubject(cam.id));

	$effect(() => {
		void loadFollowedCams();
	});

	/** Opens the room, logging it as chosen where it was shown. */
	function open() {
		if (shown) logChosen(subject, shown);
		onclick();
	}

	/** How long the room has been on, or when it was last seen. */
	function onlineDetail(current: CamRow): string | null {
		if (current.live?.startedAt) return `on for ${dayjs(current.live.startedAt).fromNow(true)}`;
		if (current.live) return null;
		if (current.lastSeenOnlineAt) return `seen ${dayjs(current.lastSeenOnlineAt).fromNow()}`;
		return 'offline';
	}

	/**
	 * Starts the preview after a pause, for a mouse or pen over a live room.
	 * The stream and the player load during the pause, so the preview finds
	 * them ready.
	 */
	function hoverStart(event: PointerEvent) {
		if (event.pointerType === 'touch' || !room.live) return;
		previewStream(room.id).catch(() => {});
		void import('hls.js');
		previewTimer = setTimeout(() => (previewing = true), PREVIEW_DELAY_MS);
	}

	/** Stops the preview, or its pending start. */
	function hoverEnd() {
		if (previewTimer) clearTimeout(previewTimer);
		previewTimer = null;
		previewing = false;
	}

	/** The ⋮ menu: follow, auto-record, record or stop, block. */
	function actionsFor(current: CamRow, isFollowed: boolean, isRecording: boolean): SheetAction[] {
		const actions: SheetAction[] = [
			{
				label: isFollowed ? 'Unfollow' : 'Follow',
				icon: isFollowed ? BellSlashIcon : BellIcon,
				run: () => void toggleCamFollow(current.id)
			},
			{
				label: current.autoRecord ? 'Stop auto-recording' : 'Auto-record',
				icon: RecordIcon,
				run: () => void toggleAutoRecord(current)
			}
		];
		if (isRecording) {
			actions.push({ label: 'Stop recording', icon: StopIcon, run: () => void stop(current) });
		} else if (current.live) {
			actions.push({ label: 'Record now', icon: RecordIcon, run: () => void record(current) });
		}
		actions.push({
			label: 'Block this room',
			icon: ProhibitIcon,
			run: () => block('camModel', current.id, current.displayName),
			destructive: true
		});
		return actions;
	}

	/** Flips auto-recording, keeping the room's quality. */
	async function toggleAutoRecord(current: CamRow) {
		changed = await setCamAutoRecord(current.id, !current.autoRecord);
	}

	/** Records the room at its quality from now. */
	async function record(current: CamRow) {
		try {
			await startCamRecording(current.id, current.recordQuality ?? 'source', 0);
			onrecordingchange?.();
		} catch (failure) {
			showUndo({ message: `Couldn't record ${current.displayName}: ${failure}` });
		}
	}

	/** Stops the room's recording; it becomes a video. */
	async function stop(current: CamRow) {
		await stopCamRecording(current.id);
		onrecordingchange?.();
	}
</script>

{#if !hidden}
	<div
		class="group flex w-full flex-col gap-2"
		role="presentation"
		onpointerenter={hoverStart}
		onpointerleave={hoverEnd}
		use:impression={{ subject, shown }}
	>
		<CoverFrame
			src={cover}
			alt={room.displayName}
			aspect="16 / 9"
			placeholder={VideoCameraIcon}
			onclick={open}
		>
			{#if previewing}
				<LiveRoomPreview camId={room.id} />
			{/if}
			{#if !room.live}
				<span class="absolute inset-0 bg-black/50"></span>
			{/if}
			{#if markLive && room.live}
				<CardBadge corner="top-left" live>● Live</CardBadge>
			{/if}
			{#if recording}
				<CardBadge corner="top-right"
					><RecordIcon size={12} weight="fill" class="text-error" /> REC</CardBadge
				>
			{:else if followed || room.autoRecord}
				<CardBadge corner="top-right">
					{#if followed}
						<BellRingingIcon size={12} weight="fill" aria-label="Followed" />
					{/if}
					{#if room.autoRecord}
						<RecordIcon size={12} class="text-error" aria-label="Auto-recorded" />
					{/if}
				</CardBadge>
			{/if}
			{#if room.live}
				<CardBadge corner="bottom-right">
					<EyeIcon size={12} />
					{compactNumber.format(room.live.viewers)}
				</CardBadge>
			{/if}
		</CoverFrame>

		<CardCaption
			title={room.displayName}
			{performers}
			{site}
			siteAvatar={site !== null}
			{detail}
			note={roomTopic(room.live?.subject)}
			onclick={open}
			onmore={(button) => (menuAnchor = button)}
		/>
		{#if shownReason}
			<span class="text-accent -mt-1 truncate text-xs leading-4">{shownReason}</span>
		{/if}
	</div>

	{#if menuAnchor}
		<CardMenu
			anchor={menuAnchor}
			title={room.displayName}
			subtitle={detail}
			imageUrl={cover}
			actions={menuActions}
			onclose={() => (menuAnchor = null)}
		/>
	{/if}
{/if}
