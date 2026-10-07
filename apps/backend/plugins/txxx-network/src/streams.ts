import { base164Decode, baseUrl } from "./http.ts";

const UA =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

interface VideoFile {
  format: string;
  video_url: string;
  is_default?: number;
}

export interface StreamSource {
  url: string;
  quality: string;
  format: "mp4";
  // The site's raw format tag (e.g. "_hq"), used to match a re-derived source.
  tag: string;
  isDefault: boolean;
}

// Friendly labels for the site's opaque format tags.
const QUALITY_LABELS: Record<string, string> = {
  _hq: "HD",
  _hd: "HD",
  _uq: "UHD",
  _lq: "SD",
  _vl: "Low",
};

// Rank used to order sources best-first.
const QUALITY_RANK = ["_uq", "_hd", "_hq", "_lq", "_vl"];

// videofile.php returns one entry per available format, each with a base164-
// obfuscated URL. Decode every non-trailer format into a playable /get_file/
// path, ordered best quality first.
export async function fetchSources(host: string, id: string, referer: string): Promise<StreamSource[]> {
  const files = await fetchVideoFiles(host, id, referer);
  const playable = files.filter((file) => !file.format.includes("_tr"));

  const sources = playable.map((file) => ({
    url: baseUrl(host) + base164Decode(file.video_url),
    quality: qualityLabel(file.format),
    format: "mp4" as const,
    tag: formatTag(file.format),
    isDefault: file.is_default === 1,
  }));

  sources.sort((a, b) => {
    if (a.isDefault !== b.isDefault) return a.isDefault ? -1 : 1;
    return rankOf(b) - rankOf(a);
  });
  return sources;
}

async function fetchVideoFiles(host: string, id: string, referer: string): Promise<VideoFile[]> {
  const target = `${baseUrl(host)}/api/videofile.php?video_id=${id}&lifetime=8640000`;
  const res = await fetch(target, {
    headers: {
      "User-Agent": UA,
      "X-Requested-With": "XMLHttpRequest",
      Referer: referer,
    },
    signal: AbortSignal.timeout(15_000),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status} fetching videofile for ${id}`);

  const data = (await res.json()) as VideoFile[] | { error?: number; msg?: string };
  if (!Array.isArray(data)) {
    throw new Error(`txxx-network: videofile error ${(data as { msg?: string }).msg}`);
  }
  return data;
}

// The bare format tag, e.g. "_hq" from "_hq.mp4".
function formatTag(format: string): string {
  const match = format.match(/(_[a-z]+)/i);
  return match ? match[1] : format;
}

function qualityLabel(format: string): string {
  for (const [tag, label] of Object.entries(QUALITY_LABELS)) {
    if (format.includes(tag)) return label;
  }
  return format.replace(/[^a-z0-9]/gi, "") || "source";
}

function rankOf(source: { quality: string }): number {
  const index = QUALITY_RANK.findIndex((tag) => QUALITY_LABELS[tag] === source.quality);
  return index === -1 ? -1 : QUALITY_RANK.length - index;
}
