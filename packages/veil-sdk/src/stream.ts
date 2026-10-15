import { writeSync } from "fs";
import type { DiscoveredItem, ItemsResult, ListArgs } from "./types.ts";

/**
 * Writes a batch of list items ahead of the final result, so the backend can
 * show them while the plugin is still crawling. The final result must still
 * carry every item; the backend drops the ones it has already seen.
 */
export function emitItems(items: DiscoveredItem[]): void {
  if (items.length === 0) return;
  writeSync(1, JSON.stringify({ items }) + "\n");
}

/** Identity of a listed item across pages and sites. */
function itemKey(item: DiscoveredItem): string {
  if (item.source_url) return item.source_url;
  return item.external_id;
}

/**
 * Collects a listing that arrives in parts (pages, sites): drops repeats,
 * keeps the args' offset/limit window and emits each new item inside it at
 * once. result() is the final ItemsResult.
 */
export class ItemSink {
  private readonly offset: number;
  private readonly limit: number;
  private readonly seen = new Set<string>();
  private readonly kept: DiscoveredItem[] = [];
  private position = 0;

  constructor(args: ListArgs) {
    this.offset = args.offset ?? 0;
    this.limit = args.limit ?? 0;
  }

  /** Adds one part and returns how many of its items were new. */
  add(items: DiscoveredItem[]): number {
    const fresh: DiscoveredItem[] = [];
    let added = 0;
    for (const item of items) {
      const key = itemKey(item);
      if (this.seen.has(key)) continue;
      this.seen.add(key);
      added++;
      if (this.inWindow(this.position)) fresh.push(item);
      this.position++;
    }
    this.kept.push(...fresh);
    emitItems(fresh);
    return added;
  }

  /** Whether the window is filled, so a crawl can stop. */
  full(): boolean {
    return this.limit > 0 && this.position >= this.offset + this.limit;
  }

  /** The kept items as the listing's final result. */
  result(): ItemsResult {
    return { items: this.kept };
  }

  /** Whether the item at position falls inside offset/limit. */
  private inWindow(position: number): boolean {
    if (position < this.offset) return false;
    return this.limit <= 0 || position < this.offset + this.limit;
  }
}
