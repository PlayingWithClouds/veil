import { usesEmbeddedBackend, waitForEmbeddedBackend } from '$lib/server';

/** Holds the first page of the native app until its on-device backend is up. */
export async function load() {
	if (usesEmbeddedBackend()) {
		await waitForEmbeddedBackend();
	}
}
