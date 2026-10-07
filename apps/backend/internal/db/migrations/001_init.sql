-- Initial SQLite schema, consolidated from the SurrealDB migrations 000–031.
--
-- Every id is "table:id". Link columns hold the linked row's id. JSON columns
-- hold arrays (ordered link lists, aliases) or free-form values. Timestamps are
-- UTC text in the fixed-width format below, so they sort as strings.

-- ---------------------------------------------------------------------------
-- Content

CREATE TABLE studio (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  aliases     JSON NOT NULL DEFAULT '[]',
  url         TEXT,
  parent      TEXT,                        -- studio id (network -> studio)
  image_path  TEXT,
  details     TEXT,
  external_id TEXT,
  source_url  TEXT,
  tags        JSON NOT NULL DEFAULT '[]',  -- tag ids
  created_by  TEXT,                        -- plugin id
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX studio_name ON studio (name);
CREATE UNIQUE INDEX studio_external_id ON studio (external_id);
CREATE UNIQUE INDEX studio_source_url ON studio (source_url);
CREATE INDEX studio_parent ON studio (parent);

CREATE TABLE tag (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  aliases     JSON NOT NULL DEFAULT '[]',
  description TEXT,
  category    TEXT,
  created_by  TEXT,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX tag_name ON tag (name);

CREATE TABLE performer (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  aliases       JSON NOT NULL DEFAULT '[]',
  details       TEXT,                      -- biography
  gender        TEXT,
  birthdate     TEXT,                      -- ISO 8601 date
  death_date    TEXT,                      -- ISO 8601 date
  country       TEXT,
  ethnicity     TEXT,
  eye_color     TEXT,
  hair_color    TEXT,
  height_cm     INTEGER,
  weight_kg     INTEGER,
  measurements  TEXT,
  fake_tits     TEXT,
  tattoos       TEXT,
  piercings     TEXT,
  career_length TEXT,
  url           TEXT,
  twitter       TEXT,
  instagram     TEXT,
  image_path    TEXT,
  tags          JSON NOT NULL DEFAULT '[]',
  favorite      BOOLEAN NOT NULL DEFAULT 0,
  rating        REAL CHECK (rating IS NULL OR (rating >= 0 AND rating <= 10)),
  external_id   TEXT,
  source_url    TEXT,
  created_by    TEXT,
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX performer_name ON performer (name);
CREATE UNIQUE INDEX performer_external_id ON performer (external_id);
CREATE INDEX performer_source_url ON performer (source_url);

CREATE TABLE scene (
  id               TEXT PRIMARY KEY,
  external_id      TEXT NOT NULL DEFAULT '',
  source_url       TEXT NOT NULL,
  title            TEXT NOT NULL DEFAULT '',
  details          TEXT,
  date             TEXT,                   -- ISO 8601 release date
  duration_seconds INTEGER CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
  studio           TEXT,
  performers       JSON NOT NULL DEFAULT '[]',
  tags             JSON NOT NULL DEFAULT '[]',
  poster_path      TEXT,
  preview_video    TEXT,
  preview_images   JSON NOT NULL DEFAULT '[]',
  rating           REAL CHECK (rating IS NULL OR (rating >= 0 AND rating <= 10)),
  view_count       INTEGER NOT NULL DEFAULT 0 CHECK (view_count >= 0),
  organized        BOOLEAN NOT NULL DEFAULT 0,
  created_by       TEXT,
  created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX scene_source_url ON scene (source_url);
CREATE INDEX scene_external_id ON scene (external_id);
CREATE INDEX scene_title ON scene (title);
CREATE INDEX scene_studio ON scene (studio);
CREATE INDEX scene_created_at ON scene (created_at);

CREATE TABLE gallery (
  id          TEXT PRIMARY KEY,
  external_id TEXT NOT NULL DEFAULT '',
  source_url  TEXT NOT NULL,
  title       TEXT NOT NULL DEFAULT '',
  details     TEXT,
  date        TEXT,
  studio      TEXT,
  performers  JSON NOT NULL DEFAULT '[]',
  tags        JSON NOT NULL DEFAULT '[]',
  cover_path  TEXT,
  image_count INTEGER NOT NULL DEFAULT 0 CHECK (image_count >= 0),
  rating      REAL CHECK (rating IS NULL OR (rating >= 0 AND rating <= 10)),
  organized   BOOLEAN NOT NULL DEFAULT 0,
  created_by  TEXT,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX gallery_source_url ON gallery (source_url);
CREATE INDEX gallery_external_id ON gallery (external_id);
CREATE INDEX gallery_studio ON gallery (studio);

-- Content images (content = 1: gallery pages, standalone images) and asset
-- images (content = 0: posters, logos, profile shots).
CREATE TABLE image (
  id           TEXT PRIMARY KEY,
  media        TEXT,                       -- owning scene/performer/studio/gallery/collection
  type         TEXT NOT NULL DEFAULT '',
  file_path    TEXT NOT NULL,
  width        INTEGER,
  height       INTEGER,
  aspect_ratio REAL,
  position     INTEGER,
  content      BOOLEAN NOT NULL DEFAULT 0,
  title        TEXT,
  details      TEXT,
  date         TEXT,
  external_id  TEXT,
  source_url   TEXT,
  studio       TEXT,
  performers   JSON NOT NULL DEFAULT '[]',
  tags         JSON NOT NULL DEFAULT '[]',
  rating       REAL CHECK (rating IS NULL OR (rating >= 0 AND rating <= 10)),
  organized    BOOLEAN NOT NULL DEFAULT 0,
  created_by   TEXT,
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX image_media ON image (media);
CREATE UNIQUE INDEX image_file_path ON image (file_path);
CREATE INDEX image_content ON image (content);

-- Performer credits on scenes, galleries and images (was the performs_in
-- graph edge: in = performer, out = media).
CREATE TABLE performs_in (
  id            TEXT PRIMARY KEY,
  performer     TEXT NOT NULL,
  media         TEXT NOT NULL,
  credited_as   TEXT,                      -- credited alias for this appearance
  billing_order INTEGER,
  created_by    TEXT,
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX performs_in_edge ON performs_in (performer, media);
CREATE INDEX performs_in_media ON performs_in (media);

CREATE TABLE stream (
  id                 TEXT PRIMARY KEY,
  media              TEXT NOT NULL,        -- scene id
  url                TEXT NOT NULL,
  kind               TEXT NOT NULL DEFAULT 'stream',
  label              TEXT,
  provider           TEXT,
  resolution         TEXT,
  width              INTEGER,
  height             INTEGER,
  language           TEXT,
  format             TEXT,
  mime_type          TEXT,
  expected_speed_bps REAL,
  file_size_bytes    REAL,
  verified           BOOLEAN NOT NULL DEFAULT 0,
  plugin_name        TEXT,
  created_by         TEXT,
  created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX stream_media_url ON stream (media, url);

CREATE TABLE collection (
  id           TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  details      TEXT,
  cover_path   TEXT,
  user_created BOOLEAN NOT NULL DEFAULT 0, -- 1 = user playlist, 0 = scraper-created
  external_id  TEXT,
  source_url   TEXT,
  tags         JSON NOT NULL DEFAULT '[]',
  created_by   TEXT,
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX collection_source_url ON collection (source_url);

CREATE TABLE collection_item (
  id         TEXT PRIMARY KEY,
  collection TEXT NOT NULL,
  media      TEXT NOT NULL,                -- scene/gallery/image/performer/studio id
  position   INTEGER NOT NULL DEFAULT 0,
  added_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX collection_item_edge ON collection_item (collection, media);
CREATE INDEX collection_item_media ON collection_item (media);

-- Tag and/or label at a point or span. personal = 1 for user-created markers,
-- 0 for global ones (e.g. a plugin position tagger, see created_by).
CREATE TABLE scene_marker (
  id          TEXT PRIMARY KEY,
  media       TEXT NOT NULL,
  personal    BOOLEAN NOT NULL DEFAULT 0,
  tag         TEXT,
  label       TEXT,
  seconds     REAL NOT NULL CHECK (seconds >= 0),
  end_seconds REAL CHECK (end_seconds IS NULL OR end_seconds >= 0),
  created_by  TEXT,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX scene_marker_media ON scene_marker (media);
CREATE INDEX scene_marker_tag ON scene_marker (tag);

-- ---------------------------------------------------------------------------
-- Ingest and jobs

CREATE TABLE job (
  id              TEXT PRIMARY KEY,
  kind            TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'pending',
  title           TEXT,
  description     TEXT,
  plugin_name     TEXT NOT NULL DEFAULT '',
  target          TEXT,
  payload         JSON,
  result          JSON,
  error           TEXT,
  attempts        INTEGER NOT NULL DEFAULT 0,
  max_attempts    INTEGER NOT NULL DEFAULT 3,
  priority        INTEGER NOT NULL DEFAULT 0,
  timeout_seconds INTEGER,
  dedupe_key      TEXT,
  run_at          TEXT,
  started_at      TEXT,
  finished_at     TEXT,
  created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX job_dedupe ON job (dedupe_key);
CREATE INDEX job_status_kind ON job (status, kind);
CREATE INDEX job_priority_queue ON job (status, priority, run_at);
CREATE INDEX job_target ON job (target);

-- Raw plugin payloads. The fixed columns are indexed; the rest of the payload
-- lives in data.
CREATE TABLE observation (
  id          TEXT PRIMARY KEY,
  target      TEXT NOT NULL,
  plugin      TEXT NOT NULL,
  source_url  TEXT NOT NULL DEFAULT '',
  confidence  REAL NOT NULL DEFAULT 1.0 CHECK (confidence >= 0 AND confidence <= 1),
  status      TEXT NOT NULL DEFAULT 'pending',
  job         TEXT,
  observed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  data        JSON NOT NULL DEFAULT '{}'
);
CREATE INDEX observation_target_plugin ON observation (target, plugin);
CREATE INDEX observation_plugin ON observation (plugin);
CREATE INDEX observation_status ON observation (status);
CREATE INDEX observation_job ON observation (job);

-- Per-field merge history of canonical records.
CREATE TABLE changes (
  id             TEXT PRIMARY KEY,
  canonical      TEXT NOT NULL,
  observation    TEXT NOT NULL,
  plugin         TEXT NOT NULL DEFAULT '',
  key            TEXT NOT NULL,
  action         TEXT NOT NULL,
  value          JSON,
  original_value JSON,
  score          REAL NOT NULL DEFAULT 1.0,
  mode           TEXT NOT NULL DEFAULT 'auto',
  changed_by     TEXT,
  created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX changes_key ON changes (canonical, key);
CREATE INDEX changes_observation ON changes (observation);

CREATE TABLE subscription (
  id               TEXT PRIMARY KEY,
  target           TEXT NOT NULL,
  job_kind         TEXT NOT NULL DEFAULT 'scrape',
  plugin_name      TEXT NOT NULL DEFAULT '',
  interval_seconds INTEGER NOT NULL DEFAULT 86400,
  enabled          BOOLEAN NOT NULL DEFAULT 1,
  last_run_at      TEXT,
  next_run_at      TEXT,
  created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX subscription_target ON subscription (target, job_kind, plugin_name);
CREATE INDEX subscription_queue ON subscription (enabled, next_run_at);

-- ---------------------------------------------------------------------------
-- Configuration

CREATE TABLE plugin (
  id           TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  display_name TEXT,
  icon_url     TEXT,
  description  TEXT,
  version      TEXT NOT NULL DEFAULT '',
  capabilities JSON NOT NULL DEFAULT '[]',
  domains      JSON NOT NULL DEFAULT '[]',
  enabled      BOOLEAN NOT NULL DEFAULT 1,
  installed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX plugin_name ON plugin (name);

CREATE TABLE setting (
  id          TEXT PRIMARY KEY,
  scope       TEXT NOT NULL,
  plugin      TEXT,                        -- plugin id for plugin-scoped settings
  key         TEXT NOT NULL,
  value       JSON,
  description TEXT,
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX setting_lookup ON setting (scope, ifnull(plugin, ''), key);

-- Expiring key-value cache (replaces Redis): stream cache index, proxy
-- sessions. expires_at is unix milliseconds; NULL never expires.
CREATE TABLE kv (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL,
  expires_at INTEGER
);
CREATE INDEX kv_expires_at ON kv (expires_at);

-- ---------------------------------------------------------------------------
-- Personal library data

CREATE TABLE watch_history (
  id               TEXT PRIMARY KEY,
  media            TEXT NOT NULL,
  started_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  finished_at      TEXT,
  progress_seconds INTEGER NOT NULL DEFAULT 0,
  duration_seconds INTEGER,
  completed        BOOLEAN NOT NULL DEFAULT 0,
  updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX watch_history_media ON watch_history (media);
CREATE INDEX watch_history_updated_at ON watch_history (updated_at);

CREATE TABLE user_rating (
  id         TEXT PRIMARY KEY,
  media      TEXT NOT NULL,
  rating     REAL NOT NULL CHECK (rating >= 1 AND rating <= 10),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX user_rating_media ON user_rating (media);

CREATE TABLE watchlist (
  id         TEXT PRIMARY KEY,
  media      TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX watchlist_media ON watchlist (media);

CREATE TABLE search_history (
  id               TEXT PRIMARY KEY,
  query            TEXT NOT NULL,
  normalized_query TEXT NOT NULL,
  plugins          JSON NOT NULL DEFAULT '[]',
  result_count     INTEGER NOT NULL DEFAULT 0,
  created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX search_history_created_at ON search_history (created_at);

CREATE TABLE search_term (
  id               TEXT PRIMARY KEY,
  normalized_query TEXT NOT NULL,
  uses             INTEGER NOT NULL DEFAULT 0,
  last_used        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX search_term_query ON search_term (normalized_query);

CREATE TABLE o_event (
  id         TEXT PRIMARY KEY,
  media      TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX o_event_media ON o_event (media);

CREATE TABLE preference_event (
  id         TEXT PRIMARY KEY,
  chosen     TEXT NOT NULL,
  rejected   TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE saved_filter (
  id           TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  search       TEXT,
  studio_id    TEXT,
  performer_id TEXT,
  tag_id       TEXT,
  min_rating   REAL,
  min_duration INTEGER,
  max_duration INTEGER,
  date_from    TEXT,
  date_to      TEXT,
  sort         TEXT,
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE blocklist_entry (
  id         TEXT PRIMARY KEY,
  kind       TEXT NOT NULL CHECK (kind IN ('tag', 'performer', 'studio')),
  target     TEXT NOT NULL,
  label      TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX blocklist_target ON blocklist_entry (target);

-- ---------------------------------------------------------------------------
-- updated_at maintenance: bump on every update that doesn't set it itself.

CREATE TRIGGER studio_updated_at AFTER UPDATE ON studio WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE studio SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER tag_updated_at AFTER UPDATE ON tag WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE tag SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER performer_updated_at AFTER UPDATE ON performer WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE performer SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER scene_updated_at AFTER UPDATE ON scene WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE scene SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER gallery_updated_at AFTER UPDATE ON gallery WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE gallery SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER image_updated_at AFTER UPDATE ON image WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE image SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER stream_updated_at AFTER UPDATE ON stream WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE stream SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER collection_updated_at AFTER UPDATE ON collection WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE collection SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER job_updated_at AFTER UPDATE ON job WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE job SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER subscription_updated_at AFTER UPDATE ON subscription WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE subscription SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER plugin_updated_at AFTER UPDATE ON plugin WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE plugin SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER setting_updated_at AFTER UPDATE ON setting WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE setting SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER watch_history_updated_at AFTER UPDATE ON watch_history WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE watch_history SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
CREATE TRIGGER user_rating_updated_at AFTER UPDATE ON user_rating WHEN NEW.updated_at IS OLD.updated_at
BEGIN UPDATE user_rating SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id; END;
