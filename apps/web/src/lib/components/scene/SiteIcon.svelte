<script lang="ts">
	// A source site's plugin icon: an image when the plugin ships one, its emoji
	// otherwise, nothing when there is neither.
	import { isImageUrl } from './format';

	interface Props {
		iconUrl: string | null | undefined;
		class?: string;
	}

	let { iconUrl, class: className = 'h-4 w-4' }: Props = $props();
</script>

{#if isImageUrl(iconUrl)}
	<img src={iconUrl} alt="" loading="lazy" class="shrink-0 rounded-sm object-contain {className}" />
{:else if iconUrl}
	<span class="shrink-0 text-xs leading-none" aria-hidden="true">{iconUrl}</span>
{/if}
