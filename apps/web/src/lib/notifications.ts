import type { RecordKind } from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';

/** A notification as the bell lists it. */
export interface NotificationRow {
	id: string;
	at: string;
	title: string;
	body: string | null;
	read: boolean;
	subject: { kind: RecordKind; id: string } | null;
}

/** How many notifications are unread. */
export async function fetchUnreadCount(): Promise<number> {
	const data = await gqlClient.query({ unreadNotificationCount: true });
	return data.unreadNotificationCount;
}

/** The newest notifications. */
export async function fetchNotifications(limit = 30): Promise<NotificationRow[]> {
	const data = await gqlClient.query({
		notifications: {
			__args: { limit },
			id: true,
			at: true,
			title: true,
			body: true,
			read: true,
			subject: { kind: true, id: true }
		}
	});
	return data.notifications;
}

/** Marks every notification read. */
export async function markAllRead(): Promise<void> {
	await gqlClient.mutation({ markAllNotificationsRead: true });
}
