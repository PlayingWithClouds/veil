# eporner plugin

Source plugin for [eporner.com](https://www.eporner.com), built on its public
[v2 JSON API](https://www.eporner.com/api/v2/) — no HTML scraping for metadata.

## Capabilities

- **scene:list** — without a query, browses newest videos (`order=latest`); with
  a query, searches `video/search` (`order=most-popular`), paged to the limit.
- **scene:list:page** — `scene:list` with a `url` lists the video grid of a
  `/channel/`, `/pornstar/` or `/cat/` page (paginated `<url><n>/`) or an
  uploader's `/profile/<name>/videos/`; subscriptions use it to follow studios
  and performers.
- **scene:find** — `video/id` → one `Scene` with tags (from `keywords`), duration,
  rating (0–5 API scale mapped to 0–10), poster and preview frames. The download
  points at the embed page, which the resolver consumes.
- **stream:resolve** — primary path calls the video-sources XHR using the hash derived
  from the embed page; falls back to the direct CDN mp4 embedded in the player.
  Both are returned with an eporner `Referer`.

## Notes

- eporner gates its CDN behind a rotating, session-bound source hash. The XHR
  path mirrors the known algorithm; if eporner rotates it, resolve falls back to
  the embed mp4. Metadata, discovery and search are unaffected.
- `external_id` is `eporner-<id>`; the id is the alphanumeric token shared by the
  `/video-<id>/`, `/hd-porn/<id>/` and `/embed/<id>/` URL forms.
