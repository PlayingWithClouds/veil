-- The comments under a video on its site, as the scene page's scrape found
-- them. Viewers often name the performer there. Rows of one (scene, plugin)
-- are replaced on every scrape; position is the site's order, reply_to the
-- position of the comment answered.
CREATE TABLE scene_comment (
  scene      TEXT    NOT NULL,
  plugin     TEXT    NOT NULL,
  position   INTEGER NOT NULL,
  author     TEXT,
  text       TEXT    NOT NULL,
  posted_at  TEXT,
  likes      INTEGER,
  reply_to   INTEGER,
  fetched_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  PRIMARY KEY (scene, plugin, position)
);
