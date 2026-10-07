<!--
	Action sheet for a scene card, opened by a long press or the card's ⋮ button
	(sceneSheet store). Mounted once in the layout on phones.
-->
<script lang="ts">
	import ClockIcon from 'phosphor-svelte/lib/ClockIcon';
	import FolderSimplePlusIcon from 'phosphor-svelte/lib/FolderSimplePlusIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import ProhibitIcon from 'phosphor-svelte/lib/ProhibitIcon';
	import ActionSheet, { type SheetAction } from './ActionSheet.svelte';
	import { quickSaveEntity } from '$lib/collectionActions';
	import { openMoveDialog } from '$lib/stores/moveDialog';
	import { downloadDefault } from '$lib/cardActions';
	import { clearVerdict, setVerdict } from '$lib/rating';
	import { notifications } from '$lib/stores/notifications';
	import {
		closeSceneSheet,
		sceneSheetTarget,
		setSceneDismissed,
		type SceneSheetTarget
	} from '$lib/stores/sceneSheet';

	/** The sheet's actions for one scene. */
	function actionsFor(target: SceneSheetTarget): SheetAction[] {
		return [
			{
				label: 'Save to Watch later',
				icon: ClockIcon,
				run: () => quickSaveEntity('scene', target.sceneId, target.title)
			},
			{
				label: 'Save to collection…',
				icon: FolderSimplePlusIcon,
				run: () => openMoveDialog('scene', target.sceneId, target.title)
			},
			{ label: 'Download', icon: DownloadSimpleIcon, run: () => download(target) },
			{
				label: 'Not interested',
				icon: ProhibitIcon,
				run: () => markNotInterested(target),
				destructive: true
			}
		];
	}

	/** Starts the default download and reports it. */
	async function download(target: SceneSheetTarget) {
		try {
			await downloadDefault(target.sceneId, target.title);
			notifications.push(`Download started: ${target.title}`, 'success');
		} catch (error) {
			notifications.push(String(error), 'error');
		}
	}

	/** Dislikes the scene, collapses its card and offers an undo. */
	async function markNotInterested(target: SceneSheetTarget) {
		setSceneDismissed(target.sceneId, true);
		try {
			await setVerdict(target.sceneId, 'down');
			notifications.pushAction(
				"Got it — you'll see less like this",
				{ label: 'Undo', run: () => undoNotInterested(target.sceneId) },
				'info'
			);
		} catch {
			setSceneDismissed(target.sceneId, false);
			notifications.push('Could not save that', 'error');
		}
	}

	/** Takes back a "Not interested". */
	async function undoNotInterested(sceneId: string) {
		setSceneDismissed(sceneId, false);
		await clearVerdict(sceneId);
	}
</script>

{#if $sceneSheetTarget}
	<ActionSheet
		title={$sceneSheetTarget.title}
		subtitle={$sceneSheetTarget.subtitle}
		imageUrl={$sceneSheetTarget.imageUrl}
		actions={actionsFor($sceneSheetTarget)}
		onclose={closeSceneSheet}
	/>
{/if}
