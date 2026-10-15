import type { CamStreamArgs, ResolveResult } from "@playingwithclouds/veil-sdk";
import { API_HEADERS, SITE, UA } from "./http.ts";

/** The endpoint's answer for one room. */
export interface EdgeUrlResponse {
  success?: boolean;
  url?: string;
  room_status?: string;
}

/** Turns the edge-url answer into a playable result, or throws a reason the room cannot be watched. */
export function toStreamResult(username: string, answer: EdgeUrlResponse): ResolveResult {
  if (answer.room_status !== undefined && answer.room_status !== "public") {
    throw new Error(`${username} is not in a public show (${answer.room_status})`);
  }
  if (!answer.success || !answer.url) {
    throw new Error(`${username} is offline`);
  }
  return {
    url: answer.url,
    mime_type: "application/x-mpegURL",
    headers: { Referer: `${SITE}/${username}/`, "User-Agent": UA },
  };
}

/** Asks the site for the room's live HLS master playlist (token included in the URL). */
export async function camStream(args: CamStreamArgs): Promise<ResolveResult> {
  const response = await fetch(`${SITE}/get_edge_hls_url_ajax/`, {
    method: "POST",
    headers: { ...API_HEADERS, "Content-Type": "application/x-www-form-urlencoded", Referer: `${SITE}/${args.username}/` },
    body: new URLSearchParams({ room_slug: args.username, bandwidth: "high" }).toString(),
    signal: AbortSignal.timeout(20000),
  });
  if (!response.ok) throw new Error(`chaturbate stream: HTTP ${response.status}`);
  return toStreamResult(args.username, (await response.json()) as EdgeUrlResponse);
}
