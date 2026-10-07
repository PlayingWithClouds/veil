<script lang="ts">
	import { gqlClient } from '$lib/veil';
	import PlugsIcon from 'phosphor-svelte/lib/PlugsIcon';
	import GearSixIcon from 'phosphor-svelte/lib/GearSixIcon';
	import WarningIcon from 'phosphor-svelte/lib/WarningIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import Toggle from '$lib/components/Toggle.svelte';
	import PluginCatalog from '$lib/components/PluginCatalog.svelte';
	import { invalidateAll } from '$app/navigation';

	let { data } = $props();

	type Plugin = (typeof data.plugins)[number];

	/** The backend loads a changed plugin folder within a second of the change. */
	const PLUGIN_RELOAD_DELAY_MS = 1200;

	// Follows the page data (reloaded after installs and updates); local edits
	// such as toggling override it until then.
	let plugins = $derived<Plugin[]>(data.plugins);
	let checkingUpdates = $state(false);
	let uninstalling = $state<string | null>(null);
	let toggling = $state<string | null>(null);
	let scrapeTarget = $state('');
	let scrapePlugin = $state('');
	let scraping = $state(false);
	let toast = $state<{ msg: string; ok: boolean } | null>(null);

	// per-plugin settings editing state: pluginName → { key → value }
	let editingSettings = $state<string | null>(null);
	let settingsDraft = $state<Record<string, string>>({});
	let savingSettings = $state(false);

	// Soft context tones for capability pills; entity-scoped capabilities
	// ("scene:find", "tag:list") share a tone per action suffix.
	const capTones: Record<string, string> = {
		scrape: 'bg-[#60a5fa]/15 text-[#60a5fa]',
		enrich: 'bg-[#4ade80]/15 text-[#4ade80]',
		resolve: 'bg-[#fbbf24]/15 text-[#fbbf24]',
		find: 'bg-[#f472b6]/15 text-[#f472b6]',
		list: 'bg-[#a78bfa]/15 text-[#a78bfa]'
	};

	function capTone(cap: string): string {
		const base = cap.includes(':') ? cap.split(':')[1] : cap;
		return capTones[base] ?? 'bg-base-300 text-base-content/60';
	}

	function showToast(msg: string, ok: boolean) {
		toast = { msg, ok };
		setTimeout(() => (toast = null), 3000);
	}

	async function toggle(plugin: Plugin) {
		toggling = plugin.name;
		try {
			await gqlClient.mutation({
				togglePlugin: { __args: { name: plugin.name, enabled: !plugin.enabled } }
			});
			plugins = plugins.map((p) => (p.name === plugin.name ? { ...p, enabled: !p.enabled } : p));
		} catch (e) {
			showToast(String(e), false);
		} finally {
			toggling = null;
		}
	}

	async function scrape() {
		if (!scrapeTarget || !scrapePlugin) return;
		scraping = true;
		try {
			const d = await gqlClient.mutation({
				scrape: { __args: { pluginName: scrapePlugin, url: scrapeTarget }, success: true, error: true }
			});
			if (d.scrape.success) {
				showToast('Scrape job queued', true);
				scrapeTarget = '';
			} else {
				showToast(d.scrape.error ?? 'Error', false);
			}
		} catch (e) {
			showToast(String(e), false);
		} finally {
			scraping = false;
		}
	}

	function openSettings(plugin: Plugin) {
		editingSettings = plugin.name;
		// seed draft with existing values, falling back to declared defaults
		const vals: Record<string, string> = {};
		for (const field of plugin.settings) {
			const existing = plugin.settingValues.find((v) => v.key === field.key);
			vals[field.key] = existing?.value ?? field.default ?? '';
		}
		settingsDraft = vals;
	}

	function closeSettings() {
		editingSettings = null;
		settingsDraft = {};
	}

	async function saveSettings(plugin: Plugin) {
		savingSettings = true;
		try {
			const values = Object.entries(settingsDraft).map(([key, value]) => ({ key, value }));
			await gqlClient.mutation({
				updatePluginSettings: { __args: { pluginName: plugin.name, values } }
			});
			// optimistically update local settingValues
			plugins = plugins.map((p) => (p.name === plugin.name ? { ...p, settingValues: values } : p));
			showToast('Settings saved', true);
			closeSettings();
		} catch (e) {
			showToast(String(e), false);
		} finally {
			savingSettings = false;
		}
	}

	/** Shows a message, then reloads the list once the backend has picked up the change. */
	async function reloadAfterChange(message: string) {
		showToast(message, true);
		await new Promise((resolve) => setTimeout(resolve, PLUGIN_RELOAD_DELAY_MS));
		await invalidateAll();
	}

	/** Installs newer npm versions of the installed plugins. */
	async function checkUpdates() {
		checkingUpdates = true;
		try {
			const result = await gqlClient.mutation({ updatePlugins: true });
			if (result.updatePlugins.length === 0) {
				showToast('All plugins are up to date', true);
				return;
			}
			await reloadAfterChange(`Updated ${result.updatePlugins.join(', ')}`);
		} catch (e) {
			showToast(String(e), false);
		} finally {
			checkingUpdates = false;
		}
	}

	/** Removes a plugin after confirmation. */
	async function uninstall(plugin: Plugin) {
		if (!confirm(`Uninstall ${plugin.displayName ?? plugin.name}?`)) return;
		uninstalling = plugin.name;
		try {
			await gqlClient.mutation({ uninstallPlugin: { __args: { name: plugin.name } } });
			await reloadAfterChange(`Uninstalled ${plugin.displayName ?? plugin.name}`);
		} catch (e) {
			showToast(String(e), false);
		} finally {
			uninstalling = null;
		}
	}

	function hasRequiredMissing(plugin: Plugin): boolean {
		return plugin.settings.some((f) => {
			if (!f.required) return false;
			const val = plugin.settingValues.find((v) => v.key === f.key)?.value ?? '';
			return val.trim() === '';
		});
	}
</script>

{#if toast}
	<div class="fixed right-4 top-4 z-[70]">
		<div
			class="rounded-lg border px-4 py-2.5 text-sm shadow-lg {toast.ok
				? 'border-[#4ade80]/30 bg-[#4ade80]/10 text-[#4ade80]'
				: 'border-[#f87171]/30 bg-[#f87171]/10 text-[#f87171]'}"
		>
			{toast.msg}
		</div>
	</div>
{/if}

<!-- Settings modal -->
{#if editingSettings}
	{@const plugin = plugins.find((p) => p.name === editingSettings)!}
	<div class="fixed inset-0 z-[60] flex items-center justify-center p-4">
		<button type="button" class="absolute inset-0 bg-black/50" aria-label="Close" onclick={closeSettings}
		></button>
		<div
			class="border-base-300 bg-base-200 relative z-10 w-full max-w-md rounded-xl border p-5 shadow-2xl"
		>
			<div class="mb-4 flex items-center justify-between">
				<div class="flex items-center gap-3">
					{#if plugin.iconUrl}
						<img
							src={plugin.iconUrl}
							alt=""
							class="h-7 w-7 rounded object-contain"
							onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
						/>
					{/if}
					<h3 class="text-base font-semibold">{plugin.displayName} settings</h3>
				</div>
				<button
					type="button"
					class="text-base-content/40 hover:text-base-content flex h-7 w-7 items-center justify-center rounded-md transition-colors"
					onclick={closeSettings}
					aria-label="Close"
				>
					<XIcon size={16} />
				</button>
			</div>

			<div class="flex flex-col gap-4">
				{#each plugin.settings as field}
					<div class="flex flex-col gap-1.5">
						<label class="text-base-content/80 text-[13px] font-medium" for="setting-{field.key}">
							{field.label}
							{#if field.required}<span class="text-[#f87171]">*</span>{/if}
						</label>
						{#if field.type === 'boolean'}
							<Toggle
								checked={settingsDraft[field.key] === 'true'}
								ariaLabel={field.label}
								onchange={(value) => (settingsDraft[field.key] = value ? 'true' : 'false')}
							/>
						{:else}
							<input
								id="setting-{field.key}"
								type={field.type === 'password' ? 'password' : 'text'}
								class="border-base-300 bg-base-100 focus:border-base-content/30 w-full rounded-md border px-3 py-2 text-sm outline-none transition-colors"
								placeholder={field.default ?? ''}
								bind:value={settingsDraft[field.key]}
							/>
						{/if}
						{#if field.description}
							<p class="text-base-content/40 text-xs">{field.description}</p>
						{/if}
					</div>
				{/each}
			</div>

			<div class="mt-5 flex justify-end gap-2">
				<button
					class="text-base-content/70 hover:bg-base-300 rounded-md px-3 py-2 text-sm font-medium transition-colors"
					onclick={closeSettings}
					disabled={savingSettings}
				>
					Cancel
				</button>
				<button
					class="flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-semibold text-black transition-opacity hover:opacity-90 disabled:opacity-50"
					onclick={() => saveSettings(plugin)}
					disabled={savingSettings}
				>
					{#if savingSettings}<span class="loading loading-spinner loading-xs"></span>{/if}
					Save
				</button>
			</div>
		</div>
	</div>
{/if}

<div class="flex w-full flex-col gap-8 py-2">
	<div class="flex items-center gap-3">
		<span class="text-base-content/70"><PlugsIcon size={26} /></span>
		<div class="flex-1">
			<h1 class="text-2xl font-bold tracking-tight">Plugins</h1>
			<p class="text-base-content/40 text-sm">{plugins.length} loaded</p>
		</div>
		<button class="btn btn-sm" onclick={checkUpdates} disabled={checkingUpdates}>
			{#if checkingUpdates}
				<span class="loading loading-spinner loading-xs"></span>
			{:else}
				<ArrowClockwiseIcon size={14} />
			{/if}
			Check for updates
		</button>
	</div>

	{#if plugins.length === 0}
		<div
			class="border-base-300 flex flex-col items-center gap-3 rounded-xl border border-dashed py-24 text-center"
		>
			<span class="text-base-content/20"><PlugsIcon size={48} /></span>
			<p class="text-base-content/50 text-sm">No plugins loaded</p>
			<p class="text-base-content/30 text-xs">Install plugins from npm below</p>
		</div>
	{:else}
		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
			{#each plugins as plugin}
				<div
					class="border-base-300 bg-base-200 flex flex-col gap-3 rounded-xl border p-4 transition-opacity"
					class:opacity-55={!plugin.enabled || !plugin.available}
				>
					<!-- Header -->
					<div class="flex items-start justify-between gap-3">
						<div class="flex min-w-0 items-center gap-3">
							{#if plugin.iconUrl}
								<img
									src={plugin.iconUrl}
									alt=""
									class="h-9 w-9 shrink-0 rounded-lg object-contain"
									onerror={(e) => ((e.currentTarget as HTMLImageElement).style.display = 'none')}
								/>
							{:else}
								<span
									class="bg-base-300 text-base-content/50 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg"
								>
									<PlugsIcon size={18} />
								</span>
							{/if}
							<div class="min-w-0">
								<h2 class="truncate text-sm font-semibold">{plugin.displayName}</h2>
								<p class="text-base-content/40 font-mono text-xs">
									v{plugin.version}{#if plugin.localBuild}&nbsp;· local build{/if}
								</p>
							</div>
						</div>
						<div class="mt-0.5">
							<Toggle
								checked={plugin.enabled}
								disabled={toggling === plugin.name}
								ariaLabel="Enable {plugin.displayName}"
								onchange={() => toggle(plugin)}
							/>
						</div>
					</div>

					{#if !plugin.available}
						<p class="flex items-center gap-1.5 text-xs font-medium text-[#f87171]">
							<WarningIcon size={14} weight="fill" /> Unavailable: no FlareSolverr configured
						</p>
					{/if}

					{#if plugin.description}
						<p class="text-base-content/55 line-clamp-2 text-[13px] leading-relaxed">
							{plugin.description}
						</p>
					{/if}

					<!-- Capabilities -->
					<div class="flex flex-wrap gap-1.5">
						{#if plugin.requiresSolver}
							<span
								class="rounded-full bg-[#fbbf24]/15 px-2 py-0.5 text-[11px] font-medium text-[#fbbf24]"
								title="Site is behind a Cloudflare challenge and is fetched through FlareSolverr"
							>
								Needs FlareSolverr
							</span>
						{/if}
						{#each plugin.capabilities as cap}
							<span class="rounded-full px-2 py-0.5 text-[11px] font-medium {capTone(cap)}">
								{cap}
							</span>
						{/each}
					</div>

					<!-- Actions -->
					<div class="mt-auto flex gap-2 pt-1">
						{#if plugin.settings.length > 0}
							<button
								class="flex flex-1 items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-[13px] font-medium transition-colors {hasRequiredMissing(
									plugin
								)
									? 'bg-[#fbbf24]/15 text-[#fbbf24] hover:bg-[#fbbf24]/25'
									: 'text-base-content/70 hover:bg-base-300'}"
								onclick={() => openSettings(plugin)}
							>
								{#if hasRequiredMissing(plugin)}
									<WarningIcon size={14} weight="fill" /> Configure
								{:else}
									<GearSixIcon size={14} /> Settings
								{/if}
							</button>
						{/if}
						<button
							class="text-base-content/50 hover:bg-base-300 hover:text-base-content ml-auto flex items-center gap-1.5 rounded-md px-3 py-1.5 text-[13px] font-medium transition-colors"
							disabled={uninstalling === plugin.name}
							onclick={() => uninstall(plugin)}
						>
							<TrashIcon size={14} /> Uninstall
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<div class="max-w-lg">
		<PluginCatalog
			packages={data.packages}
			onchange={reloadAfterChange}
			onerror={(message) => showToast(message, false)}
		/>
	</div>

	<!-- Manual scrape -->
	<div class="border-base-300 bg-base-200 flex max-w-lg flex-col gap-4 rounded-xl border p-5">
		<div>
			<h2 class="text-sm font-semibold">Scrape a URL</h2>
			<p class="text-base-content/40 text-xs">Run a one-off scrape with a specific plugin.</p>
		</div>
		<fieldset class="flex flex-col gap-3" disabled={scraping}>
			<label class="flex flex-col gap-1.5">
				<span class="text-base-content/70 text-[13px] font-medium">Plugin</span>
				<select
					class="border-base-300 bg-base-100 focus:border-base-content/30 rounded-md border px-3 py-2 text-sm outline-none"
					bind:value={scrapePlugin}
				>
					<option value="">— select plugin —</option>
					{#each plugins.filter((p) => p.enabled && (p.capabilities.includes('scrape') || p.capabilities.includes('nsfw:scrape'))) as p}
						<option value={p.name}>{p.displayName}</option>
					{/each}
				</select>
			</label>
			<label class="flex flex-col gap-1.5">
				<span class="text-base-content/70 text-[13px] font-medium">URL</span>
				<input
					class="border-base-300 bg-base-100 focus:border-base-content/30 rounded-md border px-3 py-2 text-sm outline-none"
					placeholder="https://s.to/serie/dark"
					bind:value={scrapeTarget}
				/>
			</label>
		</fieldset>
		<button
			class="flex items-center justify-center gap-2 self-start rounded-md bg-white px-4 py-2 text-sm font-semibold text-black transition-opacity hover:opacity-90 disabled:opacity-50"
			onclick={scrape}
			disabled={scraping || !scrapeTarget || !scrapePlugin}
		>
			{#if scraping}<span class="loading loading-spinner loading-xs"></span>{/if}
			Scrape
		</button>
	</div>
</div>
