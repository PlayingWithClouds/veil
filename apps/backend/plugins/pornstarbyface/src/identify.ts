import { base64ToBytes } from "@playingwithclouds/veil-sdk";
import type { FaceIdentifyArgs, FaceIdentifyResult } from "@playingwithclouds/veil-sdk";
import { parseCandidates, SITE_ORIGIN } from "./parse.ts";

/** Endpoint of the performer database (the cam-model database is a different endpoint and is not used). */
const UPLOAD_URL = `${SITE_ORIGIN}/Home/LooksLikeByPhoto`;

/** The site accepts pictures up to 6 MB. */
const MAX_IMAGE_BYTES = 6 * 1024 * 1024;

const REQUEST_TIMEOUT_MS = 30_000;

/** Uploads the picture to the site and returns the performers it looks like. */
export async function identifyFace(args: FaceIdentifyArgs): Promise<FaceIdentifyResult> {
  const imageBytes = base64ToBytes(args.image_base64);
  if (imageBytes.length > MAX_IMAGE_BYTES) throw new Error("picture is larger than 6 MB");

  const form = new FormData();
  form.append("imageUploadForm", new Blob([imageBytes], { type: args.mime_type }), "face.jpg");
  const response = await fetch(UPLOAD_URL, {
    method: "POST",
    headers: { Referer: `${SITE_ORIGIN}/`, "X-Requested-With": "XMLHttpRequest" },
    body: form,
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} from pornstarbyface`);
  return { candidates: parseCandidates(await response.text()) };
}
