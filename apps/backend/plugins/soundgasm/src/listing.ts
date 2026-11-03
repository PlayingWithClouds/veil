import { parse } from "node-html-parser";
import type { DiscoveredItem, ItemsResult, ListArgs } from "@playingwithclouds/veil-sdk";
import { externalIdOf, fetchPage, normalizeText, parseAddress, userUrl, voiceOf } from "./site.ts";
import { splitTitle } from "./titles.ts";

/** The play counter text of a card ("Play Count: 25980") as a number, or undefined. */
function playCountOf(text: string): number | undefined {
  const match = text.match(/(\d+)/);
  if (!match) return undefined;
  return Number(match[1]);
}

/** Lists the audio cards of an uploader's page, newest first as the site orders them. */
export function parseUserPage(html: string): DiscoveredItem[] {
  const items: DiscoveredItem[] = [];
  for (const card of parse(html).querySelectorAll(".sound-details")) {
    const link = card.querySelector("a");
    const address = parseAddress(link?.getAttribute("href") ?? "");
    if (!link || !address?.slug) continue;

    const item: DiscoveredItem = {
      title: splitTitle(link.text).title,
      media_type: "audio",
      source_url: link.getAttribute("href") ?? "",
      external_id: externalIdOf(address.user, address.slug),
      studio: voiceOf(address.user),
    };
    const playCount = playCountOf(normalizeText(card.querySelector(".playCount")?.text ?? ""));
    if (playCount !== undefined) item.view_count = playCount;
    items.push(item);
  }
  return items;
}

/**
 * Lists an uploader's audios (args.url = their page). The site has no search or
 * global catalog, so a query or a missing URL lists nothing. The page carries
 * every audio at once, so paging is a slice.
 */
export async function audioList(args: ListArgs): Promise<ItemsResult> {
  const address = parseAddress(args.url ?? "");
  if (!address) return { items: [] };

  const html = await fetchPage(userUrl(address.user));
  const items = parseUserPage(html);
  const offset = args.offset ?? 0;
  if (args.limit === undefined) return { items: items.slice(offset) };
  return { items: items.slice(offset, offset + args.limit) };
}
