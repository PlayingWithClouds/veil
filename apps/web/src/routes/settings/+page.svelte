<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { gqlClient } from '$lib/veil';
	import Toggle from '$lib/components/Toggle.svelte';
	import ServerForm from '$lib/components/ServerForm.svelte';
	import { isNativeApp } from '$lib/server';
	import GearSixIcon from 'phosphor-svelte/lib/GearSixIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import { fetchBlocklist, removeBlock, type BlocklistEntry } from '$lib/blocklist';

	let { data } = $props();

	let blocklist = $state<BlocklistEntry[]>([]);
	$effect(() => {
		fetchBlocklist().then((list) => (blocklist = list));
	});
	async function unblock(entry: BlocklistEntry) {
		await removeBlock(entry.targetId);
		blocklist = blocklist.filter((item) => item.id !== entry.id);
	}
	let s = $state({ ...data.settings, kindLimits: data.settings.kindLimits.map((l) => ({ ...l })) });
	let saving = $state(false);
	let saved = $state(false);
	let error = $state('');

	async function save() {
		saving = true;
		error = '';
		saved = false;
		try {
			await gqlClient.mutation({
				updateSettings: {
					__args: {
						input: {
							maxConcurrentJobs: s.maxConcurrentJobs,
							maxJobRetries: s.maxJobRetries,
							downloadSpeedLimitKBps: s.downloadSpeedLimitKBps,
							allowDownloadsWhileStreaming: s.allowDownloadsWhileStreaming,
							autoEnrichAfterScrape: s.autoEnrichAfterScrape,
							requireVpn: s.requireVpn,
							kindLimits: s.kindLimits.map((l) => ({
								kind: l.kind,
								maxConcurrent: l.maxConcurrent,
								retryInitialMs: l.retryInitialMs,
								retryMultiplier: l.retryMultiplier,
								retryMaxMs: l.retryMaxMs
							}))
						}
					},
					maxConcurrentJobs: true
				}
			});
			await invalidateAll();
			saved = true;
			setTimeout(() => (saved = false), 2000);
		} catch (e: any) {
			error = e?.message ?? 'Save failed';
		} finally {
			saving = false;
		}
	}

	const inputClass =
		'w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm outline-none transition-colors focus:border-base-content/30';
	const labelClass = 'text-[13px] font-medium text-base-content/80';
</script>

<div class="flex w-full flex-col gap-6">
	<div class="flex items-center gap-3">
		<span class="text-base-content/70"><GearSixIcon size={26} /></span>
		<h1 class="text-2xl font-bold tracking-tight">Settings</h1>
	</div>

	{#if isNativeApp()}
		<section class="border-base-300 bg-base-200 flex flex-col gap-4 rounded-xl border p-5">
			<h2 class="text-sm font-semibold">Server</h2>
			<ServerForm />
		</section>
	{/if}

	<!-- Automation -->
	<section class="border-base-300 bg-base-200 flex flex-col gap-4 rounded-xl border p-5">
		<h2 class="text-sm font-semibold">Automation</h2>

		<label class="flex items-center justify-between gap-4 py-1">
			<span class={labelClass}>Auto-enrich after scrape</span>
			<Toggle checked={s.autoEnrichAfterScrape} onchange={(v) => (s.autoEnrichAfterScrape = v)} />
		</label>
	</section>

	<!-- Jobs -->
	<section class="border-base-300 bg-base-200 flex flex-col gap-4 rounded-xl border p-5">
		<h2 class="text-sm font-semibold">Jobs</h2>

		<div class="flex flex-col gap-1.5">
			<label class={labelClass} for="max-jobs">Max concurrent jobs (global)</label>
			<input id="max-jobs" type="number" min="1" max="16" class="{inputClass} max-w-[8rem]" bind:value={s.maxConcurrentJobs} />
			<p class="text-base-content/40 text-xs">Overridden per-kind below.</p>
		</div>

		<div class="flex flex-col gap-1.5">
			<label class={labelClass} for="max-retries">Max job retries</label>
			<input id="max-retries" type="number" min="0" max="10" class="{inputClass} max-w-[8rem]" bind:value={s.maxJobRetries} />
		</div>

		<label class="flex items-center justify-between gap-4 py-1">
			<span class={labelClass}>Require VPN for downloads</span>
			<Toggle checked={s.requireVpn} onchange={(v) => (s.requireVpn = v)} />
		</label>
	</section>

	<!-- Downloads -->
	<section class="border-base-300 bg-base-200 flex flex-col gap-4 rounded-xl border p-5">
		<h2 class="text-sm font-semibold">Downloads</h2>

		<div class="flex flex-col gap-1.5">
			<label class={labelClass} for="speed-limit">Download speed limit (KB/s)</label>
			<input id="speed-limit" type="number" min="0" class="{inputClass} max-w-[8rem]" bind:value={s.downloadSpeedLimitKBps} />
			<p class="text-base-content/40 text-xs">0 = unlimited</p>
		</div>

		<label class="flex items-center justify-between gap-4 py-1">
			<span class="flex flex-col">
				<span class={labelClass}>Allow downloads while streaming</span>
				<span class="text-base-content/40 text-xs">When off, downloads pause while a video is playing.</span>
			</span>
			<Toggle
				checked={s.allowDownloadsWhileStreaming}
				onchange={(v) => (s.allowDownloadsWhileStreaming = v)}
			/>
		</label>
	</section>

	<!-- Blocklist -->
	<section class="border-base-300 bg-base-200 flex flex-col gap-4 rounded-xl border p-5">
		<div>
			<h2 class="text-sm font-semibold">Blocklist</h2>
			<p class="text-base-content/40 mt-1 text-xs">
				Blocked tags, performers, and studios are hidden from every feed. Block from a scene's page.
			</p>
		</div>

		{#if blocklist.length === 0}
			<p class="text-base-content/40 text-sm">Nothing blocked.</p>
		{:else}
			<div class="flex flex-wrap gap-2">
				{#each blocklist as entry (entry.id)}
					<span class="bg-base-300 flex items-center gap-2 rounded-full py-1 pl-3 pr-1.5 text-sm">
						<ProhibitIcon size={13} class="text-base-content/40" />
						<span>{entry.label ?? entry.targetId}</span>
						<span class="text-base-content/30 text-xs">{entry.kind}</span>
						<button
							type="button"
							class="text-base-content/40 hover:text-error flex h-5 w-5 items-center justify-center"
							onclick={() => unblock(entry)}
							aria-label="Unblock {entry.label ?? entry.targetId}"
						>
							✕
						</button>
					</span>
				{/each}
			</div>
		{/if}
	</section>

	<!-- Rate limiting -->
	{#if s.kindLimits.length > 0}
		<section class="border-base-300 bg-base-200 flex flex-col gap-4 rounded-xl border p-5">
			<div>
				<h2 class="text-sm font-semibold">Rate limiting</h2>
				<p class="text-base-content/40 mt-1 text-xs">
					Per-kind concurrency and exponential retry backoff. Max concurrent 0 falls back to the global
					setting. Backoff: delay = InitialMs × Multiplier^(attempt−1), capped at MaxMs.
				</p>
			</div>

			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="text-base-content/50 border-base-300 border-b text-left text-xs">
							<th class="py-2 pr-4 font-medium">Kind</th>
							<th class="px-2 py-2 font-medium">Max concurrent</th>
							<th class="px-2 py-2 font-medium">Initial retry (ms)</th>
							<th class="px-2 py-2 font-medium">Multiplier</th>
							<th class="px-2 py-2 font-medium">Max retry (ms)</th>
						</tr>
					</thead>
					<tbody>
						{#each s.kindLimits as limit (limit.kind)}
							<tr class="border-base-300/50 border-b last:border-0">
								<td class="py-2 pr-4">
									<span class="bg-base-300 text-base-content/70 rounded px-2 py-0.5 font-mono text-xs"
										>{limit.kind}</span
									>
								</td>
								<td class="px-2 py-2">
									<input type="number" min="0" max="32" class="{inputClass} w-20" bind:value={limit.maxConcurrent} />
								</td>
								<td class="px-2 py-2">
									<input type="number" min="0" step="100" class="{inputClass} w-28" bind:value={limit.retryInitialMs} />
								</td>
								<td class="px-2 py-2">
									<input
										type="number"
										min="1"
										max="10"
										step="0.1"
										class="{inputClass} w-20"
										bind:value={limit.retryMultiplier}
									/>
								</td>
								<td class="px-2 py-2">
									<input type="number" min="0" step="1000" class="{inputClass} w-28" bind:value={limit.retryMaxMs} />
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>
	{/if}

	{#if error}
		<div class="rounded-lg border border-[#f87171]/30 bg-[#f87171]/10 px-4 py-2.5 text-sm text-[#f87171]">
			{error}
		</div>
	{/if}

	<div class="flex items-center gap-4 pb-2">
		<button
			class="flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-semibold text-black transition-opacity hover:opacity-90 disabled:opacity-50"
			onclick={save}
			disabled={saving}
		>
			{#if saving}<span class="loading loading-spinner loading-xs"></span>{/if}
			{saving ? 'Saving…' : 'Save settings'}
		</button>
		{#if saved}
			<span class="text-sm font-medium text-[#4ade80]">Saved ✓</span>
		{/if}
	</div>
</div>
