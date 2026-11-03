<!--
	A record's tags as chips. Tags only filter: a chip opens home with the
	video grid narrowed to that tag.
-->
<script lang="ts" module>
	/** A tag as chips need it. */
	export interface TagRef {
		id: string;
		name: string;
	}

	/**
	 * Shows a tag in lower case, except words that are acronyms ("POV", "MILF",
	 * "JAV"). Sites disagree on casing, so the stored spelling is whatever arrived
	 * first.
	 */
	export function tagLabel(name: string): string {
		return name
			.replaceAll('-', ' ')
			.split(' ')
			.map((word) => {
				const isAcronym =
					word.length >= 2 && word.length <= 5 && word === word.toUpperCase() && /[A-Z]/.test(word);
				if (isAcronym) return word;
				return word.toLowerCase();
			})
			.join(' ');
	}
</script>

<script lang="ts">
	import { tagFilterUrl } from '$lib/routes';

	let { tags }: { tags: TagRef[] } = $props();
</script>

{#if tags.length > 0}
	<div class="flex flex-wrap items-center gap-1.5">
		{#each tags as tag (tag.id)}
			<a href={tagFilterUrl(tag.id)} class="chip chip-sm" title={tag.name}>
				{tagLabel(tag.name)}
			</a>
		{/each}
	</div>
{/if}
