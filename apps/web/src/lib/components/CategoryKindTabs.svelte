<!--
	The kinds of media a categories page can show (videos, galleries, audio,
	stories), as links that switch the page's ?kind=.
-->
<script lang="ts">
	import { CATEGORY_KINDS, type CategoryKind } from '$lib/categories';

	interface Props {
		current: CategoryKind;
		hrefOf: (kind: CategoryKind) => string;
	}

	let { current, hrefOf }: Props = $props();
</script>

<nav class="bg-base-200 flex gap-1 self-start rounded-full p-1" aria-label="Kind of media">
	{#each CATEGORY_KINDS as entry (entry.kind)}
		<a
			href={hrefOf(entry)}
			class="rounded-full px-4 py-1.5 text-sm font-medium transition-colors"
			class:bg-base-100={entry.kind === current.kind}
			class:shadow-sm={entry.kind === current.kind}
			class:opacity-60={entry.kind !== current.kind}
			aria-current={entry.kind === current.kind ? 'page' : undefined}
		>
			{entry.label}
		</a>
	{/each}
</nav>
