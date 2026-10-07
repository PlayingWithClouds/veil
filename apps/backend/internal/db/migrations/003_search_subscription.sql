-- Subscriptions are saved searches re-run on a schedule; the per-media refresh
-- subscriptions they replace were never wired up.
DROP TABLE subscription;

CREATE TABLE search_subscription (
  id               TEXT PRIMARY KEY,
  query            TEXT NOT NULL,
  normalized_query TEXT NOT NULL,
  -- plugin names to search; empty = every search-capable plugin
  sources          JSON NOT NULL DEFAULT '[]',
  interval_hours   INTEGER NOT NULL DEFAULT 6,
  enabled          BOOLEAN NOT NULL DEFAULT 1,
  last_run_at      TEXT,
  next_run_at      TEXT,
  last_error       TEXT,
  -- items first seen after this count as new
  seen_at          TEXT,
  created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX search_subscription_query ON search_subscription (normalized_query);
CREATE INDEX search_subscription_due ON search_subscription (enabled, next_run_at);

CREATE TRIGGER search_subscription_updated_at AFTER UPDATE ON search_subscription WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE search_subscription SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;

-- Every media record a subscription's searches have returned.
CREATE TABLE search_subscription_item (
  id            TEXT PRIMARY KEY,
  subscription  TEXT NOT NULL,
  media         TEXT NOT NULL,
  first_seen_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX search_subscription_item_media ON search_subscription_item (subscription, media);
CREATE INDEX search_subscription_item_seen ON search_subscription_item (subscription, first_seen_at);
