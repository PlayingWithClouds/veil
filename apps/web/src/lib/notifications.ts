import type {
	Collecting,
	FieldsSelection,
	Notification
} from '@playingwithclouds/veil-sdk/generated';
import { gqlClient } from '$lib/veil';

/** A notification as the bell lists it. */
const NOTIFICATION = {
	id: true,
	at: true,
	title: true,
	body: true,
	read: true,
	subject: { kind: true, id: true }
} as const;

export type NotificationRow = FieldsSelection<Notification, typeof NOTIFICATION>;

/** How many notifications are unread. */
export async function fetchUnreadCount(): Promise<number> {
	const data = await gqlClient.query({ unreadNotificationCount: true });
	return data.unreadNotificationCount;
}

/** The newest notifications. */
export async function fetchNotifications(limit = 30): Promise<NotificationRow[]> {
	const data = await gqlClient.query({ notifications: { __args: { limit }, ...NOTIFICATION } });
	return data.notifications;
}

/** Marks every notification read. */
export async function markAllRead(): Promise<void> {
	await gqlClient.mutation({ markAllNotificationsRead: true });
}

/** Takes one notification out of the inbox. */
export async function clearNotification(id: string): Promise<void> {
	await gqlClient.mutation({ clearNotifications: { __args: { ids: [id] } } });
}

/** Empties the inbox. */
export async function clearAllNotifications(): Promise<void> {
	await gqlClient.mutation({ clearAllNotifications: true });
}

/** A performer something is being fetched about, as the bell lists it. */
const COLLECTING = {
	performer: { id: true, name: true, videoCount: true },
	reading: true,
	listing: true
} as const;

export type CollectingRow = FieldsSelection<Collecting, typeof COLLECTING>;

/** Every performer something is being fetched about right now. */
export async function fetchCollecting(): Promise<CollectingRow[]> {
	const data = await gqlClient.query({ collecting: COLLECTING });
	return data.collecting;
}
