// Entry point for every "add to split view" button.
import { notifications } from '$lib/stores/notifications';
import { addSplitPane, type SplitPaneKind } from '$lib/stores/split';

/**
 * Adds a cam or scene to the split view while the user stays on the page, so
 * they can keep picking; the view's dock shows it playing. With four panes
 * open the focused one is replaced, and a notification says which.
 */
export function openInSplit(kind: SplitPaneKind, mediaId: string, title: string) {
	const result = addSplitPane({ kind, mediaId, title });
	if (result.outcome === 'replaced') {
		notifications.push(`Split view is full: replaced ${result.replacedTitle}`, 'info');
		return;
	}
	if (result.outcome === 'added') notifications.push(`Added ${title} to split view`, 'success');
}
