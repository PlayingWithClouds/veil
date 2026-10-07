// Quick-save flow for card bookmarks: add the entity to the last-used
// collection (creating a default one if none exists), then surface a
// notification whose action opens the collection dialog to reassign it.
import { notifications } from '$lib/stores/notifications';
import { getLastCollection, setLastCollection } from '$lib/stores/lastCollection';
import { openMoveDialog } from '$lib/stores/moveDialog';
import {
	fetchCollections,
	createCollection,
	addToCollection,
	type Collection,
	type CollectionEntityType
} from '$lib/collections';

const DEFAULT_COLLECTION_NAME = 'Watch Later';

// Picks the collection to save into: the last one used (if it still exists),
// else the first user-created collection, else a freshly created default.
// Only user-origin collections are quick-save targets.
async function resolveTargetCollection(): Promise<Collection | null> {
	const collections = await fetchCollections('user');
	const lastId = getLastCollection();
	if (lastId) {
		const last = collections.find((collection) => collection.id === lastId);
		if (last) return last;
	}
	if (collections.length > 0) return collections[0];
	return createCollection(DEFAULT_COLLECTION_NAME);
}

export async function quickSaveEntity(
	entityType: CollectionEntityType,
	entityId: string,
	title: string
): Promise<void> {
	const target = await resolveTargetCollection();
	if (!target) {
		notifications.push('Could not save to a collection', 'error');
		return;
	}
	try {
		await addToCollection(target.id, entityId);
		setLastCollection(target.id);
		notifications.pushAction(
			`Saved to ${target.name}`,
			{ label: 'Move to another collection', run: () => openMoveDialog(entityType, entityId, title) },
			'success'
		);
	} catch {
		notifications.push('Could not save to a collection', 'error');
	}
}
