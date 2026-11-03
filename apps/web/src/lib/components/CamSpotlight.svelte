<!--
	The top recommended rooms at the head of the cams page: one plays muted in
	a large 16:9 player (only while it's on screen) with why it's picked, its
	name, audience, Watch and Follow; beside it every pick is listed, and
	pointing at one for a moment, or clicking it, puts it in the player.
	Clicking the playing pick, or the player, opens the room. The playing
	room's picture lights the app's shared glow.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import BellRingingIcon from 'phosphor-svelte/lib/BellRingingIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import LiveRoomPreview from '$lib/components/LiveRoomPreview.svelte';
	import { camSubject, followedCams, toggleCamFollow, type CamRow } from '$lib/cams';
	import { logChosen } from '$lib/events';
	import { fadeInImage } from '$lib/fadeInImage';
	import { cacheUrl } from '$lib/img';
	import { impression } from '$lib/impression';
	import { showAmbientGlow } from '$lib/ambientGlow.svelte';

	interface Props {
		// The picks, best first; each carries its reason.
		rooms: CamRow[];
		// The surface the picks are logged as shown on, at positions 0….
		surface: string;
		onopen: (room: CamRow) => void;
	}

	let { rooms, surface, onopen }: Props = $props();

	// Pointing at a pick this long puts it in the player, so passing over the list doesn't.
	const SELECT_DELAY_MS = 350;

	const compactNumber = new Intl.NumberFormat(undefined, { notation: 'compact' });

	let selectedId = $state<string | null>(null);
	let player = $state<HTMLElement | null>(null);
	let onScreen = $state(false);
	let hoverTimer: ReturnType<typeof setTimeout> | null = null;

	let selected = $derived(rooms.find((room) => room.id === selectedId) ?? rooms[0]);
	let picture = $derived(pictureOf(selected));
	let followed = $derived($followedCams.has(selected.id));

	showAmbientGlow(() => picture);

	// The stream plays only while the player shows, so a scrolled-down page costs nothing.
	$effect(() => {
		if (!player) return;
		const observer = new IntersectionObserver(
			(entries) => {
				onScreen = entries.some((entry) => entry.isIntersecting);
			},
			{ threshold: 0.25 }
		);
		observer.observe(player);
		return () => observer.disconnect();
	});

	/** The room's live picture, else its last one. */
	function pictureOf(room: CamRow): string | null {
		return cacheUrl(room.live?.thumbnail?.uri ?? room.thumbnail?.uri);
	}

	/** "2.1K watching · 1 h", or nothing for a room that isn't live. */
	function audienceLine(room: CamRow): string {
		if (!room.live) return '';
		const parts = [`${compactNumber.format(room.live.viewers)} watching`];
		if (room.live.startedAt) parts.push(dayjs(room.live.startedAt).fromNow(true));
		return parts.join(' · ');
	}

	/** Logs the pick as chosen from the spotlight, then opens it. */
	function open(room: CamRow) {
		logChosen(camSubject(room.id), { surface, position: rooms.indexOf(room) });
		onopen(room);
	}

	/** Puts a pick in the player; clicking the one already playing opens it. */
	function choose(room: CamRow) {
		cancelHover();
		if (room.id === selected.id) {
			open(room);
			return;
		}
		selectedId = room.id;
	}

	/** Puts a pick in the player after the pointer rests on it. */
	function hoverStart(room: CamRow, event: PointerEvent) {
		if (event.pointerType === 'touch') return;
		cancelHover();
		hoverTimer = setTimeout(() => (selectedId = room.id), SELECT_DELAY_MS);
	}

	/** Forgets a pending hover selection. */
	function cancelHover() {
		if (hoverTimer) clearTimeout(hoverTimer);
		hoverTimer = null;
	}
</script>

<div class="relative">

	<section class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(280px,340px)]">
		<div
			bind:this={player}
			class="bg-base-300 relative aspect-video overflow-hidden rounded-2xl ring-1 ring-white/10"
		>
			<button
				type="button"
				class="tv-card absolute inset-0 block h-full w-full"
				aria-label="Watch {selected.displayName}"
				onclick={() => open(selected)}
			>
				{#key selected.id}
					{#if picture}
						<img src={picture} alt="" class="absolute inset-0 h-full w-full object-cover" />
					{/if}
					{#if onScreen && selected.live}
						<LiveRoomPreview camId={selected.id} />
					{/if}
				{/key}
			</button>

			<div
				class="pointer-events-none absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/85 via-black/40 to-transparent"
			></div>

			<div
				class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between gap-4 p-6 max-md:flex-col max-md:items-start max-md:p-4"
			>
				<div class="min-w-0">
					<p
						class="flex items-center gap-2 text-xs font-semibold tracking-wide text-white uppercase"
					>
						<span class="live-dot"></span>
						Top pick for you{#if selected.reason}<span class="text-white/70 normal-case"
								>· {selected.reason}</span
							>{/if}
					</p>
					<h2 class="mt-1.5 truncate text-3xl leading-tight font-bold text-white">
						{selected.displayName}
					</h2>
					<p class="mt-0.5 text-sm text-white/70">{audienceLine(selected)}</p>
				</div>
				<div class="pointer-events-auto flex shrink-0 gap-2">
					<button
						type="button"
						class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold transition-transform hover:scale-[1.03] active:scale-95"
						onclick={() => open(selected)}
					>
						<PlayIcon size={16} weight="fill" /> Watch
					</button>
					<button
						type="button"
						class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95"
						onclick={() => void toggleCamFollow(selected.id)}
					>
						{#if followed}
							<BellRingingIcon size={16} weight="fill" /> Following
						{:else}
							<BellIcon size={16} /> Follow
						{/if}
					</button>
				</div>
			</div>
		</div>

		<ol class="flex flex-col gap-2 max-lg:grid max-lg:grid-cols-2 max-md:grid-cols-1">
			{#each rooms as room, position (room.id)}
				<li
					class="min-h-0 lg:flex-1"
					use:impression={{ subject: camSubject(room.id), shown: { surface, position } }}
				>
					<button
						type="button"
						class="flex h-full w-full items-center gap-3 rounded-xl p-2 text-left transition-colors hover:bg-white/5"
						class:spotlight-on={room.id === selected.id}
						aria-current={room.id === selected.id}
						onclick={() => choose(room)}
						onpointerenter={(event) => hoverStart(room, event)}
						onpointerleave={cancelHover}
					>
						<span
							class="bg-base-300 relative aspect-video h-full max-h-24 min-h-14 shrink-0 overflow-hidden rounded-lg"
						>
							{#if pictureOf(room)}
								<img
									src={pictureOf(room)}
									alt=""
									loading="lazy"
									use:fadeInImage
									class="h-full w-full object-cover"
								/>
							{/if}
						</span>
						<span class="flex min-w-0 flex-col">
							<span class="truncate text-sm font-semibold">{room.displayName}</span>
							<span class="text-base-content/55 truncate text-xs">{audienceLine(room)}</span>
							{#if room.reason}
								<span class="text-base-content/45 truncate text-xs">{room.reason}</span>
							{/if}
						</span>
					</button>
				</li>
			{/each}
		</ol>
	</section>
</div>

<style>
	.spotlight-on {
		background-color: color-mix(in oklab, var(--color-base-content) 8%, transparent);
		box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--color-base-content) 14%, transparent);
	}
</style>
