<script lang="ts">
	// "Who is this?": performers the face-recognition services think the picture
	// of a scene or gallery image shows, best first, with their similarity.
	// Picking one credits that performer (created if the library has none yet)
	// to the media. The services only know professional performers and cam models.
	// Videos with a similar face are listed below as plain links: they name nobody
	// and cannot be credited.
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import UserCircleIcon from 'phosphor-svelte/lib/UserCircleIcon';
	import { cacheUrl } from '$lib/img';
	import { notifications } from '$lib/stores/notifications';
	import {
		creditPerformerByName,
		identifyFace,
		type FaceCandidate,
		type FaceIdentification
	} from '$lib/faceIdentify';
	import type { PerformerRef } from '$lib/sceneDetail';

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
	// The candidate being credited, so its row shows a spinner.
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

	/** Similarity as a whole percentage. */
	function percent(similarity: number): number {
		return Math.round(Math.max(0, Math.min(1, similarity)) * 100);
	}

	/** The site a profile link points to, for its label. */
	function siteOf(profileUrl: string): string {
		try {
			return new URL(profileUrl).hostname.replace(/^www\./, '');
		} catch {
			return profileUrl;
		}
	}

	/** Credits the chosen performer to the media, one at a time. */
	async function choose(candidate: FaceCandidate) {
		if (crediting) return;
		crediting = candidate.name;
		try {
			const performer = await creditPerformerByName(mediaId, candidate.name);
			onCredited?.(performer);
			notifications.push(`Credited ${performer.name}`, 'success');
			onClose();
		} catch (error) {
			console.error('creditPerformerByName failed', candidate.name, error);
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
			class="border-base-300 bg-base-200 flex max-h-[85vh] w-full max-w-xl flex-col gap-4 rounded-2xl border p-6 shadow-2xl"
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

			{#if searchedPictureUrl}
				<div class="flex flex-col items-center gap-2">
					<img
						src={searchedPictureUrl}
						alt="Searched frame"
						class="max-h-48 rounded-lg object-contain"
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

			{#if !loading && identification}
				{#each identification.failures as failure (failure)}
					<p class="text-warning text-center text-sm">{failure}</p>
				{/each}
			{/if}

			{#if loading}
				<div class="flex flex-col items-center gap-3 py-10">
					<span class="loading loading-spinner text-base-content/40"></span>
					<p class="text-base-content/40 text-sm">Comparing the face with known performers…</p>
				</div>
			{:else if errorText}
				<p class="text-error py-10 text-center text-sm">{errorText}</p>
			{:else if identification && identification.candidates.length === 0 && identification.videos.length === 0}
				<p class="text-base-content/40 py-10 text-center text-sm">No performer looks like this.</p>
			{:else if identification}
				<div class="no-scrollbar flex flex-col gap-4 overflow-y-auto">
					{#if identification.candidates.length === 0}
						<p class="text-base-content/40 text-center text-sm">No performer looks like this.</p>
					{/if}
					<ul class="flex flex-col gap-2">
						{#each identification.candidates as candidate (candidate.name)}
							<li class="relative flex items-center gap-3 rounded-xl bg-white/5 p-2">
								{#if candidate.imageUrl}
									<img
										src={cacheUrl(candidate.imageUrl)}
										alt={candidate.name}
										loading="lazy"
										class="h-20 w-16 shrink-0 rounded-lg object-cover"
									/>
								{:else}
									<span
										class="text-base-content/20 flex h-20 w-16 shrink-0 items-center justify-center"
									>
										<UserCircleIcon size={40} />
									</span>
								{/if}
								<div class="flex min-w-0 flex-1 flex-col gap-1">
									<div class="flex flex-wrap items-center gap-2">
										<span class="truncate font-semibold">{candidate.name}</span>
										<span class="bg-primary rounded px-1.5 py-0.5 text-xs font-bold text-black">
											{percent(candidate.similarity)}%
										</span>
										{#if candidate.performerId}
											<span class="rounded bg-white/10 px-1.5 py-0.5 text-xs text-white/70">
												In your library
											</span>
										{/if}
									</div>
									{#if candidate.profileUrls.length > 0}
										<div class="flex flex-wrap gap-x-3 gap-y-0.5 text-xs">
											{#each candidate.profileUrls as profileUrl (profileUrl)}
												<a
													href={profileUrl}
													target="_blank"
													rel="noopener noreferrer"
													class="text-base-content/50 hover:text-base-content underline"
												>
													{siteOf(profileUrl)}
												</a>
											{/each}
										</div>
									{/if}
								</div>
								<button
									type="button"
									class="shrink-0 rounded-full bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-white/80 disabled:opacity-50"
									disabled={crediting !== null}
									onclick={() => choose(candidate)}
								>
									{#if crediting === candidate.name}
										<span class="loading loading-xs loading-spinner"></span>
									{:else}
										Credit
									{/if}
								</button>
							</li>
						{/each}
					</ul>
					{#if identification.videos.length > 0}
						<section class="flex flex-col gap-2">
							<h3 class="text-base-content/50 text-sm font-semibold">Similar videos</h3>
							<ul class="flex flex-col gap-2">
								{#each identification.videos as video (video.url)}
									<li>
										<a
											href={video.url}
											target="_blank"
											rel="noopener noreferrer"
											class="flex items-center gap-3 rounded-xl bg-white/5 p-2 hover:bg-white/10"
										>
											{#if video.thumbnailUrl}
												<img
													src={cacheUrl(video.thumbnailUrl)}
													alt=""
													loading="lazy"
													class="h-16 w-28 shrink-0 rounded-lg object-cover"
												/>
											{:else}
												<span class="h-16 w-28 shrink-0 rounded-lg bg-white/5"></span>
											{/if}
											<div class="flex min-w-0 flex-1 flex-col gap-1">
												<span class="line-clamp-2 text-sm font-semibold">{video.title}</span>
												<div class="flex items-center gap-2 text-xs">
													<span class="bg-primary rounded px-1.5 py-0.5 font-bold text-black">
														{percent(video.similarity)}%
													</span>
													<span class="text-base-content/50">{siteOf(video.url)}</span>
												</div>
											</div>
										</a>
									</li>
								{/each}
							</ul>
						</section>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}
