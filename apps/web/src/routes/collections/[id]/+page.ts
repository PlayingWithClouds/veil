import { fetchCollection, fetchCollectionMembers } from '$lib/collections';

export async function load({ params }) {
	const [collection, members] = await Promise.all([
		fetchCollection(params.id),
		fetchCollectionMembers(params.id)
	]);
	return { id: params.id, collection, members };
}
