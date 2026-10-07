-- Recommendation engine signals: feed impressions and the furthest playback
-- position per watch.

-- progress_seconds is the resume position and moves back on seeks;
-- max_progress_seconds keeps the furthest point reached, which is what
-- engagement (completed vs. abandoned) is judged on.
ALTER TABLE watch_history ADD COLUMN max_progress_seconds INTEGER NOT NULL DEFAULT 0;
UPDATE watch_history SET max_progress_seconds = progress_seconds;

-- One row per recommended scene shown ('shown') or opened ('click'). source is
-- the candidate source the item was served from, surface where it appeared
-- (feed, row, ...). Shown rows with no later click or playback are the
-- "shown but not clicked" negative and drive feed fatigue.
CREATE TABLE recommendation_impression (
  id       TEXT PRIMARY KEY,
  media    TEXT NOT NULL,
  kind     TEXT NOT NULL DEFAULT 'shown' CHECK (kind IN ('shown', 'click')),
  source   TEXT NOT NULL DEFAULT '',
  surface  TEXT NOT NULL DEFAULT '',
  position INTEGER,
  shown_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX recommendation_impression_media ON recommendation_impression (media, kind, shown_at);
CREATE INDEX recommendation_impression_shown_at ON recommendation_impression (shown_at);

-- Visits are a click signal, read by time window.
CREATE INDEX scene_detail_fetched_at ON scene (detail_fetched_at);
