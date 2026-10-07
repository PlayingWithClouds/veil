-- Searches the user submitted, one row per normalized query (trimmed,
-- whitespace-collapsed, case-folded), for search suggestions. search_history
-- stays the per-run log the recommender reads; this table is what the user
-- sees as "recent searches" and can remove entries from.
CREATE TABLE recent_search (
  id               TEXT PRIMARY KEY,
  query            TEXT NOT NULL,             -- display text of the latest run
  normalized_query TEXT NOT NULL,
  search_count     INTEGER NOT NULL DEFAULT 1 CHECK (search_count >= 1),
  last_searched_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX recent_search_query ON recent_search (normalized_query);
CREATE INDEX recent_search_last_searched_at ON recent_search (last_searched_at);

-- Seed from the searches already logged, so recent searches aren't empty
-- after the upgrade. The display text is the latest spelling.
INSERT INTO recent_search (id, query, normalized_query, search_count, last_searched_at)
SELECT 'recent_search:' || lower(hex(randomblob(10))),
  (SELECT latest.query FROM search_history AS latest
    WHERE latest.normalized_query = logged.normalized_query
    ORDER BY latest.created_at DESC LIMIT 1),
  logged.normalized_query, count(*), max(logged.created_at)
FROM search_history AS logged
WHERE trim(logged.normalized_query) <> ''
GROUP BY logged.normalized_query;
