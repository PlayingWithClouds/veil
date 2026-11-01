<!--
	Plugins: the installed ones (switch, settings, browse their catalog) and
	the published packages to install.
-->
<script lang="ts">
	import PageHeader from '$lib/components/PageHeader.svelte';
	import PluginRow from '$lib/components/PluginRow.svelte';
	import {
		fetchPlugins,
		fetchPluginPackages,
		installPlugin,
		updatePlugins,
		type InstalledPlugin,
		type PluginPackage
	} from '$lib/plugins';

	let installed = $state<InstalledPlugin[]>([]);
	let packages = $state<PluginPackage[]>([]);
	let packagesError = $state<string | null>(null);
	let busy = $state<string | null>(null);
	let message = $state<string | null>(null);

	let sortedInstalled = $derived(
		[...installed].sort((first, second) => first.name.localeCompare(second.name))
	);

	$effect(() => {
		void refresh();
	});

	/** Reloads the installed plugins and the package index. */
	async function refresh() {
		installed = await fetchPlugins();
		try {
			packages = await fetchPluginPackages();
			packagesError = null;
		} catch (failure) {
			packagesError = String(failure);
		}
	}

	/** Installs a package; it loads within a second. */
	async function install(name: string) {
		busy = name;
		try {
			await installPlugin(name);
			// The plugin directory watcher loads it shortly after the files land.
			setTimeout(refresh, 1500);
		} finally {
			busy = null;
		}
	}

	/** Updates every installed plugin with a newer published version. */
	async function updateAll() {
		busy = 'update';
		try {
			const updated = await updatePlugins();
			message =
				updated.length === 0 ? 'Everything is up to date.' : `Updated ${updated.join(', ')}.`;
			setTimeout(refresh, 1500);
		} catch (failure) {
			message = String(failure);
		} finally {
			busy = null;
		}
	}
</script>

<svelte:head><title>Plugins · Veil</title></svelte:head>

<div class="flex max-w-4xl flex-col gap-6">
	<PageHeader title="Plugins" description="Sites Veil reads from">
		{#snippet actions()}
			<button
				type="button"
				class="bg-base-200 h-9 rounded-lg px-3 text-sm disabled:opacity-50"
				disabled={busy !== null}
				onclick={updateAll}
			>
				{busy === 'update' ? 'Checking…' : 'Check for updates'}
			</button>
		{/snippet}
	</PageHeader>
	{#if message}<p class="text-base-content/60 text-sm">{message}</p>{/if}

	<section class="flex flex-col gap-3">
		<h2 class="text-lg font-semibold">Installed</h2>
		<ul class="flex flex-col gap-2">
			{#each sortedInstalled as plugin (plugin.name)}
				<PluginRow {plugin} onchanged={refresh} />
			{/each}
		</ul>
		{#if installed.length === 0}
			<p class="text-base-content/50 text-sm">No plugins installed yet.</p>
		{/if}
	</section>

	<section class="flex flex-col gap-3">
		<h2 class="text-lg font-semibold">Available</h2>
		{#if packagesError}
			<p class="text-base-content/50 text-sm">The plugin index couldn't be read: {packagesError}</p>
		{/if}
		<ul class="flex flex-col gap-2">
			{#each packages as found (found.name)}
				<li class="bg-base-200 flex items-center gap-3 rounded-xl p-4">
					<div class="flex min-w-0 flex-1 flex-col">
						<span class="font-medium"
							>{found.name} <span class="text-base-content/40 text-xs">v{found.version}</span></span
						>
						{#if found.description}<span class="text-base-content/50 truncate text-xs"
								>{found.description}</span
							>{/if}
					</div>
					{#if found.installedVersion === found.version}
						<span class="text-base-content/40 text-sm">Installed</span>
					{:else}
						<button
							type="button"
							class="bg-primary text-primary-content h-9 rounded-lg px-3 text-sm font-medium disabled:opacity-50"
							disabled={busy !== null}
							onclick={() => install(found.name)}
						>
							{found.installedVersion ? 'Update' : 'Install'}
						</button>
					{/if}
				</li>
			{/each}
		</ul>
	</section>
</div>
