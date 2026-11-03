// Canonical route builders. Every internal link goes through these so a route
// rename is a one-line change and stray literals are grep-able.

/** One video's page. */
export function videoUrl(id: string): string {
	return `/video/${encodeURIComponent(id)}`;
}

/** A plugin on the plugins page, opened at its login. */
export function pluginLoginUrl(name: string): string {
	return `/plugins#${encodeURIComponent(name)}`;
}

/** One performer's page. */
export function performerUrl(id: string): string {
	return `/performers/${encodeURIComponent(id)}`;
}

/** One studio's page. */
export function studioUrl(id: string): string {
	return `/studios/${encodeURIComponent(id)}`;
}

/** The galleries page. */
export function galleriesUrl(): string {
	return '/galleries';
}

/** One gallery's page. */
export function galleryUrl(id: string): string {
	return `/galleries/${encodeURIComponent(id)}`;
}

/** The page of any record by kind; null for kinds without a page. */
export function recordUrl(kind: string, id: string): string | null {
	const builders: Record<string, (recordId: string) => string> = {
		video: videoUrl,
		gallery: galleryUrl,
		performer: performerUrl,
		studio: studioUrl,
		tag: tagFilterUrl,
		audio: audioUrl,
		story: storyUrl,
		post: postUrl,
		group: groupUrl,
		category: categoryUrl,
		collection: collectionUrl,
		camModel: camUrl
	};
	const build = builders[kind];
	if (!build) return null;
	return build(id);
}

/** One audio's page. */
export function audioUrl(id: string): string {
	return `/audio/${encodeURIComponent(id)}`;
}

/** The audio in the library. */
export function audiosUrl(): string {
	return '/audio';
}

/** One story's page. */
export function storyUrl(id: string): string {
	return `/stories/${encodeURIComponent(id)}`;
}

/** The stories in the library. */
export function storiesUrl(): string {
	return '/stories';
}

/** The reader at a story's chapter (by position), else where reading stopped. */
export function storyReaderUrl(id: string, position?: number): string {
	const base = `${storyUrl(id)}/read`;
	if (position === undefined) return base;
	return `${base}?${new URLSearchParams({ chapter: String(position) })}`;
}

/** One creator post's page. */
export function postUrl(id: string): string {
	return `/posts/${encodeURIComponent(id)}`;
}

/** One site group's page (a movie, series, compilation). */
export function groupUrl(id: string): string {
	return `/groups/${encodeURIComponent(id)}`;
}

/** One category's page. */
export function categoryUrl(id: string): string {
	return `/categories/${encodeURIComponent(id)}`;
}

/** One collection's page. */
export function collectionUrl(id: string): string {
	return `/collections/${encodeURIComponent(id)}`;
}

/** One cam room's live stream. */
export function camUrl(id: string): string {
	return `/cams/${encodeURIComponent(id)}`;
}

/** Home with the video grid playing one saved filter. */
export function savedFilterUrl(id: string): string {
	return `/?${new URLSearchParams({ saved: id })}`;
}

/** Home with the video grid filtered to one tag. Tags have no page of their own. */
export function tagFilterUrl(id: string): string {
	return `/?${new URLSearchParams({ tag: id })}`;
}

/** The plugins page. */
export function pluginsUrl(): string {
	return '/plugins';
}
