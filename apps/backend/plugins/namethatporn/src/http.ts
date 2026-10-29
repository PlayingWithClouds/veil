import { SITE_ORIGIN } from "./parse.ts";

export const SEARCH_URL = `${SITE_ORIGIN}/search/images.html`;

const REQUEST_TIMEOUT_MS = 30_000;

/** Cloudflare answers 403 without a browser's User-Agent. */
const USER_AGENT =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

/** Fetches a page of the site the way a browser on it would. */
export async function fetchPage(url: string, init: RequestInit = {}): Promise<string> {
  const response = await fetch(url, {
    ...init,
    headers: { Referer: SEARCH_URL, "User-Agent": USER_AGENT },
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} from namethatporn`);
  return response.text();
}
