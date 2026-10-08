-- Where a scene gets watched and scrubbed to, in 100 buckets across its runtime.
-- Feeds the player's most-replayed graph and the best-moment thumbnail.
CREATE TABLE scene_heat (
  scene    TEXT    NOT NULL,
  bucket   INTEGER NOT NULL,
  watched  REAL    NOT NULL DEFAULT 0,
  scrubbed REAL    NOT NULL DEFAULT 0,
  PRIMARY KEY (scene, bucket)
);

-- A frame from the scene's best moment, replacing the site's poster in listings.
-- thumbnail_seconds is the moment it was taken at, so it is only redone when the best moment moves.
ALTER TABLE scene ADD COLUMN thumbnail_path TEXT;
ALTER TABLE scene ADD COLUMN thumbnail_seconds REAL;
