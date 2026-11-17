import type { CamListArgs, CamListResult, CamRoom } from "@playingwithclouds/veil-sdk";
import { API_HEADERS, SITE } from "./http.ts";

/** The site's room-list endpoint rejects pages larger than this. */
const MAX_PAGE_SIZE = 100;
const DEFAULT_PAGE_SIZE = 48;

/** One room as the room-list endpoint returns it (only the fields used). */
export interface ChaturbateRoom {
  username: string;
  gender?: string;
  tags?: string[];
  num_users?: number;
  img?: string;
  room_subject?: string;
  start_dt_utc?: string;
  current_show?: string;
  has_password?: boolean;
}

const GENDER_NAMES: Record<string, string> = {
  f: "female",
  m: "male",
  c: "couple",
  s: "trans",
  t: "trans",
};

const GENDER_FILTERS: Record<string, string> = {
  female: "f",
  male: "m",
  couple: "c",
  trans: "t",
};

/** Maps one API room to the plugin protocol's CamRoom. */
export function toCamRoom(room: ChaturbateRoom): CamRoom {
  const camRoom: CamRoom = {
    username: room.username,
    display_name: room.username,
    source_url: `${SITE}/${room.username}/`,
    online: room.current_show === "public" && room.has_password !== true,
  };
  if (room.img) camRoom.thumbnail_url = room.img;
  if (typeof room.num_users === "number") camRoom.viewers = room.num_users;
  if (room.tags && room.tags.length > 0) camRoom.tags = room.tags;
  if (room.room_subject) camRoom.subject = room.room_subject;
  if (room.start_dt_utc) camRoom.started_at = room.start_dt_utc;
  if (room.gender && GENDER_NAMES[room.gender]) camRoom.gender = GENDER_NAMES[room.gender];
  return camRoom;
}

/** Builds the room-list query string for a cam:list call. */
export function buildRoomListQuery(args: CamListArgs): string {
  const params = new URLSearchParams();
  let limit = args.limit;
  if (limit === undefined || limit <= 0) limit = DEFAULT_PAGE_SIZE;
  params.set("limit", String(Math.min(limit, MAX_PAGE_SIZE)));
  let offset = args.offset;
  if (offset === undefined) offset = 0;
  params.set("offset", String(offset));
  if (args.query) params.set("keywords", args.query);
  if (args.tag) params.set("hashtags", args.tag.toLowerCase().replace(/^#/, ""));
  if (args.gender && GENDER_FILTERS[args.gender]) params.set("genders", GENDER_FILTERS[args.gender]);
  return params.toString();
}

/** One answer of the room-list endpoint. */
export interface RoomListBody {
  rooms?: ChaturbateRoom[];
  total_count?: number;
}

/**
 * The total the site reports when it answers a keyword search loosely: mostly
 * rooms that don't match, capped at this many. It answers the same request
 * either loosely or exactly at random, so a search asks again on this total.
 */
const LOOSE_SEARCH_TOTAL = 500;

/** Times a keyword search is asked while the site answers it loosely. */
const SEARCH_TRIES = 3;

/** Lists the rooms online now, most viewed first (the site's default order). */
export async function camList(args: CamListArgs): Promise<CamListResult> {
  let body = await fetchRoomList(args);
  for (let tries = 1; args.query && isLooseAnswer(body) && tries < SEARCH_TRIES; tries++) {
    body = await fetchRoomList(args);
  }
  return toListResult(body, args);
}

/** Fetches one page of the room list. */
async function fetchRoomList(args: CamListArgs): Promise<RoomListBody> {
  const response = await fetch(`${SITE}/api/ts/roomlist/room-list/?${buildRoomListQuery(args)}`, {
    headers: API_HEADERS,
    signal: AbortSignal.timeout(20000),
  });
  if (!response.ok) throw new Error(`chaturbate room list: HTTP ${response.status}`);
  return (await response.json()) as RoomListBody;
}

/** Whether a search answer is the loose one (a word common enough to have 500 exact matches looks the same). */
export function isLooseAnswer(body: RoomListBody): boolean {
  return body.total_count === LOOSE_SEARCH_TOTAL;
}

/**
 * The plugin result of one answer. An exact search answer pads its page with
 * unrelated rooms past its total, so those are cut off.
 */
export function toListResult(body: RoomListBody, args: CamListArgs): CamListResult {
  const result: CamListResult = { rooms: [] };
  let rooms = body.rooms ?? [];
  if (typeof body.total_count === "number") {
    result.total = body.total_count;
    rooms = rooms.slice(0, Math.max(0, body.total_count - (args.offset ?? 0)));
  }
  result.rooms = rooms.map(toCamRoom);
  return result;
}
