<!--
	The bell in the top bar. Its badge counts unread notifications; a ring
	spins while performers are being collected. The panel lists what's being
	collected now (checked every few seconds while anything is), then every
	notification until it's cleared, each linking to its record. Opening the
	panel marks them read; they stay listed until cleared, one by one or all.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import {
		clearAllNotifications,
		clearNotification,
		fetchCollecting,
		fetchUnreadCount,
		fetchNotifications,
		markAllRead,
		type CollectingRow,
		type NotificationRow
	} from '$lib/notifications';
	import { siteDisplayName } from '$lib/plugins';
	import { sites } from '$lib/stores/sites';
	import { performerUrl, recordUrl } from '$lib/routes';
	import { pushOverlay } from '$lib/stores/overlayStack';

	const POLL_MS = 60_000;
	// How often collections are checked while any is running.
	const COLLECTING_POLL_MS = 4000;

	let unread = $state(0);
	let open = $state(false);
	let notifications = $state<NotificationRow[]>([]);
	let collecting = $state<CollectingRow[]>([]);

	$effect(() => {
		void refresh();
		const timer = setInterval(() => void refresh(), POLL_MS);
		return () => clearInterval(timer);
	});

	// While anything is being collected, look again often; a collection ending
	// brings its notification.
	$effect(() => {
		if (collecting.length === 0) return;
		const timer = setInterval(() => void refresh(), COLLECTING_POLL_MS);
		return () => clearInterval(timer);
	});

	$effect(() => {
		if (open) return pushOverlay(() => (open = false));
	});

	/** Reads the unread count and what's being collected; a failed check keeps the last. */
	async function refresh() {
		try {
			const [count, running] = await Promise.all([fetchUnreadCount(), fetchCollecting()]);
			const finished = running.length < collecting.length;
			unread = count;
			collecting = running;
			if (open && finished) notifications = await fetchNotifications();
		} catch {
			return;
		}
	}

	/** Opens the panel and marks everything in it read. */
	async function toggle() {
		open = !open;
		if (!open) return;
		notifications = await fetchNotifications();
		if (unread === 0) return;
		await markAllRead();
		unread = 0;
	}

	/** Takes one notification out of the list. */
	async function clear(notification: NotificationRow) {
		notifications = notifications.filter((other) => other.id !== notification.id);
		await clearNotification(notification.id);
	}

	/** Empties the list. */
	async function clearAll() {
		notifications = [];
		await clearAllNotifications();
	}

	/** Where a notification leads: its record's page, when it has one. */
	function hrefOf(notification: NotificationRow): string | null {
		if (!notification.subject) return null;
		return recordUrl(notification.subject.kind, notification.subject.id);
	}

	/** "Reading their profiles…", or "Listing videos on xVideos, EPorner". */
	function collectingLine(row: CollectingRow): string {
		if (row.reading) return 'Reading their profiles and asking the sites…';
		const names = row.listing.map((plugin) => siteDisplayName(plugin, $sites));
		return `Listing videos on ${names.join(', ')} · ${row.performer.videoCount} so far`;
	}
</script>

<div class="relative">
	<button
		type="button"
		aria-label="Notifications"
		class="glass-round hover:bg-base-content/15 relative flex h-10 w-10 items-center justify-center rounded-full transition-colors"
		onclick={toggle}
	>
		{#if collecting.length > 0}
			<span
				class="border-t-primary absolute inset-1 animate-spin rounded-full border-2 border-transparent"
				aria-hidden="true"
			></span>
		{/if}
		<BellIcon size={22} weight={open ? 'fill' : 'regular'} />
		{#if unread > 0}
			<span
				class="bg-error absolute top-1 right-1 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-bold text-white"
			>
				{unread > 99 ? '99+' : unread}
			</span>
		{/if}
	</button>

	{#if open}
		<button
			type="button"
			class="z-overlay fixed inset-0 cursor-default"
			aria-label="Close notifications"
			tabindex="-1"
			onclick={() => (open = false)}
		></button>
		<div
			class="bg-base-200 z-overlay absolute top-12 right-0 flex max-h-[70vh] w-96 flex-col overflow-hidden rounded-2xl shadow-xl ring-1 ring-white/10"
		>
			{#if collecting.length > 0}
				<div class="border-base-content/10 flex flex-col border-b py-2">
					<h2 class="px-4 pb-1 text-xs font-semibold tracking-wide uppercase opacity-60">
						Collecting
					</h2>
					{#each collecting as row (row.performer.id)}
						<a
							href={performerUrl(row.performer.id)}
							class="hover:bg-base-300 flex items-center gap-3 px-4 py-2"
							onclick={() => (open = false)}
						>
							<span class="loading loading-spinner loading-xs shrink-0"></span>
							<span class="flex min-w-0 flex-col">
								<span class="text-sm font-medium">{row.performer.name}</span>
								<span class="text-base-content/60 truncate text-xs">{collectingLine(row)}</span>
							</span>
						</a>
					{/each}
				</div>
			{/if}
			<div class="flex items-center justify-between px-4 pt-3 pb-2">
				<h2 class="text-sm font-semibold">Notifications</h2>
				{#if notifications.length > 0}
					<button
						type="button"
						class="text-base-content/60 hover:text-base-content text-xs"
						onclick={clearAll}
					>
						Clear all
					</button>
				{/if}
			</div>
			<div class="flex flex-col overflow-y-auto pb-2">
				{#each notifications as notification (notification.id)}
					{@const href = hrefOf(notification)}
					<div class="group hover:bg-base-300 flex items-start gap-2 px-4 py-2">
						<a
							href={href ?? undefined}
							class="flex min-w-0 flex-1 flex-col gap-0.5"
							class:pointer-events-none={!href}
							onclick={() => (open = false)}
						>
							<span class="flex items-baseline justify-between gap-3">
								<span class="text-sm font-medium" class:opacity-70={notification.read}>
									{notification.title}
								</span>
								<span class="text-base-content/40 shrink-0 text-xs">
									{dayjs(notification.at).fromNow()}
								</span>
							</span>
							{#if notification.body}
								<span class="text-base-content/60 line-clamp-2 text-xs">{notification.body}</span>
							{/if}
						</a>
						<button
							type="button"
							aria-label="Clear"
							class="text-base-content/40 hover:text-base-content mt-0.5 opacity-0 group-hover:opacity-100 focus:opacity-100"
							onclick={() => clear(notification)}
						>
							<XIcon size={14} />
						</button>
					</div>
				{:else}
					<p class="text-base-content/50 px-4 py-3 text-sm">Nothing here.</p>
				{/each}
			</div>
		</div>
	{/if}
</div>
