-- Live cam rooms seen in a cam plugin's listings. A room comes and goes, so the
-- row keeps what the listing last said and when it was last online; `followed`
-- pins it to the Following row. source_url is the identity (the room's page).
CREATE TABLE cam_model (
  id           TEXT PRIMARY KEY,
  plugin       TEXT NOT NULL,
  username     TEXT NOT NULL,
  display_name TEXT NOT NULL,
  source_url   TEXT NOT NULL,
  thumbnail_url TEXT,
  tags         JSON NOT NULL DEFAULT '[]',  -- tag names as the site writes them
  gender       TEXT,
  subject      TEXT,
  viewers      INTEGER NOT NULL DEFAULT 0,
  online       BOOLEAN NOT NULL DEFAULT 0,
  last_seen_online_at TEXT,
  followed     BOOLEAN NOT NULL DEFAULT 0,
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX cam_model_source_url ON cam_model (source_url);
CREATE INDEX cam_model_plugin_username ON cam_model (plugin, username);
CREATE INDEX cam_model_followed ON cam_model (followed) WHERE followed = 1;
