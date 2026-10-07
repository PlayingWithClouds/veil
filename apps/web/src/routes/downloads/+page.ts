import { redirect } from '@sveltejs/kit';

/** The download queue lives in the library now; old links land there. */
export function load() {
	redirect(307, '/library');
}
