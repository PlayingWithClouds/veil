import type { ListArgs, ItemsResult, DiscoveredItem, ScrapeResult, Performer } from "@playingwithclouds/veil-sdk";
import { SITE } from "./http.ts";
import { fetchDoc, absoluteUrl, imageSrc } from "./html.ts";
import { fetchPornstarByUrl } from "./pornstar.ts";

const MAX_LIST_PAGES = 15;
const DEFAULT_TARGET = 30;

// eporner's pornstar slug is the last path segment of /pornstar/<slug>/.
function slugFromUrl(url: string): string {
  const match = url.match(/\/pornstar\/([^/]+)/);
  return match ? match[1] : "";
}

function externalId(url: string): string {
  return `eporner-ps-${slugFromUrl(url)}`;
}

// Lists pornstars from the /pornstar-list/ index, paginating until the target is
// met. A query filters the crawled names (eporner has no pornstar search API).
export async function performerList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim().toLowerCase();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;
  const target = limit > 0 ? offset + limit : DEFAULT_TARGET;

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (let page = 1; page <= MAX_LIST_PAGES; page++) {
    const doc = await fetchDoc(`${SITE}/pornstar-list/${page}/`);
    if (!doc) break;

    const blocks = doc.querySelectorAll(".mbprofile");
    if (blocks.length === 0) break;

    for (const block of blocks) {
      const anchor = block.querySelector("a[href*='/pornstar/']");
      const href = anchor?.getAttribute("href");
      if (!href) continue;
      const name = block.querySelector("p.mbtit")?.text?.trim() ?? anchor?.getAttribute("title") ?? "";
      if (!name) continue;
      if (query && !name.toLowerCase().includes(query)) continue;

      const sourceUrl = absoluteUrl(href);
      if (seen.has(sourceUrl)) continue;
      seen.add(sourceUrl);

      items.push({
        title: name,
        media_type: "performer",
        source_url: sourceUrl,
        external_id: externalId(sourceUrl),
        poster_path: imageSrc(block.querySelector(".mbprofileimg img")),
      });
    }
    if (items.length >= target) break;
  }

  if (limit > 0) return { items: items.slice(offset, offset + limit) };
  return { items: items.slice(offset) };
}

// Fetches one pornstar's full profile (bio, stats, avatar) as a Performer.
export async function performerFind(url: string): Promise<ScrapeResult[]> {
  const slug = slugFromUrl(url);
  if (!slug) throw new Error(`eporner performer: no slug in ${url}`);

  const partial = await fetchPornstarByUrl(url);

  const performer: Performer = {
    ...partial,
    type: "performer",
    external_id: externalId(url),
    source_url: url,
    name: partial?.name ?? nameFromSlug(slug),
  };
  return [{ type: "performer", performer }];
}

// "riley-reid-abc12" → "Riley Reid" (drops eporner's trailing id suffix heuristically).
function nameFromSlug(slug: string): string {
  return slug
    .split("-")
    .filter((part) => !/^[a-z0-9]{4,6}$/i.test(part) || part.length < 5)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ")
    .trim();
}
