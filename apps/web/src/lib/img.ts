import { apiBase, backendUrl } from '$lib/server';

// Routes a remote image through the backend MinIO cache. Adult-source CDNs
// hotlink-protect their images, so the browser can't load them directly; the
// proxy fetches and caches them server-side. Non-absolute paths pass through,
// and the backend's own blob URLs are only pointed at its reachable address.
export function cacheUrl(url: string | null | undefined): string | null {
	if (!url) return null;
	if (!url.startsWith('http://') && !url.startsWith('https://')) return url;
	const rebased = backendUrl(url);
	if (rebased !== url) return rebased;
	return `${apiBase()}/api/img?url=${encodeURIComponent(url)}`;
}
