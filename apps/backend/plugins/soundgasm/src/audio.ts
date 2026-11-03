import { parse } from "node-html-parser";
import type { Audio, ScrapeResult } from "@playingwithclouds/veil-sdk";
import { audioUrl, externalIdOf, fetchPage, normalizeText, parseAddress, voiceOf } from "./site.ts";
import { splitTitle } from "./titles.ts";

/** The m4a address the page's player script plays, or an empty string. */
export function mediaUrlOf(html: string): string {
  const match = html.match(/m4a:\s*"([^"]+)"/);
  if (!match) return "";
  return match[1];
}

/** Builds the Audio of one audio page, or undefined when the page has no player. */
export function audioFromPage(html: string, user: string, slug: string): Audio | undefined {
  const mediaUrl = mediaUrlOf(html);
  if (mediaUrl === "") return undefined;

  const root = parse(html);
  const rawTitle = root.querySelector(".jp-title")?.text ?? "";
  const { title, tags } = splitTitle(rawTitle);
  const details = normalizeText(root.querySelector(".jp-description")?.text ?? "");

  const audio: Audio = {
    type: "audio",
    external_id: externalIdOf(user, slug),
    source_url: audioUrl(user, slug),
    title: title || slug,
    performers: [voiceOf(user)],
    downloads: [{ label: "m4a", url: mediaUrl, format: "m4a" }],
  };
  if (details !== "") audio.details = details;
  if (tags.length > 0) audio.tags = tags;
  return audio;
}

/** Scrapes one audio page (soundgasm.net/u/<user>/<slug>). */
export async function findAudio(url: string): Promise<ScrapeResult[]> {
  const address = parseAddress(url);
  if (!address?.slug) return [];

  const html = await fetchPage(audioUrl(address.user, address.slug));
  const audio = audioFromPage(html, address.user, address.slug);
  if (!audio) return [];
  return [{ type: "audio", audio }];
}
