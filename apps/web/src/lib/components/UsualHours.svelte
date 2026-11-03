<!--
	When a cam room is usually on: a week of hours, Monday first, in local
	time, each cell as strong as the share of weeks the room was on then.
-->
<script lang="ts">
	import { usualHours, type CamShowRow } from '$lib/cams';

	interface Props {
		shows: CamShowRow[];
	}

	let { shows }: Props = $props();

	// Weekdays as Date.getDay counts them (0 = Sunday), in the order rows show.
	const WEEK = [1, 2, 3, 4, 5, 6, 0];
	const HOUR_LABELS = [0, 6, 12, 18];
	// A cell this faint still shows the room was there once.
	const MIN_OPACITY = 0.15;

	const dayName = new Intl.DateTimeFormat(undefined, { weekday: 'short' });
	const hourName = new Intl.DateTimeFormat(undefined, { hour: 'numeric' });

	let hours = $derived(usualHours(shows));

	/** The short name of a weekday, by Date.getDay. */
	function weekdayLabel(day: number): string {
		// 2023-01-01 was a Sunday.
		return dayName.format(new Date(2023, 0, 1 + day));
	}

	/** An hour of the day in the locale's style ("6 PM", "18"). */
	function hourLabel(hour: number): string {
		return hourName.format(new Date(2023, 0, 1, hour));
	}

	/** How strong a cell is drawn; 0 when the room was never on then. */
	function cellOpacity(share: number): number {
		if (share === 0) return 0;
		return MIN_OPACITY + share * (1 - MIN_OPACITY);
	}
</script>

<div class="flex flex-col gap-1 text-xs">
	{#each WEEK as day (day)}
		<div class="flex items-center gap-2">
			<span class="text-base-content/50 w-8 shrink-0">{weekdayLabel(day)}</span>
			<div class="grid flex-1 grid-cols-24 gap-px">
				{#each hours[day] as share, hour (hour)}
					<span
						class="bg-base-300 relative h-4 overflow-hidden rounded-sm"
						title="{weekdayLabel(day)} {hourLabel(hour)}: {Math.round(share * 100)}% of weeks"
					>
						<span class="bg-primary absolute inset-0" style:opacity={cellOpacity(share)}></span>
					</span>
				{/each}
			</div>
		</div>
	{/each}
	<div class="flex items-center gap-2">
		<span class="w-8 shrink-0"></span>
		<div class="text-base-content/50 grid flex-1 grid-cols-4">
			{#each HOUR_LABELS as hour (hour)}
				<span>{hourLabel(hour)}</span>
			{/each}
		</div>
	</div>
</div>
