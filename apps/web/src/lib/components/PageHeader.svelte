<!--
	The heading every top-level page starts with: title, an optional one-line
	description, and optional actions on the right.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		description?: string | null;
		actions?: Snippet;
	}

	let { title, description = null, actions }: Props = $props();
</script>

<div class="flex items-center justify-between gap-4 pt-2">
	<div class="min-w-0">
		<h1 class="text-2xl font-bold tracking-tight">{title}</h1>
		{#if description}
			<p class="text-base-content/50 text-sm">{description}</p>
		{/if}
	</div>
	{#if actions}
		<div class="flex shrink-0 items-center gap-2">
			{@render actions()}
		</div>
	{/if}
</div>
