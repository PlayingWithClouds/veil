<script lang="ts">
	// A source site's plugin icon: an image when the plugin ships one, its emoji
	// otherwise, nothing when there is neither.

	interface Props {
		iconUrl: string | null | undefined;
		class?: string;
	}

	let { iconUrl, class: className = 'h-4 w-4' }: Props = $props();

	// Some plugins give an emoji instead of an image URL.
	let isImage = $derived(
		Boolean(
			iconUrl &&
				(iconUrl.startsWith('http://') || iconUrl.startsWith('https://') || iconUrl.startsWith('/'))
		)
	);
</script>

{#if isImage}
	<img src={iconUrl} alt="" loading="lazy" class="shrink-0 rounded-sm object-contain {className}" />
{:else if iconUrl}
	<span class="shrink-0 text-xs leading-none" aria-hidden="true">{iconUrl}</span>
{/if}
