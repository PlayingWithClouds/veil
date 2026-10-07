<!--
	Backend choice of the native app: the backend running on the phone (default)
	or another Veil server. Checks a server answers, saves it and reloads, so
	every client (GraphQL, SSE, image proxy) picks it up.
-->
<script lang="ts">
	import { checkServer, clearServerUrl, saveServerUrl, storedServerUrl } from '$lib/server';

	const savedAddress = storedServerUrl();
	let address = $state(savedAddress ?? '');
	let checking = $state(false);
	let error = $state('');

	/** Verifies the backend is reachable, then saves the address and restarts the app on it. */
	async function connect(event: SubmitEvent) {
		event.preventDefault();
		checking = true;
		error = '';
		const reachable = await checkServer(address);
		checking = false;
		if (!reachable) {
			error = 'No Veil backend answered at this address.';
			return;
		}
		saveServerUrl(address);
		window.location.href = '/';
	}

	/** Switches back to the backend on this phone. */
	function useThisDevice() {
		clearServerUrl();
		window.location.href = '/';
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center justify-between gap-3">
		<p class="text-base-content/70 text-sm">
			{#if savedAddress}
				Connected to <span class="font-mono">{savedAddress}</span>
			{:else}
				Using the library on this device
			{/if}
		</p>
		{#if savedAddress}
			<button type="button" class="btn btn-sm" onclick={useThisDevice}>Use this device</button>
		{/if}
	</div>

	<form class="flex flex-col gap-2" onsubmit={connect}>
		<label class="text-base-content/80 text-[13px] font-medium" for="server-address"
			>Other server</label
		>
		<div class="flex gap-2">
			<input
				id="server-address"
				type="url"
				inputmode="url"
				autocapitalize="off"
				autocomplete="off"
				spellcheck="false"
				placeholder="http://192.168.1.10:8080"
				bind:value={address}
				class="border-base-300 bg-base-100 focus:border-base-content/30 min-w-0 flex-1 rounded-md border px-3 py-2 text-sm transition-colors outline-none"
			/>
			<button type="submit" class="btn btn-primary" disabled={checking || address.trim() === ''}>
				{#if checking}
					<span class="loading loading-spinner loading-sm"></span>
				{:else}
					Connect
				{/if}
			</button>
		</div>
		{#if error}
			<p class="text-error text-xs">{error}</p>
		{/if}
	</form>
</div>
