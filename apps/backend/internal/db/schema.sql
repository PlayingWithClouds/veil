CREATE TABLE UserState (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  rating REAL,
  favorite INTEGER NOT NULL,
  notes TEXT,
  blocked INTEGER NOT NULL,
  progress REAL,
  viewCount INTEGER NOT NULL,
  lastViewedAt TEXT,
  fields TEXT NOT NULL CHECK (json_valid(fields))
);
CREATE UNIQUE INDEX UserState_subjectKind_subjectId_idx ON UserState (subjectKind, subjectId);

CREATE TABLE Collection (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('user', 'watchlist', 'queue')),
  details TEXT,
  cover TEXT CHECK (json_valid(cover))
);

CREATE TABLE CollectionItem (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  collection TEXT NOT NULL REFERENCES Collection(id),
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  position INTEGER NOT NULL
);
CREATE INDEX CollectionItem_collection_position_idx ON CollectionItem (collection, position);

CREATE TABLE SavedFilter (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  "filter" TEXT NOT NULL CHECK (json_valid("filter")),
  pinned INTEGER NOT NULL
);

CREATE TABLE Follow (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  "filter" TEXT NOT NULL CHECK (json_valid("filter")),
  targetKind TEXT CHECK (targetKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  targetId TEXT,
  intervalHours INTEGER NOT NULL,
  enabled INTEGER NOT NULL,
  notify INTEGER NOT NULL,
  lastRunAt TEXT,
  nextRunAt TEXT,
  lastError TEXT
);

CREATE TABLE Event (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  at TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('view', 'progress', 'finish', 'o', 'search', 'impression', 'choice', 'heatSpan', 'skip', 'like')),
  subjectKind TEXT CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT,
  data TEXT CHECK (json_valid(data)),
  device TEXT REFERENCES Device(id),
  sessionId TEXT
);
CREATE INDEX Event_kind_at_idx ON Event (kind, at);
CREATE INDEX Event_subjectKind_subjectId_idx ON Event (subjectKind, subjectId);

CREATE TABLE "Limit" (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  kind TEXT NOT NULL CHECK (kind IN ('watchTime', 'sessions')),
  "window" TEXT NOT NULL CHECK ("window" IN ('day', 'week')),
  amount INTEGER NOT NULL,
  "action" TEXT NOT NULL CHECK ("action" IN ('remind', 'lock')),
  enabled INTEGER NOT NULL
);

CREATE TABLE Account (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  plugin TEXT NOT NULL,
  label TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('ok', 'expired', 'needsLogin', 'failed')),
  device TEXT NOT NULL REFERENCES Device(id),
  credentials BLOB,
  session BLOB,
  expiresAt TEXT,
  lastCheckedAt TEXT
);

CREATE TABLE Device (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  platform TEXT NOT NULL CHECK (platform IN ('server', 'desktop', 'android', 'tv', 'vr')),
  lastSeenAt TEXT,
  lastSyncAt TEXT,
  "current" INTEGER NOT NULL
);

CREATE TABLE Job (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  kind TEXT NOT NULL CHECK (kind IN ('download', 'record', 'scrape', 'followRun', 'fingerprint', 'embed', 'trickplay', 'transcode', 'libraryScan', 'automationRun', 'snapshot')),
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled')),
  subject TEXT CHECK (json_valid(subject)),
  payload TEXT CHECK (json_valid(payload)),
  progress REAL NOT NULL,
  error TEXT,
  attempts INTEGER NOT NULL,
  runAt TEXT NOT NULL,
  startedAt TEXT,
  finishedAt TEXT,
  device TEXT REFERENCES Device(id)
);
CREATE INDEX Job_status_runAt_idx ON Job (status, runAt);

CREATE TABLE ApiToken (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  scopes TEXT NOT NULL CHECK (json_valid(scopes)),
  tokenHash BLOB NOT NULL,
  lastUsedAt TEXT,
  expiresAt TEXT
);

CREATE TABLE Plugin (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  displayName TEXT,
  version TEXT NOT NULL,
  enabled INTEGER NOT NULL,
  origin TEXT NOT NULL CHECK (origin IN ('store', 'local')),
  contributions TEXT NOT NULL CHECK (json_valid(contributions)),
  settings TEXT CHECK (json_valid(settings))
);
CREATE UNIQUE INDEX Plugin_name_idx ON Plugin (name);

CREATE TABLE Source (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  plugin TEXT NOT NULL,
  externalId TEXT NOT NULL,
  url TEXT NOT NULL,
  title TEXT,
  image TEXT CHECK (json_valid(image)),
  snapshot TEXT CHECK (json_valid(snapshot)),
  fetchedAt TEXT
);
CREATE UNIQUE INDEX Source_plugin_externalId_idx ON Source (plugin, externalId);

CREATE TABLE Video (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  title TEXT NOT NULL,
  details TEXT,
  date TEXT,
  cover TEXT CHECK (json_valid(cover)),
  studio TEXT REFERENCES Studio(id),
  performers TEXT NOT NULL CHECK (json_valid(performers)),
  tags TEXT NOT NULL CHECK (json_valid(tags)),
  status TEXT NOT NULL CHECK (status IN ('available', 'wanted', 'unavailable')),
  durationSeconds REAL,
  width INTEGER,
  height INTEGER,
  vr TEXT CHECK (json_valid(vr)),
  preview TEXT CHECK (json_valid(preview)),
  trickplay TEXT CHECK (json_valid(trickplay)),
  gallery TEXT REFERENCES Gallery(id),
  recordedFrom TEXT REFERENCES CamModel(id)
);

CREATE TABLE GroupItem (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  "group" TEXT NOT NULL REFERENCES "Group"(id),
  video TEXT NOT NULL REFERENCES Video(id),
  position INTEGER NOT NULL
);
CREATE UNIQUE INDEX GroupItem_group_video_idx ON GroupItem ("group", video);

CREATE TABLE Gallery (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  title TEXT NOT NULL,
  details TEXT,
  date TEXT,
  cover TEXT CHECK (json_valid(cover)),
  studio TEXT REFERENCES Studio(id),
  performers TEXT NOT NULL CHECK (json_valid(performers)),
  tags TEXT NOT NULL CHECK (json_valid(tags)),
  status TEXT NOT NULL CHECK (status IN ('available', 'wanted', 'unavailable')),
  imageCount INTEGER NOT NULL,
  video TEXT REFERENCES Video(id)
);

CREATE TABLE Image (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  gallery TEXT NOT NULL REFERENCES Gallery(id),
  position INTEGER NOT NULL,
  asset TEXT NOT NULL CHECK (json_valid(asset)),
  width INTEGER,
  height INTEGER
);
CREATE INDEX Image_gallery_position_idx ON Image (gallery, position);

CREATE TABLE Audio (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  title TEXT NOT NULL,
  details TEXT,
  date TEXT,
  cover TEXT CHECK (json_valid(cover)),
  studio TEXT REFERENCES Studio(id),
  performers TEXT NOT NULL CHECK (json_valid(performers)),
  tags TEXT NOT NULL CHECK (json_valid(tags)),
  status TEXT NOT NULL CHECK (status IN ('available', 'wanted', 'unavailable')),
  durationSeconds REAL
);

CREATE TABLE Story (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  title TEXT NOT NULL,
  details TEXT,
  date TEXT,
  cover TEXT CHECK (json_valid(cover)),
  studio TEXT REFERENCES Studio(id),
  performers TEXT NOT NULL CHECK (json_valid(performers)),
  tags TEXT NOT NULL CHECK (json_valid(tags)),
  status TEXT NOT NULL CHECK (status IN ('available', 'wanted', 'unavailable')),
  chapterCount INTEGER NOT NULL,
  wordCount INTEGER,
  complete INTEGER NOT NULL
);

CREATE TABLE StoryChapter (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  story TEXT NOT NULL REFERENCES Story(id),
  position INTEGER NOT NULL,
  title TEXT,
  text TEXT NOT NULL,
  wordCount INTEGER NOT NULL
);
CREATE UNIQUE INDEX StoryChapter_story_position_idx ON StoryChapter (story, position);

CREATE TABLE Post (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  title TEXT NOT NULL,
  details TEXT,
  date TEXT,
  cover TEXT CHECK (json_valid(cover)),
  studio TEXT REFERENCES Studio(id),
  performers TEXT NOT NULL CHECK (json_valid(performers)),
  tags TEXT NOT NULL CHECK (json_valid(tags)),
  status TEXT NOT NULL CHECK (status IN ('available', 'wanted', 'unavailable')),
  creator TEXT NOT NULL REFERENCES Performer(id),
  text TEXT,
  attachments TEXT NOT NULL CHECK (json_valid(attachments)),
  locked INTEGER NOT NULL,
  priceCents INTEGER,
  currency TEXT
);

CREATE TABLE Performer (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  aliases TEXT NOT NULL CHECK (json_valid(aliases)),
  image TEXT CHECK (json_valid(image)),
  details TEXT,
  attributes TEXT NOT NULL CHECK (json_valid(attributes)),
  links TEXT NOT NULL CHECK (json_valid(links)),
  tags TEXT NOT NULL CHECK (json_valid(tags))
);

CREATE TABLE Studio (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  aliases TEXT NOT NULL CHECK (json_valid(aliases)),
  image TEXT CHECK (json_valid(image)),
  details TEXT,
  parent TEXT REFERENCES Studio(id),
  url TEXT,
  tags TEXT NOT NULL CHECK (json_valid(tags))
);

CREATE TABLE CamModel (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  plugin TEXT NOT NULL,
  username TEXT NOT NULL,
  displayName TEXT NOT NULL,
  performer TEXT REFERENCES Performer(id),
  thumbnail TEXT CHECK (json_valid(thumbnail)),
  tags TEXT NOT NULL CHECK (json_valid(tags)),
  gender TEXT,
  live TEXT CHECK (json_valid(live)),
  lastSeenOnlineAt TEXT,
  autoRecord INTEGER NOT NULL,
  recordQuality TEXT
);
CREATE UNIQUE INDEX CamModel_plugin_username_idx ON CamModel (plugin, username);

CREATE TABLE "Group" (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  title TEXT NOT NULL,
  details TEXT,
  date TEXT,
  cover TEXT CHECK (json_valid(cover)),
  studio TEXT REFERENCES Studio(id)
);

CREATE TABLE Tag (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL UNIQUE,
  aliases TEXT NOT NULL CHECK (json_valid(aliases))
);

CREATE TABLE Category (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  parent TEXT REFERENCES Category(id),
  matchTags TEXT NOT NULL CHECK (json_valid(matchTags)),
  cover TEXT CHECK (json_valid(cover)),
  position INTEGER NOT NULL
);

CREATE TABLE FieldDef (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  kinds TEXT NOT NULL CHECK (json_valid(kinds)),
  fieldType TEXT NOT NULL CHECK (fieldType IN ('text', 'number', 'boolean', 'date', 'choice', 'choices', 'url')),
  options TEXT CHECK (json_valid(options))
);

CREATE TABLE Automation (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  enabled INTEGER NOT NULL,
  "trigger" TEXT NOT NULL CHECK (json_valid("trigger")),
  steps TEXT NOT NULL CHECK (json_valid(steps)),
  builtIn INTEGER NOT NULL
);

CREATE TABLE AutomationRun (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  automation TEXT NOT NULL REFERENCES Automation(id),
  startedAt TEXT NOT NULL,
  finishedAt TEXT,
  status TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed', 'undone')),
  dryRun INTEGER NOT NULL,
  triggerData TEXT CHECK (json_valid(triggerData)),
  steps TEXT NOT NULL CHECK (json_valid(steps))
);
CREATE INDEX AutomationRun_automation_startedAt_idx ON AutomationRun (automation, startedAt);

CREATE TABLE NotificationChannel (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  kind TEXT NOT NULL CHECK (kind IN ('inApp', 'push', 'ntfy', 'webhook')),
  name TEXT NOT NULL,
  config TEXT CHECK (json_valid(config)),
  enabled INTEGER NOT NULL
);

CREATE TABLE Notification (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  at TEXT NOT NULL,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT,
  subject TEXT CHECK (json_valid(subject)),
  read INTEGER NOT NULL
);
CREATE INDEX Notification_read_at_idx ON Notification (read, at);

CREATE TABLE Stream (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  source TEXT NOT NULL REFERENCES Source(id),
  url TEXT NOT NULL,
  mimeType TEXT NOT NULL,
  quality TEXT,
  height INTEGER,
  headers TEXT NOT NULL CHECK (json_valid(headers)),
  expiresAt TEXT,
  speedKbps INTEGER
);
CREATE INDEX Stream_subjectKind_subjectId_idx ON Stream (subjectKind, subjectId);

CREATE TABLE StorageLocation (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  device TEXT NOT NULL REFERENCES Device(id),
  kind TEXT NOT NULL CHECK (kind IN ('library', 'cache', 'archive')),
  uri TEXT NOT NULL,
  quotaBytes INTEGER,
  usedBytes INTEGER NOT NULL
);

CREATE TABLE LibraryRoot (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  location TEXT NOT NULL REFERENCES StorageLocation(id),
  path TEXT NOT NULL,
  scanIntervalMinutes INTEGER,
  lastScanAt TEXT,
  organizeTemplate TEXT
);

CREATE TABLE MediaFile (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  location TEXT NOT NULL REFERENCES StorageLocation(id),
  path TEXT NOT NULL,
  sizeBytes INTEGER NOT NULL,
  mimeType TEXT NOT NULL,
  quality TEXT,
  width INTEGER,
  height INTEGER,
  durationSeconds REAL,
  libraryRoot TEXT REFERENCES LibraryRoot(id),
  missing INTEGER NOT NULL
);
CREATE UNIQUE INDEX MediaFile_location_path_idx ON MediaFile (location, path);

CREATE TABLE Segment (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  startSeconds REAL NOT NULL,
  endSeconds REAL,
  title TEXT,
  tag TEXT REFERENCES Tag(id),
  performers TEXT NOT NULL CHECK (json_valid(performers)),
  origin TEXT NOT NULL CHECK (origin IN ('user', 'plugin', 'ai'))
);
CREATE INDEX Segment_subjectKind_subjectId_idx ON Segment (subjectKind, subjectId);

CREATE TABLE Clip (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  video TEXT NOT NULL REFERENCES Video(id),
  startSeconds REAL NOT NULL,
  endSeconds REAL NOT NULL,
  title TEXT,
  cover TEXT CHECK (json_valid(cover)),
  file TEXT REFERENCES MediaFile(id)
);

CREATE TABLE Subtitle (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  language TEXT NOT NULL,
  format TEXT NOT NULL,
  asset TEXT NOT NULL CHECK (json_valid(asset)),
  origin TEXT NOT NULL CHECK (origin IN ('site', 'generated', 'translated'))
);
CREATE INDEX Subtitle_subjectKind_subjectId_idx ON Subtitle (subjectKind, subjectId);

CREATE TABLE Script (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('funscript')),
  asset TEXT NOT NULL CHECK (json_valid(asset)),
  axis TEXT,
  writtenBy TEXT NOT NULL CHECK (writtenBy IN ('user', 'plugin', 'automation', 'system'))
);
CREATE INDEX Script_subjectKind_subjectId_idx ON Script (subjectKind, subjectId);

CREATE TABLE FieldSource (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  fieldOrdinal INTEGER NOT NULL,
  writtenBy TEXT NOT NULL CHECK (writtenBy IN ('user', 'plugin', 'automation', 'system')),
  source TEXT REFERENCES Source(id),
  locked INTEGER NOT NULL
);
CREATE UNIQUE INDEX FieldSource_subjectKind_subjectId_fieldOrdinal_idx ON FieldSource (subjectKind, subjectId, fieldOrdinal);

CREATE TABLE Revision (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  fieldOrdinal INTEGER NOT NULL,
  "before" TEXT CHECK (json_valid("before")),
  "after" TEXT CHECK (json_valid("after")),
  writtenBy TEXT NOT NULL CHECK (writtenBy IN ('user', 'plugin', 'automation', 'system')),
  actorName TEXT,
  automationRunId TEXT
);
CREATE INDEX Revision_subjectKind_subjectId_idx ON Revision (subjectKind, subjectId);

CREATE TABLE Fingerprint (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  fileId TEXT,
  algorithm TEXT NOT NULL CHECK (algorithm IN ('phash', 'oshash', 'md5', 'audio')),
  value TEXT NOT NULL,
  durationSeconds REAL
);
CREATE INDEX Fingerprint_algorithm_value_idx ON Fingerprint (algorithm, value);

CREATE TABLE Embedding (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  subjectId TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('face', 'visual', 'text')),
  model TEXT NOT NULL,
  vector BLOB NOT NULL,
  atSeconds REAL,
  region TEXT CHECK (json_valid(region))
);
CREATE INDEX Embedding_kind_model_idx ON Embedding (kind, model);

CREATE TABLE Alike (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  subjectKind TEXT NOT NULL CHECK (subjectKind IN ('video', 'gallery', 'image', 'audio', 'story', 'post', 'performer', 'studio', 'camModel', 'group', 'tag', 'category', 'collection', 'segment', 'clip')),
  firstId TEXT NOT NULL,
  secondId TEXT NOT NULL,
  score REAL NOT NULL,
  reason TEXT,
  dismissed INTEGER NOT NULL
);
CREATE UNIQUE INDEX Alike_subjectKind_firstId_secondId_idx ON Alike (subjectKind, firstId, secondId);

CREATE TABLE Comment (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  source TEXT NOT NULL REFERENCES Source(id),
  author TEXT,
  text TEXT NOT NULL,
  postedAt TEXT,
  likes INTEGER,
  replyTo TEXT,
  mentions TEXT NOT NULL CHECK (json_valid(mentions))
);
CREATE INDEX Comment_source_idx ON Comment (source);
