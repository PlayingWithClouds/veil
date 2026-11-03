<!--
	A select you can type into: an input whose list of options narrows as
	you type (options starting with the text first). Picking an option or
	typing anything else sets the value. Arrow keys move through the list,
	Enter picks, Escape closes it. The list is a popover, so it opens above
	modal dialogs and isn't clipped by scrolling containers.
-->
<script lang="ts">
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';

	interface Props {
		value: string | null | undefined;
		options: string[];
		placeholder?: string;
		label?: string;
	}

	let { value = $bindable(), options, placeholder = '', label }: Props = $props();

	const VIEWPORT_MARGIN = 8;
	const listId = `combobox-${Math.random().toString(36).slice(2)}`;

	let input = $state<HTMLInputElement>();
	let list = $state<HTMLElement>();
	let open = $state(false);
	// The text typed since the list opened; empty shows every option.
	let typed = $state('');
	let highlighted = $state(-1);
	let position = $state({ top: 0, left: 0, width: 0, maxHeight: 256 });

	let matches = $derived(filterOptions(options, typed));

	$effect(() => {
		if (!list) return;
		if (open && matches.length > 0) {
			place();
			if (!list.matches(':popover-open')) list.showPopover();
			return;
		}
		if (list.matches(':popover-open')) list.hidePopover();
	});

	$effect(() => {
		if (highlighted < 0 || !list) return;
		list.children[highlighted]?.scrollIntoView({ block: 'nearest' });
	});

	/** The options containing text, those starting with it first. */
	function filterOptions(all: string[], text: string): string[] {
		const needle = text.trim().toLowerCase();
		if (needle === '') return all;
		const starting = all.filter((option) => option.toLowerCase().startsWith(needle));
		const containing = all.filter(
			(option) => !option.toLowerCase().startsWith(needle) && option.toLowerCase().includes(needle)
		);
		return [...starting, ...containing];
	}

	/** Puts the list under the input, or above it when there's more room there. */
	function place() {
		if (!input) return;
		const box = input.getBoundingClientRect();
		const below = window.innerHeight - box.bottom - VIEWPORT_MARGIN;
		const above = box.top - VIEWPORT_MARGIN;
		const maxHeight = Math.min(256, Math.max(below, above) - 4);
		const top = below >= Math.min(256, above) ? box.bottom + 4 : box.top - 4 - maxHeight;
		position = { top, left: box.left, width: box.width, maxHeight };
	}

	/** Opens the list with every option, the current value highlighted. */
	function openList() {
		typed = '';
		open = true;
		highlighted = options.findIndex((option) => option === value);
	}

	/** Sets the value and closes the list. */
	function pick(option: string) {
		value = option;
		open = false;
		highlighted = -1;
	}

	/** Takes typed text as the value and narrows the list to it. */
	function onInput(event: Event) {
		value = (event.currentTarget as HTMLInputElement).value;
		typed = value;
		open = true;
		highlighted = -1;
	}

	/** Arrow keys move through the list, Enter picks, Escape closes it. */
	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();
			if (!open) {
				openList();
				return;
			}
			move(event.key === 'ArrowDown' ? 1 : -1);
			return;
		}
		if (event.key === 'Enter' && open && highlighted >= 0) {
			event.preventDefault();
			pick(matches[highlighted]);
			return;
		}
		if (event.key === 'Escape' && open) {
			// Closes the list, not the dialog around it.
			event.preventDefault();
			event.stopPropagation();
			open = false;
		}
	}

	/** Moves the highlight by step, wrapping around the list. */
	function move(step: number) {
		if (matches.length === 0) return;
		highlighted = (highlighted + step + matches.length) % matches.length;
	}
</script>

<svelte:window onresize={() => open && place()} onscrollcapture={() => open && place()} />

<span class="relative flex items-center">
	<input
		bind:this={input}
		class="input w-full pr-9"
		role="combobox"
		aria-label={label}
		aria-expanded={open}
		aria-controls={listId}
		aria-autocomplete="list"
		aria-activedescendant={highlighted >= 0 ? `${listId}-${highlighted}` : undefined}
		autocomplete="off"
		{placeholder}
		value={value ?? ''}
		oninput={onInput}
		onkeydown={onKeydown}
		onfocus={openList}
		onblur={() => (open = false)}
	/>
	<button
		type="button"
		tabindex="-1"
		aria-label="Show options"
		class="text-base-content/50 hover:text-base-content absolute right-2 flex h-6 w-6 items-center justify-center"
		onmousedown={(event) => event.preventDefault()}
		onclick={() => {
			if (open) {
				open = false;
				return;
			}
			input?.focus();
			openList();
		}}
	>
		<CaretDownIcon size={14} class="transition-transform" />
	</button>
</span>

<ul
	bind:this={list}
	id={listId}
	popover="manual"
	role="listbox"
	aria-label={label}
	class="menu-panel m-0 overflow-y-auto p-1"
	style:transform-origin="top left"
	style:position="fixed"
	style:inset="auto"
	style:top="{position.top}px"
	style:left="{position.left}px"
	style:width="{position.width}px"
	style:max-height="{position.maxHeight}px"
>
	{#each matches as option, index (option)}
		<li
			id="{listId}-{index}"
			role="option"
			aria-selected={option === value}
			class="menu-item cursor-pointer"
			class:menu-item-highlighted={index === highlighted}
			class:font-medium={option === value}
			onmousedown={(event) => {
				// Picks on press, so the input keeps focus; keys are handled on the input.
				event.preventDefault();
				pick(option);
			}}
			onmouseenter={() => (highlighted = index)}
		>
			<span class="truncate">{option}</span>
		</li>
	{/each}
</ul>
