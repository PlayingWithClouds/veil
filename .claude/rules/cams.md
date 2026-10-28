---
paths:
  - "apps/backend/internal/cams/**"
  - "apps/backend/internal/livebuffer/**"
  - "apps/backend/plugins/chaturbate/**"
  - "apps/web/src/routes/cams/**"
  - "apps/web/src/lib/stores/camRecordings*"
  - "apps/mobile/app/src/main/java/**/ui/cams/**"
---

# Cams

A cam model is a live room, not a scene.

## Rules

- Cam plugins never produce scenes; scene discovery, ingest and recommendations never see them.
- Private and password shows are skipped.
- Live sessions bypass the blob store and kv cache.
- Token URLs expire within about a minute: resolve right before playing, on every retry, and on every recorder reconnect.
- One room never has two captures.
- Recordings live under `recordings/`, outside the cache prefixes (the evictor must never touch them), encrypted at rest like `stream-cache/`.
- Recording scenes (`veil-recording:<id>`, `ingest.RecordingSourcePrefix`) are exempt from duplicate merging (two sessions share performer and runtime) and never copied into the stream cache (`stream.CacheService.IsRecording`).
- Footage under 5 s is dropped.
- Auto-record on also follows; unfollow turns it off.
- The recorder never creates a live buffer, only uses an existing one.
- A rendition that missed segments restarts at the live edge instead of keeping a hole.
- Live mode: web has no seek bar, speed, markers or progress saving; Android no frame step, speed or loop.

## Rooms (`internal/cams`)

- `cam_model` (id `cam_model:<id>`, identity `source_url`): name, thumbnail, tags, gender, subject, viewers, `online`, `last_seen_online_at`, `followed`. Upserted on every listing fetch, so followed rooms stay known offline.
- `camRooms(sources, tag, query, gender, limit, offset)` queries plugins live, 100 per chunk, chunks cached 60 s for stable paging; plugin popularity order kept, merged by viewers. `camSources` feeds the site filter.
- Online status (`camModel`, `followedCamModels` online first) is a `cam:stream` probe reused 45 s. `followCamModel` / `unfollowCamModel`.
- `camStream(id)` → `Proxy.LiveManifestURL`: playlists fetched fresh, LL-HLS tags (parts, preload hints, blocking reload) stripped, segments relayed.

## Recording (`recorder.go`, `recording_scene.go`, `autorecord.go`)

- `startCamRecording(id, quality, fromSecondsAgo?)`: `quality` = `source` or a target height (tallest ≤ it, else shortest above; `chooseVariant`). Re-picks on every reconnect. Follows separate audio renditions (Chaturbate has them). Segments stored once (by URL path), discontinuity after a gap.
- Ends on `stopCamRecording`, broadcast end, `offlineGrace` (90 s unreachable), or `camRecordingMaxMinutes` (default 240).
- Files: `recordings/<id>/` (`index.m3u8`, `seg-NNNNN.ts|m4s`, `init-N.mp4`; with separate audio `video.m3u8`, `audio.m3u8`, `aseg-*`, `ainit-*`; `poster.jpg`). `index.m3u8` rewritten every 10 segments so `Recorder.Recover` files crash-interrupted captures at start. `cam_recording` rows: status `recording|finished|failed`, stop reason, quality, duration, size, scene.
- Finished = normal scene: title "<name> — live <date time>", performer found/created by `name_key`, cam site as studio, tags "Live cam" + room tags, room thumbnail poster, `detail_fetched_at` set, one `hls` stream to `index.m3u8`, played straight from the blob store. `downloadedScenes` lists them with downloads, newest first.
- `camRecordings(camModelId, activeOnly)`: running (live duration, size, quality) then finished. `setCamAutoRecord(id, enabled, quality)` → `cam_model.auto_record`, `record_quality`. `AutoRecorder` polls followed auto-record models every 2 min via the cached probe, staggered.
- From the past: `fromSecondsAgo` seeds with that much buffer (whole segments, capped, none if unbuffered), then continues live in the same recording/scene in the buffered rendition (`quality` ignored; label says which). A break is marked if the buffer's last segment left the live playlist.

## Rewind buffer (`internal/livebuffer`)

- While a room is watched (`camStream` sessions carry the room id, `Proxy.SetLiveBuffer`) or recorded, every fetched rendition's segments (video and separate audio) are kept 180 s (`WindowSeconds`) as blobs under `live-buffer/<cam id>/` (outside cache prefixes, encrypted). Metadata in memory only.
- The player gets a sliding live playlist with the origin's sequence numbers and discontinuity count (`EXT-X-DISCONTINUITY-SEQUENCE` + tags, identical across renditions so audio matches video), `#EXT-X-ENDLIST` when the origin ends.
- Failed segment downloads stay listed and relay from the origin. Small worker pool downloads each once, newest first; a requested one jumps the queue; served from disk, origin fallback.
- A rendition holds only what was fetched since it was first played.
- Dropped 2 min after the last viewer request or recorder poll (a rendition 2 min after its last playlist); all dropped at startup.
- The recorder reads segments the buffer already holds.

## Web

- Cams in sidebar and phone drawer. `/cams`: live grid (`ScrollSentinel`), search, gender/site filters, `?tag=`, Following row, "Recording" strip, REC/Auto badges.
- `/cams/[id]`: `VideoPlayer` in `live` mode (LIVE badge, "Go live"); quality menu from hls levels, last pick in `localStorage` `veil.cam.quality`; unmuted start with "Tap to unmute" fallback; errors refetch the URL up to 3 times, then offline. Rewind bar over hls.js's playlist range (up to 3 min, "-m:ss"). Record menu: quality, "From here" (behind live), "From 1 min ago", "From 3 min ago" when buffered. Follow, auto-record + quality, finished recordings as `SceneCard`s.
- `ActiveRecordings` (sidebar, drawer): avatar, name, pulsing dot, time, quality, Stop; polled every 10 s by `stores/camRecordings.ts`.

## Android (`ui/cams`)

Cams in the gear menu and Library hub; followed rooms on Following. `CamsScreen` (`cams?tag=`): grid, search, gender/site chips, tag chip, Recording strip, Followed cams, REC/AUTO badges. `CamRoomScreen` (`cam/{id}`): scene player in live mode (`PlayerExtras.live`: LIVE badge, rewind within the buffer window with "-m:ss", Go live), URL refetch up to 3 times, Follow, Record sheet (past start via `recordStarts`, Source or a height), Auto-record + quality, finished recordings.
