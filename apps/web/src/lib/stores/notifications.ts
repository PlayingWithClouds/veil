import { writable } from 'svelte/store';

// App-wide transient notifications surfaced by the top-right NotificationManager
// (e.g. "Download started"). Entries auto-dismiss after a timeout.
export type NotificationKind = 'info' | 'success' | 'error';

// Optional inline action rendered as a button in the toast (e.g. "Move").
export interface NotificationAction {
	label: string;
	run: () => void;
}

export interface Notification {
	id: string;
	message: string;
	kind: NotificationKind;
	action?: NotificationAction;
}

const AUTO_DISMISS_MS = 6000;

function createNotifications() {
	const { subscribe, update } = writable<Notification[]>([]);
	let counter = 0;

	function dismiss(id: string) {
		update((list) => list.filter((entry) => entry.id !== id));
	}

	function add(message: string, kind: NotificationKind, action?: NotificationAction): string {
		counter += 1;
		const id = `${Date.now()}-${counter}`;
		update((list) => [{ id, message, kind, action }, ...list]);
		setTimeout(() => dismiss(id), AUTO_DISMISS_MS);
		return id;
	}

	function push(message: string, kind: NotificationKind = 'info'): string {
		return add(message, kind, undefined);
	}

	// Like push, but with an action button. Running the action also dismisses the
	// toast.
	function pushAction(
		message: string,
		action: NotificationAction,
		kind: NotificationKind = 'info'
	): string {
		let id = '';
		const wrapped: NotificationAction = {
			label: action.label,
			run: () => {
				action.run();
				dismiss(id);
			}
		};
		id = add(message, kind, wrapped);
		return id;
	}

	return { subscribe, push, pushAction, dismiss };
}

export const notifications = createNotifications();
