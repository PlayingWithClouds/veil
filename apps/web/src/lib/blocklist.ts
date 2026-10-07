import { gqlClient } from '$lib/veil';

export type BlockKind = 'tag' | 'performer' | 'studio';

export interface BlocklistEntry {
	id: string;
	kind: BlockKind;
	targetId: string;
	label: string | null;
	createdAt: string;
}

export async function fetchBlocklist(): Promise<BlocklistEntry[]> {
	const result = await gqlClient.query({
		blocklist: {
			id: true,
			kind: true,
			targetId: true,
			label: true,
			createdAt: true
		}
	});
	return (result.blocklist ?? []) as BlocklistEntry[];
}

export async function addBlock(
	kind: BlockKind,
	targetId: string,
	label?: string | null
): Promise<BlocklistEntry> {
	const result = await gqlClient.mutation({
		addBlock: {
			__args: { kind, targetId, label: label ?? null },
			id: true,
			kind: true,
			targetId: true,
			label: true,
			createdAt: true
		}
	});
	return result.addBlock as BlocklistEntry;
}

export async function removeBlock(targetId: string): Promise<void> {
	await gqlClient.mutation({ removeBlock: { __args: { targetId } } });
}
