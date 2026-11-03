<!--
	Desktop side menu: the top-level pages in sections (Watch, Library, You),
	the current one highlighted. Cams shows how many of the user's rooms
	(followed or auto-recorded) are live now. Starts below the see-through
	top bar.
-->
<script lang="ts">
	import { page } from '$app/state';
	import { camsUrl } from '$lib/routes';
	import { fetchYourRooms } from '$lib/cams';
	import { NAV_GROUPS, navLinks, isNavActive, type NavGroup } from '$lib/navigation';

	// The sites' live state changes; the count follows it.
	const LIVE_REFRESH_MS = 60_000;

	let liveRooms = $state(0);

	$effect(() => {
		void refreshLiveRooms();
		const timer = setInterval(refreshLiveRooms, LIVE_REFRESH_MS);
		return () => clearInterval(timer);
	});

	/** Counts the user's rooms live now; a failed check keeps the last count. */
	async function refreshLiveRooms() {
		if (document.hidden) return;
		try {
			const rooms = await fetchYourRooms();
			liveRooms = rooms.filter((room) => room.live).length;
		} catch (failure) {
			console.warn('live room count failed', failure);
		}
	}

	/** The links listed under one section. */
	function linksOf(group: NavGroup) {
		return navLinks.filter((link) => link.group === group);
	}
</script>

<aside
	class="no-scrollbar flex w-60 shrink-0 flex-col overflow-y-auto px-3 pb-3"
	style:padding-top="var(--desktop-header-height)"
	aria-label="Main menu"
>
	<nav class="flex flex-col gap-5 pt-1">
		{#each NAV_GROUPS as group (group)}
			<div class="flex flex-col gap-0.5">
				<span
					class="text-base-content/35 px-3 pb-1 text-[10px] font-semibold tracking-wider uppercase"
				>
					{group}
				</span>
				{#each linksOf(group) as link (link.href)}
					{@const active = isNavActive(link, page.url.pathname)}
					<a
						href={link.href}
						title={link.label}
						class="sidebar-link text-base-content/70 flex items-center rounded-lg"
						class:sidebar-link-active={active}
						aria-current={active ? 'page' : undefined}
					>
						<link.icon size={20} weight={active ? 'fill' : 'regular'} />
						<span class="flex-1 truncate">{link.label}</span>
						{#if link.href === camsUrl() && liveRooms > 0}
							<span
								class="text-base-content/60 flex items-center gap-1.5 text-xs font-medium tabular-nums"
								title="{liveRooms} of your rooms live"
							>
								<span class="live-dot" style:width="6px" style:height="6px"></span>
								{liveRooms}
							</span>
						{/if}
					</a>
				{/each}
			</div>
		{/each}
	</nav>
	<div class="mt-auto flex justify-end px-3 pt-4">
		<span class="text-base-content/20 text-[9px] font-medium">v0.1.0</span>
	</div>
</aside>

<style>
	.sidebar-link {
		height: 2.5rem;
		gap: 1.25rem;
		padding-inline: 0.75rem;
		font-size: 14px;
		transition:
			background-color 150ms,
			color 150ms;
	}

	.sidebar-link:hover {
		background-color: color-mix(in oklab, var(--color-base-content) 5%, transparent);
		color: var(--color-base-content);
	}

	.sidebar-link-active,
	.sidebar-link-active:hover {
		background-color: color-mix(in oklab, var(--color-base-content) 9%, transparent);
		color: var(--color-base-content);
		font-weight: 600;
	}
</style>
