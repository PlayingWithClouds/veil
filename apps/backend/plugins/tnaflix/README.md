# tnaflix plugin

Scraper for [tnaflix.com](https://www.tnaflix.com). Capabilities: `scene:find`, `scene:list`, `stream:resolve`.

## URL patterns

- Newest listing: `/new/<page>` — cards as `div[data-vid]` with `a.video-thumb[href][data-trailer]`, poster `img[src|data-src]`, title `a.video-title`, runtime chip `.video-duration` (`MM:SS` or `HH:MM:SS`)
- Search: `/search?what=<query>&page=<n>` — same card grid
- Video page: `/<category>/<Slug>/video<id>`
  - title: `og:title` / `h1`
  - poster: `og:image`
  - duration: `video#video-player[data-duration]` (seconds)
  - date: `"uploadDate"` in the VideoObject JSON-LD
  - badges: `div.video-detail-badges` — uploader (`a.badge-video-info`), performers (`a[href*="/profile/"]`), category chips (site-section paths), tag chips (`/search?what=<tag>`)
- Streams: `<video><source src="…mp4?…secure=<token>,<expiry>" size="720">` — signed, expiring; `stream:resolve` re-fetches the page and returns the highest-resolution mp4. No Referer required on the CDN.
