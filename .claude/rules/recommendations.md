---
paths:
  - "apps/backend/internal/recommend/**"
  - "apps/backend/internal/recommendeval/**"
  - "apps/backend/cmd/recommend-eval/**"
  - "apps/backend/internal/updater/**"
  - "apps/web/src/lib/recommendations*"
  - "apps/mobile/app/src/main/java/**/data/ImpressionLogger*"
  - "apps/mobile/app/src/main/java/**/ui/components/ImpressionTracking*"
---

# Recommendations

`internal/recommend`: candidate generation → ranking → re-ranking per request.

## Rules

- All tunables live in `weights.go`.
- The profile is computed on demand; no cache table.
- Watches are judged by time played, not position reached (viewers skip).
- Blocked entities pin to −1. Blocked tags keep full rarity weight.
- "Watch again" is a row only, never a feed candidate.
- Only recommendation surfaces (`recommendationSurfaces` in `impressions.go`: `home`, legacy `feed` and empty, `row`, `preview-feed`) feed learning, shown-but-ignored and fatigue. Elsewhere the user chose the page or query.
- Every list of scenes logs each showing with `surface` and 0-based `position`, and resets on replace. New list UI must do the same.

## Signals

Per scene, 30-day half-life: o-event (5) > session-ending watch (4) > like > save (watchlist, user collection) ≈ A/B pick > followed watch (full 1, partial 0.5) ≈ click (first visit or logged click).

- Session-ending: ≥ 1 min played and no other scene started or played within 30 min after. A watch under 30 min old counts as followed. Sessions from `watch_history` started/updated (`session.go`).
- Negative: dislike, early abandon (< 10 s, or < 5% of runtime and < 1 min), A/B reject, shown-but-not-clicked (`recommendation_impression`).
- `watch_history.played_seconds` sums gaps between the players' 15 s progress saves (saved on playing time; a save > 45 s after the last starts a new sitting, counting ≤ 15 s). 5 min or 80% of runtime = full watch.
- A fresh followed watch plus its click seeds related scenes (`seedMinimumWeight` 0.9).

## Profile

Signalled scenes spread weight onto performers, tags (÷√count), studio and site (first observing plugin); each dimension normalized to [−1, 1].

- Tag rarity (`rarity.go`): log inverse share of tagged scenes (1 at ≤ 5%, floor 0.05; all 1 below 50 tagged scenes), applied when spreading and in `tagAffinity`. `libraryStats` cached 10 min.
- Moments (`moments.go`): for the 100 latest session enders, marker tags from 120 s before to 20 s after the stop point (`progress_seconds`), and around the `scene_heat` peak (bucket ≥ 2× average, ±30 s), get an extra 0.5× of the scene's weight.
- Runtime (`runtime.go`): signals (not shown-but-not-clicked) spread over bands 3/6/12/25/50/90 min, shrunk by `runtimeMinMass`. Match weights: performer 0.36, tag 0.27, studio 0.18, site 0.09, runtime 0.10 (neutral if unknown).
- Looks (`attributes.go`): `Profile.Attributes` per attribute (hair, ethnicity, eyes, country, natural/fake, tattoos, piercings, cup, height band) from rated performers. A performer without own affinity inherits 0.5× the mean of their attributes' (`castAffinity`).

## Candidates

Each tagged with its source: `scene_related` of the strongest signalled scenes (site-related highest prior); new subscription finds; newest of top performers/studios/tags; feeder finds (`discover:taste`, `discover:catalog`, `discover:popular`, `explore:site`); library exploration (random scenes of tags adjacent to top tags, plus random); unvisited stubs from recent searches; newest.

- `related:walk` (`walk.go`, personal, prior 0.5): 3-hop walk from seeds over `scene_related` (forward, reverse ×0.5) + shared-performer and shared-rare-tag edges, ≤ 100 per hop, only scenes the direct lists don't propose.
- `explore:attribute` (prior 0.35): newest scenes of up to 8 unknown performers with inherited looks affinity ≥ 0.15.
- Cold start: catalog finds, searches, newest.

## Feeder (`updater.Feeder`)

The engine only ranks what's stored. Every 12 h (and when < 100 unwatched, unshown candidates, at most hourly) it lists: one catalog page of every scene plugin (one deeper per run, 3 deep); site pages, else name searches, of top 3 performers, 2 studios, 3 tags (`taste`) and 3 random adjacent tags (`explore`); 2 most-viewed pages rotating through `defaultPopularPages` (`popular`, entity = host, via `scene:list:page`, kvs sites only today). Ingested as stubs, recorded in `feed_discovery` (kept 30 days, counted 14; ≤ 20 taste / 30 explore / 40 popular+catalog per entity). A taste find's prior scales with the entity's affinity down to half (`tasteAffinityFloor`). A run that stored anything clears the memoized feed.

## Ranking

Affinity + source prior + freshness + site popularity − watched/fatigue/disliked-neighbour penalties.

- Popularity: view-count percentile within the site (first observing plugin) blended 70/30 with rating percentile; 0.5 if neither or < 20 samples.
- Rewatch (`rewatch.go`): a session ender keeps the watched penalty 3 days, loses it linearly over 11, ends at +0.3. A followed watch keeps it.
- Excluded: blocked, disliked, inactive-plugin stubs.
- Re-ranking: each 24-slot page split by kind (`kindQuotas`: personal = `related:*`/`affinity:*`/subscription 8, taste 5, catalog+popular 5, `explore:*` 4, search 2; newest is filler). Kinds spread evenly (≤ 2 slots ahead), score order within; unfilled slots go to the best remaining, search last. Per-page caps per performer/studio/site; no source or site three in a row.
- Duplicates: `dropDuplicates` keeps the best-ranked copy (`alike.SameScene`).
- `recommendations(excludeTagIds)` / `scenes(includeTagIds, excludeTagIds)` filter by own, studio or performer tags. `tasteProfile` exposes strongest liked/disliked tags, performers, studios, sites.
- Memo: feed memoized 30 min for `offset > 0`; `offset 0` re-ranks after 10 min or on `refresh: true`.
- Watch again: `Rows` adds row `rewatch` (second position) of session enders last watched 7+ days ago, strongest first.

## Learning (`learning.go`)

Per ranking, one SQL aggregation counts each source's showings (`recommendation_impression` kind `shown`, 60 days, decayed by `AffinityHalfLife`, < 6 h old skipped) and rewards. Opened = click or first visit within 6 h with no later showing in between (that one gets credit). Reward 0.6 opened, +0.25 session-ending watch, +0.25 like within 7 days, cap 1. Beta posterior centred on the pooled rate, 30 pseudo-rewards (`learnPriorClicks`). Thompson sample once per ranking (`Engine.random`, injectable); draw ÷ pooled rate, clamped [0.5, 1.5], multiplies `sourcePriors`. Kinds sampled the same way shift `kindQuotas` (between `kindQuotaMinimums` and 1.5× default, summing to `pageSize`; `allocateQuotas`). Nothing learned below 300 decayed showings. Logged at most every 30 min. Position bias: rewards ÷ per-surface examination propensity per position bucket (`examination.go`), shrunk toward a decaying prior; no position = surface mean.

## Impressions

Surfaces: `home`, `row` (position within the row), `preview-feed`, `related` (up-next), `performer`, `studio`, `tag`, `search`.

- Web (`src/lib/recommendations.ts`): `use:impression={{ mediaId, surface, position, source? }}`, `recordClicked` with the same fields, `resetShownImpressions(surface)` on list replace; batched via `recordImpressions`.
- Android: `data/ImpressionLogger`, `ui/components/ImpressionTracking.kt` (`LogShownScenes`, `logSceneClicked`). Incognito sends nothing.

## Offline eval

`go run ./cmd/recommend-eval -db path/to/veil.db` in `apps/backend` (`-variants default,no-learning,baseline-newest,baseline-random`, `-window-days`, `-step-days`, `-repeats`, `-json`). Works on a `VACUUM INTO` copy. Per cutoff: rolls state back (events, stubs, related links, visit-given tags/performers), ranks with `recommend.NewWithOptions`, scores against the next window (hit rate and recall @24/96, MRR for session enders, watches, clicks; site/tag coverage; source shares). Targets created after the cutoff are unreachable. Tables without timestamps are taken as-is.
