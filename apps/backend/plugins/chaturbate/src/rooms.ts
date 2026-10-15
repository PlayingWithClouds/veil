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

/** Lists the rooms online now, most viewed first (the site's default order). */
export async function camList(args: CamListArgs): Promise<CamListResult> {
  const response = await fetch(`${SITE}/api/ts/roomlist/room-list/?${buildRoomListQuery(args)}`, {
    headers: API_HEADERS,
    signal: AbortSignal.timeout(20000),
  });
  if (!response.ok) throw new Error(`chaturbate room list: HTTP ${response.status}`);
  const body = (await response.json()) as { rooms?: ChaturbateRoom[]; total_count?: number };
  const result: CamListResult = { rooms: [] };
  if (body.rooms) result.rooms = body.rooms.map(toCamRoom);
  if (typeof body.total_count === "number") result.total = body.total_count;
  return result;
}
