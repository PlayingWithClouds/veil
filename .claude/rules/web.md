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

## Split view

`SplitView` is mounted by the root layout so players survive navigation. On `/split` (`routes/split`, `splitUrl()`, empty route) it fills the page; elsewhere it's a muted dock at bottom left that opens `/split`.

- Up to 4 cams/scenes: 1 full, 2 side by side, 3–4 in 2×2. Drag-to-reorder tab bar, fullscreen button.
- State in `lib/stores/split.ts` (localStorage `veil:split`: `{kind, mediaId, title}` per pane + focused index; playback restarts on reload).
- `SplitAddButton` (hover on `SceneCard`, `ExploreSceneCard`, `CamCard`, sidebar followed cams, `ActiveRecordings`) → `openInSplit`: adds a pane, stays on the page. With 4 open it replaces the focused pane and says so.
- Each `SplitPane` resolves its own stream (cam: `fetchCamStream`, re-resolved on error; scene: best stream) into a private `VideoPlayer`, never the shared one, so the mini player is untouched.
- Only the focused pane is unmuted, unless "All audio" (`allAudio`, saved). The selecting click doesn't pause (`silenced`, which also isolates it from the session's live-audio choice). Only the focused pane takes keyboard shortcuts (`keyboard`).
- Panes keep their DOM position and are placed with CSS `order`, because moving a player's element stalls its stream.
