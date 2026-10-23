-- Full-text index behind library search, ranked by bm25. One row per scene
-- holds the text a search matches: its title plus the names of its performers,
-- studio and tags (the scene row only stores their ids), one name per line. Triggers keep it in
-- step with scene, performer, studio and tag edits.
CREATE VIRTUAL TABLE scene_search USING fts5(
  scene_id UNINDEXED,
  title,
  performers,
  studio,
  tags,
  tokenize = 'unicode61 remove_diacritics 2'
);

INSERT INTO scene_search (scene_id, title, performers, studio, tags)
SELECT scene.id, scene.title,
  coalesce((SELECT group_concat(performer.name, char(10)) FROM performer
            WHERE performer.id IN (SELECT value FROM json_each(scene.performers))), ''),
  coalesce((SELECT studio.name FROM studio WHERE studio.id = scene.studio), ''),
  coalesce((SELECT group_concat(tag.name, char(10)) FROM tag
            WHERE tag.id IN (SELECT value FROM json_each(scene.tags))), '')
FROM scene;

CREATE TRIGGER scene_search_insert AFTER INSERT ON scene
BEGIN
  INSERT INTO scene_search (scene_id, title, performers, studio, tags)
  VALUES (NEW.id, NEW.title,
    coalesce((SELECT group_concat(performer.name, char(10)) FROM performer
              WHERE performer.id IN (SELECT value FROM json_each(NEW.performers))), ''),
    coalesce((SELECT studio.name FROM studio WHERE studio.id = NEW.studio), ''),
    coalesce((SELECT group_concat(tag.name, char(10)) FROM tag
              WHERE tag.id IN (SELECT value FROM json_each(NEW.tags))), ''));
END;

CREATE TRIGGER scene_search_update AFTER UPDATE OF title, performers, studio, tags ON scene
BEGIN
  DELETE FROM scene_search WHERE scene_id = NEW.id;
  INSERT INTO scene_search (scene_id, title, performers, studio, tags)
  VALUES (NEW.id, NEW.title,
    coalesce((SELECT group_concat(performer.name, char(10)) FROM performer
              WHERE performer.id IN (SELECT value FROM json_each(NEW.performers))), ''),
    coalesce((SELECT studio.name FROM studio WHERE studio.id = NEW.studio), ''),
    coalesce((SELECT group_concat(tag.name, char(10)) FROM tag
              WHERE tag.id IN (SELECT value FROM json_each(NEW.tags))), ''));
END;

CREATE TRIGGER scene_search_delete AFTER DELETE ON scene
BEGIN
  DELETE FROM scene_search WHERE scene_id = OLD.id;
END;

-- A renamed performer, studio or tag changes the text of the scenes using it:
-- touching the scene's title column re-runs scene_search_update.
CREATE TRIGGER scene_search_performer_renamed AFTER UPDATE OF name ON performer
BEGIN
  UPDATE scene SET title = title
  WHERE EXISTS (SELECT 1 FROM json_each(scene.performers) WHERE value = NEW.id);
END;

CREATE TRIGGER scene_search_studio_renamed AFTER UPDATE OF name ON studio
BEGIN
  UPDATE scene SET title = title WHERE studio = NEW.id;
END;

CREATE TRIGGER scene_search_tag_renamed AFTER UPDATE OF name ON tag
BEGIN
  UPDATE scene SET title = title
  WHERE EXISTS (SELECT 1 FROM json_each(scene.tags) WHERE value = NEW.id);
END;
