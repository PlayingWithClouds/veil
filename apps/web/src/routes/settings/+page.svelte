<script lang="ts">
	// Settings: one column of grouped rows (what a setting does on the left,
	// its control right beside it), a section index on wide screens, and a save
	// bar that appears only while something differs from what is stored.
	import { invalidateAll } from '$app/navigation';
	import { gqlClient } from '$lib/veil';
	import Toggle from '$lib/components/Toggle.svelte';
	import ServerForm from '$lib/components/ServerForm.svelte';
	import { isNativeApp } from '$lib/server';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { fetchBlocklist, removeBlock, type BlockKind, type BlocklistEntry } from '$lib/blocklist';
	import type { Snippet } from 'svelte';

	let { data } = $props();

	type Settings = typeof data.settings;

	/** A deep copy of the settings, so edits never touch the loaded ones. */
	function copySettings(settings: Settings): Settings {
		return { ...settings, kindLimits: settings.kindLimits.map((limit) => ({ ...limit })) };
	}

	let stored = $state(copySettings(data.settings));
	let edited = $state(copySettings(data.settings));
	let dirty = $derived(JSON.stringify(edited) !== JSON.stringify(stored));
	let saving = $state(false);
	let saved = $state(false);
	let error = $state('');

	const SECTIONS = [
		{ id: 'downloads', label: 'Downloads' },
		{ id: 'jobs', label: 'Background jobs' },
		{ id: 'library', label: 'Library' },
		{ id: 'cams', label: 'Live cams' },
		{ id: 'blocklist', label: 'Blocklist' },
		{ id: 'advanced', label: 'Advanced' }
	];

	const BLOCK_KIND_LABELS: Record<BlockKind, string> = {
		performer: 'Performers',
		studio: 'Studios',
		tag: 'Tags'
	};

	let blocklist = $state<BlocklistEntry[]>([]);
	$effect(() => {
		fetchBlocklist().then((list) => (blocklist = list));
	});
	let blockGroups = $derived(
		(Object.keys(BLOCK_KIND_LABELS) as BlockKind[])
			.map((kind) => ({ kind, entries: blocklist.filter((entry) => entry.kind === kind) }))
			.filter((group) => group.entries.length > 0)
	);

	// Whether a VPN interface is up, for the VPN setting's status line.
	let vpnConnected = $state<boolean | null>(null);
	$effect(() => {
		gqlClient
			.query({ vpnStatus: { connected: true } })
			.then((result) => (vpnConnected = result.vpnStatus.connected))
			.catch(() => (vpnConnected = null));
	});

	/** Unblocks one entry. */
	async function unblock(entry: BlocklistEntry) {
		await removeBlock(entry.targetId);
		blocklist = blocklist.filter((item) => item.id !== entry.id);
	}

	/** Throws the unsaved edits away. */
	function discard() {
		edited = copySettings(stored);
		error = '';
	}

	/** Stores the edited settings. */
	async function save() {
		saving = true;
		error = '';
		saved = false;
		try {
			await gqlClient.mutation({
				updateSettings: {
					__args: {
						input: {
							maxConcurrentJobs: edited.maxConcurrentJobs,
							maxJobRetries: edited.maxJobRetries,
							downloadSpeedLimitKBps: edited.downloadSpeedLimitKBps,
							allowDownloadsWhileStreaming: edited.allowDownloadsWhileStreaming,
							camRecordingMaxMinutes: edited.camRecordingMaxMinutes,
							autoEnrichAfterScrape: edited.autoEnrichAfterScrape,
							requireVpn: edited.requireVpn,
							kindLimits: edited.kindLimits.map((limit) => ({
								kind: limit.kind,
								maxConcurrent: limit.maxConcurrent,
								retryInitialMs: limit.retryInitialMs,
								retryMultiplier: limit.retryMultiplier,
								retryMaxMs: limit.retryMaxMs
							}))
						}
					},
					maxConcurrentJobs: true
				}
			});
			stored = copySettings(edited);
			await invalidateAll();
			saved = true;
			setTimeout(() => (saved = false), 2000);
		} catch (caught) {
			error = 'Save failed';
			if (caught instanceof Error) error = caught.message;
		} finally {
			saving = false;
		}
	}

	/** Status line of the VPN setting. */
	function vpnStatusLabel(connected: boolean | null): string {
		if (connected === null) return 'VPN status unknown';
		if (connected) return 'A VPN is connected now';
		return 'No VPN is connected now';
	}

	const numberInputClass =
		'w-24 rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-right text-sm tabular-nums outline-none transition-colors focus:border-base-content/30';
</script>

<div class="flex w-full justify-center pb-24">
	<div class="grid w-full max-w-5xl grid-cols-1 gap-8 lg:grid-cols-[11rem_minmax(0,1fr)]">
		<nav class="hidden lg:block" aria-label="Settings sections">
			<div class="sticky top-4 flex flex-col gap-0.5">
				<h1 class="mb-3 text-2xl font-bold tracking-tight">Settings</h1>
				{#each SECTIONS as section (section.id)}
					<a
						href="#{section.id}"
						class="text-base-content/60 hover:text-base-content hover:bg-base-200 rounded-lg px-3 py-1.5 text-sm transition-colors"
					>
						{section.label}
					</a>
				{/each}
			</div>
		</nav>

		<div class="flex min-w-0 flex-col gap-8">
			<h1 class="text-2xl font-bold tracking-tight lg:hidden">Settings</h1>

			{#if isNativeApp()}
				{#snippet serverBody()}
					<div class="p-4"><ServerForm /></div>
				{/snippet}
				{@render section('server', 'Server', null, serverBody)}
			{/if}

			{#snippet downloadsBody()}
				{#snippet speedControl()}
					<span class="flex items-center gap-2">
						<input
							id="speed-limit"
							type="number"
							min="0"
							class={numberInputClass}
							bind:value={edited.downloadSpeedLimitKBps}
						/>
						<span class="text-base-content/50 text-xs">KB/s</span>
					</span>
				{/snippet}
				{@render row('Speed limit', '0 means unlimited.', speedControl, 'speed-limit')}
				{#snippet streamingControl()}
					<Toggle
						checked={edited.allowDownloadsWhileStreaming}
						onchange={(value) => (edited.allowDownloadsWhileStreaming = value)}
						ariaLabel="Allow downloads while streaming"
					/>
				{/snippet}
				{@render row(
					'Download while watching',
					'When off, downloads pause while a video is playing, so playback gets the bandwidth.',
					streamingControl
				)}
			{/snippet}
			{@render section('downloads', 'Downloads', null, downloadsBody)}

			{#snippet jobsBody()}
				{#snippet concurrencyControl()}
					<input
						id="max-jobs"
						type="number"
						min="1"
						max="16"
						class={numberInputClass}
						bind:value={edited.maxConcurrentJobs}
					/>
				{/snippet}
				{@render row(
					'Jobs at once',
					'How many scrapes, downloads and checks run in parallel. Per-kind limits under Advanced take precedence.',
					concurrencyControl,
					'max-jobs'
				)}
				{#snippet retriesControl()}
					<input
						id="max-retries"
						type="number"
						min="0"
						max="10"
						class={numberInputClass}
						bind:value={edited.maxJobRetries}
					/>
				{/snippet}
				{@render row('Retries', 'How often a failed job is tried again before it gives up.', retriesControl, 'max-retries')}
				{#snippet vpnControl()}
					<Toggle
						checked={edited.requireVpn}
						onchange={(value) => (edited.requireVpn = value)}
						ariaLabel="Only run jobs on a VPN"
					/>
				{/snippet}
				{@render row(
					'Only run jobs on a VPN',
					`While on, no background job (downloads, scrapes, checks) starts without a VPN. ${vpnStatusLabel(vpnConnected)}.`,
					vpnControl
				)}
			{/snippet}
			{@render section('jobs', 'Background jobs', null, jobsBody)}

			{#snippet libraryBody()}
				{#snippet enrichControl()}
					<Toggle
						checked={edited.autoEnrichAfterScrape}
						onchange={(value) => (edited.autoEnrichAfterScrape = value)}
						ariaLabel="Enrich after scraping"
					/>
				{/snippet}
				{@render row(
					'Enrich after scraping',
					'Every scraped scene also gets an enrich job, which fills in missing details through the plugins that can.',
					enrichControl
				)}
			{/snippet}
			{@render section('library', 'Library', null, libraryBody)}

			{#snippet camsBody()}
				{#snippet recordingControl()}
					<span class="flex items-center gap-2">
						<input
							id="cam-recording-max"
							type="number"
							min="1"
							class={numberInputClass}
							bind:value={edited.camRecordingMaxMinutes}
						/>
						<span class="text-base-content/50 text-xs">min</span>
					</span>
				{/snippet}
				{@render row(
					'Longest recording',
					'A longer session is cut here; with auto-record on, a new recording starts. Default 240.',
					recordingControl,
					'cam-recording-max'
				)}
			{/snippet}
			{@render section('cams', 'Live cams', null, camsBody)}

			{#snippet blocklistBody()}
				{#if blockGroups.length === 0}
					<p class="text-base-content/40 px-4 py-3.5 text-sm">
						Nothing blocked. Block a performer, studio or tag from a scene's ⋮ menu.
					</p>
				{:else}
					{#each blockGroups as group (group.kind)}
						<div class="flex flex-col gap-2 px-4 py-3.5">
							<h3 class="text-base-content/50 text-xs font-semibold">
								{BLOCK_KIND_LABELS[group.kind]}
								<span class="font-normal">{group.entries.length}</span>
							</h3>
							<div class="flex flex-wrap gap-2">
								{#each group.entries as entry (entry.id)}
									<span class="bg-base-300 flex items-center gap-1 rounded-full py-1 pr-1 pl-3 text-sm">
										{entry.label ?? entry.targetId}
										<button
											type="button"
											class="text-base-content/40 hover:text-base-content flex h-6 w-6 items-center justify-center rounded-full hover:bg-white/10"
											onclick={() => unblock(entry)}
											aria-label="Unblock {entry.label ?? entry.targetId}"
											title="Unblock"
										>
											<XIcon size={12} weight="bold" />
										</button>
									</span>
								{/each}
							</div>
						</div>
					{/each}
				{/if}
			{/snippet}
			{@render section(
				'blocklist',
				'Blocklist',
				'Blocked performers, studios and tags are left out of every feed and search. Unblocking takes effect right away.',
				blocklistBody
			)}

			{#snippet advancedBody()}
				<details class="group">
					<summary
						class="flex cursor-pointer list-none items-center justify-between gap-4 px-4 py-3.5 text-sm font-medium"
					>
						Limits per job kind
						<span class="transition-transform group-open:rotate-90">
							<CaretRightIcon size={14} weight="bold" />
						</span>
					</summary>
					<div class="flex flex-col gap-3 px-4 pb-4">
						<p class="text-base-content/50 text-xs">
							Max concurrent 0 uses "Jobs at once". A retry waits Initial × Multiplier^(attempt − 1)
							milliseconds, at most Max.
						</p>
						<div class="overflow-x-auto">
							<table class="w-full text-sm">
								<thead>
									<tr class="text-base-content/50 border-base-300 border-b text-left text-xs">
										<th class="py-2 pr-4 font-medium">Kind</th>
										<th class="px-2 py-2 font-medium">Max concurrent</th>
										<th class="px-2 py-2 font-medium">Initial (ms)</th>
										<th class="px-2 py-2 font-medium">Multiplier</th>
										<th class="px-2 py-2 font-medium">Max (ms)</th>
									</tr>
								</thead>
								<tbody>
									{#each edited.kindLimits as limit (limit.kind)}
										<tr class="border-base-300/50 border-b last:border-0">
											<td class="py-2 pr-4">
												<span class="bg-base-300 text-base-content/70 rounded px-2 py-0.5 font-mono text-xs">
													{limit.kind}
												</span>
											</td>
											<td class="px-2 py-2">
												<input type="number" min="0" max="32" class={numberInputClass} bind:value={limit.maxConcurrent} />
											</td>
											<td class="px-2 py-2">
												<input type="number" min="0" step="100" class={numberInputClass} bind:value={limit.retryInitialMs} />
											</td>
											<td class="px-2 py-2">
												<input
													type="number"
													min="1"
													max="10"
													step="0.1"
													class={numberInputClass}
													bind:value={limit.retryMultiplier}
												/>
											</td>
											<td class="px-2 py-2">
												<input type="number" min="0" step="1000" class={numberInputClass} bind:value={limit.retryMaxMs} />
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				</details>
			{/snippet}
			{#if edited.kindLimits.length > 0}
				{@render section('advanced', 'Advanced', null, advancedBody)}
			{/if}
		</div>
	</div>
</div>

{#if dirty || saved || error}
	<div class="pointer-events-none fixed inset-x-0 bottom-6 z-40 flex justify-center px-4">
		<div
			class="bg-base-300 pointer-events-auto flex items-center gap-4 rounded-2xl border border-white/10 py-2.5 pr-2.5 pl-5 text-sm shadow-2xl"
			role="status"
		>
			{#if error}
				<span class="text-error">{error}</span>
			{:else if saved && !dirty}
				<span class="text-success">Saved</span>
			{:else}
				<span class="text-base-content/80">Unsaved changes</span>
			{/if}
			{#if dirty}
				<button
					type="button"
					class="hover:bg-base-200 rounded-lg px-3 py-1.5 font-medium transition-colors disabled:opacity-50"
					onclick={discard}
					disabled={saving}
				>
					Discard
				</button>
				<button
					type="button"
					class="flex items-center gap-2 rounded-lg bg-white px-4 py-1.5 font-semibold text-black transition-opacity hover:opacity-90 disabled:opacity-50"
					onclick={save}
					disabled={saving}
				>
					{#if saving}<span class="loading loading-spinner loading-xs"></span>{/if}
					Save
				</button>
			{/if}
		</div>
	</div>
{/if}

{#snippet section(id: string, title: string, description: string | null, body: Snippet)}
	<section {id} class="flex scroll-mt-4 flex-col gap-2">
		<div class="px-1">
			<h2 class="text-base font-semibold">{title}</h2>
			{#if description}
				<p class="text-base-content/50 mt-0.5 text-[13px]">{description}</p>
			{/if}
		</div>
		<div class="bg-base-200 divide-base-300/60 flex flex-col divide-y rounded-xl">
			{@render body()}
		</div>
	</section>
{/snippet}

{#snippet row(title: string, description: string, control: Snippet, controlId?: string)}
	<div class="flex items-center justify-between gap-6 px-4 py-3.5">
		<div class="flex min-w-0 flex-col gap-0.5">
			<label class="text-sm font-medium" for={controlId}>{title}</label>
			<p class="text-base-content/50 text-[13px] leading-snug">{description}</p>
		</div>
		<div class="shrink-0">{@render control()}</div>
	</div>
{/snippet}
