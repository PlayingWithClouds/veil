export const BASE_URL = "https://www.pornpics.com";

export function resolveUrl(path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  return BASE_URL + (path.startsWith("/") ? path : "/" + path);
}

// Gallery URLs are /galleries/<slug>-<gid>/ — the id is the trailing number.
export function galleryIdFromUrl(url: string): string {
  const match = url.match(/-(\d+)\/?(?:[?#].*)?$/);
  return match ? match[1] : "";
}

export function isGalleryUrl(url: string): boolean {
  return /\/galleries\//.test(url);
}

// Thumbnails come as /460/ or /300/ CDN paths; the full-size image is the same
// path under /1280/.
export function fullSizeImage(url: string): string {
  return url.replace(/\/(?:300|460|880)\//, "/1280/");
}

// The search API appends the gallery id to the description; strip it for a clean
// title.
export function cleanTitle(desc: string): string {
  return decodeEntities(desc.replace(/\s+\d+$/, "").trim());
}

export function decodeEntities(text: string): string {
  return text
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#0?39;|&#x27;/gi, "'")
    .replace(/&#(\d+);/g, (_, code) => String.fromCodePoint(parseInt(code, 10)))
    .replace(/&#x([0-9a-f]+);/gi, (_, code) => String.fromCodePoint(parseInt(code, 16)));
}
