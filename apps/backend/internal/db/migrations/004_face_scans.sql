-- Generated migration — 3 operation(s).

CREATE TABLE FaceScan (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  model TEXT NOT NULL,
  picture TEXT,
  frames INTEGER NOT NULL,
  faces INTEGER NOT NULL
);

ALTER TABLE Embedding ADD COLUMN scan TEXT REFERENCES FaceScan(id);

CREATE INDEX FaceScan_subjectKind_subjectId_model_idx ON FaceScan (subjectKind, subjectId, model);
