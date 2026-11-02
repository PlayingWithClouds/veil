<!--
	Choosing a performer's picture: upload a file (click or drop), paste a
	URL, take a video's cover or one of its preview frames, or play a video
	and take the frame on screen. Hands the chosen picture's URI to onpicked.
-->
<script lang="ts">
	import UploadSimpleIcon from 'phosphor-svelte/lib/UploadSimpleIcon';
	import FilmStripIcon from 'phosphor-svelte/lib/FilmStripIcon';
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import FrameGrabber from '$lib/components/FrameGrabber.svelte';
	import ScrollSentinel from '$lib/components/ScrollSentinel.svelte';
	import { fetchVideos, type VideoCardRow } from '$lib/library';
	import { uploadPicture } from '$lib/performers';
	import { cacheUrl } from '$lib/img';

	interface Props {
		performerId: string;
		onpicked: (uri: string) => void;
		onclose: () => void;
	}

	let { performerId, onpicked, onclose }: Props = $props();

	const PAGE_SIZE = 12;

	let videos = $state<VideoCardRow[]>([]);
	let loadingVideos = $state(false);
	let exhausted = $state(false);
	let grabbing = $state<VideoCardRow | null>(null);
	let url = $state('');
	let uploading = $state(false);
	let dragging = $state(false);
	let error = $state<string | null>(null);
	let fileInput = $state<HTMLInputElement>();

	/** Opens the dialog as a modal once it's in the page. */
	function showModal(dialog: HTMLDialogElement) {
		dialog.showModal();
	}

	/** Appends the next page of the performer's videos. */
	async function loadMore() {
		if (loadingVideos || exhausted) return;
		loadingVideos = true;
		try {
			const page = await fetchVideos({ performerId }, PAGE_SIZE, videos.length);
			videos = [...videos, ...page];
			exhausted = page.length < PAGE_SIZE;
		} finally {
			loadingVideos = false;
		}
	}

	/** Uploads a file and picks it. */
	async function upload(file: File | undefined) {
		if (!file) return;
		if (!file.type.startsWith('image/')) {
			error = 'That file is not a picture.';
			return;
		}
		uploading = true;
		error = null;
		try {
			onpicked(await uploadPicture(file));
		} catch (failure) {
			error = String(failure);
		} finally {
			uploading = false;
		}
	}

	/** Takes a file dropped on the upload area. */
	function drop(event: DragEvent) {
		event.preventDefault();
		dragging = false;
		void upload(event.dataTransfer?.files[0]);
	}

	/** The stills a video offers: its cover, then its preview frames. */
	function stillsOf(video: VideoCardRow): string[] {
		const stills = (video.previewFrames ?? []).map((frame) => frame.uri);
		if (video.cover) stills.unshift(video.cover.uri);
		return [...new Set(stills)];
	}
</script>

<dialog use:showModal class="modal" {onclose} aria-label="Choose a picture">
	<div class="modal-box flex max-h-[90vh] !max-w-4xl flex-col gap-5 overflow-y-auto">
		{#if grabbing}
			<div class="flex items-center gap-2">
				<button type="button" class="btn btn-ghost btn-sm btn-square" aria-label="Back" onclick={() => (grabbing = null)}>
					<CaretLeftIcon size={18} />
				</button>
				<h2 class="line-clamp-1 text-lg font-semibold">{grabbing.title}</h2>
			</div>
			<FrameGrabber videoId={grabbing.id} onpick={(frame) => upload(frame)} />
		{:else}
			<h2 class="text-lg font-semibold">Choose a picture</h2>

			<div class="grid grid-cols-[1fr_1fr] gap-3 max-md:grid-cols-1">
				<button
					type="button"
					class="bg-base-200 hover:bg-base-300 border-base-content/10 flex h-28 flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed text-sm"
					class:!border-primary={dragging}
					disabled={uploading}
					onclick={() => fileInput?.click()}
					ondragover={(event) => {
						event.preventDefault();
						dragging = true;
					}}
					ondragleave={() => (dragging = false)}
					ondrop={drop}
				>
					<UploadSimpleIcon size={24} />
					{uploading ? 'Uploading…' : 'Upload or drop a picture'}
				</button>
				<form
					class="flex flex-col justify-center gap-2"
					onsubmit={(event) => {
						event.preventDefault();
						if (url.trim()) onpicked(url.trim());
					}}
				>
					<label class="form-control">
						<span class="label-text">Or a picture's URL</span>
						<input class="input" bind:value={url} placeholder="https://…" />
					</label>
					<button type="submit" class="btn btn-sm self-end" disabled={!url.trim()}>Use URL</button>
				</form>
			</div>
			<input
				bind:this={fileInput}
				type="file"
				accept="image/*"
				class="hidden"
				onchange={(event) => upload(event.currentTarget.files?.[0])}
			/>

			<section class="flex flex-col gap-3">
				<h3 class="text-sm font-semibold">From their videos</h3>
				<div class="flex flex-col gap-4">
					{#each videos as video (video.id)}
						<div class="flex flex-col gap-1.5">
							<div class="flex items-center justify-between gap-3">
								<span class="text-base-content/70 line-clamp-1 text-xs">{video.title}</span>
								<button type="button" class="btn btn-ghost btn-xs shrink-0" onclick={() => (grabbing = video)}>
									<FilmStripIcon size={14} /> Pick a frame…
								</button>
							</div>
							<div class="flex gap-2 overflow-x-auto pb-1">
								{#each stillsOf(video) as still (still)}
									<button
										type="button"
										class="hover:ring-primary aspect-video w-40 shrink-0 overflow-hidden rounded-lg ring-2 ring-transparent"
										aria-label="Use this still"
										onclick={() => onpicked(still)}
									>
										<img src={cacheUrl(still)} alt="" loading="lazy" class="h-full w-full object-cover" />
									</button>
								{/each}
							</div>
						</div>
					{/each}
				</div>
				<ScrollSentinel onvisible={loadMore} loading={loadingVideos} done={exhausted} />
				{#if exhausted && videos.length === 0}
					<p class="text-base-content/50 text-sm">No videos of them in the library yet.</p>
				{/if}
			</section>
		{/if}

		{#if error}
			<p class="text-error text-sm">{error}</p>
		{/if}
		<div class="modal-action mt-0">
			<button type="button" class="btn btn-ghost" onclick={onclose}>Cancel</button>
		</div>
	</div>
</dialog>
