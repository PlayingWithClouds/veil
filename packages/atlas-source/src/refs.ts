const IMAGE_PREFIX = "veil:image:";
const SCENE_PREFIX = "veil:scene:";

export type VeilRef = { type: "image"; remoteUrl: string } | { type: "scene"; sceneId: string };

export function imageRef(remoteUrl: string): string {
  return `${IMAGE_PREFIX}${remoteUrl}`;
}

export function sceneRef(sceneId: string): string {
  return `${SCENE_PREFIX}${sceneId}`;
}

export function parseVeilRef(ref: string): VeilRef | undefined {
  if (ref.startsWith(IMAGE_PREFIX)) {
    return { type: "image", remoteUrl: ref.slice(IMAGE_PREFIX.length) };
  }
  if (ref.startsWith(SCENE_PREFIX)) {
    return { type: "scene", sceneId: ref.slice(SCENE_PREFIX.length) };
  }
  return undefined;
}
