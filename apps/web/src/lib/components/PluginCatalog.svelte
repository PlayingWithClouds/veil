<!--
	Plugins published on GitHub: install the ones not installed yet, or any package
	by name. The backend loads an installed plugin within a second; onchange
	lets the page reload its list.
-->
<script lang="ts">
	import { gqlClient } from '$lib/veil';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';

	interface PluginPackage {
		name: string;
		version: string;
		description: string | null;
		installedVersion: string | null;
	}

	let {
		packages,
		onchange,
		onerror
	}: {
		packages: PluginPackage[];
		onchange: (message: string) => void;
		onerror: (message: string) => void;
	} = $props();

	let installing = $state<string | null>(null);
	let packageName = $state('');

	let notInstalled = $derived(packages.filter((found) => found.installedVersion === null));

	/** Installs a published plugin package. */
	async function install(name: string) {
		installing = name;
		try {
			await gqlClient.mutation({ installPlugin: { __args: { packageName: name } } });
			onchange(`Installed ${name}`);
			packageName = '';
		} catch (error) {
			onerror(String(error));
		} finally {
			installing = null;
		}
	}

	/** Installs the package typed into the field. */
	function installTyped(event: SubmitEvent) {
		event.preventDefault();
		install(packageName.trim());
	}
</script>

<section class="border-base-300 bg-base-200 flex flex-col gap-4 rounded-xl border p-5">
	<div>
		<h2 class="text-sm font-semibold">Get plugins</h2>
		<p class="text-base-content/40 text-xs">Plugins published on the Veil GitHub plugin index.</p>
	</div>

	{#if notInstalled.length > 0}
		<ul class="flex flex-col gap-2">
			{#each notInstalled as found (found.name)}
				<li class="flex items-center justify-between gap-3">
					<div class="min-w-0">
						<p class="truncate font-mono text-[13px]">{found.name}</p>
						{#if found.description}
							<p class="text-base-content/50 truncate text-xs">{found.description}</p>
						{/if}
					</div>
					<button
						class="btn btn-sm shrink-0"
						disabled={installing !== null}
						onclick={() => install(found.name)}
					>
						{#if installing === found.name}
							<span class="loading loading-spinner loading-xs"></span>
						{:else}
							<DownloadSimpleIcon size={14} />
						{/if}
						Install
					</button>
				</li>
			{/each}
		</ul>
	{/if}

	<form class="flex gap-2" onsubmit={installTyped}>
		<input
			class="border-base-300 bg-base-100 focus:border-base-content/30 min-w-0 flex-1 rounded-md border px-3 py-2 font-mono text-sm outline-none"
			placeholder="@scope/veil-plugin-name"
			autocapitalize="off"
			spellcheck="false"
			bind:value={packageName}
		/>
		<button type="submit" class="btn btn-primary" disabled={installing !== null || packageName.trim() === ''}>
			Install
		</button>
	</form>
</section>
