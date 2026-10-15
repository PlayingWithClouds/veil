// Canonical route builders. Every internal link goes through these so a route
// rename is a one-line change and stray literals are grep-able.

export function sceneUrl(id: string): string {
	return `/scene/${encodeURIComponent(id)}`;
}

export function performerUrl(id: string): string {
	return `/performers/${encodeURIComponent(id)}`;
}

export function studioUrl(id: string): string {
	return `/studios/${encodeURIComponent(id)}`;
}

export function galleryUrl(id: string): string {
	return `/galleries/${encodeURIComponent(id)}`;
}

export function collectionUrl(id: string): string {
	return `/collections/${encodeURIComponent(id)}`;
}

export function tagUrl(id: string): string {
	return `/tags/${encodeURIComponent(id)}`;
}

/** Subscriptions page, optionally with one subscription's feed open. */
export function subscriptionsUrl(id?: string): string {
	if (id === undefined) return '/subscriptions';
	return `/subscriptions?id=${encodeURIComponent(id)}`;
}

/** The live cams grid. */
export function camsUrl(): string {
	return '/cams';
}

/** One live cam room. */
export function camUrl(id: string): string {
	return `/cams/${encodeURIComponent(id)}`;
}

/** The split view that plays several cams and videos side by side. */
export function splitUrl(): string {
	return '/split';
}
