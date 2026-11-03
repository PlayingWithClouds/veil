<!--
	A row of exclusive choices (gender, sort) whose white thumb slides under
	the chosen one instead of jumping.
-->
<script lang="ts" generics="Value extends string">
	interface Props {
		// Names the group for screen readers.
		label: string;
		choices: { value: Value; label: string }[];
		current: Value;
		onchoose: (value: Value) => void;
	}

	let { label, choices, current, onchoose }: Props = $props();

	let buttons = $state<Record<string, HTMLButtonElement>>({});
	let thumbLeft = $state(0);
	let thumbWidth = $state(0);

	// Follows the chosen button, also when fonts load or the row wraps.
	$effect(() => {
		const chosen = buttons[current];
		if (!chosen) return;
		const place = () => {
			thumbLeft = chosen.offsetLeft;
			thumbWidth = chosen.offsetWidth;
		};
		place();
		const observer = new ResizeObserver(place);
		observer.observe(chosen);
		return () => observer.disconnect();
	});
</script>

<div class="segmented" role="group" aria-label={label}>
	{#if thumbWidth > 0}
		<span class="segmented-thumb" style:left="{thumbLeft}px" style:width="{thumbWidth}px"></span>
	{/if}
	{#each choices as choice (choice.value)}
		<button
			type="button"
			class="segmented-option"
			bind:this={buttons[choice.value]}
			aria-pressed={current === choice.value}
			onclick={() => onchoose(choice.value)}
		>
			{choice.label}
		</button>
	{/each}
</div>
