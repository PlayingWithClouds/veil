<script lang="ts">
	import { onMount } from 'svelte';
	import ShieldCheckIcon from 'phosphor-svelte/lib/ShieldCheckIcon';
	import ShieldWarningIcon from 'phosphor-svelte/lib/ShieldWarningIcon';
	import { gqlClient } from '$lib/veil';

	interface Props {
		// Spell the state out next to the shield instead of only in the tooltip.
		showLabel?: boolean;
	}

	let { showLabel = false }: Props = $props();

	const POLL_INTERVAL_MS = 10000;

	let connected = $state(false);
	let interfaceName = $state<string | null>(null);

	async function refresh() {
		try {
			const data = await gqlClient.query({
				vpnStatus: { connected: true, interface: true }
			});
			connected = data.vpnStatus.connected;
			interfaceName = data.vpnStatus.interface ?? null;
		} catch {
			connected = false;
			interfaceName = null;
		}
	}

	onMount(() => {
		refresh();
		const timer = setInterval(refresh, POLL_INTERVAL_MS);
		return () => clearInterval(timer);
	});

	let label = $derived(
		connected ? `VPN connected${interfaceName ? ` (${interfaceName})` : ''}` : 'VPN disconnected'
	);
</script>

<div class="tooltip tooltip-right flex items-center gap-2" data-tip={label}>
	{#if connected}
		<ShieldCheckIcon size={20} weight="fill" class="text-success" />
	{:else}
		<ShieldWarningIcon size={20} weight="fill" class="text-error" />
	{/if}
	{#if showLabel}
		<span class="text-base-content/70 text-xs font-medium">
			{#if connected}VPN connected{:else}VPN off{/if}
		</span>
	{/if}
</div>
