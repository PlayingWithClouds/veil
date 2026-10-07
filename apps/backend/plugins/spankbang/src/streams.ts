// Highest to lowest, so callers can present or pick sources in a stable order.
const QUALITY_ORDER = ["4k", "2160p", "1080p", "720p", "480p", "320p", "240p"];

export interface StreamSource {
  url: string;
  quality: string;
  format: "mp4" | "hls";
}

// spankbang embeds every progressive quality in a `var stream_data = { '<q>':
// ['<mp4>'], ... }` object plus one or more adaptive HLS masters. Return all of
// them, highest quality first, mp4 before HLS.
export function extractSources(html: string): StreamSource[] {
  const sources: StreamSource[] = [];
  for (const mp4 of extractMp4Sources(html)) sources.push(mp4);

  const master = pickMaster(html);
  if (master) sources.push({ url: master, quality: "auto", format: "hls" });

  return sources;
}

function extractMp4Sources(html: string): StreamSource[] {
  const block = html.match(/var\s+stream_data\s*=\s*\{([\s\S]*?)\};/);
  if (!block) return [];
  const body = block[1];

  const sources: StreamSource[] = [];
  for (const quality of QUALITY_ORDER) {
    const entry = new RegExp(`'${quality}'\\s*:\\s*\\[([^\\]]*)\\]`).exec(body);
    if (!entry) continue;
    const urlMatch = entry[1].match(/'(https?:\/\/[^']+\.mp4[^']*)'/);
    if (urlMatch) sources.push({ url: urlMatch[1], quality, format: "mp4" });
  }
  return sources;
}

// The video's own master playlist carries a _tid marker; prefer the multi-
// quality one (its path lists more than one resolution).
function pickMaster(html: string): string | null {
  const masters = Array.from(
    html.matchAll(/https:\/\/[^\s"'\\]+?master\.m3u8[^\s"'\\]*_tid=\d+/g),
    (match) => match[0]
  );
  if (masters.length === 0) return null;

  masters.sort((a, b) => resolutionCount(b) - resolutionCount(a));
  return masters[0];
}

function resolutionCount(url: string): number {
  return (url.match(/\d{3,4}p/g) ?? []).length;
}
