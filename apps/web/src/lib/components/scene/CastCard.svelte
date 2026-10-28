<script lang="ts">
	// One credited performer on the watch page: photo (initials without one),
	// name, and what the profile knows: age when the video came out, country and
	// how many scenes the library has with them.
	import { cacheUrl } from '$lib/img';
	import { performerUrl } from '$lib/routes';
	import type { PerformerRef } from '$lib/sceneDetail';
	import { ageAt, formatCount, initials } from './format';

	interface Props {
		performer: PerformerRef;
		// The scene's release date, for the performer's age in it.
		sceneDate: string | null;
	}

	let { performer, sceneDate }: Props = $props();

	let photoUrl = $derived(cacheUrl(performer.imagePath));
	let facts = $derived(performerFacts(performer, sceneDate));

	/** The short facts line: age (at release when the date is known), country, scene count. */
	function performerFacts(current: PerformerRef, date: string | null): string[] {
		const parts: string[] = [];
		const ageThen = ageAt(current.birthdate, date);
		if (date && ageThen !== null) {
			parts.push(`${ageThen} at release`);
		} else {
			const ageNow = ageAt(current.birthdate, null);
			if (ageNow !== null) parts.push(`${ageNow} years`);
		}
		if (current.country) parts.push(current.country);
		if (current.sceneCount && current.sceneCount > 1) {
			parts.push(`${formatCount(current.sceneCount)} scenes`);
		}
		return parts;
	}
</script>

<a
	href={performerUrl(performer.id)}
	class="bg-base-300/60 flex min-w-0 items-center gap-3 rounded-xl p-2 pr-3 transition-colors hover:bg-white/10"
>
	<span
		class="bg-base-100 text-base-content/60 flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-full text-sm font-semibold"
	>
		{#if photoUrl}
			<img src={photoUrl} alt="" loading="lazy" class="h-full w-full object-cover object-top" />
		{:else}
			{initials(performer.name)}
		{/if}
	</span>
	<span class="flex min-w-0 flex-col">
		<span class="truncate font-semibold capitalize">{performer.name}</span>
		{#if facts.length > 0}
			<span class="text-base-content/50 truncate text-xs">{facts.join(' · ')}</span>
		{/if}
	</span>
</a>
