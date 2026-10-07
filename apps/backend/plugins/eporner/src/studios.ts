import type { ListArgs, ItemsResult, DiscoveredItem, ScrapeResult, Studio } from "@playingwithclouds/veil-sdk";
import { SITE } from "./http.ts";
import { fetchDoc, absoluteUrl, imageSrc } from "./html.ts";

// eporner models studios as channels: /channel/<slug>/.
function slugFromUrl(url: string): string {
  const match = url.match(/\/channel\/([^/]+)/);
  return match ? match[1] : "";
}

function externalId(url: string): string {
  return `eporner-ch-${slugFromUrl(url)}`;
}

function titleCaseSlug(slug: string): string {
  return slug
    .split("-")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ")
    .trim();
}

// Lists studios from the /channels/ index. A query filters by channel name
// (eporner has no channel search API).
export async function studioList(args: ListArgs): Promise<ItemsResult> {
  const query = (args.query ?? "").trim().toLowerCase();
  const limit = args.limit ?? 0;
  const offset = args.offset ?? 0;

  const doc = await fetchDoc(`${SITE}/channels/`);
  if (!doc) return { items: [] };

  const items: DiscoveredItem[] = [];
  const seen = new Set<string>();

  for (const box of doc.querySelectorAll(".channelscategoriesbox")) {
    const anchor = box.querySelector("a[href*='/channel/']");
    const href = anchor?.getAttribute("href");
    if (!href) continue;
    const name = box.querySelector("h2")?.text?.trim() ?? "";
    if (!name) continue;
    if (query && !name.toLowerCase().includes(query)) continue;

    const sourceUrl = absoluteUrl(href);
    if (seen.has(sourceUrl)) continue;
    seen.add(sourceUrl);

    items.push({
      title: name,
      media_type: "studio",
      source_url: sourceUrl,
      external_id: externalId(sourceUrl),
      poster_path: imageSrc(box.querySelector("img.cclogo")),
    });
  }

  if (limit > 0) return { items: items.slice(offset, offset + limit) };
  return { items: items.slice(offset) };
}

// Fetches one channel page as a Studio (name + logo).
export async function studioFind(url: string): Promise<ScrapeResult[]> {
  const slug = slugFromUrl(url);
  if (!slug) throw new Error(`eporner studio: no slug in ${url}`);

  const doc = await fetchDoc(url);
  const name = doc?.querySelector("h1")?.text?.trim() || titleCaseSlug(slug);
  // Prefer a real channel logo; skip the generic og:image so a find never
  // overwrites the listing's own logo with eporner's fallback.
  const logo = imageSrc(doc?.querySelector(".channelheader img") ?? doc?.querySelector("img.cclogo"));

  const studio: Studio = {
    type: "studio",
    external_id: externalId(url),
    source_url: url,
    name,
    url,
  };
  if (logo) studio.image_path = logo;
  return [{ type: "studio", studio }];
}
