<!--
	One installed plugin: name, version, what it does, an on/off switch, its
	settings form, its site login, and buttons to store a page of its catalog.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import Toggle from '$lib/components/Toggle.svelte';
	import SiteIcon from '$lib/components/scene/SiteIcon.svelte';
	import PluginAccount from '$lib/components/PluginAccount.svelte';
	import {
		setPluginEnabled,
		savePluginSettings,
		uninstallPlugin,
		type InstalledPlugin,
		type PluginSettingField
	} from '$lib/plugins';
	import { browse, type Listed } from '$lib/library';
	import { audiosUrl, galleriesUrl, storiesUrl } from '$lib/routes';

	interface Props {
		plugin: InstalledPlugin;
		// Opened by a link to it (e.g. its login): shown open and scrolled to.
		opened?: boolean;
		onchanged: () => void;
	}

	let { plugin, opened = false, onchanged }: Props = $props();

	type BrowseKind = 'VIDEO' | 'GALLERY' | 'AUDIO' | 'STORY';

	// Where a browsed kind lands in the answer, and the page that shows it.
	const BROWSE_TARGETS: Record<BrowseKind, { listed: keyof Listed; page: string }> = {
		VIDEO: { listed: 'videos', page: '/' },
		GALLERY: { listed: 'galleries', page: galleriesUrl() },
		AUDIO: { listed: 'audios', page: audiosUrl() },
		STORY: { listed: 'stories', page: storiesUrl() }
	};

	let open = $state(false);
	let row = $state<HTMLLIElement>();

	$effect(() => {
		if (!opened) return;
		open = true;
		row?.scrollIntoView({ block: 'center' });
	});
	let busy = $state(false);
	let message = $state<string | null>(null);
	let values = $state<Record<string, string>>({});

	let canBrowseVideos = $derived(plugin.capabilities.includes('scene:list'));
	let canBrowseGalleries = $derived(plugin.capabilities.includes('gallery:list'));
	let canBrowseAudio = $derived(plugin.capabilities.includes('audio:list'));
	let canBrowseStories = $derived(plugin.capabilities.includes('story:list'));
	// The site has a login, and this device has none that works.
	let needsLogin = $derived(plugin.login !== null && plugin.account?.status !== 'ok');
	// On/off settings show as switches that save at once; the rest as a form.
	let switches = $derived(plugin.settings.filter((field) => field.type === 'boolean'));
	let textFields = $derived(plugin.settings.filter((field) => field.type !== 'boolean'));

	$effect(() => {
		values = Object.fromEntries(plugin.settingValues.map((entry) => [entry.key, entry.value]));
	});

	/** Turns the plugin on or off. */
	async function toggle(enabled: boolean) {
		await setPluginEnabled(plugin.name, enabled);
		onchanged();
	}

	/** Saves the settings form. */
	async function save(event: SubmitEvent) {
		event.preventDefault();
		await saveValues();
	}

	/** Stores every setting's value as shown. */
	async function saveValues() {
		await savePluginSettings(
			plugin.name,
			Object.entries(values).map(([key, value]) => ({ key, value }))
		);
		message = 'Saved.';
	}

	/** Whether a switch setting is on: its stored value, else its default. */
	function isOn(field: PluginSettingField): boolean {
		const value = values[field.key];
		if (value === undefined || value === '') return field.default === 'true';
		return value === 'true';
	}

	/** Flips a switch setting and saves at once, as a switch promises. */
	async function flip(field: PluginSettingField, on: boolean) {
		values[field.key] = String(on);
		await saveValues();
	}

	/** Stores a page of the plugin's catalog and opens where it shows. */
	async function browseCatalog(kind: BrowseKind) {
		busy = true;
		message = null;
		try {
			const listed = await browse(plugin.name, kind);
			const target = BROWSE_TARGETS[kind];
			message = `Stored ${listed[target.listed].length}.`;
			goto(target.page);
		} catch (failure) {
			message = String(failure);
		} finally {
			busy = false;
		}
	}

	/** Removes the plugin. */
	async function remove() {
		busy = true;
		try {
			await uninstallPlugin(plugin.name);
			onchanged();
		} finally {
			busy = false;
		}
	}
</script>

<li bind:this={row} id={plugin.name} class="bg-base-200 flex flex-col gap-3 rounded-xl p-4">
	<div class="flex items-center gap-3">
		<SiteIcon iconUrl={plugin.iconUrl} class="h-8 w-8 shrink-0" />
		<button
			type="button"
			class="flex min-w-0 flex-1 flex-col text-left"
			onclick={() => (open = !open)}
		>
			<span class="flex items-center gap-2 font-medium">
				{plugin.displayName ?? plugin.name}
				<span class="text-base-content/40 text-xs">v{plugin.version}</span>
				{#if plugin.localBuild}<span class="text-base-content/40 text-xs">dev build</span>{/if}
				<span class:rotate-180={open} class="text-base-content/40 transition-transform"
					><CaretDownIcon size={12} /></span
				>
			</span>
			<span class="text-base-content/50 truncate text-xs">
				{#if !plugin.available}
					Needs FlareSolverr, which isn't set up here ·
				{/if}
				{#if needsLogin}
					Needs a login ·
				{/if}
				{plugin.description ?? plugin.capabilities.join(', ')}
			</span>
		</button>
		<Toggle checked={plugin.enabled} onchange={toggle} ariaLabel={`Enable ${plugin.name}`} />
	</div>

	{#if open}
		<div class="flex flex-col gap-3 pl-11">
			<p class="text-base-content/50 text-xs">{plugin.capabilities.join(' · ')}</p>
			<div class="flex flex-wrap gap-2">
				{#if canBrowseVideos}
					<button
						type="button"
						class="bg-base-300 rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
						disabled={busy || !plugin.enabled}
						onclick={() => browseCatalog('VIDEO')}
					>
						Browse videos
					</button>
				{/if}
				{#if canBrowseGalleries}
					<button
						type="button"
						class="bg-base-300 rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
						disabled={busy || !plugin.enabled}
						onclick={() => browseCatalog('GALLERY')}
					>
						Browse galleries
					</button>
				{/if}
				{#if canBrowseAudio}
					<button
						type="button"
						class="bg-base-300 rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
						disabled={busy || !plugin.enabled}
						onclick={() => browseCatalog('AUDIO')}
					>
						Browse audio
					</button>
				{/if}
				{#if canBrowseStories}
					<button
						type="button"
						class="bg-base-300 rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
						disabled={busy || !plugin.enabled}
						onclick={() => browseCatalog('STORY')}
					>
						Browse stories
					</button>
				{/if}
				{#if !plugin.localBuild}
					<button
						type="button"
						class="text-error rounded-lg px-3 py-1.5 text-sm disabled:opacity-50"
						disabled={busy}
						onclick={remove}>Uninstall</button
					>
				{/if}
			</div>
			{#if switches.length > 0}
				<div class="grid max-w-3xl gap-x-6 gap-y-2 sm:grid-cols-2">
					{#each switches as field (field.key)}
						<div class="flex items-center justify-between gap-3 text-sm">
							<span class="flex min-w-0 flex-col">
								<span class="truncate">{field.label}</span>
								{#if field.description}
									<span class="text-base-content/50 truncate text-xs">{field.description}</span>
								{/if}
							</span>
							<Toggle
								checked={isOn(field)}
								onchange={(on) => flip(field, on)}
								ariaLabel={field.label}
							/>
						</div>
					{/each}
				</div>
			{/if}
			{#if textFields.length > 0}
				<form class="flex max-w-md flex-col gap-2" onsubmit={save}>
					{#each textFields as field (field.key)}
						<label class="flex flex-col gap-1 text-sm">
							<span>{field.label}{field.required ? ' *' : ''}</span>
							<input
								type={field.type === 'password' ? 'password' : 'text'}
								bind:value={values[field.key]}
								placeholder={field.default ?? ''}
								class="bg-base-300 h-9 rounded-lg px-3 outline-none"
							/>
							{#if field.description}<span class="text-base-content/50 text-xs"
									>{field.description}</span
								>{/if}
						</label>
					{/each}
					<button
						type="submit"
						class="bg-primary text-primary-content h-9 self-start rounded-lg px-4 text-sm font-medium"
						>Save</button
					>
				</form>
			{/if}
			{#if plugin.login}
				<PluginAccount
					plugin={plugin.name}
					login={plugin.login}
					account={plugin.account}
					{onchanged}
				/>
			{/if}
			{#if message}<p class="text-base-content/60 text-xs">{message}</p>{/if}
		</div>
	{/if}
</li>
