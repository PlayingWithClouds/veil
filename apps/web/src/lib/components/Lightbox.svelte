<script lang="ts">
	import XIcon from 'phosphor-svelte/lib/XIcon';

	interface Props {
		src: string;
		alt?: string;
		onclose: () => void;
	}

	let { src, alt = '', onclose }: Props = $props();

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' || event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			event.stopPropagation();
			onclose();
		}
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div
	class="fixed inset-0 z-[100] flex items-center justify-center bg-black/85 p-6"
	role="dialog"
	aria-modal="true"
	aria-label={alt || 'Image'}
>
	<button type="button" class="absolute inset-0 cursor-zoom-out" aria-label="Close" onclick={onclose}
	></button>
	<img
		{src}
		{alt}
		class="relative max-h-full max-w-full rounded-xl object-contain shadow-2xl"
	/>
	<button
		type="button"
		class="absolute top-5 right-5 flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
		aria-label="Close image"
		onclick={onclose}
	>
		<XIcon size={20} />
	</button>
</div>
