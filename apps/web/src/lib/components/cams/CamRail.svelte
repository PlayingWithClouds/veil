<script lang="ts">
	// The cam page's "More live rooms" rail: rooms live now that share the
	// room's tags, then its gender, then the most watched, never the room itself.
	// Compact cards: thumbnail with viewers on the left, name and topic beside it.
	import { goto } from '$app/navigation';
	import EyeIcon from 'phosphor-svelte/lib/EyeIcon';
	import { cacheUrl } from '$lib/img';
	import { camUrl } from '$lib/routes';
	import { fetchCamRooms, formatViewers, roomTopic, type CamModel } from '$lib/cams';

	interface Props {
		// The room on screen; its tags and gender pick the neighbours.
		camModel: CamModel;
	}

	let { camModel }: Props = $props();

	// How many rooms the rail shows, and how many tags of the room it asks about.
	const RAIL_SIZE = 30;
	const TAGS_ASKED = 2;

	let rooms = $state<CamModel[]>([]);
	let loading = $state(false);

	$effect(() => {
		const current = camModel.id;
		const tags = camModel.tags.slice(0, TAGS_ASKED);
		let gender = '';
		if (camModel.gender) gender = camModel.gender;
		loading = true;
		loadRail(current, tags, gender).then((found) => {
			if (camModel.id !== current) return;
			rooms = found;
			loading = false;
		});
	});

	/** Rooms for the rail: by each tag, then by gender, then any, until it is full. */
	async function loadRail(currentId: string, tags: string[], gender: string): Promise<CamModel[]> {
		const picked: CamModel[] = [];
		const seen = new Set([currentId]);
		const queries = [
			...tags.map((tag) => ({ tag, gender })),
			{ tag: '', gender },
			{ tag: '', gender: '' }
		];
		for (const query of queries) {
			if (picked.length >= RAIL_SIZE) break;
			const page = await fetchCamRooms({ sources: [], query: '', ...query }, RAIL_SIZE, 0).catch(() => null);
			if (!page) continue;
			for (const room of page.rooms) {
				if (seen.has(room.id) || picked.length >= RAIL_SIZE) continue;
				seen.add(room.id);
				picked.push(room);
			}
		}
		return picked;
	}
</script>

<!-- Beside the player (lg) the rail scrolls on its own, held at the top of the
     page, like the watch page's up-next rail. -->
<aside
	class="no-scrollbar flex flex-col gap-3 lg:sticky lg:top-0 lg:max-h-[calc(100dvh-3.5rem)] lg:self-start lg:overflow-y-auto lg:overscroll-y-contain lg:pb-4"
	aria-label="More live rooms"
>
	<h2 class="font-semibold">More live rooms</h2>
	{#if loading && rooms.length === 0}
		<div class="flex justify-center py-6">
			<span class="loading loading-spinner loading-sm text-base-content/40"></span>
		</div>
	{/if}
	{#each rooms as room (room.id)}
		{@const topic = roomTopic(room.subject, room.username)}
		<button
			type="button"
			class="group flex gap-3 text-left"
			onclick={() => goto(camUrl(room.id))}
		>
			<span class="bg-base-200 relative aspect-video w-[60%] min-w-[168px] shrink-0 overflow-hidden rounded-xl">
				{#if cacheUrl(room.thumbnailUrl)}
					<img
						src={cacheUrl(room.thumbnailUrl)}
						alt=""
						loading="lazy"
						class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.04]"
					/>
				{/if}
				<span
					class="absolute right-1.5 bottom-1.5 flex items-center gap-1 rounded bg-black/75 px-1.5 py-0.5 text-[11px] font-medium text-white"
				>
					<EyeIcon size={11} weight="fill" />
					{formatViewers(room.viewers)}
				</span>
			</span>
			<span class="flex min-w-0 flex-col gap-1 py-0.5">
				<span class="truncate text-[15px] leading-[21px] font-medium group-hover:text-white">
					{room.displayName}
				</span>
				{#if topic}
					<span class="text-base-content/60 line-clamp-3 text-[13px] leading-snug">{topic}</span>
				{/if}
				{#if room.tags.length > 0}
					<span class="text-base-content/40 truncate text-xs">
						{room.tags.slice(0, 3).map((tag) => `#${tag}`).join(' ')}
					</span>
				{/if}
			</span>
		</button>
	{/each}
</aside>
