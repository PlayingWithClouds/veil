-- Generated migration — 2 operation(s).

CREATE TABLE MergeJournal (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  keptId TEXT NOT NULL,
  mergedId TEXT NOT NULL,
  mergedName TEXT,
  "rows" TEXT NOT NULL CHECK (json_valid("rows")),
  writtenBy TEXT NOT NULL CHECK (writtenBy IN ('user', 'plugin', 'automation', 'system')),
  undoneAt TEXT
);

CREATE INDEX MergeJournal_subjectKind_keptId_idx ON MergeJournal (subjectKind, keptId);
