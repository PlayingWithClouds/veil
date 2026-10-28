-- The scenes one search returned, best first. The recommender's search source
-- used to guess them as the scenes first stored around the time of the search,
-- which caught whatever else was stored then (the feeder's finds, an earlier
-- search's results) and missed results the library already had.
CREATE TABLE search_result (
  search     TEXT NOT NULL,   -- search_history id
  scene      TEXT NOT NULL,
  rank       INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  PRIMARY KEY (search, scene)
);
CREATE INDEX search_result_created_at ON search_result (created_at);
