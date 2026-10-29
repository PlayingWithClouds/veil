import { base64ToBytes } from "@playingwithclouds/veil-sdk";
import type { FaceIdentifyArgs, FaceIdentifyResult, FacePostMatch } from "@playingwithclouds/veil-sdk";
import { fetchPage, SEARCH_URL } from "./http.ts";
import { parseAnswerNames, parsePosts } from "./parse.ts";

/** The site takes uploads of a few MB; the backend already caps pictures at 5 MB. */
const MAX_IMAGE_BYTES = 8 * 1024 * 1024;

/** Solved posts whose answers are read; each costs one page fetch. */
const MAX_ANSWERED_POSTS = 10;

/** File extensions the upload form accepts, by MIME type. */
const EXTENSIONS: Record<string, string> = {
  "image/jpeg": "jpg",
  "image/png": "png",
  "image/webp": "webp",
  "image/gif": "gif",
  "image/bmp": "bmp",
};

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
