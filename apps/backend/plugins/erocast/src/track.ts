import type { Audio, DiscoveredItem, Download, ScenePerformer } from "@playingwithclouds/veil-sdk";
import { decodeEntities, externalIdOf, normalizeText } from "./http.ts";

/** The part of the site's track JSON the plugin reads; the site sends many more fields. */
export interface RawTrack {
  id: number;
  title: string;
  description?: string | null;
  duration?: number | null;
  created_at?: string | null;
  artwork_url?: string | null;
  permalink_url: string;
  stream_url?: string | null;
  file_url?: string | null;
  /** 1 once the track is converted to an HLS playlist */
  hls?: number;
  /** 1 for a track served as a signed mp3 */
  mp3?: number;
  plays?: number | null;
  tags?: { tag?: string }[] | null;
  user?: { name?: string; username?: string; permalink_url?: string } | null;
}

const SONG_DATA = /var song_data_(\d+) = (\{[^\n]*?\})<\/script>/;
const PLACEHOLDER_COVER = "/common/default/";

/**
 * The tracks a page or listing fragment embeds. Each card carries its track as a
 * `song_data_<id>` script, which holds more than the card shows (tags, uploader).
 */
export function tracksFromPage(html: string): RawTrack[] {
  const tracks: RawTrack[] = [];
  const pattern = new RegExp(SONG_DATA.source, "g");
  let match = pattern.exec(html);
  while (match !== null) {
    tracks.push(JSON.parse(match[2]) as RawTrack);
    match = pattern.exec(html);
  }
  return tracks;
}

/** The downloads of a track; none while the site is still converting a new upload. */
export function downloadsOf(track: RawTrack): Download[] {
  if (track.hls === 1 && track.stream_url) {
    return [{ label: "HLS", url: track.stream_url, format: "hls" }];
  }
  if (track.mp3 === 1 && track.stream_url) {
    return [{ label: "MP3", url: track.stream_url, format: "mp3" }];
  }
  return [];
}

/** The listing card of a track. */
export function cardOf(track: RawTrack): DiscoveredItem {
  const card: DiscoveredItem = {
    title: titleOf(track),
    media_type: "audio",
    source_url: track.permalink_url,
    external_id: externalIdOf(track.id),
  };
  const date = dateOf(track);
  if (date !== undefined) card.date = date;
  const cover = coverOf(track);
  if (cover !== undefined) card.poster_path = cover;
  if (track.duration) card.duration_seconds = track.duration;
  if (track.plays) card.view_count = track.plays;
  const downloads = downloadsOf(track);
  if (downloads.length > 0) card.downloads = downloads;
  return card;
}

/** The audio record of a track; genres are the page's genre names, which the track JSON only has as ids. */
export function audioOf(track: RawTrack, genres: string[]): Audio {
  const audio: Audio = {
    type: "audio",
    external_id: externalIdOf(track.id),
    source_url: track.permalink_url,
    title: titleOf(track),
    tags: tagsOf(track, genres),
    downloads: downloadsOf(track),
  };
  const details = detailsOf(track);
  if (details !== undefined) audio.details = details;
  const date = dateOf(track);
  if (date !== undefined) audio.date = date;
  const cover = coverOf(track);
  if (cover !== undefined) audio.cover_path = cover;
  if (track.duration) audio.duration_seconds = track.duration;
  const voice = voiceOf(track);
  if (voice !== undefined) audio.performers = [voice];
  return audio;
}

/** The title without the HTML entities the site leaves in it. */
function titleOf(track: RawTrack): string {
  return normalizeText(decodeEntities(track.title));
}

/** The description, entities decoded, line breaks kept; undefined when the track has none. */
function detailsOf(track: RawTrack): string | undefined {
  if (!track.description) return undefined;
  const details = decodeEntities(track.description).trim();
  if (details === "") return undefined;
  return details;
}

/** The upload day as an ISO date; undefined when the track has no timestamp. */
function dateOf(track: RawTrack): string | undefined {
  if (!track.created_at) return undefined;
  return track.created_at.slice(0, 10);
}

/** The cover image; undefined for the site's grey placeholder. */
function coverOf(track: RawTrack): string | undefined {
  if (!track.artwork_url) return undefined;
  if (track.artwork_url.includes(PLACEHOLDER_COVER)) return undefined;
  return track.artwork_url;
}

/** The uploader, who is also the voice; undefined when the track JSON names none. */
function voiceOf(track: RawTrack): ScenePerformer | undefined {
  const user = track.user;
  if (!user || !user.name || !user.username) return undefined;
  const voice: ScenePerformer = { name: user.name, external_id: user.username };
  if (user.permalink_url) voice.source_url = user.permalink_url;
  return voice;
}

/** Hashtags and genre names, each once, lower-cased the way the site's tag pages spell them. */
function tagsOf(track: RawTrack, genres: string[]): string[] {
  const tags: string[] = [];
  const seen = new Set<string>();
  const names = [...genres, ...hashtagsOf(track)];
  for (const name of names) {
    const key = name.toLowerCase();
    if (key === "" || seen.has(key)) continue;
    seen.add(key);
    tags.push(name);
  }
  return tags;
}

/** The hashtag texts of a track. */
function hashtagsOf(track: RawTrack): string[] {
  const names: string[] = [];
  if (!track.tags) return names;
  for (const entry of track.tags) {
    if (entry.tag) names.push(normalizeText(entry.tag));
  }
  return names;
}
