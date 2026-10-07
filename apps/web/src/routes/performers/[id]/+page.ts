import { loadPerformerDetail } from '$lib/performerDetail';

export async function load({ params }) {
	const detail = await loadPerformerDetail(params.id);
	return { ...detail, id: params.id };
}
