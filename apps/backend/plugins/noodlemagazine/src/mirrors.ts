import { BROWSER_HEADERS, fetchHtml, fetchHtmlSmart, hasSolver } from "@playingwithclouds/veil-sdk";

/** One copy of the site: the same videos under the same ids on another domain. */
export interface Mirror {
  origin: string;
  /** Whether the domain answers plain requests with a Cloudflare challenge. */
  needsSolver: boolean;
}

/**
 * The site's copies in the order they are tried: a request goes to the first
 * that answers with a page of the site. noodlemagazine.com and ukdevilz.com sit
 * behind Cloudflare and are skipped without FlareSolverr (e.g. on a phone);
 * mat6tube.com answers plain requests.
 */
export const MIRRORS: Mirror[] = [
  { origin: "https://noodlemagazine.com", needsSolver: true },
  { origin: "https://ukdevilz.com", needsSolver: true },
  { origin: "https://mat6tube.com", needsSolver: false },
];

/** The origin the plugin's identities (source URLs) use, whichever copy served the page. */
export const CANONICAL_ORIGIN = MIRRORS[0].origin;

/** A page and the mirror that served it. */
export interface MirrorPage {
  body: string;
  origin: string;
}

/** Decides whether a body is the page asked for, not a challenge or error page. */
export type PageCheck = (body: string) => boolean;

/** Whether a body is an HTML page of the site: each copy links its own static assets. */
export function isSitePage(body: string): boolean {
  return body.includes("/static/extend/");
}

/** The mirrors a request may try here: those behind Cloudflare only with a solver. */
export function usableMirrors(solverAvailable: boolean): Mirror[] {
  return MIRRORS.filter((mirror) => solverAvailable || !mirror.needsSolver);
}

/**
 * Fetches a path (with its query) from the first mirror that serves it, in
 * MIRRORS order; a mirror that fails or answers something check rejects is
 * passed over for the next. Throws when none served it.
 */
export async function fetchFromMirrors(path: string, check: PageCheck = isSitePage): Promise<MirrorPage> {
  const failures: string[] = [];
  for (const mirror of usableMirrors(hasSolver())) {
    try {
      const body = await fetchFrom(mirror, path);
      if (check(body)) return { body, origin: mirror.origin };
      failures.push(`${mirror.origin}: not a page of the site`);
    } catch (error) {
      failures.push(`${mirror.origin}: ${(error as Error).message}`);
    }
  }
  throw new Error(`noodlemagazine: no mirror served ${path} (${failures.join("; ")})`);
}

/** Fetches a path from one mirror, through FlareSolverr when the mirror needs it. */
function fetchFrom(mirror: Mirror, path: string): Promise<string> {
  const url = mirror.origin + path;
  if (mirror.needsSolver) return fetchHtmlSmart(url);
  return fetchHtml(url);
}

/** The headers a mirror's CDN links are played with. */
export function playbackHeaders(origin: string): Record<string, string> {
  return { Referer: origin + "/", "User-Agent": BROWSER_HEADERS["User-Agent"] };
}

/** The path and query of a URL on any mirror, or undefined when it is not on one. */
export function mirrorPath(url: string): string | undefined {
  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    return undefined;
  }
  if (!MIRRORS.some((mirror) => isMirrorHost(parsed.hostname, mirror))) return undefined;
  return parsed.pathname + parsed.search;
}

/** Whether a hostname (with or without www) is the mirror's. */
function isMirrorHost(hostname: string, mirror: Mirror): boolean {
  const host = new URL(mirror.origin).hostname;
  return hostname === host || hostname === `www.${host}`;
}
