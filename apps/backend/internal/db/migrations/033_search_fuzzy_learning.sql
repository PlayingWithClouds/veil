-- Typo-tolerant library search and what the search page learns from clicks.

-- Trigram twin of scene_search: the same four columns, tokenized into
-- three-letter pieces, so a misspelled word still shares most of its pieces
-- with the right one. It only supplies candidates when the exact index finds
-- little; the code then keeps those whose words are within a typo or two of
-- the query's. The triggers mirror scene_search's (the rename triggers there
-- touch scene.title, which fires scene_search_fuzzy_update too).
CREATE VIRTUAL TABLE scene_search_fuzzy USING fts5(
  scene_id UNINDEXED,
  title,
  performers,
  studio,
  tags,
  tokenize = 'trigram remove_diacritics 1'
);

INSERT INTO scene_search_fuzzy (scene_id, title, performers, studio, tags)
SELECT scene_id, title, performers, studio, tags FROM scene_search;

CREATE TRIGGER scene_search_fuzzy_insert AFTER INSERT ON scene
BEGIN
  INSERT INTO scene_search_fuzzy (scene_id, title, performers, studio, tags)
  VALUES (NEW.id, NEW.title,
    coalesce((SELECT group_concat(performer.name, char(10)) FROM performer
              WHERE performer.id IN (SELECT value FROM json_each(NEW.performers))), ''),
    coalesce((SELECT studio.name FROM studio WHERE studio.id = NEW.studio), ''),
    coalesce((SELECT group_concat(tag.name, char(10)) FROM tag
              WHERE tag.id IN (SELECT value FROM json_each(NEW.tags))), ''));
END;

CREATE TRIGGER scene_search_fuzzy_update AFTER UPDATE OF title, performers, studio, tags ON scene
BEGIN
  DELETE FROM scene_search_fuzzy WHERE scene_id = NEW.id;
  INSERT INTO scene_search_fuzzy (scene_id, title, performers, studio, tags)
  VALUES (NEW.id, NEW.title,
    coalesce((SELECT group_concat(performer.name, char(10)) FROM performer
              WHERE performer.id IN (SELECT value FROM json_each(NEW.performers))), ''),
    coalesce((SELECT studio.name FROM studio WHERE studio.id = NEW.studio), ''),
    coalesce((SELECT group_concat(tag.name, char(10)) FROM tag
              WHERE tag.id IN (SELECT value FROM json_each(NEW.tags))), ''));
END;

CREATE TRIGGER scene_search_fuzzy_delete AFTER DELETE ON scene
BEGIN
  DELETE FROM scene_search_fuzzy WHERE scene_id = OLD.id;
END;

-- What the user did with search results, logged by the search page through
-- recordImpressions (surface "search", with the query). tokens is the query's
-- distinct words, sorted, as a JSON array; plugin is the result's source
-- ("db" for the library). Kept 90 days. Read by searchrank's history: a click
-- memory per (query, scene) and a click rate per (query word, plugin).
CREATE TABLE search_event (
  id         TEXT PRIMARY KEY,
  tokens     JSON NOT NULL,
  media      TEXT NOT NULL,
  plugin     TEXT NOT NULL DEFAULT '',
  kind       TEXT NOT NULL CHECK (kind IN ('shown', 'click')),
  position   INTEGER,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX search_event_kind_created ON search_event (kind, created_at);
CREATE INDEX search_event_media ON search_event (media);
