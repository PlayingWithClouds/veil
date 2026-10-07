-- Search-driven discovery: a scene's detail page is fetched once, when it is
-- first visited, and that visit links related scenes.

-- When the scene's detail page was last fetched; NULL for listing stubs.
ALTER TABLE scene ADD COLUMN detail_fetched_at TEXT;

-- Directed "related to" edges from a visited scene. source says where the
-- edge came from: 'site' (the site's own related list), or a fallback search
-- by 'performer', 'tag' or 'title'. rank is the position within that source.
CREATE TABLE scene_related (
  id         TEXT PRIMARY KEY,
  scene      TEXT NOT NULL,
  related    TEXT NOT NULL,
  source     TEXT NOT NULL CHECK (source IN ('site', 'performer', 'tag', 'title')),
  rank       INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX scene_related_edge ON scene_related (scene, related);
CREATE INDEX scene_related_related ON scene_related (related);
