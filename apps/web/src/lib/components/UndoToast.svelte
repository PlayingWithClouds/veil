<!--
	The undo toast (stores/undoToast) at the bottom right; above the bottom
	tab bar on phones.
-->
<script lang="ts">
	import { fly } from 'svelte/transition';
	import { undoToast, undoCurrent, settleCurrent } from '$lib/stores/undoToast';
</script>

{#if $undoToast}
	<div class="toast max-md:bottom-[var(--bottom-nav-clearance)]" role="status">
		<div class="alert" transition:fly={{ y: 12, duration: 180 }}>
			<span>{$undoToast.message}</span>
			{#if $undoToast.link}
				<a class="btn btn-ghost btn-sm" href={$undoToast.link.href} onclick={settleCurrent}>
					{$undoToast.link.label}
				</a>
			{/if}
			{#if $undoToast.undo}
				<button type="button" class="btn btn-ghost btn-sm" onclick={undoCurrent}>Undo</button>
			{/if}
		</div>
	</div>
{/if}
