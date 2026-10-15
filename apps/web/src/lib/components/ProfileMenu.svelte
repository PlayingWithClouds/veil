<!--
	The top bar's account avatar with its dropdown. There is no login yet, so the
	profile is a guest; the menu holds the personal pages and a disabled sign-in.
-->
<script lang="ts">
	import { page } from '$app/stores';
	import UserIcon from 'phosphor-svelte/lib/UserIcon';
	import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
	import GearSixIcon from 'phosphor-svelte/lib/GearSixIcon';
	import SignInIcon from 'phosphor-svelte/lib/SignInIcon';
	import type { Component } from 'svelte';

	type MenuLink = { label: string; icon: Component; href: string };

	const links: MenuLink[] = [
		{ label: 'Library', icon: FilmStripIcon, href: '/library' },
		{ label: 'Downloads', icon: DownloadSimpleIcon, href: '/library#downloads' },
		{ label: 'History', icon: ClockCounterClockwiseIcon, href: '/history' },
		{ label: 'Settings', icon: GearSixIcon, href: '/settings' }
	];

	let open = $state(false);
	let container = $state<HTMLElement | null>(null);

	// Following a link closes the menu.
	$effect(() => {
		$page.url;
		open = false;
	});

	/** Closes the menu when a click lands outside it. */
	function onWindowClick(event: MouseEvent) {
		if (!open || !container) return;
		if (event.target instanceof Node && container.contains(event.target)) return;
		open = false;
	}

	/** Closes the menu on Escape. */
	function onWindowKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') open = false;
	}
</script>

<svelte:window onclick={onWindowClick} onkeydown={onWindowKeydown} />

<div class="relative" bind:this={container}>
	<button
		type="button"
		onclick={() => (open = !open)}
		aria-label="Account"
		aria-haspopup="menu"
		aria-expanded={open}
		class="bg-base-300 hover:bg-base-content/20 text-base-content/70 flex h-9 w-9 items-center justify-center rounded-full transition-colors"
		class:bg-base-content={open}
	>
		<UserIcon size={20} weight="fill" />
	</button>

	{#if open}
		<div
			role="menu"
			class="bg-base-200 border-base-300 absolute top-full right-0 z-40 mt-2 w-64 rounded-xl border p-2 shadow-xl"
		>
			<div class="flex items-center gap-3 px-2 py-2">
				<span
					class="bg-base-300 text-base-content/70 flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
				>
					<UserIcon size={22} weight="fill" />
				</span>
				<div class="min-w-0">
					<p class="truncate text-sm font-semibold">Guest</p>
					<p class="text-base-content/50 truncate text-xs">Not signed in</p>
				</div>
			</div>
			<hr class="border-base-300 my-1" />
			{#each links as link (link.href)}
				<a
					href={link.href}
					role="menuitem"
					class="hover:bg-base-300 flex h-10 items-center gap-3 rounded-lg px-2 text-sm transition-colors"
				>
					<link.icon size={20} />
					{link.label}
				</a>
			{/each}
			<hr class="border-base-300 my-1" />
			<button
				type="button"
				role="menuitem"
				disabled
				title="Login is coming soon"
				class="text-base-content/40 flex h-10 w-full cursor-not-allowed items-center gap-3 rounded-lg px-2 text-sm"
			>
				<SignInIcon size={20} />
				Sign in
				<span class="bg-base-300 ml-auto rounded px-1.5 py-0.5 text-[10px] font-medium">Soon</span>
			</button>
		</div>
	{/if}
</div>
