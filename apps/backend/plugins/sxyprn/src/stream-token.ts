import { SITE_HOST, BASE_URL } from "./http.ts";

const BASE64_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

/**
 * The playable URL of a hosted video's source path, rebuilt the way the site's
 * own player script builds it. The path is "/cdn/<a>/<token>/<expiry>/<b>/<file>":
 * the first segment gets a signature of the digit sums of the token and file
 * names, and the expiry is shifted by the same digit sums.
 */
export function playableUrl(sourcePath: string): string {
  const parts = sourcePath.split("/");
  if (parts.length < 8) throw new Error(`sxyprn: unexpected video source path ${sourcePath}`);
  const tokenDigits = digitSum(parts[6]);
  const fileDigits = digitSum(parts[7]);
  parts[1] = `${parts[1]}8/${signatureOf(tokenDigits, fileDigits)}`;
  parts[5] = String(Number(parts[5]) - (tokenDigits + fileDigits));
  return `${BASE_URL}${parts.join("/")}`;
}

/** The sum of the decimal digits in a text (letters and other characters are skipped). */
export function digitSum(text: string): number {
  let sum = 0;
  for (const character of text.replace(/[^0-9]/g, "")) {
    sum += Number(character);
  }
  return sum;
}

/** The URL-safe base64 of "<tokenDigits>-<site host>-<fileDigits>", with "=" written as ".". */
function signatureOf(tokenDigits: number, fileDigits: number): string {
  const encoded = base64Encode(`${tokenDigits}-${SITE_HOST}-${fileDigits}`);
  return encoded.replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, ".");
}

/** Standard base64 of an ASCII text (the runtime has no btoa). */
export function base64Encode(text: string): string {
  let encoded = "";
  for (let offset = 0; offset < text.length; offset += 3) {
    encoded += encodeGroup(text, offset);
  }
  return encoded;
}

/** The four base64 characters of up to three bytes starting at an offset, padded with "=". */
function encodeGroup(text: string, offset: number): string {
  const first = text.charCodeAt(offset);
  const hasSecond = offset + 1 < text.length;
  const hasThird = offset + 2 < text.length;
  // A character code past the end is NaN, which the bit operators read as 0.
  const second = text.charCodeAt(offset + 1);
  const third = text.charCodeAt(offset + 2);
  const bits = (first << 16) | (second << 8) | third;
  let encoded = BASE64_ALPHABET[(bits >> 18) & 63] + BASE64_ALPHABET[(bits >> 12) & 63];
  if (hasSecond) {
    encoded += BASE64_ALPHABET[(bits >> 6) & 63];
  } else {
    encoded += "=";
  }
  if (hasThird) {
    encoded += BASE64_ALPHABET[bits & 63];
  } else {
    encoded += "=";
  }
  return encoded;
}
