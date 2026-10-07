<!--
	Top of a performer or studio page: the image, blurred and darkened, fills the
	area behind the phone top bar and fades into the page; the sharp image, the
	name, an actions row and a few stat tiles sit centred on it. While the name is
	on screen the top bar stays transparent; once it scrolls away the bar shows
	the name instead (pageTitle store).
-->
<script module lang="ts">
	/** One stat tile: a short value over a small label. */
	export type HeroStat = { label: string; value: string };
</script>

<script lang="ts">
	import type { Snippet } from 'svelte';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
	import { fadeInImage } from '$lib/fadeInImage';
	import { isCompact } from '$lib/stores/viewport';
	import { pageHeroVisible, pageTitle } from '$lib/stores/pageTitle';

	interface Props {
		name: string;
		imageUrl: string | null;
		// A performer's portrait photo or a studio's logo.
		imageShape: 'portrait' | 'logo';
		// Shown under the name (aliases, parent network, …).
		subtitle?: Snippet;
		// Buttons row under the name (Subscribe, favourite, ⋮).
		actions: Snippet;
		// Only filled-in stats; the first four are shown.
		stats: HeroStat[];
	}

	let { name, imageUrl, imageShape, subtitle, actions, stats }: Props = $props();

	const MAX_STATS = 4;

	let nameElement = $state<HTMLElement | null>(null);

	$effect(() => {
		let shape: 'avatar' | 'logo' = 'avatar';
		if (imageShape === 'logo') shape = 'logo';
		pageTitle.set({ title: name, imageUrl, imageShape: shape });
	});

	// Leaving the page hands the top bar back.
	$effect(() => () => {
		pageTitle.set(null);
		pageHeroVisible.set(false);
	});

	// The name counts as scrolled away once it slides under the top bar.
	$effect(() => {
		if (!nameElement) return;
		const observer = new IntersectionObserver(
			(entries) => pageHeroVisible.set(entries[0]?.isIntersecting ?? false),
			{ rootMargin: '-64px 0px 0px 0px' }
		);
		observer.observe(nameElement);
		return () => observer.disconnect();
	});
</script>

<section class="entity-hero relative flex flex-col items-center gap-4 overflow-hidden pb-2" class:entity-hero-compact={$isCompact}>
	{#if imageUrl}
		<div class="pointer-events-none absolute inset-0 -z-10 overflow-hidden" aria-hidden="true">
			<img src={imageUrl} alt="" class="hero-backdrop h-full w-full object-cover" />
			<div class="hero-fade absolute inset-0"></div>
		</div>
	{/if}

	{#if imageShape === 'portrait'}
		<div class="hero-portrait bg-base-300 ring-base-content/10 aspect-[3/4] w-36 overflow-hidden rounded-3xl shadow-2xl ring-1">
			{#if imageUrl}
				<img src={imageUrl} alt={name} use:fadeInImage class="h-full w-full object-cover" />
			{:else}
				<div class="flex h-full w-full items-center justify-center">
					<UserIcon size={48} class="opacity-20" />
				</div>
			{/if}
		</div>
	{:else}
		<div class="hero-logo ring-base-content/10 flex aspect-video w-44 items-center justify-center overflow-hidden rounded-2xl p-3 shadow-2xl ring-1">
			{#if imageUrl}
				<img src={imageUrl} alt={name} use:fadeInImage class="h-full w-full object-contain" />
			{:else}
				<FilmSlateIcon size={44} class="opacity-25" />
			{/if}
		</div>
	{/if}

	<div class="flex max-w-full flex-col items-center gap-1 px-2 text-center">
		<h1 bind:this={nameElement} class="text-3xl leading-tight font-bold tracking-tight">{name}</h1>
		{@render subtitle?.()}
	</div>

	<div class="flex items-center gap-2">
		{@render actions()}
	</div>

	{#if stats.length > 0}
		<div class="grid w-full max-w-md auto-cols-fr grid-flow-col gap-2 px-1">
			{#each stats.slice(0, MAX_STATS) as stat (stat.label)}
				<div class="hero-stat flex min-w-0 flex-col items-center rounded-2xl px-2 py-2.5">
					<span class="max-w-full truncate text-base font-semibold tabular-nums">{stat.value}</span>
					<span class="text-base-content/55 text-[11px] font-medium tracking-wide uppercase">{stat.label}</span>
				</div>
			{/each}
		</div>
	{/if}
</section>

<style>
	.entity-hero {
		isolation: isolate;
		padding-top: 1.5rem;
		border-radius: 1.75rem;
	}

	/* Phones: bleed to the screen edges and up behind the transparent top bar. */
	.entity-hero-compact {
		margin-inline: -0.75rem;
		margin-top: calc(-1 * var(--compact-header-height));
		padding-top: calc(var(--compact-header-height) + 0.75rem);
		border-radius: 0;
	}

	.hero-backdrop {
		scale: 1.3;
		filter: blur(40px) saturate(1.4);
		opacity: 0.55;
	}

	.hero-fade {
		background: linear-gradient(
			to bottom,
			color-mix(in oklab, var(--color-base-100) 30%, transparent) 0%,
			color-mix(in oklab, var(--color-base-100) 55%, transparent) 55%,
			var(--color-base-100) 100%
		);
	}

	.hero-logo {
		background-color: color-mix(in oklab, var(--color-base-content) 8%, transparent);
		backdrop-filter: blur(12px);
	}

	.hero-stat {
		background-color: color-mix(in oklab, var(--color-base-content) 7%, transparent);
		backdrop-filter: blur(12px);
	}

	/* Portrait and name settle in when the page opens. */
	.hero-portrait,
	.hero-logo {
		animation: hero-enter 420ms cubic-bezier(0.22, 1, 0.36, 1) both;
	}

	@keyframes hero-enter {
		from {
			opacity: 0;
			scale: 0.94;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.hero-portrait,
		.hero-logo {
			animation: none;
		}
	}
</style>
