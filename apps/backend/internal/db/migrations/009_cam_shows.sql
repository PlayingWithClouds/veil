-- Generated migration — 2 operation(s).

CREATE TABLE CamShow (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  camModel TEXT NOT NULL REFERENCES CamModel(id),
  startedAt TEXT NOT NULL,
  endedAt TEXT NOT NULL
);

CREATE INDEX CamShow_camModel_startedAt_idx ON CamShow (camModel, startedAt);
