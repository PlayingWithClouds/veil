import { resolveUrl, BASE_URL, fetchPage } from "./http.ts";
import type { ResolveResult } from "@playingwithclouds/veil-sdk";

// The playable stream lives in a dean-edwards packed script on the video page.
// Unpack it, read the surrit.com HLS master, and return it. The master URL is
// served from a Cloudflare-fronted CDN and expects a missav referer at playback.
export async function resolve(url: string): Promise<ResolveResult> {
  const html = await fetchPage(resolveUrl(url));
  const script = extractPackedScript(html);
  if (!script) throw new Error(`missav resolver: no packed player script at ${url}`);

  const decoded = unpackPacked(script);
  const master = decoded.match(/https:\/\/surrit\.com\/[0-9a-f-]+\/playlist\.m3u8/i);
  if (!master) throw new Error(`missav resolver: no HLS master in player at ${url}`);

  return {
    url: master[0],
    mime_type: "application/x-mpegURL",
    headers: { Referer: `${BASE_URL}/`, Origin: BASE_URL },
  };
}

function extractPackedScript(html: string): string | null {
  const start = html.indexOf("eval(function(p,a,c,k,e,d)");
  if (start === -1) return null;
  // The packer body itself contains "))", so anchor the end past the dictionary
  // `.split('|')` argument before looking for the closing "))".
  const dictSplit = html.indexOf(".split('|')", start);
  if (dictSplit === -1) return null;
  const end = html.indexOf("))", dictSplit);
  if (end === -1) return null;
  return html.slice(start, end + 2);
}

// Reverse the dean-edwards packer: substitute each base-36 symbol back to its
// dictionary word. The symbol table matches the packer's `c.toString(36)`.
function unpackPacked(packed: string): string {
  const match = packed.match(/\}\('(.*)',(\d+),(\d+),'(.*?)'\.split\('\|'\)/s);
  if (!match) return "";

  const payload = match[1].replace(/\\'/g, "'").replace(/\\\\/g, "\\");
  const count = parseInt(match[3], 10);
  const dictionary = match[4].split("|");

  let decoded = payload;
  for (let index = count - 1; index >= 0; index--) {
    const word = dictionary[index];
    if (!word) continue;
    const symbol = index.toString(36);
    decoded = decoded.replace(new RegExp(`\\b${symbol}\\b`, "g"), word);
  }
  return decoded;
}
