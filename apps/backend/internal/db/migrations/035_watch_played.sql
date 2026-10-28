-- How long a scene has actually played, summed over its sittings. Viewers skip
-- around to find what they like, so the furthest position reached says little
-- about how much of a scene was watched. The backend adds up the time between
-- the progress saves the players send every 15 seconds while playing.
ALTER TABLE watch_history ADD COLUMN played_seconds INTEGER NOT NULL DEFAULT 0;

-- Rows from before: the time between first start and last save, at most the
-- furthest position (a row reused over days would otherwise count the days).
-- A row saved only once says nothing about how long it played, so it keeps
-- the furthest position, as watches were read until now. The updated_at
-- trigger is lifted meanwhile, since updated_at is when the scene last played
-- and the recommender reads it as such.
DROP TRIGGER watch_history_updated_at;
UPDATE watch_history SET played_seconds = max(0, CASE
  WHEN (julianday(updated_at) - julianday(started_at)) * 86400 < 15 THEN max_progress_seconds
  ELSE min(max_progress_seconds, cast((julianday(updated_at) - julianday(started_at)) * 86400 AS INTEGER))
END);
CREATE TRIGGER watch_history_updated_at AFTER UPDATE ON watch_history WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE watch_history SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
