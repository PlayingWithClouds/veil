<script lang="ts">
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { cacheUrl } from '$lib/img';
	import { notifications } from '$lib/stores/notifications';
	import { fetchPluginIcons } from '$lib/search';
	import { attachAlikeSource, type AlikeCandidate } from '$lib/alike';
	import SceneCard, { type SceneCardData } from '$lib/components/SceneCard.svelte';
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

	// The source URL currently being attached, so its card shows a spinner.
	let attaching = $state<string | null>(null);
	let pluginIcons = $state<Record<string, string | null>>({});

	// Provider icons are shared across the app; fetch them once the dialog opens.
	$effect(() => {
		if (!open || Object.keys(pluginIcons).length > 0) return;
		fetchPluginIcons().then((icons) => (pluginIcons = icons));
	});

	function matchPercent(score: number): number {
		return Math.round(Math.max(0, Math.min(1, score)) * 100);
	}

	function toCard(candidate: AlikeCandidate): SceneCardData {
		const previewImages = candidate.previewImages
			.map((url) => cacheUrl(url))
			.filter((url): url is string => url !== null);
		return {
			// No id: these are external matches, not ingested scenes.
			title: candidate.title,
			bannerUrl: cacheUrl(candidate.posterUrl),
			previewImages,
			previewVideo: cacheUrl(candidate.previewVideo),
			providerIconUrl: pluginIcons[candidate.plugin] ?? null,
			durationSeconds: candidate.durationSeconds
		};
	}

	async function attach(candidate: AlikeCandidate) {
		if (attaching) return;
		attaching = candidate.sourceUrl;
		try {
			const streams = await attachAlikeSource(sceneId, candidate.plugin, candidate.sourceUrl);
			onAttached(streams, candidate.plugin);
			notifications.push(`Added source from ${candidate.plugin}`, 'success');
			onClose();
		} catch (err) {
			console.error('attachAlikeSource failed', err);
			notifications.push('Could not attach that source', 'error');
		} finally {
			attaching = null;
		}
	}
</script>

{#if open}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4"
		role="button"
		tabindex="-1"
		onclick={onClose}
		onkeydown={(event) => event.key === 'Escape' && onClose()}
	>
		<!-- Dialog -->
		<div
			class="border-base-300 bg-base-200 flex max-h-[85vh] w-full max-w-3xl flex-col gap-4 rounded-2xl border p-6 shadow-2xl"
			role="dialog"
			aria-modal="true"
			tabindex="-1"
			onclick={(event) => event.stopPropagation()}
			onkeydown={() => {}}
		>
			<div class="flex items-start justify-between gap-3">
				<div class="flex flex-col">
					<h2 class="text-lg font-bold">Alternate sources</h2>
					<p class="text-base-content/50 line-clamp-1 text-sm">{sceneTitle}</p>
				</div>
				<button
					type="button"
					class="text-base-content/40 hover:text-base-content"
					aria-label="Close"
					onclick={onClose}
				>
					<XIcon size={20} />
				</button>
			</div>

			{#if loading}
				<div class="flex flex-col items-center gap-3 py-10">
					<span class="loading loading-spinner text-base-content/40"></span>
					<p class="text-base-content/40 text-sm">Searching every source for this scene…</p>
				</div>
			{:else if candidates.length === 0}
				<p class="text-base-content/40 py-10 text-center text-sm">No alternate sources found.</p>
			{:else}
				<div class="no-scrollbar grid grid-cols-1 gap-4 overflow-y-auto sm:grid-cols-2">
					{#each candidates as candidate (candidate.sourceUrl)}
						<div class="relative">
							<SceneCard item={toCard(candidate)} onclick={() => attach(candidate)} />

							<!-- Match confidence, mirrored on the card's free top-left corner. -->
							<span
								class="bg-primary pointer-events-none absolute left-1.5 top-1.5 rounded px-1.5 py-0.5 text-xs font-bold text-black"
							>
								{matchPercent(candidate.matchScore)}%
							</span>

							{#if attaching === candidate.sourceUrl}
								<div
									class="absolute inset-0 flex items-center justify-center rounded-[inherit] bg-black/60"
								>
									<span class="loading loading-spinner text-white"></span>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/if}
