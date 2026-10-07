import { HTMLElement } from "node-html-parser";
import { decodeEntities } from "./http.ts";

export interface StreamSource {
  url: string;
  height: number;
  quality: string;
  format: "mp4";
}

// tnaflix inlines every playable quality as a <source> tag on the video page,
// each signed with an expiring token and carrying its vertical resolution in the
// `size` attribute (e.g. size="720"). Return them all, highest resolution first.
export function extractSources(root: HTMLElement): StreamSource[] {
  const sources: StreamSource[] = [];

  for (const element of root.querySelectorAll("video source[src]")) {
    const src = element.getAttribute("src") ?? "";
    if (!src) continue;
    const height = parseInt(element.getAttribute("size") ?? "0", 10);
    const safeHeight = Number.isFinite(height) ? height : 0;
    sources.push({
      url: decodeEntities(src),
      height: safeHeight,
      quality: safeHeight > 0 ? `${safeHeight}p` : "unknown",
      format: "mp4",
    });
  }

  sources.sort((a, b) => b.height - a.height);
  return sources;
}
