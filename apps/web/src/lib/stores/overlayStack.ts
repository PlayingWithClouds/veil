// Open sheets and overlays register how to close themselves here, so the
// Android back button closes the topmost one before navigating back.

const closers: (() => void)[] = [];

/** Registers an open overlay's close function. Returns the unregister. */
export function pushOverlay(close: () => void): () => void {
	closers.push(close);
	return () => {
		const index = closers.lastIndexOf(close);
		if (index !== -1) closers.splice(index, 1);
	};
}

/** Closes the most recently opened overlay; false when none is open. */
export function closeTopOverlay(): boolean {
	const close = closers.at(-1);
	if (!close) return false;
	close();
	return true;
}
