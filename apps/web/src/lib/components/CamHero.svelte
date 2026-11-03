<!--
	The top recommended room at the head of the cams page: its stream plays
	muted across a wide banner (only while the banner is on screen), with why
	it's picked, the room's name, audience, topic and tags over it, and Watch
	and Follow. Its picture, blurred, lights the page behind it.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import BellRingingIcon from 'phosphor-svelte/lib/BellRingingIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import LiveRoomPreview from '$lib/components/LiveRoomPreview.svelte';
	import { camSubject, followedCams, roomTopic, toggleCamFollow, type CamRow } from '$lib/cams';
	import { logChosen, type Shown } from '$lib/events';
	import { cacheUrl } from '$lib/img';
	import { impression } from '$lib/impression';
	import { camsUrl } from '$lib/routes';

	interface Props {
		cam: CamRow;
		// Why the room is picked, e.g. "Because you like #latina".
		reason?: string | null;
		// Where the banner is shown; logs it as shown there and, on Watch, chosen.
		shown?: Shown | null;
		onopen: () => void;
	}

	let { cam, reason = null, shown = null, onopen }: Props = $props();

	// Tags named on the banner; the room page lists them all.
	const SHOWN_TAGS = 6;

	const compactNumber = new Intl.NumberFormat(undefined, { notation: 'compact' });

	let banner = $state<HTMLElement | null>(null);
	let onScreen = $state(false);

	let picture = $derived(cacheUrl(cam.live?.thumbnail?.uri ?? cam.thumbnail?.uri));
	let followed = $derived($followedCams.has(cam.id));
	let topic = $derived(roomTopic(cam.live?.subject));
	let tags = $derived(cam.tags.slice(0, SHOWN_TAGS));
	let subject = $derived(camSubject(cam.id));

	/** Logs the pick where the banner is shown, then opens the room. */
	function open() {
		if (shown) logChosen(subject, shown);
		onopen();
	}

	// The stream plays only while the banner shows, so a scrolled-down page costs nothing.
	$effect(() => {
		if (!banner) return;
		const observer = new IntersectionObserver(
			(entries) => {
				onScreen = entries.some((entry) => entry.isIntersecting);
			},
			{ threshold: 0.25 }
		);
		observer.observe(banner);
		return () => observer.disconnect();
	});

	/** The link that narrows the live rooms to a tag. */
	function tagHref(name: string): string {
		return `${camsUrl()}?${new URLSearchParams({ tag: name })}`;
	}
</script>

<div class="relative">
	{#if picture}
		<img
			src={picture}
			alt=""
			aria-hidden="true"
			class="pointer-events-none absolute -inset-x-24 -top-48 -z-10 h-[150%] w-[calc(100%+12rem)] max-w-none scale-110 object-cover opacity-35 blur-[90px] saturate-150"
			style:mask-image="radial-gradient(ellipse at 50% 40%, black 30%, transparent 70%)"
		/>
	{/if}

	<section
		bind:this={banner}
		use:impression={{ subject, shown }}
		class="bg-base-300 relative h-[clamp(320px,48vh,500px)] overflow-hidden rounded-2xl ring-1 ring-white/10 max-md:aspect-[4/5] max-md:h-auto"
	>
		<button
			type="button"
			class="tv-card absolute inset-0 block h-full w-full"
			aria-label="Watch {cam.displayName}"
			onclick={open}
		>
			{#if picture}
				<img src={picture} alt="" class="absolute inset-0 h-full w-full object-cover" />
			{/if}
			{#if onScreen && cam.live}
				<LiveRoomPreview camId={cam.id} />
			{/if}
		</button>

		<div
			class="pointer-events-none absolute inset-0 bg-gradient-to-r from-black/85 via-black/45 to-transparent max-md:bg-gradient-to-t max-md:from-black/90 max-md:via-black/40"
		></div>

		<div
			class="pointer-events-none absolute inset-y-0 left-0 flex max-w-xl flex-col justify-end gap-3 p-8 max-md:inset-x-0 max-md:max-w-none max-md:p-5"
		>
			<p class="text-live flex items-center gap-2 text-xs font-semibold tracking-wide uppercase">
				<span class="live-dot"></span>
				Top pick for you{#if reason}<span class="text-white/70 normal-case">· {reason}</span>{/if}
			</p>
			<h2 class="truncate text-4xl leading-tight font-bold text-white max-md:text-3xl">
				{cam.displayName}
			</h2>
			<p class="flex flex-wrap items-center gap-x-2 text-sm text-white/75">
				{#if cam.live}
					<span>{compactNumber.format(cam.live.viewers)} watching</span>
					{#if cam.live.startedAt}
						<span aria-hidden="true">·</span>
						<span>on for {dayjs(cam.live.startedAt).fromNow(true)}</span>
					{/if}
				{/if}
				{#if cam.performer}
					<span aria-hidden="true">·</span>
					<span>{cam.performer.name}</span>
				{/if}
			</p>
			{#if topic}
				<p class="line-clamp-2 text-sm text-white/60">{topic}</p>
			{/if}
			{#if tags.length > 0}
				<div class="pointer-events-auto flex flex-wrap gap-1.5">
					{#each tags as tag (tag)}
						<a
							href={tagHref(tag)}
							class="rounded-full border border-white/20 bg-black/30 px-2.5 py-0.5 text-xs text-white/80 backdrop-blur-sm transition-colors hover:border-white/40 hover:text-white"
						>
							#{tag}
						</a>
					{/each}
				</div>
			{/if}
			<div class="pointer-events-auto mt-1 flex gap-2">
				<button
					type="button"
					class="bg-primary text-primary-content flex h-10 items-center gap-2 rounded-full px-5 text-sm font-semibold transition-transform hover:scale-[1.03] active:scale-95"
					onclick={open}
				>
					<PlayIcon size={16} weight="fill" /> Watch
				</button>
				<button
					type="button"
					class="glass-pill flex h-10 items-center gap-2 rounded-full px-4 text-sm font-medium transition-transform hover:scale-[1.03] active:scale-95"
					onclick={() => void toggleCamFollow(cam.id)}
				>
					{#if followed}
						<BellRingingIcon size={16} weight="fill" /> Following
					{:else}
						<BellIcon size={16} /> Follow
					{/if}
				</button>
			</div>
		</div>
	</section>
</div>
