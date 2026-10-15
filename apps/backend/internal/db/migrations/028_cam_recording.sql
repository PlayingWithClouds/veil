-- Recording of live cam rooms. auto_record makes the watcher start a recording
-- whenever a followed model comes online. A cam_recording row exists from the
-- start of a capture; when it ends, the recorded segments become a scene.
ALTER TABLE cam_model ADD COLUMN auto_record BOOLEAN NOT NULL DEFAULT 0;

CREATE TABLE cam_recording (
  id               TEXT PRIMARY KEY,
  cam_model        TEXT NOT NULL,           -- cam_model id
  scene            TEXT,                    -- scene created from the recording
  status           TEXT NOT NULL,           -- recording | finished | failed
  stop_reason      TEXT,                    -- stopped | offline | max_duration | restart | too_short | error
  started_at       TEXT NOT NULL,
  ended_at         TEXT,
  duration_seconds REAL NOT NULL DEFAULT 0,
  size_bytes       INTEGER NOT NULL DEFAULT 0,
  created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX cam_recording_model ON cam_recording (cam_model, started_at);
CREATE INDEX cam_recording_status ON cam_recording (status) WHERE status = 'recording';
