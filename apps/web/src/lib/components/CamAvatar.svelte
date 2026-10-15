<script lang="ts">
	// Round avatar of a followed cam model: the room thumbnail (else the name's
	// first letter) with a red dot while the model is live.
	import { cacheUrl } from '$lib/img';
	import type { CamModel } from '$lib/cams';

	interface Props {
		camModel: CamModel;
		size?: number;
	}

	let { camModel, size = 24 }: Props = $props();

	let imageUrl = $derived(cacheUrl(camModel.thumbnailUrl));
</script>

<span class="relative shrink-0" style:width="{size}px" style:height="{size}px">
	<span
		class="bg-base-300 flex h-full w-full items-center justify-center overflow-hidden rounded-full font-semibold uppercase"
		style:font-size="{Math.round(size * 0.45)}px"
	>
		{#if imageUrl}
			<img src={imageUrl} alt="" loading="lazy" class="h-full w-full object-cover" />
		{:else}
			{camModel.displayName.charAt(0)}
		{/if}
	</span>
	{#if camModel.online}
		<span
			class="bg-error ring-base-100 absolute right-0 bottom-0 h-2 w-2 rounded-full ring-2"
			title="Live now"
		></span>
	{/if}
</span>
