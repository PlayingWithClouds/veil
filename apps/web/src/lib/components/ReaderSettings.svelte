<!--
	The reader's display settings: theme, typeface, text size, line spacing
	and column width. Changes apply at once and are kept for every story.
-->
<script lang="ts">
	import MinusIcon from 'phosphor-svelte/lib/MinusIcon';
	import PlusIcon from 'phosphor-svelte/lib/PlusIcon';
	import {
		FONT_SIZE_RANGE,
		LINE_HEIGHT_RANGE,
		READER_FONTS,
		READER_THEMES,
		READER_WIDTHS,
		readerSettings
	} from '$lib/stores/readerSettings';

	const LINE_HEIGHT_STEP = 0.1;

	/** Moves the text size by step px, within its range. */
	function changeFontSize(step: number) {
		readerSettings.update((settings) => ({
			...settings,
			fontSize: Math.min(
				Math.max(settings.fontSize + step, FONT_SIZE_RANGE.min),
				FONT_SIZE_RANGE.max
			)
		}));
	}

	/** Moves the line spacing by one step, within its range. */
	function changeLineHeight(direction: number) {
		readerSettings.update((settings) => {
			const next = settings.lineHeight + direction * LINE_HEIGHT_STEP;
			const bounded = Math.min(Math.max(next, LINE_HEIGHT_RANGE.min), LINE_HEIGHT_RANGE.max);
			return { ...settings, lineHeight: Math.round(bounded * 10) / 10 };
		});
	}
</script>

<div class="flex flex-col gap-5 text-sm" style:font-family="var(--font-sans)">
	<section class="flex flex-col gap-2">
		<h3 class="text-xs font-medium tracking-wide uppercase opacity-60">Theme</h3>
		<div class="grid grid-cols-4 gap-2">
			{#each READER_THEMES as theme (theme.value)}
				<button
					type="button"
					data-reader-theme={theme.value}
					class="reader flex h-14 flex-col items-center justify-center gap-0.5 rounded-xl border-2"
					style:border-color={$readerSettings.theme === theme.value
						? 'var(--reader-accent)'
						: 'var(--reader-line)'}
					onclick={() => readerSettings.update((settings) => ({ ...settings, theme: theme.value }))}
					aria-pressed={$readerSettings.theme === theme.value}
				>
					<span class="text-base font-semibold" style:font-family="Georgia, serif">Aa</span>
					<span class="text-[10px]">{theme.label}</span>
				</button>
			{/each}
		</div>
	</section>

	<section class="flex flex-col gap-2">
		<h3 class="text-xs font-medium tracking-wide uppercase opacity-60">Typeface</h3>
		<div class="grid grid-cols-5 gap-2">
			{#each READER_FONTS as font (font.value)}
				<button
					type="button"
					class="flex h-14 flex-col items-center justify-center gap-0.5 rounded-xl border-2"
					style:border-color={$readerSettings.font === font.value
						? 'var(--reader-accent)'
						: 'var(--reader-line)'}
					style:background-color="var(--reader-surface)"
					onclick={() => readerSettings.update((settings) => ({ ...settings, font: font.value }))}
					aria-pressed={$readerSettings.font === font.value}
				>
					<span class="text-lg" style:font-family={font.stack}>Aa</span>
					<span class="text-[10px]">{font.label}</span>
				</button>
			{/each}
		</div>
	</section>

	<section class="flex items-center justify-between gap-3">
		<h3 class="text-xs font-medium tracking-wide uppercase opacity-60">Text size</h3>
		<div class="flex items-center gap-2">
			<button
				type="button"
				class="flex h-9 w-9 items-center justify-center rounded-lg disabled:opacity-40"
				style:background-color="var(--reader-surface)"
				disabled={$readerSettings.fontSize <= FONT_SIZE_RANGE.min}
				onclick={() => changeFontSize(-1)}
				aria-label="Smaller text"><MinusIcon size={14} /></button
			>
			<span class="w-10 text-center tabular-nums">{$readerSettings.fontSize}</span>
			<button
				type="button"
				class="flex h-9 w-9 items-center justify-center rounded-lg disabled:opacity-40"
				style:background-color="var(--reader-surface)"
				disabled={$readerSettings.fontSize >= FONT_SIZE_RANGE.max}
				onclick={() => changeFontSize(1)}
				aria-label="Larger text"><PlusIcon size={14} /></button
			>
		</div>
	</section>

	<section class="flex items-center justify-between gap-3">
		<h3 class="text-xs font-medium tracking-wide uppercase opacity-60">Line spacing</h3>
		<div class="flex items-center gap-2">
			<button
				type="button"
				class="flex h-9 w-9 items-center justify-center rounded-lg disabled:opacity-40"
				style:background-color="var(--reader-surface)"
				disabled={$readerSettings.lineHeight <= LINE_HEIGHT_RANGE.min}
				onclick={() => changeLineHeight(-1)}
				aria-label="Tighter lines"><MinusIcon size={14} /></button
			>
			<span class="w-10 text-center tabular-nums">{$readerSettings.lineHeight.toFixed(1)}</span>
			<button
				type="button"
				class="flex h-9 w-9 items-center justify-center rounded-lg disabled:opacity-40"
				style:background-color="var(--reader-surface)"
				disabled={$readerSettings.lineHeight >= LINE_HEIGHT_RANGE.max}
				onclick={() => changeLineHeight(1)}
				aria-label="Looser lines"><PlusIcon size={14} /></button
			>
		</div>
	</section>

	<section class="flex flex-col gap-2">
		<h3 class="text-xs font-medium tracking-wide uppercase opacity-60">Width</h3>
		<div class="grid grid-cols-3 gap-2">
			{#each READER_WIDTHS as width (width.value)}
				<button
					type="button"
					class="h-9 rounded-lg border-2"
					style:border-color={$readerSettings.width === width.value
						? 'var(--reader-accent)'
						: 'var(--reader-line)'}
					style:background-color="var(--reader-surface)"
					onclick={() => readerSettings.update((settings) => ({ ...settings, width: width.value }))}
					aria-pressed={$readerSettings.width === width.value}>{width.label}</button
				>
			{/each}
		</div>
	</section>
</div>
