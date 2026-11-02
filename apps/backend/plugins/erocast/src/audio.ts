import { parse } from "node-html-parser";
import type { Audio, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { BASE_URL, fetchText, normalizeText, toSiteUrl, trackIdFromUrl } from "./http.ts";
import { audioOf, tracksFromPage } from "./track.ts";

/** Reads one track page into its audio record. */
export async function scrapeAudio(url: string): Promise<ScrapeResult[]> {
  const trackId = trackIdFromUrl(url);
  if (trackId === "") throw new Error(`erocast: not a track URL: ${url}`);
  // The site routes /track/<id>/<anything>; without a slug it answers 404.
  const html = await fetchText(`${BASE_URL}/track/${trackId}/audio`);
  return [{ type: "audio", audio: audioFromTrackPage(html, trackId) }];
}

/** The audio of the track a page is about; a page also embeds the tracks it recommends. */
export function audioFromTrackPage(html: string, trackId: string): Audio {
  const track = tracksFromPage(html).find((candidate) => String(candidate.id) === trackId);
  if (track === undefined) throw new Error(`erocast: track ${trackId} not found on its page`);
  track.permalink_url = toSiteUrl(track.permalink_url);
  return audioOf(track, genreNamesOf(html));
}

/** The genres listed under the track's title, e.g. Fdom or Narrative. */
export function genreNamesOf(html: string): string[] {
  const names: string[] = [];
  const holder = parse(html).querySelector("li.tags");
  if (holder === null) return names;
  for (const genre of holder.querySelectorAll(".genre-link .tag")) {
    names.push(normalizeText(genre.text));
  }
  return names;
}
