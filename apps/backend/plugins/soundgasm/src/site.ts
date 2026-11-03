import type { StudioRef } from "@playingwithclouds/veil-sdk";
import { fetchHtml } from "@playingwithclouds/veil-sdk";

export const SITE_URL = "https://soundgasm.net";

/** A Soundgasm address: an uploader's page, or one of their audios when the slug is set. */
export interface SiteAddress {
  user: string;
  slug?: string;
}

/** Fetches a page with a plain request; the site answers without a challenge. */
export function fetchPage(url: string): Promise<string> {
  return fetchHtml(url);
}

/** Reads the uploader and audio slug out of a soundgasm.net/u/... URL, or undefined for other URLs. */
export function parseAddress(url: string): SiteAddress | undefined {
  const match = url.match(/^https?:\/\/(?:www\.)?soundgasm\.net\/u\/([^/?#]+)(?:\/([^/?#]+))?/i);
  if (!match) return undefined;
  return { user: match[1], slug: match[2] };
}

/** The page listing all audios of an uploader. */
export function userUrl(user: string): string {
  return `${SITE_URL}/u/${user}`;
}

/** The page of one audio. */
export function audioUrl(user: string, slug: string): string {
  return `${SITE_URL}/u/${user}/${slug}`;
}

/** The uploader as the voice credit; the audio page and the listing card share it so both resolve to one performer. */
export function voiceOf(user: string): StudioRef {
  return { name: user, external_id: user, source_url: userUrl(user) };
}

/** The identity an audio carries; slugs are only unique within an uploader. */
export function externalIdOf(user: string, slug: string): string {
  return `soundgasm-${user}/${slug}`;
}

/** Collapses whitespace runs and trims. */
export function normalizeText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}
