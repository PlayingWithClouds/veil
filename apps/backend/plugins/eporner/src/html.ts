import { parse, HTMLElement } from "node-html-parser";
import { SITE, UA } from "./http.ts";

// Fetches a page and parses it, following eporner's canonical-URL redirects.
// Returns null on any non-2xx so callers can degrade gracefully.
export async function fetchDoc(url: string): Promise<HTMLElement | null> {
  const response = await fetch(url, {
    headers: { "User-Agent": UA },
    redirect: "follow",
    signal: AbortSignal.timeout(15_000),
  });
  if (!response.ok) return null;
  return parse(await response.text());
}

// Resolves a site-relative href to an absolute eporner URL.
export function absoluteUrl(href: string): string {
  if (href.startsWith("http")) return href;
  if (href.startsWith("/")) return `${SITE}${href}`;
  return `${SITE}/${href}`;
}

// The real image src, preferring the lazy-loaded data-src over the 1x1 placeholder.
export function imageSrc(img: HTMLElement | null | undefined): string | undefined {
  if (!img) return undefined;
  const dataSrc = img.getAttribute("data-src");
  if (dataSrc && dataSrc.startsWith("http")) return dataSrc;
  const src = img.getAttribute("src");
  if (src && src.startsWith("http")) return src;
  return undefined;
}
