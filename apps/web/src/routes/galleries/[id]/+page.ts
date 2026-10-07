import { loadGalleryDetail } from '$lib/galleries';

export async function load({ params }) {
	const gallery = await loadGalleryDetail(params.id);
	return { gallery, id: params.id };
}
