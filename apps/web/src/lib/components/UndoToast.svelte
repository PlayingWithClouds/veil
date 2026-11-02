<!--
	The undo toast (stores/undoToast) at the bottom right; above the bottom
	tab bar on phones.
-->
<script lang="ts">
	import { fly } from 'svelte/transition';
	import { undoToast, undoCurrent } from '$lib/stores/undoToast';
</script>

{#if $undoToast}
	<div class="toast max-md:bottom-[var(--bottom-nav-clearance)]" role="status">
		<div class="alert" transition:fly={{ y: 12, duration: 180 }}>
			<span>{$undoToast.message}</span>
			{#if $undoToast.undo}
				<button type="button" class="btn btn-ghost btn-sm" onclick={undoCurrent}>Undo</button>
			{/if}
		</div>
	</div>
{/if}
