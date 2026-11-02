// Plugin accounts: the session the backend hands a plugin with a login, and
// cookie jars for sites that log in with cookies. A login (account:login)
// fills a fresh jar through jarFetch and returns it as the session; later
// calls fetch through sessionFetch, which sends the jar's cookies and keeps
// what the site sets, so a site that rotates its cookies stays logged in.
import type { AccountArgs, AccountResult, Cookie, CookieJar } from "./types.ts";

/** Most redirects jarFetch follows before it gives up. */
const MAX_REDIRECTS = 10;

/**
 * Thrown when the site wants a login the plugin doesn't have, or no longer
 * takes its session. The backend logs in again with the stored login and
 * retries the call once; without one, the user is asked to log in.
 */
export class NeedsLoginError extends Error {
  constructor(message = "This needs a login") {
    super(message);
    this.name = "NeedsLoginError";
  }
}

/**
 * The session of the plugin's account on this device (what account:login
 * returned), or undefined when the user hasn't logged in.
 */
export function accountSession(): AccountArgs | undefined {
  const encoded = process.env.VEIL_ACCOUNT;
  if (encoded === undefined || encoded === "") {
    return undefined;
  }
  return JSON.parse(encoded) as AccountArgs;
}

// The account's jar, read once per call, and whether the site changed it.
let accountJarCache: CookieJar | undefined;
let accountJarRenewed = false;

/** A jar without cookies, to log in with. */
export function newJar(): CookieJar {
  return { cookies: [] };
}

/** A jar as the session string the backend keeps. */
export function serializeJar(jar: CookieJar): string {
  return JSON.stringify(jar);
}

/** The jar a session string holds; empty when it holds none (not logged in, or not cookies). */
export function jarOf(session: string | undefined): CookieJar {
  if (session === undefined || session === "") return newJar();
  try {
    const parsed = JSON.parse(session) as Partial<CookieJar>;
    if (Array.isArray(parsed.cookies)) return { cookies: parsed.cookies };
  } catch {
    return newJar();
  }
  return newJar();
}

/** The account's cookie jar for this call; empty when the user hasn't logged in. */
export function accountJar(): CookieJar {
  if (accountJarCache === undefined) {
    accountJarCache = jarOf(accountSession()?.session);
  }
  return accountJarCache;
}

/** Whether the account's jar has a cookie by this name, e.g. to tell logged-in calls apart. */
export function hasAccountCookie(name: string): boolean {
  return accountJar().cookies.some((cookie) => cookie.name === name);
}

/**
 * Fetches with the account's cookies. Cookies the site sets are kept and
 * handed back to the backend as the renewed session when the call ends.
 */
export async function sessionFetch(url: string, init: RequestInit = {}): Promise<Response> {
  const jar = accountJar();
  const before = serializeJar(jar);
  const response = await jarFetch(jar, url, init);
  if (serializeJar(jar) !== before) {
    accountJarRenewed = true;
  }
  return response;
}

/** The session to hand back as renewed, when sessionFetch changed the account's cookies. */
export function renewedSession(): string | undefined {
  if (!accountJarRenewed || accountJarCache === undefined) return undefined;
  return serializeJar(accountJarCache);
}

/** The result of a login that filled jar, for account:login. */
export function loggedIn(jar: CookieJar, label?: string): AccountResult {
  return { status: "ok", session: serializeJar(jar), label };
}

/**
 * Fetches with a jar's cookies and keeps the cookies the site sets, following
 * redirects itself so the cookies set along the way (as logins do) are kept.
 */
export async function jarFetch(jar: CookieJar, url: string, init: RequestInit = {}): Promise<Response> {
  let current = url;
  let request = init;
  for (let hop = 0; hop <= MAX_REDIRECTS; hop++) {
    const response = await fetch(current, withCookies(jar, current, request));
    storeCookies(jar, current, response);
    const location = response.headers.get("location");
    if (!isRedirect(response.status) || location === null) return response;
    current = new URL(location, current).toString();
    request = redirectedRequest(request, response.status);
  }
  throw new Error(`Too many redirects from ${url}`);
}

/** The request with the jar's cookies for url added, redirects left to jarFetch. */
function withCookies(jar: CookieJar, url: string, init: RequestInit): RequestInit {
  const headers: Record<string, string> = { ...((init.headers ?? {}) as Record<string, string>) };
  const cookies = cookieHeader(jar, url);
  if (cookies !== "") {
    const given = headers.Cookie ?? headers.cookie;
    headers.Cookie = given ? `${given}; ${cookies}` : cookies;
  }
  return { ...init, headers, redirect: "manual" };
}

/** Whether a status sends the client elsewhere. */
function isRedirect(status: number): boolean {
  return status === 301 || status === 302 || status === 303 || status === 307 || status === 308;
}

/** The request to repeat at a redirect's target: a 303, or a 301/302 after a POST, becomes a bodiless GET. */
function redirectedRequest(init: RequestInit, status: number): RequestInit {
  const method = (init.method ?? "GET").toUpperCase();
  if (status === 303 || ((status === 301 || status === 302) && method === "POST")) {
    const headers: Record<string, string> = { ...((init.headers ?? {}) as Record<string, string>) };
    delete headers["Content-Type"];
    delete headers["content-type"];
    return { ...init, method: "GET", body: undefined, headers };
  }
  return init;
}

/** The Cookie header for url: the jar's live cookies whose domain and path match. */
export function cookieHeader(jar: CookieJar, url: string): string {
  const target = new URL(url);
  const now = Date.now() / 1000;
  return jar.cookies
    .filter((cookie) => cookieMatches(cookie, target, now))
    .map((cookie) => `${cookie.name}=${cookie.value}`)
    .join("; ");
}

/** Whether a cookie goes with a request to target. */
function cookieMatches(cookie: Cookie, target: URL, now: number): boolean {
  if (cookie.expires && cookie.expires <= now) return false;
  if (cookie.secure && target.protocol !== "https:") return false;
  if (!target.pathname.startsWith(cookie.path ?? "/")) return false;
  const domain = bareDomain(cookie.domain);
  const host = target.hostname.toLowerCase();
  return host === domain || host.endsWith(`.${domain}`);
}

/** A cookie domain without its leading dot (browsers export ".site.com"). */
function bareDomain(domain: string): string {
  return domain.replace(/^\./, "").toLowerCase();
}

/** Keeps the cookies a response sets, dropping the ones it expires. */
function storeCookies(jar: CookieJar, url: string, response: Response): void {
  const headers = response.headers as Headers & { getSetCookie?: () => string[] };
  const setCookies = headers.getSetCookie?.() ?? [];
  for (const header of setCookies) {
    const cookie = parseSetCookie(header, url);
    if (cookie !== undefined) putCookie(jar, cookie);
  }
}

/** Replaces the jar's cookie of the same name, domain and path, or drops it when the new one has expired. */
function putCookie(jar: CookieJar, cookie: Cookie): void {
  jar.cookies = jar.cookies.filter((kept) => !sameCookie(kept, cookie));
  if (cookie.expires && cookie.expires <= Date.now() / 1000) return;
  jar.cookies.push(cookie);
}

/** Whether two cookies are the same one: name, domain and path. */
function sameCookie(first: Cookie, second: Cookie): boolean {
  if (first.name !== second.name || (first.path ?? "/") !== (second.path ?? "/")) return false;
  return bareDomain(first.domain) === bareDomain(second.domain);
}

/** Reads one Set-Cookie header; the domain defaults to the request's host. */
export function parseSetCookie(header: string, url: string): Cookie | undefined {
  const [pair, ...attributes] = header.split(";");
  const separator = pair.indexOf("=");
  if (separator <= 0) return undefined;
  const cookie: Cookie = {
    name: pair.slice(0, separator).trim(),
    value: pair.slice(separator + 1).trim(),
    domain: new URL(url).hostname,
    path: "/",
  };
  let maxAge: number | undefined;
  for (const attribute of attributes) {
    const [rawKey, ...rest] = attribute.split("=");
    const key = rawKey.trim().toLowerCase();
    const value = rest.join("=").trim();
    if (key === "domain" && value !== "") cookie.domain = value.replace(/^\./, "");
    if (key === "path" && value !== "") cookie.path = value;
    if (key === "secure") cookie.secure = true;
    if (key === "expires") cookie.expires = expiresAt(value);
    if (key === "max-age") maxAge = Number(value);
  }
  if (maxAge !== undefined && Number.isFinite(maxAge)) {
    cookie.expires = Math.floor(Date.now() / 1000) + maxAge;
  }
  return cookie;
}

/** An Expires date as Unix seconds; 0 when it can't be read (kept for the session). */
function expiresAt(value: string): number {
  const time = Date.parse(value);
  if (Number.isNaN(time)) return 0;
  return Math.floor(time / 1000);
}
