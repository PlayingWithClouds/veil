import { SITE, UA, idFromUrl, embedUrl } from "./http.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

// eporner streams live on a token-gated CDN. Primary path: the video-sources XHR
// (needs a hash derived from the page). Fallback: the direct CDN mp4 embedded in
// the player page. Both are returned with an eporner Referer so the CDN serves
// them.
export async function resolve(url: string): Promise<ResolveResult> {
  const id = idFromUrl(url);
  if (!id) throw new Error(`eporner resolve: no video id in ${url}`);

  const page = embedUrl(id);
  const html = await fetchPage(page);

  const viaXhr = await resolveViaXhr(id, html, page, wantedHeight(url)).catch(() => null);
  if (viaXhr) return viaXhr;

  const mp4 = html.match(/https?:\/\/[a-z0-9.-]*eporner\.com\/[^\s"'\\]+?\.mp4/i)?.[0];
  if (mp4 && !mp4.includes("/na.mp4")) {
    return { url: mp4, mime_type: "video/mp4", headers: refererHeaders() };
  }
  throw new Error(`eporner resolve: no source for ${url}`);
}

interface Source {
  src: string;
  height: number;
  quality: string;
}

/** The height a `#quality=360p` hint on the stream URL asks for; undefined for the best. */
export function wantedHeight(url: string): number | undefined {
  const match = url.match(/#quality=(\d+)p?$/);
  return match ? Number(match[1]) : undefined;
}

/** A stream URL that resolves to the given height (see wantedHeight). */
export function qualityUrl(url: string, height: number): string {
  return `${url}#quality=${height}p`;
}

async function resolveViaXhr(id: string, html: string, referer: string, wanted?: number): Promise<ResolveResult | null> {
  const rawHash = html.match(/player\.hash\s*=\s*['"]([0-9a-f]{32})['"]/i)?.[1];
  if (!rawHash) return null;

  // Query params mirror the player's XHR (vjs*.js). The signed sources are bound
  // to the requesting IP, so the backend that resolves must also fetch the stream.
  const query = new URLSearchParams({
    hash: calcHash(rawHash),
    domain: "www.eporner.com",
    pixelRatio: "1",
    playerWidth: "640",
    playerHeight: "360",
    fallback: "false",
    embed: "true",
    supportedFormats: "hls,mp4",
    _: String(Date.now()),
  }).toString();

  const res = await fetch(`${SITE}/xhr/video/${id}?${query}`, {
    headers: {
      "User-Agent": UA,
      "X-Requested-With": "XMLHttpRequest",
      Referer: referer,
      Cookie: "age_verified=1",
    },
    signal: AbortSignal.timeout(12_000),
  });
  if (!res.ok) return null;

  const data: any = await res.json();
  const best = pickMp4(data?.sources?.mp4, wanted);
  if (!best) return null;
  return {
    url: best.src,
    mime_type: "video/mp4",
    quality: best.quality,
    headers: refererHeaders(),
  };
}

// Reverse-engineered from eporner's player (vjs*.js): split the 32-hex page hash
// into four 8-char chunks, parse each as a base-16 int, and re-encode it in
// base-36; concatenate the four.
function calcHash(hash: string): string {
  let out = "";
  for (let offset = 0; offset < 32; offset += 8) {
    out += parseInt(hash.slice(offset, offset + 8), 16).toString(36);
  }
  return out;
}

// sources.mp4 is a label→{src,labelShort} map ("1080p HD", "720p HD", "auto",
// ...). Pick the highest real resolution, ignoring the na.mp4 placeholder; with
// `wanted`, the highest at or below that height (else the lowest above it).
export function pickMp4(
  mp4: Record<string, { src?: string; labelShort?: string }> | undefined,
  wanted?: number,
): Source | null {
  if (!mp4) return null;
  const sources: Source[] = [];
  for (const [label, entry] of Object.entries(mp4)) {
    const src = entry?.src;
    if (!src || src.includes("/na.mp4")) continue;
    const quality = entry.labelShort || label;
    const height = parseInt(quality, 10);
    sources.push({ src, height: Number.isFinite(height) ? height : 0, quality });
  }
  if (sources.length === 0) return null;
  sources.sort((a, b) => b.height - a.height);
  if (wanted === undefined) return sources[0];
  const fitting = sources.find((source) => source.height > 0 && source.height <= wanted);
  return fitting ?? sources.filter((source) => source.height > 0).at(-1) ?? sources[0];
}

function refererHeaders(): Record<string, string> {
  return { Referer: `${SITE}/`, "User-Agent": UA };
}

async function fetchPage(url: string): Promise<string> {
  const res = await fetch(url, {
    headers: {
      "User-Agent": UA,
      Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
      "Accept-Language": "en-US,en;q=0.9",
      Cookie: "age_verified=1",
    },
    signal: AbortSignal.timeout(12_000),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status} fetching ${url}`);
  return res.text();
}
