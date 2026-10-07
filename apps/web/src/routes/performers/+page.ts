import { loadPerformers } from '$lib/performerDetail';

export async function load() {
	const performers = await loadPerformers();
	return { performers };
}
