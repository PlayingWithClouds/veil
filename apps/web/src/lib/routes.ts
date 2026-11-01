// Canonical route builders. Every internal link goes through these so a route
// rename is a one-line change and stray literals are grep-able.

/** One video's page. */
export function videoUrl(id: string): string {
	return `/video/${encodeURIComponent(id)}`;
}

/** One performer's page. */
export function performerUrl(id: string): string {
	return `/performers/${encodeURIComponent(id)}`;
}

/** One studio's page. */
export function studioUrl(id: string): string {
	return `/studios/${encodeURIComponent(id)}`;
}

/** One gallery's page. */
export function galleryUrl(id: string): string {
	return `/galleries/${encodeURIComponent(id)}`;
}

/** Home with the video grid filtered to one tag. Tags have no page of their own. */
export function tagFilterUrl(id: string): string {
	return `/?${new URLSearchParams({ tag: id })}`;
}

/** The plugins page. */
export function pluginsUrl(): string {
	return '/plugins';
}
