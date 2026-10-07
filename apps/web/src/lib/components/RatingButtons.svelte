<script lang="ts">
	import ThumbsUpIcon from 'phosphor-svelte/lib/ThumbsUpIcon';
	import ThumbsDownIcon from 'phosphor-svelte/lib/ThumbsDownIcon';
	import { notifications } from '$lib/stores/notifications';
	import { fetchVerdict, setVerdict, clearVerdict, type Verdict } from '$lib/rating';

	interface Props {
		mediaId: string;
		// One pill split down the middle instead of two separate pills.
		joined?: boolean;
		class?: string;
	}

	let { mediaId, joined = false, class: className = '' }: Props = $props();

	let verdict = $state<Verdict>(null);
	let busy = $state(false);
	// Restarts the thumb's animation each time a verdict is set; 0 = none yet.
	let animationKey = $state(0);
	const SPARKS = 7;

	// Reflect the stored verdict; re-checks when the media changes.
	$effect(() => {
		if (!mediaId) {
			verdict = null;
			return;
		}
		let cancelled = false;
		fetchVerdict(mediaId).then((value) => {
			if (!cancelled) verdict = value;
		});
		return () => {
			cancelled = true;
		};
	});

	// Clicking the active thumb clears the verdict; otherwise sets it.
	async function choose(next: Exclude<Verdict, null>) {
		if (!mediaId || busy) return;
		busy = true;
		const previous = verdict;
		const target: Verdict = verdict === next ? null : next;
		verdict = target;
		if (target !== null) {
			animationKey = performance.now();
			navigator.vibrate?.(target === 'up' ? 12 : [8, 40, 8]);
		}
		try {
			if (target === null) await clearVerdict(mediaId);
			else await setVerdict(mediaId, target);
		} catch {
			verdict = previous;
			notifications.push('Could not save your rating', 'error');
		} finally {
			busy = false;
		}
	}
</script>

{#if joined}
	<!-- One glass pill split by a hairline, like the tab bar. -->
	<div class="glass-pill flex h-10 items-center overflow-hidden rounded-full {className}">
		<button
			type="button"
			aria-label="Thumbs up"
			aria-pressed={verdict === 'up'}
			disabled={busy}
			onclick={() => choose('up')}
			class="flex h-full items-center justify-center pr-3.5 pl-4 transition-colors disabled:opacity-50"
			class:glass-pill-on={verdict === 'up'}
			class:text-success={verdict === 'up'}
		>
			{@render thumb('up')}
		</button>
		<span class="bg-base-content/15 h-5 w-px"></span>
		<button
			type="button"
			aria-label="Thumbs down"
			aria-pressed={verdict === 'down'}
			disabled={busy}
			onclick={() => choose('down')}
			class="flex h-full items-center justify-center pr-4 pl-3.5 transition-colors disabled:opacity-50"
			class:glass-pill-on={verdict === 'down'}
			class:text-error={verdict === 'down'}
		>
			{@render thumb('down')}
		</button>
	</div>
{:else}
<div class="flex items-center gap-2 {className}">
	<button
		type="button"
		aria-label="Thumbs up"
		aria-pressed={verdict === 'up'}
		disabled={busy}
		onclick={() => choose('up')}
		class="flex items-center justify-center rounded-full px-4 py-2.5 text-sm font-semibold transition-colors disabled:opacity-50 {verdict ===
		'up'
			? 'bg-success text-black'
			: 'bg-white/10 text-white hover:bg-white/20'}"
	>
		{@render thumb('up')}
	</button>
	<button
		type="button"
		aria-label="Thumbs down"
		aria-pressed={verdict === 'down'}
		disabled={busy}
		onclick={() => choose('down')}
		class="flex items-center justify-center rounded-full px-4 py-2.5 text-sm font-semibold transition-colors disabled:opacity-50 {verdict ===
		'down'
			? 'bg-error text-white'
			: 'bg-white/10 text-white hover:bg-white/20'}"
	>
		{@render thumb('down')}
	</button>
</div>
{/if}

<!-- A thumb icon that animates when its verdict is set: up springs up with a
     tilt and a ring of green sparks, down drops and shakes. -->
{#snippet thumb(kind: 'up' | 'down')}
	{#key verdict === kind ? animationKey : 0}
		<span
			class="relative flex"
			class:thumb-up-pop={verdict === kind && kind === 'up' && animationKey > 0}
			class:thumb-down-shake={verdict === kind && kind === 'down' && animationKey > 0}
		>
			{#if kind === 'up'}
				<ThumbsUpIcon size={18} weight={verdict === 'up' ? 'fill' : 'regular'} />
			{:else}
				<ThumbsDownIcon size={18} weight={verdict === 'down' ? 'fill' : 'regular'} />
			{/if}
			{#if kind === 'up' && verdict === 'up' && animationKey > 0}
				{#each { length: SPARKS } as _, index (index)}
					<span class="thumb-spark" style:--spark-angle="{(index / SPARKS) * 360}deg"></span>
				{/each}
			{/if}
		</span>
	{/key}
{/snippet}

<style>
	.thumb-up-pop {
		animation: thumb-up-pop 520ms cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	@keyframes thumb-up-pop {
		0% {
			scale: 1;
			rotate: 0deg;
		}
		35% {
			scale: 1.45;
			rotate: -18deg;
			translate: 0 -3px;
		}
		100% {
			scale: 1;
			rotate: 0deg;
			translate: 0 0;
		}
	}

	.thumb-down-shake {
		animation: thumb-down-shake 480ms ease-out;
	}

	@keyframes thumb-down-shake {
		0% {
			translate: 0 0;
		}
		20% {
			translate: 0 3px;
			scale: 1.2;
		}
		40% {
			translate: -2px 3px;
			rotate: -10deg;
		}
		60% {
			translate: 2px 2px;
			rotate: 8deg;
		}
		80% {
			translate: -1px 1px;
			rotate: -4deg;
		}
		100% {
			translate: 0 0;
			rotate: 0deg;
			scale: 1;
		}
	}

	.thumb-spark {
		position: absolute;
		left: 50%;
		top: 50%;
		width: 3px;
		height: 3px;
		margin: -1.5px;
		border-radius: 9999px;
		background: var(--color-success);
		animation: thumb-spark 480ms ease-out forwards;
	}

	/* Rotate first, then push out: each spark flies along its own angle. */
	@keyframes thumb-spark {
		0% {
			transform: rotate(var(--spark-angle)) translateY(-4px);
			opacity: 1;
		}
		100% {
			transform: rotate(var(--spark-angle)) translateY(-14px);
			opacity: 0;
		}
	}
</style>
