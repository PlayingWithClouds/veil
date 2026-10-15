import type { FaceCandidate } from "@playingwithclouds/veil-sdk";

/** One entry of a result category as FaceDigger returns it. */
interface RawMatch {
  url?: string;
  title?: string;
  thumbnail?: string;
  similarity?: string;
}

/** Turns "66.5%" into 0..1; unreadable text is 0. */
export function similarityFrom(text: string | undefined): number {
  const percent = parseFloat(String(text));
  if (!Number.isFinite(percent)) return 0;
  return Math.min(Math.max(percent / 100, 0), 1);
}

/** Keeps a link only when it is plain https. */
function httpsOnly(url: string | undefined): string | undefined {
  if (typeof url !== "string" || !url.startsWith("https://")) return undefined;
  return url;
}

/**
 * Reads the job result of FaceDigger. Only the "cams" category (live-cam
 * models, whose title is the model's name) is used: "porn" lists videos by
 * their titles, possibly of private people, and "onlyfans" lists creator
 * accounts, so neither is read even though the API returns them. Each
 * candidate carries the name, similarity, picture and cam profile link only.
 */
export function parseCamMatches(result: unknown): FaceCandidate[] {
  const categories = (result as { results?: { cams?: { results?: RawMatch[] } } }).results;
  const matches = categories?.cams?.results;
  if (!Array.isArray(matches)) return [];

  const candidates: FaceCandidate[] = [];
  for (const match of matches) {
    const name = typeof match.title === "string" ? match.title.trim() : "";
    if (name === "") continue;
    const candidate: FaceCandidate = {
      name,
      similarity: similarityFrom(match.similarity),
      profile_urls: [],
    };
    const profileUrl = httpsOnly(match.url);
    if (profileUrl) candidate.profile_urls.push(profileUrl);
    const imageUrl = httpsOnly(match.thumbnail);
    if (imageUrl) candidate.image_url = imageUrl;
    candidates.push(candidate);
  }
  return candidates.sort((left, right) => right.similarity - left.similarity);
}
