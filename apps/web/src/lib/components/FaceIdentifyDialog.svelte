<script lang="ts">
	// "Who is this?": performers the face-recognition services think the picture
	// of a scene or gallery image shows, best first, with their similarity.
	// Picking one credits that performer (created if the library has none yet)
	// to the media. The services only know professional performers and cam models.
	// Videos with a similar face are listed below: they name nobody and cannot be
	// credited. One a scene plugin serves opens as a scene in the app, others as a link.
	// Forum posts with a similar picture (NameThatPorn) get their own view: the
	// pornstars their answers name (creditable), then the posts as links.
	// Each kind of result is a view of its own, offered when it has any.
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import UserCircleIcon from 'phosphor-svelte/lib/UserCircleIcon';
	import { goto } from '$app/navigation';
	import { cacheUrl } from '$lib/img';
	import { sceneUrl } from '$lib/routes';
	import { ingestPluginResult } from '$lib/search';
	import { notifications } from '$lib/stores/notifications';
	import {
		creditPerformerByName,
		identifyFace,
		type FaceIdentification,
		type FacePostMatch,
		type FaceVideoMatch
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
	// The similar video being scraped to open, so its row shows a spinner.
	let openingVideoUrl = $state<string | null>(null);
	// The view the user picked; null shows the first one with results.
	let pickedView = $state<ResultView | null>(null);

	type ResultView = 'performers' | 'videos' | 'posts';

	// The views that have results, in the order they are offered.
	let views = $derived.by(() => {
		const offered: { key: ResultView; label: string; count: number }[] = [];
		if (!identification) return offered;
		if (identification.candidates.length > 0) {
			offered.push({
				key: 'performers',
				label: 'Performers',
				count: identification.candidates.length
			});
		}
		if (identification.videos.length > 0) {
			offered.push({ key: 'videos', label: 'Similar videos', count: identification.videos.length });
		}
		if (identification.posts.length > 0) {
			offered.push({ key: 'posts', label: 'Posts', count: identification.posts.length });
		}
		return offered;
	});

	let activeView = $derived.by(() => {
		if (pickedView && views.some((view) => view.key === pickedView)) return pickedView;
		if (views.length === 0) return null;
		return views[0].key;
	});

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
		pickedView = null;
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

	/** "Named in 1 post" / "Named in 3 posts". */
	function postCountLabel(posts: number): string {
		if (posts === 1) return 'Named in 1 post';
		return `Named in ${posts} posts`;
	}

	/** Scrapes a similar video through its plugin and opens it as a scene. */
	async function openVideo(video: FaceVideoMatch) {
		if (!video.scenePlugin || openingVideoUrl) return;
		openingVideoUrl = video.url;
		try {
			const sceneId = await ingestPluginResult(video.scenePlugin, video.url, video.thumbnailUrl);
			onClose();
			await goto(sceneUrl(sceneId));
		} catch (error) {
			console.error('open similar video failed', video.url, error);
			notifications.push('Could not open this video', 'error');
		} finally {
			openingVideoUrl = null;
		}
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

{#snippet videoRow(video: FaceVideoMatch, opening: boolean)}
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
			{#if opening}
				<span class="loading loading-xs loading-spinner"></span>
			{/if}
		</div>
	</div>
{/snippet}

{#snippet creditButton(name: string)}
	<button
		type="button"
		class="shrink-0 rounded-full bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-white/80 disabled:opacity-50"
		disabled={crediting !== null}
		onclick={() => choose(name)}
	>
		{#if crediting === name}
			<span class="loading loading-xs loading-spinner"></span>
		{:else}
			Credit
		{/if}
	</button>
{/snippet}

{#snippet libraryBadge()}
	<span class="rounded bg-white/10 px-1.5 py-0.5 text-xs text-white/70">In your library</span>
{/snippet}

{#snippet postRow(post: FacePostMatch)}
	<a
		href={post.url}
		target="_blank"
		rel="noopener noreferrer"
		class="flex items-center gap-3 rounded-xl bg-white/5 p-2 hover:bg-white/10"
	>
		{#if post.thumbnailUrl}
			<img
				src={cacheUrl(post.thumbnailUrl)}
				alt=""
				loading="lazy"
				class="size-16 shrink-0 rounded-lg object-cover"
			/>
		{:else}
			<span class="size-16 shrink-0 rounded-lg bg-white/5"></span>
		{/if}
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<span class="line-clamp-2 text-sm font-semibold">{post.title}</span>
			<div class="flex flex-wrap items-center gap-2 text-xs">
				{#if post.solved}
					<span class="bg-primary rounded px-1.5 py-0.5 font-bold text-black">Answered</span>
				{/if}
				{#if post.names.length > 0}
					<span class="text-base-content/70">{post.names.join(', ')}</span>
				{/if}
				<span class="text-base-content/50">{siteOf(post.url)}</span>
			</div>
		</div>
	</a>
{/snippet}

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
			{:else if identification && views.length === 0}
				<p class="text-base-content/40 py-10 text-center text-sm">No performer looks like this.</p>
			{:else if identification}
				{#if views.length > 1}
					<div class="flex flex-wrap items-center gap-2" role="tablist">
						{#each views as view (view.key)}
							<button
								type="button"
								role="tab"
								aria-selected={activeView === view.key}
								onclick={() => (pickedView = view.key)}
								class="border-base-300 text-base-content/40 hover:text-base-content/70 rounded-full border px-3 py-1 text-sm font-medium transition-colors"
								class:view-active={activeView === view.key}
							>
								{view.label}
								<span class="opacity-60">{view.count}</span>
							</button>
						{/each}
					</div>
				{/if}
				<div class="no-scrollbar flex flex-col gap-4 overflow-y-auto">
					{#if activeView === 'performers'}
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
												{@render libraryBadge()}
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
									{@render creditButton(candidate.name)}
								</li>
							{/each}
						</ul>
					{:else if activeView === 'videos'}
						<ul class="flex flex-col gap-2">
							{#each identification.videos as video (video.url)}
								<li>
									{#if video.scenePlugin}
										<button
											type="button"
											class="flex w-full items-center gap-3 rounded-xl bg-white/5 p-2 text-left hover:bg-white/10 disabled:opacity-50"
											disabled={openingVideoUrl !== null}
											onclick={() => openVideo(video)}
										>
											{@render videoRow(video, openingVideoUrl === video.url)}
										</button>
									{:else}
										<a
											href={video.url}
											target="_blank"
											rel="noopener noreferrer"
											class="flex items-center gap-3 rounded-xl bg-white/5 p-2 hover:bg-white/10"
										>
											{@render videoRow(video, false)}
										</a>
									{/if}
								</li>
							{/each}
						</ul>
					{:else if activeView === 'posts'}
						{#if identification.postPerformers.length > 0}
							<section class="flex flex-col gap-2">
								<h3 class="text-base-content/50 text-sm font-semibold">Named in these posts</h3>
								<ul class="flex flex-col gap-2">
									{#each identification.postPerformers as performer (performer.name)}
										<li class="flex items-center gap-3 rounded-xl bg-white/5 p-2">
											{#if performer.imageUrl}
												<img
													src={cacheUrl(performer.imageUrl)}
													alt={performer.name}
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
													<span class="truncate font-semibold">{performer.name}</span>
													{#if performer.performerId}
														{@render libraryBadge()}
													{/if}
												</div>
												<a
													href={performer.postUrl}
													target="_blank"
													rel="noopener noreferrer"
													class="text-base-content/50 hover:text-base-content text-xs underline"
												>
													{postCountLabel(performer.posts)}
												</a>
											</div>
											{@render creditButton(performer.name)}
										</li>
									{/each}
								</ul>
							</section>
						{/if}
						<section class="flex flex-col gap-2">
							<h3 class="text-base-content/50 text-sm font-semibold">
								Posts with a similar picture
							</h3>
							<ul class="flex flex-col gap-2">
								{#each identification.posts as post (post.url)}
									<li>{@render postRow(post)}</li>
								{/each}
							</ul>
						</section>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.view-active {
		border-color: var(--color-primary);
		background-color: color-mix(in oklab, var(--color-primary) 15%, transparent);
		color: var(--color-primary);
	}
</style>
