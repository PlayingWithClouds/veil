import type { FaceIdentifyArgs, FaceIdentifyResult } from "@playingwithclouds/veil-sdk";
import { parseCamMatches, parseVideoMatches } from "./parse.ts";

const API_ORIGIN = "https://facedigger.com";
const AGREEMENT_SETTING = "FACEDIGGER_ACCEPT_AGREEMENT";

/** The site compresses to ~300 KB; refuse anything far beyond that. */
const MAX_BASE64_CHARS = 1_000_000;
const POLL_INTERVAL_MS = 1500;
const MAX_POLLS = 40;
const REQUEST_TIMEOUT_MS = 30_000;

/** Whether the owner confirmed FaceDigger's user agreement in the plugin settings. */
export function agreementAccepted(): boolean {
  const value = String(process.env[AGREEMENT_SETTING] || "").trim().toLowerCase();
  return value === "true" || value === "1" || value === "yes" || value === "on";
}

/** Waits before the next status check. */
function sleep(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

/** Calls the site's JSON API the way its own page does (same-origin headers, no key). */
async function callApi(path: string, body?: unknown): Promise<any> {
  const response = await fetch(`${API_ORIGIN}${path}`, {
    method: body === undefined ? "GET" : "POST",
    headers: { "Content-Type": "application/json", Referer: `${API_ORIGIN}/` },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });
  if (response.status === 429) throw new Error("FaceDigger is rate limiting requests, try again in a minute");
  if (!response.ok) throw new Error(`HTTP ${response.status} from facedigger`);
  return response.json();
}

/**
 * Searches FaceDigger like its page does: create a job with the picture,
 * poll its status until completed, fetch the result. Cam models come back as
 * named candidates, tube-site videos as links without names.
 */
export async function identifyFace(args: FaceIdentifyArgs): Promise<FaceIdentifyResult> {
  if (!agreementAccepted()) {
    throw new Error(
      "FaceDigger's user agreement has not been accepted: confirm it in this plugin's settings first (it excludes EU residents)",
    );
  }
  if (args.image_base64.length > MAX_BASE64_CHARS) throw new Error("picture is too large for FaceDigger");

  const picture = `data:${args.mime_type};base64,${args.image_base64}`;
  const job = await callApi("/api/face-search", { image: picture, thumbnail: picture });
  if (!job.success || !job.job_id) throw new Error(String(job.message || "FaceDigger refused the picture"));

  for (let poll = 0; poll < MAX_POLLS; poll++) {
    const status = await callApi(`/api/job-status/${job.job_id}`);
    if (status.status === "completed" || status.status === "complete") {
      const result = await callApi(`/api/job-result/${job.job_id}`);
      return { candidates: parseCamMatches(result), videos: parseVideoMatches(result) };
    }
    if (status.status === "failed" || status.success === false) {
      throw new Error(String(status.message || "FaceDigger could not process the picture"));
    }
    await sleep(POLL_INTERVAL_MS);
  }
  throw new Error("FaceDigger took too long");
}
