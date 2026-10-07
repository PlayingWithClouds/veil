import { notifications } from '$lib/stores/notifications';

/**
 * Shares a link through the system share sheet, else copies it. Uses the
 * entity's own site URL when it has one (the app's URL means nothing outside
 * it), else the current page.
 */
export async function shareLink(title: string, url: string | null): Promise<void> {
	let link = url;
	if (!link) link = window.location.href;
	if (navigator.share) {
		try {
			await navigator.share({ title, url: link });
		} catch {
			// Dismissing the share sheet rejects; nothing to report.
		}
		return;
	}
	try {
		await navigator.clipboard.writeText(link);
		notifications.push('Link copied', 'success');
	} catch {
		notifications.push('Could not copy the link', 'error');
	}
}
