<script lang="ts" module>
	// Shared performer tile: a round avatar with the name and age beneath it.
	export interface PerformerCardData {
		id: string;
		name: string;
		imagePath: string | null;
		// ISO date string; used to derive the displayed age.
		birthdate?: string | null;
	}

	// Whole-year age from an ISO birthdate. Returns null for missing or unparseable
	// dates so the caller can omit the line entirely.
	export function ageFromBirthdate(birthdate: string | null | undefined): number | null {
		if (!birthdate) return null;
		const born = new Date(birthdate);
		if (Number.isNaN(born.getTime())) return null;
		const now = new Date();
		let age = now.getFullYear() - born.getFullYear();
		const monthDiff = now.getMonth() - born.getMonth();
		const beforeBirthday = monthDiff < 0 || (monthDiff === 0 && now.getDate() < born.getDate());
		if (beforeBirthday) age -= 1;
		if (age < 0) return null;
		return age;
	}
</script>

<script lang="ts">
	import { cacheUrl } from '$lib/img';

	interface Props {
		performer: PerformerCardData;
		onclick?: () => void;
	}

	let { performer, onclick }: Props = $props();

	let age = $derived(ageFromBirthdate(performer.birthdate));
	let avatarUrl = $derived(cacheUrl(performer.imagePath));
</script>

<button type="button" class="flex flex-col items-center gap-2 text-center" {onclick}>
	<div class="tv-card aspect-square w-full overflow-hidden rounded-full">
		{#if avatarUrl}
			<img
				src={avatarUrl}
				alt={performer.name}
				loading="lazy"
				class="h-full w-full object-cover"
			/>
		{:else}
			<div class="flex h-full w-full items-center justify-center">
				<span class="text-3xl opacity-20">👤</span>
			</div>
		{/if}
	</div>
	<span class="line-clamp-2 text-xs font-medium">{performer.name}</span>
	{#if age !== null}
		<span class="text-base-content/40 text-xs">{age}</span>
	{/if}
</button>
