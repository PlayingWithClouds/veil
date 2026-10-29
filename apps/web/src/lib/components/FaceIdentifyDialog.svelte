<script lang="ts">
	// "Who is this?": what the face searches answer for the picture of a scene or
	// gallery image (FaceResults lays it out). Picking a performer credits them
	// (created if the library has none yet) to the media.
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import { cacheUrl } from '$lib/img';
	import { notifications } from '$lib/stores/notifications';
	import { creditPerformerByName, identifyFace, type FaceIdentification } from '$lib/faceIdentify';
	import type { PerformerRef } from '$lib/sceneDetail';
	import FaceResults from './FaceResults.svelte';

	interface Props {
		open: boolean;
		subtitle: string;
		/** The scene, gallery or image whose picture is identified and which gets the credit. */
		mediaId: string;
		/** The gallery image on screen; absent for scenes. */
		imageId?: string | null;
		/** A paused video frame (JPEG data URL), searched instead of the stored picture. */
		frameUrl?: string | null;
		onClose: () => void;
		/** Offered with a frame: closes the dialog so the user can scrub to another moment. */
		onPickAnother?: () => void;
		onCredited?: (performer: PerformerRef) => void;
	}

	let { open, subtitle, mediaId, imageId, frameUrl, onClose, onPickAnother, onCredited }: Props =
		$props();

	let loading = $state(false);
	let identification = $state<FaceIdentification | null>(null);
	let errorText = $state<string | null>(null);
	// The name being credited, so its row shows a spinner.
	let crediting = $state<string | null>(null);

	// The picture being searched: the frame right away, a stored one once the backend names it.
	let searchedPictureUrl = $derived.by(() => {
		if (frameUrl) return frameUrl;
		if (identification && identification.pictureUrl) return cacheUrl(identification.pictureUrl);
		return null;
	});

	// Identify each time the dialog opens, or when another picture is asked about.
	$effect(() => {
		if (!open) return;
		search(mediaId, imageId, frameUrl);
	});

	/** Runs the identification for one picture. */
	async function search(
		forMediaId: string,
		forImageId: string | null | undefined,
		forFrameUrl: string | null | undefined
	) {
		loading = true;
		identification = null;
		errorText = null;
		try {
			identification = await identifyFace(forMediaId, forImageId, frameBase64(forFrameUrl));
		} catch (error) {
			console.error('identifyFace failed', forMediaId, error);
			errorText = error instanceof Error ? error.message : 'Could not identify this picture.';
		} finally {
			loading = false;
		}
	}

	/** The base64 payload of a data URL, or null without one. */
	function frameBase64(dataUrl: string | null | undefined): string | null {
		if (!dataUrl) return null;
		return dataUrl.slice(dataUrl.indexOf(',') + 1);
	}

	/** Credits the performer called name to the media, one at a time. */
	async function choose(name: string) {
		if (crediting) return;
		crediting = name;
		try {
			const performer = await creditPerformerByName(mediaId, name);
			onCredited?.(performer);
			notifications.push(`Credited ${performer.name}`, 'success');
			onClose();
		} catch (error) {
			console.error('creditPerformerByName failed', name, error);
			notifications.push('Could not credit that performer', 'error');
		} finally {
			crediting = null;
		}
	}
</script>

{#if open}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-[70] flex items-center justify-center bg-black/60 p-4"
		role="button"
		tabindex="-1"
		onclick={onClose}
		onkeydown={(event) => event.key === 'Escape' && onClose()}
	>
		<!-- Dialog -->
		<div
			class="border-base-300 bg-base-200 flex max-h-[90vh] w-full max-w-3xl flex-col gap-4 rounded-2xl border p-6 shadow-2xl max-md:p-4"
			role="dialog"
			aria-modal="true"
			tabindex="-1"
			onclick={(event) => event.stopPropagation()}
			onkeydown={() => {}}
		>
			<div class="flex items-start justify-between gap-3">
				<div class="flex flex-col">
					<h2 class="text-lg font-bold">Who is this?</h2>
					<p class="text-base-content/50 line-clamp-1 text-sm">{subtitle}</p>
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

			<div class="no-scrollbar flex flex-col gap-4 overflow-y-auto">
				{#if searchedPictureUrl}
					<div class="flex flex-col items-center gap-2">
						<img
							src={searchedPictureUrl}
							alt="Searched frame"
							class="max-h-64 rounded-lg object-contain"
						/>
						{#if frameUrl && onPickAnother}
							<button
								type="button"
								class="text-base-content/60 hover:text-base-content text-sm underline"
								onclick={onPickAnother}
							>
								Pick another moment
							</button>
						{/if}
					</div>
				{/if}

				{#if loading}
					<div class="flex flex-col items-center gap-3 py-10">
						<span class="loading loading-spinner text-base-content/40"></span>
						<p class="text-base-content/40 text-sm">Comparing the face with known performers…</p>
					</div>
				{:else if errorText}
					<p class="text-error py-10 text-center text-sm">{errorText}</p>
				{:else if identification}
					{#key identification}
						<FaceResults {identification} onCredit={choose} {crediting} onLeave={onClose} />
					{/key}
				{/if}
			</div>
		</div>
	</div>
{/if}
