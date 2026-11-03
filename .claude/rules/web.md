---
paths:
  - "apps/web/**"
---

# Web

## Rules

- GraphQL only, via `gqlClient` (`src/lib/veil.ts`) with inline genql selections in `src/lib/*.ts` data modules. Subscriptions over graphql-ws (`streamsChanged`).
- Internal links go through `src/lib/routes.ts` (`sceneUrl`, `performerUrl`, `tagUrl`, …).
- Streams play through `/api/stream/` (CORS `*`); images through `/api/img` (`cacheUrl`).
- Any list of scenes logs impressions (`use:impression`, `recordClicked`, `resetShownImpressions` on replace; see recommendations rules).
- New UI follows `src/routes/layout.css` tokens and patterns. TV/D-pad focus: `src/lib/tv.css`, `FocusGrid`, `FocusList`.

## Cams

`lib/cams.ts` is the data module; `CamCard` is the one room card (hover plays the muted stream via `LiveRoomPreview`, ⋮ menu follows, auto-records, records, blocks).

- Follows of rooms live in the `followedCams` store (room id → follow id), loaded once by `loadFollowedCams`; change them only through `toggleCamFollow`.
- The live grid on `/cams` refreshes shown rooms in place every minute rather than reloading, so it never reorders under the user.
- There is no split view; don't add one.
