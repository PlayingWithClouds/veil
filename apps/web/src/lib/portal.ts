// Svelte action that moves an element to another parent (default: <body>), so
// fixed overlays escape an ancestor's stacking context (a sticky or z-indexed
// wrapper would otherwise keep them under the tab bar). Updating the target
// moves it again, e.g. into a fullscreen element, the only subtree drawn then.

/** Moves the node into the target and keeps it there until it is destroyed. */
export function portal(node: HTMLElement, target: HTMLElement | null = null) {
	/** Appends the node to the target, or to <body> when there is none. */
	function move(next: HTMLElement | null) {
		let parent = next;
		if (!parent) parent = document.body;
		parent.appendChild(node);
	}

	move(target);
	return {
		update: move,
		destroy() {
			node.remove();
		}
	};
}
