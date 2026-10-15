export const BROWSER_HEADERS = {
  "User-Agent":
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
  Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
  "Accept-Language": "de-DE,de;q=0.9,en;q=0.8",
  "Accept-Encoding": "gzip, deflate, br",
};

export function resolveUrl(base: string, path: string): string {
  if (path.startsWith("http")) return path;
  if (path.startsWith("//")) return "https:" + path;
  const origin = new URL(base).origin;
  return origin + (path.startsWith("/") ? path : "/" + path);
}

export async function fetchHtml(
  url: string,
  extraHeaders?: Record<string, string>
): Promise<string> {
  const res = await fetch(url, {
    headers: { ...BROWSER_HEADERS, ...extraHeaders },
  });
  if (!res.ok) throw new Error(`HTTP ${res.status} fetching ${url}`);
  return res.text();
}

// FlareSolverr endpoint (e.g. http://127.0.0.1:8191/v1), injected via env by the
// backend. Absent = no solver configured.
function solverEndpoint(): string | undefined {
  const url = process.env.FLARESOLVERR_URL;
  return url && url.trim() ? url.trim() : undefined;
}

export function hasSolver(): boolean {
  return solverEndpoint() !== undefined;
}

// Fetches a URL through FlareSolverr, which drives a headless browser to clear
// Cloudflare's challenge and returns the fully rendered HTML. Throws when no
// solver is configured or the solver reports a non-ok status.
export async function fetchViaSolver(url: string, maxTimeoutMs = 60_000): Promise<string> {
  const endpoint = solverEndpoint();
  if (!endpoint) throw new Error("no FLARESOLVERR_URL configured");

  const res = await fetch(endpoint, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ cmd: "request.get", url, maxTimeout: maxTimeoutMs }),
    signal: AbortSignal.timeout(maxTimeoutMs + 5_000),
  });
  if (!res.ok) throw new Error(`solver HTTP ${res.status} fetching ${url}`);

  const body = (await res.json()) as {
    status?: string;
    message?: string;
    solution?: { response?: string };
  };
  if (body.status !== "ok" || !body.solution?.response) {
    throw new Error(`solver failed for ${url}: ${body.message ?? "unknown"}`);
  }
  return body.solution.response;
}

// Fetches HTML with a plain request, transparently retrying through the solver
// when the origin answers 403 (Cloudflare block) and a solver is available.
export async function fetchHtmlSmart(
  url: string,
  extraHeaders?: Record<string, string>
): Promise<string> {
  try {
    return await fetchHtml(url, extraHeaders);
  } catch (error) {
    const blocked = error instanceof Error && error.message.includes("403");
    if (blocked && hasSolver()) return fetchViaSolver(url);
    throw error;
  }
}

export async function followRedirect(url: string): Promise<string> {
  try {
    const res = await fetch(url, {
      headers: BROWSER_HEADERS,
      redirect: "follow",
    });
    return res.url || url;
  } catch {
    return url;
  }
}

const BASE64_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

/**
 * Decodes base64 into bytes in plain JS: the embedded plugin runtime has no
 * atob or Buffer. Used for the picture a face:identify call carries.
 */
export function base64ToBytes(base64: string): Uint8Array {
  const clean = base64.replace(/[^A-Za-z0-9+/]/g, "");
  const bytes = new Uint8Array(Math.floor((clean.length * 3) / 4));
  let byteIndex = 0;
  let buffer = 0;
  let bitCount = 0;
  for (const character of clean) {
    buffer = (buffer << 6) | BASE64_ALPHABET.indexOf(character);
    bitCount += 6;
    if (bitCount >= 8) {
      bitCount -= 8;
      bytes[byteIndex] = (buffer >> bitCount) & 0xff;
      byteIndex += 1;
    }
  }
  return bytes;
}

export function domainOf(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return "";
  }
}
