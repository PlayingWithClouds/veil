-- Generated migration — 3 operation(s).

CREATE TABLE SiteRelated (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  source TEXT NOT NULL REFERENCES Source(id),
  video TEXT NOT NULL REFERENCES Video(id),
  position INTEGER NOT NULL
);

CREATE INDEX SiteRelated_source_position_idx ON SiteRelated (source, position);

CREATE INDEX SiteRelated_video_idx ON SiteRelated (video);
