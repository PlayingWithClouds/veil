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
- Incognito: `saveProgress`, `recordSearch`, `ImpressionLogger` send nothing. New tracking must respect it.
- The app keeps its real name until the owner picks a disguise.
- The backend has no ffmpeg; frames come only from the app.
- `FLAG_SECURE` blacks out screenshots, `adb screencap` and recordings; turn "Hide in recents and screenshots" off to take them.

## Backend on device

- `backend/EmbeddedBackend.kt` starts it from `VeilApplication` on `127.0.0.1:47831`, `DATA_DIR` in app storage; extracts `assets/plugins` as `PLUGIN_SEED_DIR` once per install/update (`PluginSeeder`); keeps `DNS_SERVERS_FILE` current; restarts on crash; logs to tag `VeilBackend`. Env/DNS rules in `BackendEnvironment` (unit tested).
- Another server can be saved in Settings (`ServerSettings`, `ServerForm`); saving restarts the app. `BootGate` (`ui/VeilApp.kt`) waits for `/health`, falls back to the server form.
- GraphQL: Apollo Kotlin from `apps/backend/internal/api/graphql/schema/*.graphql`; operations in `app/src/main/graphql/com/playingwithclouds/veil/*.graphql`; `data/*Repository.kt` map to UI models. `/api/search` SSE via `api/LiveSearch.kt`.

## UI

- `ui/VeilShell.kt`: Home, Following, Galleries, Library in one swipeable `HorizontalPager` (`tabs` route); floating tab bar `ui/FloatingNavBar.kt` (solid, hides on scroll down); search in each tab's top bar (`SearchButton`).
- Home (`ui/home`): slim top bar (wordmark, search, menu); chip row hiding on scroll (`HomeFilterRow`: site sheet, All, presets, runtime windows, top tags; `HomeFilter` picks ranked feed vs tag listing; presets = `saved_filter` rows incl. sites); edge-to-edge `SceneCard`s with meta line (studio · performer · rating · views · age), ⋮ quick actions, swipe right = dislike (`SwipeToDismissCard`), a shelf after every 5 cards (`recommendedRows`, feed credits).
- Continue watching is in the Library hub (downloads, watchlist, collections, history, performers, studios, tags). Galleries = the gallery plugin's category index. Gear menu: Random, Plugins, Settings. Transitions in `ui/PageTransitions.kt`; back via navigation-compose.
- Playback: Media3 ExoPlayer (`ui/scene/ScenePlayer.kt`), HLS and progressive via `/api/stream/`, resume, progress saving, fullscreen with landscape lock (vertical videos stay portrait, `rememberFullscreenState`).
- Player extras (`ui/scene`): `PlayerSettings` gear (speed 0.25–2×, quality, A–B loop, zoom to fill, PiP privacy) kept by `PlayerOptions` (loop per page; zoom, PiP privacy in `player` prefs; helpers in `PlayerMath`). Frame step while paused (`Player.stepFrame`). `MarkerChips` + seek-bar notches from `scene_marker`, lazy. PiP on leaving (auto from Android 12) or by button, blurred by default (`PictureInPicturePrivacy`), closing pauses. Headphone unplug pauses. `data/PlaybackQueue`: play next, add, Play all / Shuffle, auto-advance (`SceneScreen`, `QueueSheet`).
- Replay data: `PlaybackHeatReporting` → `recordSceneHeat` (`scene_heat`, 100 buckets). `sceneHeat` gives the graph, `bestMomentSeconds` (most-scrubbed, else a marker), `thumbnailSeconds`. If best moment is > 30 s from the thumbnail's, `ThumbnailCapture` (`MediaMetadataRetriever`, progressive only) grabs and uploads a frame (`setSceneThumbnail`); `thumbnail_path` beats `poster_path` in `sceneFromMap` and cards.
- `ui/previews` (`openPreviewFeed()`, `previews`): `VerticalPager` over `recommendations` with preview clips, only the visible page plays (`ScenePreview` draws nothing on failure), tag button hides tags (`excludeTagIds`).
- `ui/taste` (`openTaste()`) renders `tasteProfile`. Performer/studio pages: `ui/entity/ChannelHeader` (banner, bell, "N new", bio sheet), Newest/Top rated (`EntityFilter.sort`).
- `data/TagFilter` + `TagFilterSheet`: include/exclude (live results checked client-side). `SceneQuickActionsSheet` "Don't recommend" uses the backend blocklist, which also hides blocked entities from index lists. Search appends `sourceSuggestions` under library suggestions.
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
