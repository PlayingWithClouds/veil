import { parse, type HTMLElement } from "node-html-parser";
import type { FaceCandidate } from "@playingwithclouds/veil-sdk";

export const SITE_ORIGIN = "https://pornstarbyface.com";

/** Turns the site's percentage text ("45") into a 0..1 similarity. */
function similarityFrom(percentText: string | undefined): number {
  const percent = Number(percentText);
  if (!Number.isFinite(percent)) return 0;
  return Math.min(Math.max(percent / 100, 0), 1);
}

/** Resolves an image path of the site ("/ImgFiles/Lexi Lore/1.jpg") to an absolute, encoded URL. */
function absoluteImageUrl(source: string | undefined): string | undefined {
  if (!source) return undefined;
  if (source.startsWith("http")) return encodeURI(source);
  if (!source.startsWith("/")) return undefined;
  return encodeURI(SITE_ORIGIN + source);
}

/** Keeps only absolute http(s) links to a performer's pages on other sites (not tracking redirects). */
function profileUrlsFrom(candidateElement: HTMLElement): string[] {
  const urls: string[] = [];
  for (const link of candidateElement.querySelectorAll("a.candidate-link")) {
    const href = link.getAttribute("href");
    if (!href || !href.startsWith("https://")) continue;
    if (urls.includes(href)) continue;
    urls.push(href);
  }
  return urls;
}

/**
 * Parses the HTML fragment of candidate blocks the site answers an upload with.
 * Returns the candidates best first; an error answer ({"error":true,...}) throws.
 */
export function parseCandidates(body: string): FaceCandidate[] {
  const trimmed = body.trim();
  if (trimmed.startsWith("{")) {
    const failure = JSON.parse(trimmed) as { errorMsg?: string };
    throw new Error(failure.errorMsg || "pornstarbyface rejected the picture");
  }

  const candidates: FaceCandidate[] = [];
  for (const block of parse(body).querySelectorAll(".realCandidate")) {
    const name = block.querySelector(".candidate-main p")?.text.trim();
    if (!name) continue;
    const similarityText = block.querySelector(".progress-bar")?.getAttribute("similarity");
    const candidate: FaceCandidate = {
      name,
      similarity: similarityFrom(similarityText),
      profile_urls: profileUrlsFrom(block),
    };
    const imageUrl = absoluteImageUrl(block.querySelector("img.img-thumbnail")?.getAttribute("src"));
    if (imageUrl) candidate.image_url = imageUrl;
    candidates.push(candidate);
  }
  return candidates.sort((left, right) => right.similarity - left.similarity);
}
