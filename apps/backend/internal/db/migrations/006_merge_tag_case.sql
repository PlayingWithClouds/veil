-- Tags used to be matched case-sensitively, so "Anal" and "anal" became two
-- tags. Merge every case-insensitive group into one survivor (an all-lowercase
-- spelling when there is one, else the oldest), repoint every reference to it,
-- and drop listing noise that isn't a tag at all ("Uncategorized", "Sex", …).
-- Ingest matches tag names case-insensitively and skips the noise from here on.

CREATE TEMP TABLE tag_survivor AS
SELECT tag.id AS old_id,
  (SELECT candidate.id FROM tag AS candidate
    WHERE lower(candidate.name) = lower(tag.name)
    ORDER BY candidate.name = lower(candidate.name) DESC, candidate.created_at, candidate.id
    LIMIT 1) AS new_id
FROM tag;

-- Noise tags map to nothing: they are removed rather than merged.
UPDATE tag_survivor SET new_id = NULL
WHERE old_id IN (
  SELECT id FROM tag WHERE lower(trim(name)) IN (
    'uncategorized', 'other', 'others', 'sex', 'porn', 'porno', 'xxx', 'video', 'videos',
    'hd', 'hd porn', 'free porn', 'porn videos', 'sex videos', 'full video', 'full movie'
  )
);

-- Only tags that change (merged away or removed).
CREATE TEMP TABLE tag_changed AS
SELECT old_id, new_id FROM tag_survivor WHERE new_id IS NULL OR new_id <> old_id;

-- Rewrites a JSON tag-id list: surviving ids, first occurrence order, no repeats.
UPDATE scene SET tags = (
  SELECT json_group_array(new_id) FROM (
    SELECT survivor.new_id, min(entry.key) AS position
    FROM json_each(scene.tags) AS entry JOIN tag_survivor AS survivor ON survivor.old_id = entry.value
    WHERE survivor.new_id IS NOT NULL GROUP BY survivor.new_id ORDER BY position))
WHERE EXISTS (SELECT 1 FROM json_each(scene.tags) AS entry WHERE entry.value IN (SELECT old_id FROM tag_changed));

UPDATE gallery SET tags = (
  SELECT json_group_array(new_id) FROM (
    SELECT survivor.new_id, min(entry.key) AS position
    FROM json_each(gallery.tags) AS entry JOIN tag_survivor AS survivor ON survivor.old_id = entry.value
    WHERE survivor.new_id IS NOT NULL GROUP BY survivor.new_id ORDER BY position))
WHERE EXISTS (SELECT 1 FROM json_each(gallery.tags) AS entry WHERE entry.value IN (SELECT old_id FROM tag_changed));

UPDATE image SET tags = (
  SELECT json_group_array(new_id) FROM (
    SELECT survivor.new_id, min(entry.key) AS position
    FROM json_each(image.tags) AS entry JOIN tag_survivor AS survivor ON survivor.old_id = entry.value
    WHERE survivor.new_id IS NOT NULL GROUP BY survivor.new_id ORDER BY position))
WHERE EXISTS (SELECT 1 FROM json_each(image.tags) AS entry WHERE entry.value IN (SELECT old_id FROM tag_changed));

UPDATE performer SET tags = (
  SELECT json_group_array(new_id) FROM (
    SELECT survivor.new_id, min(entry.key) AS position
    FROM json_each(performer.tags) AS entry JOIN tag_survivor AS survivor ON survivor.old_id = entry.value
    WHERE survivor.new_id IS NOT NULL GROUP BY survivor.new_id ORDER BY position))
WHERE EXISTS (SELECT 1 FROM json_each(performer.tags) AS entry WHERE entry.value IN (SELECT old_id FROM tag_changed));

UPDATE studio SET tags = (
  SELECT json_group_array(new_id) FROM (
    SELECT survivor.new_id, min(entry.key) AS position
    FROM json_each(studio.tags) AS entry JOIN tag_survivor AS survivor ON survivor.old_id = entry.value
    WHERE survivor.new_id IS NOT NULL GROUP BY survivor.new_id ORDER BY position))
WHERE EXISTS (SELECT 1 FROM json_each(studio.tags) AS entry WHERE entry.value IN (SELECT old_id FROM tag_changed));

UPDATE collection SET tags = (
  SELECT json_group_array(new_id) FROM (
    SELECT survivor.new_id, min(entry.key) AS position
    FROM json_each(collection.tags) AS entry JOIN tag_survivor AS survivor ON survivor.old_id = entry.value
    WHERE survivor.new_id IS NOT NULL GROUP BY survivor.new_id ORDER BY position))
WHERE EXISTS (SELECT 1 FROM json_each(collection.tags) AS entry WHERE entry.value IN (SELECT old_id FROM tag_changed));

-- Markers keep their label; a removed tag just leaves them untagged.
UPDATE scene_marker SET tag = (SELECT new_id FROM tag_changed WHERE old_id = scene_marker.tag)
WHERE tag IN (SELECT old_id FROM tag_changed);

UPDATE saved_filter SET tag_id = (SELECT new_id FROM tag_changed WHERE old_id = saved_filter.tag_id)
WHERE tag_id IN (SELECT old_id FROM tag_changed);

-- Blocklist and subscriptions are unique per target: when the survivor is
-- already there the merged-away row goes, otherwise it is repointed.
DELETE FROM blocklist_entry
WHERE target IN (SELECT old_id FROM tag_changed)
  AND ((SELECT new_id FROM tag_changed WHERE old_id = target) IS NULL
    OR (SELECT new_id FROM tag_changed WHERE old_id = target) IN (SELECT target FROM blocklist_entry));
UPDATE blocklist_entry SET target = (SELECT new_id FROM tag_changed WHERE old_id = target)
WHERE target IN (SELECT old_id FROM tag_changed);

DELETE FROM search_subscription
WHERE entity IN (SELECT old_id FROM tag_changed WHERE new_id IS NOT NULL)
  AND (SELECT new_id FROM tag_changed WHERE old_id = entity) IN (SELECT entity FROM search_subscription WHERE entity IS NOT NULL);
UPDATE search_subscription SET entity = (SELECT new_id FROM tag_changed WHERE old_id = entity)
WHERE entity IN (SELECT old_id FROM tag_changed WHERE new_id IS NOT NULL);
DELETE FROM search_subscription_item WHERE subscription NOT IN (SELECT id FROM search_subscription);

DELETE FROM tag WHERE id IN (SELECT old_id FROM tag_changed);

DROP TABLE tag_changed;
DROP TABLE tag_survivor;
