import { fetchTag } from '$lib/tags';

export async function load({ params }) {
	const tag = await fetchTag(params.id);
	return { id: params.id, tag };
}
