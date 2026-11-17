<script lang="ts">
	import type { Project } from '@atlas/contracts';
	import VeilBrowser from './VeilBrowser.svelte';

	let { open }: { project: Project; open: (params: Record<string, unknown>) => Promise<void> } = $props();
</script>

<VeilBrowser kind="scene" label="scenes" {open} />
