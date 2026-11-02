import type { SceneMarkerInput } from "@playingwithclouds/veil-sdk";

/**
 * Pornhub's timed acts: the player config's `actionTags` is a comma list of
 * `Name:startSecond` pairs ("Blowjob:379,Reverse Cowgirl:921"), empty when the
 * video has none. The site gives start times only, so the markers are points.
 */
export function actionTagMarkers(actionTags: unknown): SceneMarkerInput[] {
  if (typeof actionTags !== "string") return [];
  const markers: SceneMarkerInput[] = [];
  for (const entry of actionTags.split(",")) {
    const marker = actionTagMarker(entry);
    if (marker) markers.push(marker);
  }
  return markers;
}

/** One `Name:seconds` entry, undefined when it isn't that shape. */
function actionTagMarker(entry: string): SceneMarkerInput | undefined {
  const split = entry.lastIndexOf(":");
  if (split === -1) return undefined;
  const tag = entry.slice(0, split).trim();
  const seconds = Number(entry.slice(split + 1).trim());
  if (tag === "" || !Number.isFinite(seconds) || seconds < 0) return undefined;
  return { tag, seconds };
}
