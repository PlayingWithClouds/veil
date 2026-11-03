<!--
	"Refresh from sites" as a split button: the left part refreshes from
	every site; the caret opens a drop-down listing the sites (each with what
	it would do: read the profile, or look the performer up), where any one
	or several can be ticked and refreshed alone.
-->
<script lang="ts" module>
	/** A site the refresh can ask. */
	export interface RefreshSite {
		// The plugin name.
		name: string;
		label: string;
		iconUrl: string | null;
		// What refreshing it does, e.g. "Profile read 3 days ago", "Looks them up".
		detail: string;
	}
</script>

<script lang="ts">
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import SiteIcon from '$lib/components/scene/SiteIcon.svelte';
	import { pushOverlay } from '$lib/stores/overlayStack';

	interface Props {
		sites: RefreshSite[];
		// A refresh is running.
		busy: boolean;
		// Refreshes from these plugins; empty is every site.
		onrefresh: (plugins: string[]) => void;
	}

	let { sites, busy, onrefresh }: Props = $props();

	let open = $state(false);
	let chosen = $state<string[]>([]);
	let container = $state<HTMLElement | null>(null);

	$effect(() => {
		if (open) return pushOverlay(() => (open = false));
	});

	/** Ticks or unticks one site. */
	function toggle(name: string) {
		if (chosen.includes(name)) {
			chosen = chosen.filter((entry) => entry !== name);
			return;
		}
		chosen = [...chosen, name];
	}

	/** Refreshes from the ticked sites and closes the drop-down. */
	function refreshChosen() {
		if (chosen.length === 0) return;
		open = false;
		onrefresh(chosen);
		chosen = [];
	}

	/** Refreshes from every site. */
	function refreshAll() {
		open = false;
		onrefresh([]);
	}

	/** Closes on a press outside. */
	function onWindowPointerDown(event: PointerEvent) {
		if (!open || !container) return;
		if (container.contains(event.target as Node)) return;
		open = false;
	}
</script>

<svelte:window onpointerdown={onWindowPointerDown} />

<div class="relative" bind:this={container}>
	<div class="glass-pill flex h-10 items-center rounded-full">
		<button
			type="button"
			class="flex h-full items-center gap-2 rounded-l-full pr-3 pl-4 text-sm font-medium transition-colors hover:bg-white/8 disabled:opacity-60"
			disabled={busy}
			onclick={refreshAll}
			title="Read every profile and ask every site again"
		>
			<span class="flex" class:animate-spin={busy}><ArrowClockwiseIcon size={16} /></span>
			{busy ? 'Reading sites…' : 'Refresh'}
		</button>
		<span class="h-5 w-px bg-white/15"></span>
		<button
			type="button"
			class="flex h-full items-center rounded-r-full pr-3 pl-2 transition-colors hover:bg-white/8 disabled:opacity-60"
			disabled={busy || sites.length === 0}
			aria-label="Choose sites to refresh from"
			aria-expanded={open}
			onclick={() => (open = !open)}
		>
			<span class="flex transition-transform" class:rotate-180={open}>
				<CaretDownIcon size={14} weight="bold" />
			</span>
		</button>
	</div>

	{#if open}
		<div class="menu-panel z-overlay absolute top-12 right-0 flex w-80 flex-col p-1.5">
			<p class="text-base-content/45 px-2.5 pt-1.5 pb-2 text-xs">Refresh from just these sites:</p>
			<div class="flex max-h-80 flex-col gap-0.5 overflow-y-auto">
				{#each sites as site (site.name)}
					<label class="menu-item cursor-pointer py-1.5">
						<input
							type="checkbox"
							checked={chosen.includes(site.name)}
							onchange={() => toggle(site.name)}
						/>
						<SiteIcon iconUrl={site.iconUrl} class="h-5 w-5 shrink-0 rounded" />
						<span class="flex min-w-0 flex-col">
							<span class="truncate">{site.label}</span>
							<span class="text-base-content/45 truncate text-xs">{site.detail}</span>
						</span>
					</label>
				{/each}
			</div>
			<div
				class="mt-1 flex items-center justify-between gap-2 border-t border-white/5 px-1 pt-2 pb-0.5"
			>
				<button type="button" class="chip chip-sm" onclick={refreshAll}>All sites</button>
				<button
					type="button"
					class="bg-primary text-primary-content h-7 rounded-full px-3.5 text-xs font-semibold transition-transform active:scale-95 disabled:opacity-40"
					disabled={chosen.length === 0}
					onclick={refreshChosen}
				>
					Refresh {chosen.length}
					{chosen.length === 1 ? 'site' : 'sites'}
				</button>
			</div>
		</div>
	{/if}
</div>
