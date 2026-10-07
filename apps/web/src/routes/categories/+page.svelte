<script lang="ts">
	import { goto } from '$app/navigation';
	import { NSFW_CATEGORIES } from '$lib/nsfw';
</script>

<div class="flex flex-col gap-6">
	<h1 class="text-3xl font-bold">Categories</h1>

	<div class="flex flex-wrap gap-2">
		{#each NSFW_CATEGORIES as category (category.slug)}
			<button
				type="button"
				class="bg-base-200 hover:bg-base-300 rounded-lg px-4 py-2 text-sm font-medium transition-colors"
				onclick={() => goto(`/nsfw/${category.slug}`)}
			>
				{category.label}
			</button>
		{/each}
	</div>
</div>
