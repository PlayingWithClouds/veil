// Entry point for every "add to split view" button.
import { goto } from '$app/navigation';
import { splitUrl } from '$lib/routes';
import { notifications } from '$lib/stores/notifications';
import { addSplitPane, type SplitPaneKind } from '$lib/stores/split';

/**
 * Adds a cam or scene to the split view and opens it. With four panes open the
 * focused one is replaced, and a notification says which.
 */
export async function openInSplit(kind: SplitPaneKind, mediaId: string, title: string) {
	const result = addSplitPane({ kind, mediaId, title });
	if (result.outcome === 'replaced') {
		notifications.push(`Split view is full: replaced ${result.replacedTitle}`, 'info');
	}
	await goto(splitUrl());
}
