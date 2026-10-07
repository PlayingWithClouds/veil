import { loadStudios } from '$lib/studioDetail';

export async function load() {
	const studios = await loadStudios();
	return { studios };
}
