import type { EnrichArgs, EnrichResult } from "@playingwithclouds/veil-sdk";
import { fetchPornstar } from "./pornstar.ts";

// Fills performer details (photo, aliases, country, measurements, height, etc.)
// by scraping the matching eporner pornstar page. Only performers are handled;
// other media types return an empty result so other enrichers can try.
export async function enrich(args: EnrichArgs): Promise<EnrichResult> {
  if (args.media_type !== "performer") return {};

  const existing = (args.existing ?? {}) as { name?: string };
  const name = existing.name?.trim();
  if (!name) return {};

  const performer = await fetchPornstar(name);
  return performer ?? {};
}
