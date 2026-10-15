<script lang="ts">
	// Alternate sources of a scene: picking one attaches its streams.
	import { notifications } from '$lib/stores/notifications';
	import { attachAlikeSource, type AlikeCandidate } from '$lib/alike';
	import AlikeMatchesDialog from '$lib/components/AlikeMatchesDialog.svelte';
	import type { SceneStream } from '$lib/sceneDetail';

	interface Props {
		open: boolean;
		sceneId: string;
		sceneTitle: string;
		candidates: AlikeCandidate[];
		loading: boolean;
		onClose: () => void;
		onAttached: (streams: SceneStream[], pluginName: string) => void;
	}

	let { open, sceneId, sceneTitle, candidates, loading, onClose, onAttached }: Props = $props();

	/** Attaches the picked copy's streams to the scene. */
	async function attach(candidate: AlikeCandidate) {
		try {
			const streams = await attachAlikeSource(sceneId, candidate.plugin, candidate.sourceUrl);
			onAttached(streams, candidate.plugin);
			notifications.push(`Added source from ${candidate.plugin}`, 'success');
			onClose();
		} catch (err) {
			console.error('attachAlikeSource failed', err);
			notifications.push('Could not attach that source', 'error');
		}
	}
</script>

<AlikeMatchesDialog
	{open}
	heading="Alternate sources"
	subtitle={sceneTitle}
	searchingText="Searching every source for this scene…"
	emptyText="No alternate sources found."
	{candidates}
	{loading}
	{onClose}
	pick={attach}
/>
