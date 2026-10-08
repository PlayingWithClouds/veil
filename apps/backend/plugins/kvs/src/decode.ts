// KVS's kt_player hides the real media URL when the site enables link
// protection: flashvars then carry `function/0/<url>` where the 32-character
// hash segment of <url> has been shuffled. The shuffle is derived from the
// page's `license_code`, so the player (and this module) can undo it.

const OBFUSCATED_PREFIX = "function/0/";
const HASH_LENGTH = 32;
// The shuffled hash is the fourth path segment: /get_file/<group>/<hash>/...
const HASH_SEGMENT_INDEX = 3;

/** Whether a flashvars URL is obfuscated and needs `decodeVideoUrl`. */
export function isObfuscated(videoUrl: string): boolean {
  return videoUrl.startsWith(OBFUSCATED_PREFIX);
}

/**
 * The playable URL behind a flashvars `video_url`: obfuscated
 * `function/0/...` values are unshuffled with the license code, anything else
 * is returned unchanged.
 */
export function decodeVideoUrl(videoUrl: string, licenseCode: string): string {
  if (!isObfuscated(videoUrl)) return videoUrl;

  const url = new URL(videoUrl.slice(OBFUSCATED_PREFIX.length));
  const segments = url.pathname.split("/");
  const segment = segments[HASH_SEGMENT_INDEX];
  if (segment === undefined || segment.length < HASH_LENGTH) {
    throw new Error(`kvs: obfuscated url has no hash segment: ${videoUrl}`);
  }

  const token = licenseToken(licenseCode);
  if (token.length < HASH_LENGTH || token.some((digit) => Number.isNaN(digit))) {
    throw new Error(`kvs: unusable license code ${licenseCode}`);
  }
  const hash = segment.slice(0, HASH_LENGTH);
  const order = unshuffleOrder(token);
  const restored = order.map((index) => hash.charAt(index)).join("");
  segments[HASH_SEGMENT_INDEX] = restored + segment.slice(HASH_LENGTH);
  url.pathname = segments.join("/");
  return url.toString();
}

/**
 * The digit sequence kt_player derives from the license code: the code's
 * digits (zeros read as ones) are split into overlapping halves, four times
 * their difference gives a key, and each key digit is combined with four
 * consecutive license digits.
 */
export function licenseToken(licenseCode: string): number[] {
  const license = licenseCode.replace("$", "");
  const licenseDigits = Array.from(license, (character) => parseInt(character, 10));
  const nonZero = license.replace(/0/g, "1");
  const center = Math.floor(nonZero.length / 2);
  const frontHalf = parseInt(nonZero.slice(0, center + 1), 10);
  const backHalf = parseInt(nonZero.slice(center), 10);
  const key = String(4 * Math.abs(frontHalf - backHalf)).slice(0, center + 1);

  const token: number[] = [];
  for (let index = 0; index < key.length; index++) {
    const keyDigit = parseInt(key.charAt(index), 10);
    for (let offset = 0; offset < 4; offset++) {
      token.push((licenseDigits[index + offset] + keyDigit) % 10);
    }
  }
  return token;
}

/** The hash positions in restored order, swapping from the back as kt_player does. */
function unshuffleOrder(token: number[]): number[] {
  const order = Array.from({ length: HASH_LENGTH }, (_, index) => index);
  let accumulated = 0;
  for (let source = HASH_LENGTH - 1; source >= 0; source--) {
    accumulated += token[source];
    const destination = (source + accumulated) % HASH_LENGTH;
    const swapped = order[source];
    order[source] = order[destination];
    order[destination] = swapped;
  }
  return order;
}
