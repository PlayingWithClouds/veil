<!--
	The comments under the video on its sites. Viewers often say who is in it,
	so comments that ask or tell (or name a performer of the library) can be
	shown on their own, named performers link to their page and can be credited
	in one click, and any selected words of a comment can be credited as a name.
-->
<script lang="ts">
	import ArrowClockwiseIcon from 'phosphor-svelte/lib/ArrowClockwiseIcon';
	import ChatCircleTextIcon from 'phosphor-svelte/lib/ChatCircleTextIcon';
	import ThumbsUpIcon from 'phosphor-svelte/lib/ThumbsUpIcon';
	import UserPlusIcon from 'phosphor-svelte/lib/UserPlusIcon';
	import { performerUrl } from '$lib/routes';
	import { formatRelativeTime } from '$lib/time';
	import { creditPerformerByName } from '$lib/faceIdentify';
	import { fetchSearchPlugins } from '$lib/search';
	import { notifications } from '$lib/stores/notifications';
	import type { PerformerRef } from '$lib/sceneDetail';
	import {
		commentSegments,
		commentThreads,
		fetchSceneComments,
		isAboutIdentity,
		refreshSceneComments,
		type CommentMention,
		type SceneComment
	} from '$lib/sceneComments';
	import { initials } from './format';

	interface Props {
		sceneId: string;
		// Bumped when the scene's page was fetched, so newly stored comments load.
		reloadKey: number;
		// Ids of the performers credited on the scene.
		creditedIds: string[];
		oncredited: (performer: PerformerRef) => void;
	}

	let { sceneId, reloadKey, creditedIds, oncredited }: Props = $props();

	// Threads shown before "Show all".
	const COLLAPSED_THREADS = 5;

	let comments = $state<SceneComment[]>([]);
	let loaded = $state(false);
	let refreshing = $state(false);
	let expanded = $state(false);
	let identityOnly = $state(false);
	let crediting = $state<string | null>(null);
	// The comment whose name field is open, and what it holds.
	let namingId = $state<string | null>(null);
	let typedName = $state('');

	let siteNames = $state<Map<string, string>>(new Map());
	$effect(() => {
		fetchSearchPlugins().then((plugins) => {
			siteNames = new Map(plugins.map((plugin) => [plugin.name, plugin.displayName || plugin.name]));
		});
	});

	$effect(() => {
		const id = sceneId;
		void reloadKey;
		fetchSceneComments(id)
			.then((found) => {
				if (id !== sceneId) return;
				comments = found;
				loaded = true;
			})
			.catch(() => (loaded = true));
	});

	let identityCount = $derived(comments.filter(isAboutIdentity).length);
	let threads = $derived.by(() => {
		const all = commentThreads(comments);
		if (!identityOnly) return all;
		return all.filter(
			(thread) => isAboutIdentity(thread.comment) || thread.replies.some(isAboutIdentity)
		);
	});
	let visibleThreads = $derived.by(() => {
		if (expanded || identityOnly) return threads;
		return threads.slice(0, COLLAPSED_THREADS);
	});
	let sites = $derived([...new Set(comments.map((comment) => siteNames.get(comment.plugin) || comment.plugin))]);

	/** Reads the comments from the site again. */
	async function refresh() {
		refreshing = true;
		try {
			comments = await refreshSceneComments(sceneId);
			if (comments.length === 0) notifications.push('The site shows no comments for this video', 'info');
		} catch {
			notifications.push('Could not read the comments from the site', 'error');
		} finally {
			refreshing = false;
		}
	}

	/** Credits a name to the scene and tells the page. */
	async function credit(name: string) {
		const trimmed = name.trim();
		if (!trimmed || crediting) return;
		crediting = trimmed;
		try {
			const performer = await creditPerformerByName(sceneId, trimmed);
			oncredited(performer);
			notifications.push(`Credited ${performer.name}`, 'success');
			namingId = null;
		} catch {
			notifications.push(`Could not credit ${trimmed}`, 'error');
		} finally {
			crediting = null;
		}
	}

	/** Opens a comment's name field, filled with the words selected in it. */
	function startNaming(comment: SceneComment, element: HTMLElement | null) {
		namingId = comment.id;
		typedName = selectedTextWithin(element);
	}

	/** The page's selected text when it lies inside element, else "". */
	function selectedTextWithin(element: HTMLElement | null): string {
		const selection = window.getSelection();
		if (!element || !selection || selection.isCollapsed) return '';
		if (!element.contains(selection.anchorNode) || !element.contains(selection.focusNode)) return '';
		return selection.toString().trim().slice(0, 80);
	}

	/** Whether a mentioned performer is credited on the scene already. */
	function isCredited(performer: CommentMention): boolean {
		return creditedIds.includes(performer.id);
	}
</script>

{#if loaded}
	<section class="flex flex-col gap-4" aria-label="Comments">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<h2 class="flex items-baseline gap-2 text-lg font-semibold">
				Comments
				{#if comments.length > 0}
					<span class="text-base-content/40 text-sm font-normal">
						{comments.length} · {sites.join(', ')}
					</span>
				{/if}
			</h2>
			<div class="flex items-center gap-2">
				{#if identityCount > 0}
					<button
						type="button"
						class="comment-chip"
						class:comment-chip-active={identityOnly}
						aria-pressed={identityOnly}
						onclick={() => (identityOnly = !identityOnly)}
						title="Comments that ask or tell who is in the video"
					>
						Who's who
						<span class="opacity-60">{identityCount}</span>
					</button>
				{/if}
				<button
					type="button"
					class="comment-chip"
					disabled={refreshing}
					onclick={refresh}
					title="Read the comments from the site again"
				>
					{#if refreshing}
						<span class="loading loading-spinner loading-xs"></span>
					{:else}
						<ArrowClockwiseIcon size={14} weight="bold" />
					{/if}
					{#if comments.length === 0}Load comments{:else}Refresh{/if}
				</button>
			</div>
		</div>

		{#if comments.length === 0}
			<p class="text-base-content/40 flex items-center gap-2 text-sm">
				<ChatCircleTextIcon size={16} />
				No comments stored. Load them from the site; viewers often say who is in the video.
			</p>
		{:else}
			<ul class="flex flex-col gap-4">
				{#each visibleThreads as thread (thread.comment.id)}
					<li class="flex flex-col gap-3">
						{@render commentRow(thread.comment)}
						{#if thread.replies.length > 0}
							<ul class="border-base-300 ml-11 flex flex-col gap-3 border-l pl-4">
								{#each thread.replies as reply (reply.id)}
									<li>{@render commentRow(reply)}</li>
								{/each}
							</ul>
						{/if}
					</li>
				{/each}
			</ul>
			{#if !expanded && !identityOnly && threads.length > COLLAPSED_THREADS}
				<button type="button" class="comment-chip w-fit" onclick={() => (expanded = true)}>
					Show all {comments.length} comments
				</button>
			{/if}
		{/if}
	</section>
{/if}

{#snippet commentRow(comment: SceneComment)}
	{@const about = isAboutIdentity(comment)}
	<div class="group flex gap-3">
		<span
			class="bg-base-300 text-base-content/60 flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold"
			aria-hidden="true"
		>
			{initials(comment.author || '?')}
		</span>
		<div class="flex min-w-0 flex-1 flex-col gap-1">
			<p class="text-base-content/50 flex flex-wrap items-center gap-x-2 text-xs">
				<span class="text-base-content/80 font-semibold">{comment.author || 'Anonymous'}</span>
				{#if comment.postedAt}
					<span>{formatRelativeTime(comment.postedAt)}</span>
				{/if}
				{#if about}
					<span class="text-accent font-medium">Who's who</span>
				{/if}
			</p>
			<p class="comment-text text-sm leading-relaxed break-words whitespace-pre-line" data-comment={comment.id}>
				{#each commentSegments(comment.text, comment.mentions) as segment, index (index)}
					{#if segment.performer}
						<a href={performerUrl(segment.performer.id)} class="text-accent font-medium hover:underline">
							{segment.text}
						</a>
					{:else}
						{segment.text}
					{/if}
				{/each}
			</p>
			<div class="text-base-content/50 flex flex-wrap items-center gap-3 text-xs">
				{#if comment.likes}
					<span class="flex items-center gap-1"><ThumbsUpIcon size={12} />{comment.likes}</span>
				{/if}
				{#each comment.mentions.filter((performer) => !isCredited(performer)) as performer (performer.id)}
					<button
						type="button"
						class="hover:text-base-content flex items-center gap-1 font-medium"
						disabled={crediting !== null}
						onclick={() => credit(performer.name)}
					>
						<UserPlusIcon size={12} />
						Credit {performer.name}
					</button>
				{/each}
				<button
					type="button"
					class="hover:text-base-content flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
					class:opacity-100={namingId === comment.id}
					onmousedown={(event) => event.preventDefault()}
					onclick={(event) =>
						startNaming(
							comment,
							(event.currentTarget as HTMLElement).closest('.group')?.querySelector('.comment-text') ?? null
						)}
					title="Select the name in the comment first to fill it in"
				>
					<UserPlusIcon size={12} />
					Credit a name
				</button>
			</div>
			{#if namingId === comment.id}
				<form
					class="flex items-center gap-2"
					onsubmit={(event) => {
						event.preventDefault();
						credit(typedName);
					}}
				>
					<!-- svelte-ignore a11y_autofocus -->
					<input
						type="text"
						bind:value={typedName}
						placeholder="Performer's name"
						class="bg-base-200 border-base-300 h-8 w-56 rounded-lg border px-2 text-sm outline-none"
						autofocus
					/>
					<button type="submit" class="comment-chip" disabled={!typedName.trim() || crediting !== null}>
						Credit
					</button>
					<button type="button" class="text-base-content/50 text-xs" onclick={() => (namingId = null)}>
						Cancel
					</button>
				</form>
			{/if}
		</div>
	</div>
{/snippet}

<style>
	.comment-chip {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		height: 2rem;
		padding-inline: 0.75rem;
		border-radius: 0.5rem;
		background-color: var(--color-base-200);
		font-size: 13px;
		font-weight: 500;
		transition: background-color 150ms;
	}

	.comment-chip:hover:not(:disabled) {
		background-color: var(--color-base-300);
	}

	.comment-chip:disabled {
		opacity: 0.5;
	}

	.comment-chip-active,
	.comment-chip-active:hover:not(:disabled) {
		background-color: var(--color-base-content);
		color: var(--color-base-100);
	}
</style>
