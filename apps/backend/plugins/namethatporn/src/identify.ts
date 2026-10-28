import { base64ToBytes } from "@playingwithclouds/veil-sdk";
import type { FaceIdentifyArgs, FaceIdentifyResult, FacePostMatch } from "@playingwithclouds/veil-sdk";
import { parseAnswerNames, parsePosts, SITE_ORIGIN } from "./parse.ts";

const SEARCH_URL = `${SITE_ORIGIN}/search/images.html`;

/** The site takes uploads of a few MB; the backend already caps pictures at 5 MB. */
const MAX_IMAGE_BYTES = 8 * 1024 * 1024;

/** Solved posts whose answers are read; each costs one page fetch. */
const MAX_ANSWERED_POSTS = 10;

const REQUEST_TIMEOUT_MS = 30_000;

/** Cloudflare answers 403 without a browser's User-Agent. */
const USER_AGENT =
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

/** File extensions the upload form accepts, by MIME type. */
const EXTENSIONS: Record<string, string> = {
  "image/jpeg": "jpg",
  "image/png": "png",
  "image/webp": "webp",
  "image/gif": "gif",
  "image/bmp": "bmp",
};

/** Fetches a page of the site the way a browser on it would. */
async function fetchPage(url: string, init: RequestInit = {}): Promise<string> {
  const response = await fetch(url, {
    ...init,
    headers: { Referer: SEARCH_URL, "User-Agent": USER_AGENT },
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status} from namethatporn`);
  return response.text();
}

/** Reads the accepted answer of a solved post; a post that fails to load keeps no names. */
async function withAnswerNames(post: FacePostMatch): Promise<FacePostMatch> {
  try {
    return { ...post, names: parseAnswerNames(await fetchPage(post.url)) };
  } catch {
    return post;
  }
}

/**
 * Uploads the picture to the site's reverse image search and returns the
 * matching posts, best first. The first solved ones are opened to read the
 * pornstars their accepted answers name.
 */
export async function identifyFace(args: FaceIdentifyArgs): Promise<FaceIdentifyResult> {
  const imageBytes = base64ToBytes(args.image_base64);
  if (imageBytes.length > MAX_IMAGE_BYTES) throw new Error("picture is too large for NameThatPorn");
  const extension = EXTENSIONS[args.mime_type];
  if (extension === undefined) throw new Error(`NameThatPorn does not search ${args.mime_type} pictures`);

  const form = new FormData();
  form.append("imgfile", new Blob([imageBytes], { type: args.mime_type }), `picture.${extension}`);
  const posts = parsePosts(await fetchPage(SEARCH_URL, { method: "POST", body: form }));

  const answered = posts.filter((post) => post.solved).slice(0, MAX_ANSWERED_POSTS);
  const named = new Map<string, FacePostMatch>();
  for (const post of await Promise.all(answered.map(withAnswerNames))) {
    named.set(post.url, post);
  }
  return {
    candidates: [],
    posts: posts.map((post) => named.get(post.url) || post),
  };
}
