---
paths:
  - "apps/mobile/**"
---

# Android app

Native Kotlin + Jetpack Compose (Material 3) plus the Go backend on the phone. Gradle root `apps/mobile` (Kotlin DSL, catalog `gradle/libs.versions.toml`, one `:app` module, package `com.playingwithclouds.veil`). `local.properties` (gitignored) must point `sdk.dir` at an SDK with platform 37.

Commands (in `apps/mobile`): `bun run build` (arm64 backend into `jniLibs` + release plugin bundles into `assets/plugins`), `bun run apk` (+ `./gradlew assembleDebug` → `app/build/outputs/apk/debug/app-debug.apk`), `bun run android` (+ install, launch), `bun run test` (JVM).

## Rules

- No WebView, no Capacitor.
- Backend is pure Go (`CGO_ENABLED=0`) at `jniLibs/arm64-v8a/libveil.so`, the only place Android executes app files from.
- `VeilApi` is the only place that builds the Apollo client.
- Backend-built URLs carry `PUBLIC_URL` (default localhost): pass them through `ServerSettings.backendUrl()`, `imageUrl()` or `RemoteImage` (proxies via `/api/img`) so loopback is rebased (`BackendUrls`).
- Screens use only `ui/theme` tokens and `ui/design` components: no raw colours, shapes, `sp`, or literal dp in paddings/gaps (`DesignTokenUsageTest`).
- Accent marks selection and progress; main actions are white; things floating over content use blurring `glass`; the tab bar is solid; every other button, pill, switch is `glassControl`.
- One package per screen under `ui/`, `ViewModel` + `StateFlow` + `LoadState`, endless lists via `ui/paging/PagedList`.
- Media items carry only "Playing" + app name, no artwork; notifications use `NeutralNotifications` wording, `VISIBILITY_SECRET`.
- Incognito: `AppServices.events` (the `EventLog`: impressions, watch progress, likes, O, searches) sends nothing. New tracking must go through the event log or respect it.
- The app keeps its real name until the owner picks a disguise.
- The backend has no ffmpeg; frames come only from the app.
- `FLAG_SECURE` blacks out screenshots, `adb screencap` and recordings; turn "Hide in recents and screenshots" off to take them.

## Backend on device

- `backend/EmbeddedBackend.kt` starts it from `VeilApplication` on `127.0.0.1:47831`, `DATA_DIR` in app storage; extracts `assets/plugins` as `PLUGIN_SEED_DIR` once per install/update (`PluginSeeder`); keeps `DNS_SERVERS_FILE` current; restarts on crash; logs to tag `VeilBackend`. Env/DNS rules in `BackendEnvironment` (unit tested).
- Another server can be saved in Settings (`ServerSettings`, `ServerForm`); saving restarts the app. `BootGate` (`ui/VeilApp.kt`) waits for `/health`, falls back to the server form.
- GraphQL: Apollo Kotlin from `apps/backend/internal/api/graphql/schema/*.graphql`; operations in `app/src/main/graphql/com/playingwithclouds/veil/*.graphql`; `data/*Repository.kt` map to UI models. Subscriptions (search) go over Apollo's WebSocket transport (`VeilApi`).

## UI

- Rebuild in progress: the app was reset to the shell + Home after the backend's schema rework. Screens not rebuilt yet open `ui/pending/PendingScreen` through `AppNavigator` (`openSearch`, `openPerformer`, …); porting one = add its route and change that function. The notes below on previews, taste, entity pages, tag filter, "Who is this?" and the gallery page describe the old screens (in git history), not current code.
- `ui/VeilShell.kt`: side menu `ui/MainDrawer.kt` (`Destination` entries in Watch / Library / You as the web's `navLinks`, Settings at the bottom; ☰ `MenuButton` via `LocalOpenDrawer` or a left-edge swipe on the tabs) over Home, Cams, Library in a `HorizontalPager` (`tabs` route) changed only by the small centred floating bar `ui/FloatingNavBar.kt` (solid, hides on scroll down), so the edge swipe stays the drawer's. Tab screens start with `TabTopBar` (☰, wordmark or title, search).
- Home (`ui/home`) mirrors the web home (`apps/web/src/lib/home.ts`, `HomeRows.svelte`): top picks as a snapping row of large `VideoCard`s (the one in front plays its preview), Continue watching, Live now, More for you (full-width cards), Continue reading, Stories, Audio, Performers. Rows load on their own (`HomeRepository`, `Home.graphql`) and hide while empty; pull to refresh. Cards in `ui/components/MediaCards.kt`, rows in `Shelf`, seen cards via `TrackShown` → `ImpressionLogger` (`recordEvents`, surfaces `home` / `continue` as on the web).
- Cams (`ui/cams`, mirrors the web's `cams.ts`): the Cams tab is `CamsScreen("")` (top picks from `camRecommendations`, your rooms = followed + auto-recorded, running recordings, then the live grid with gender / order / site / tag chips); a room's tag opens `CamsScreen(tag)` as a detail screen. Follows go through `followRecord` / `unfollow` (targetKind `camModel`); finished recordings are videos (`camModel.recordings`). `CamRoomScreen` plays `watchCam` in the player's live mode and logs `view` + `progress {watched}` like the web's `trackRoomWatch`. Live is the pink accent, REC red.
- Library tab (`ui/library`, the web's library page): Watchlist / Queue / Collections / Downloads segments, each loaded the first time it shows; long press a card to remove it (`CollectionEntries`); `CollectionScreen` for one collection; downloads poll while one runs. Data in `LibraryRepository` (`builtInCollection`, `collections`, `downloads`).
- Search (`ui/search`, the web's `search.ts`): the `searchResults` subscription over Apollo's WebSocket transport, one page at a time, each site's batch added as it comes (videos + performers), next page on scroll until every site is exhausted; recent searches from `recentSearches` while the box is empty; a submit logs a `search` event (no subject).
- People (`ui/people`): `MembersScreen` lists performers or studios (name filter, debounced); `PersonScreen` is one performer (opened with `fetchPerformer`, which reads their site pages like the web) or studio: header with facts, Favourite, bio, tags, a galleries row and their videos paged (`videos(filter: performerIds|studioIds)`). Editing, merging and face matches are web-only.
- Transitions in `ui/PageTransitions.kt`; back via navigation-compose.
- Video page (`ui/video`, mirrors the web's video page): `fetchVideo` (`VideoRepository.open`) reads the source pages; local files play first, else site streams resolved by `play(url)`, the next source when one fails (`VideoViewModel`); resume from `userState.progress` unless ≥ 95 % watched. Under the player: title and facts, actions (like / dislike with Undo / favourite / O / watchlist / download, as the web's `VideoActions`), 5-star rating (0..10), source pills, people, tags, description, Up next (library `related` → `siteRelated` pages → recommendations, surface `related`).
- Playback: Media3 ExoPlayer (`ui/video/VideoPlayer.kt`), HLS and progressive, fullscreen with landscape lock (vertical videos stay portrait, `rememberFullscreenState`). Watching is logged like the web's `watchLog.ts` (`WatchLog.kt`: progress every 15 s played and on pause, `heatSpan` per stretch, `finish`, one session id per page).
- Player extras (`ui/video`): `PlayerSettings` gear (speed 0.25–2×, quality, A–B loop, zoom to fill, PiP privacy) kept by `PlayerOptions` (loop per page; zoom, PiP privacy in `player` prefs; helpers in `PlayerMath`). Frame step while paused (`Player.stepFrame`). `MarkerChips` + seek-bar notches from the video's segments. PiP on leaving (auto from Android 12) or by button, blurred by default (`PictureInPicturePrivacy`), closing pauses. Headphone unplug pauses. Fullscreen: swipe up for the related panel.
- `ui/previews` (`openPreviewFeed()`, `previews`): `VerticalPager` over `recommendations` with preview clips, only the visible page plays (`ScenePreview` draws nothing on failure), tag button hides tags (`excludeTagIds`).
- `ui/taste` (`openTaste()`) renders `tasteProfile`. Performer/studio pages: `ui/entity/ChannelHeader` (banner, bell, "N new", bio sheet), Newest/Top rated (`EntityFilter.sort`).
- `data/TagFilter` + `TagFilterSheet`: include/exclude (live results checked client-side). `SceneQuickActionsSheet` "Don't recommend" uses the backend blocklist, which also hides blocked entities from index lists.
- "Who is this?" (`ui/face/FaceIdentifySheet`): frame at playhead (progressive only; HLS uses the thumbnail) or the gallery image on screen. Gallery page: like/dislike, O, Favorite (`collection:favorites`), Save, Watch / Find the video (`findGalleryVideos` → `linkGalleryVideo`). ⋮ → "Merge into…" (`MergeDialog`) on performer/studio/tag pages.
- Design kit: Settings → Design kit (`ui/designkit`). Tokens: `VeilColors`, `VeilSpacing` (16 dp `gutter`), `VeilShapes`, `Theme.kt`, `VeilType`. Components: buttons, pills (flat `FeedChip`), headers, surfaces, glass, `bleed`/`gutterPadding`.

## Privacy (Settings → Privacy, `ui/settings/PrivacySection.kt`)

Stored by `privacy/PrivacyPreferences` (StateFlows over SharedPreferences), applied by `MainActivity` (a `FragmentActivity` for biometrics).

- Recents/screenshots: `FLAG_SECURE` while "Hide in recents and screenshots" (default on) or app lock is on.
- App lock (`AppLock`, `LockPolicy`, `PinHasher`, `ui/privacy/LockScreen`): 4–8 digit PIN, salted PBKDF2, optional biometrics (`BiometricUnlock`), delay (immediate, 1, 5 min), throttled from the 5th wrong try. Drawn over the app in `VeilApp`; cold start is locked. Lock and panic pause playback via `PrivacyEvents.pauseRequests` (`rememberScenePlayer`).
- Panic (`PanicController`, `PanicDetectors`): face-down (foreground only) or two-finger double tap (`MainActivity.dispatchTouchEvent`, never consumed) → pause, lock if set, go home.
- Disguise: one `activity-alias` per entry (`.alias.DefaultAlias` real; `Files`/`Gallery`/`Notes`/`Calculator` share `@mipmap/ic_disguise`); `Disguises.apply` enables one. Labels in `strings.xml`.
- `NeutralMedia`: no media session or notification producer exists yet.
- Incognito: in memory, ends with the process or by tapping "You're incognito". Likes, ratings, lists, downloads still count; a first visit still stamps `detail_fetched_at` (read as a click).
- Encrypted downloads: see storage rules. `privacy/BlobKeyProvider` generates the key, seals it with an Android Keystore AES key, passes it via `BackendEnvironment`, replaces it if unopenable (old blobs then unreadable). "Encrypt downloads" (default on) applies after restart.
