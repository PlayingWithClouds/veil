import { loadStudioDetail } from '$lib/studioDetail';

export async function load({ params }) {
	const detail = await loadStudioDetail(params.id);
	return { ...detail, id: params.id };
}
