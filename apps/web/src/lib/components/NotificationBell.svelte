<!--
	The bell in the top bar: the unread count as a badge (checked every
	minute), and on click the newest notifications. Opening it marks them read;
	a notification about a record opens its page.
-->
<script lang="ts">
	import dayjs from 'dayjs';
	import { goto } from '$app/navigation';
	import BellIcon from 'phosphor-svelte/lib/BellIcon';
	import {
		fetchUnreadCount,
		fetchNotifications,
		markAllRead,
		type NotificationRow
	} from '$lib/notifications';
	import { recordUrl } from '$lib/routes';
	import { pushOverlay } from '$lib/stores/overlayStack';

	const POLL_MS = 60_000;

	let unread = $state(0);
	let open = $state(false);
	let notifications = $state<NotificationRow[]>([]);

	$effect(() => {
		void refreshCount();
		const timer = setInterval(() => void refreshCount(), POLL_MS);
		return () => clearInterval(timer);
	});

	$effect(() => {
		if (open) return pushOverlay(() => (open = false));
	});

	/** Reads the unread count; a failed check keeps the last one. */
	async function refreshCount() {
		try {
			unread = await fetchUnreadCount();
		} catch {
			return;
		}
	}

	/** Opens the list and marks everything in it read. */
	async function toggle() {
		open = !open;
		if (!open) return;
		notifications = await fetchNotifications();
		if (unread === 0) return;
		await markAllRead();
		unread = 0;
	}

	/** Goes to the page of the notification's record, when it has one. */
	function follow(notification: NotificationRow) {
		open = false;
		if (!notification.subject) return;
		const href = recordUrl(notification.subject.kind, notification.subject.id);
		if (href) goto(href);
	}
</script>

<div class="relative">
	<button
		type="button"
		aria-label="Notifications"
		class="hover:bg-base-content/8 relative flex h-10 w-10 items-center justify-center rounded-full"
		onclick={toggle}
	>
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
			<h2 class="px-4 pt-3 pb-2 text-sm font-semibold">Notifications</h2>
			<div class="flex flex-col overflow-y-auto pb-2">
				{#each notifications as notification (notification.id)}
					<button
						type="button"
						class="hover:bg-base-300 flex flex-col gap-0.5 px-4 py-2 text-left"
						onclick={() => follow(notification)}
					>
						<span class="flex items-baseline justify-between gap-3">
							<span class="text-sm font-medium" class:text-base-content={!notification.read}>
								{notification.title}
							</span>
							<span class="text-base-content/40 shrink-0 text-xs">
								{dayjs(notification.at).fromNow()}
							</span>
						</span>
						{#if notification.body}
							<span class="text-base-content/60 line-clamp-2 text-xs">{notification.body}</span>
						{/if}
					</button>
				{:else}
					<p class="text-base-content/50 px-4 py-3 text-sm">Nothing new.</p>
				{/each}
			</div>
		</div>
	{/if}
</div>
