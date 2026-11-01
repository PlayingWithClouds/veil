<!-- Desktop side menu: the top-level pages. -->
<script lang="ts">
	import { page } from '$app/state';
	import { navLinks, isNavActive } from '$lib/navigation';
</script>

<aside
	class="no-scrollbar flex w-60 shrink-0 flex-col overflow-y-auto px-3 pb-3"
	aria-label="Main menu"
>
	<nav class="flex flex-col gap-0.5">
		{#each navLinks as link (link.href)}
			{@const active = isNavActive(link, page.url.pathname)}
			<a
				href={link.href}
				title={link.label}
				class="text-base-content/80 hover:bg-base-200 hover:text-base-content sidebar-link-full flex items-center rounded-lg transition-colors"
				class:sidebar-link-active={active}
			>
				<link.icon size={20} weight={active ? 'fill' : 'regular'} />
				<span class="truncate">{link.label}</span>
			</a>
		{/each}
	</nav>
	<div class="mt-auto flex justify-end px-3 pt-4">
		<span class="text-base-content/20 text-[9px] font-medium">v0.1.0</span>
	</div>
</aside>

<style>
	.sidebar-link-full {
		height: 2.5rem;
		gap: 1.25rem;
		padding-inline: 0.75rem;
		font-size: 14px;
	}

	.sidebar-link-active {
		background-color: var(--color-base-200);
		color: var(--color-base-content);
		font-weight: 600;
	}
</style>
