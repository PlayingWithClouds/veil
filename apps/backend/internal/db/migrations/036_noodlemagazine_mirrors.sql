-- ukdevilz.com and mat6tube.com are copies of noodlemagazine.com (same videos,
-- same ids), and their plugins were merged into the noodlemagazine plugin,
-- which fetches from the first copy that answers. What the old plugins stored
-- becomes the merged plugin's: a visit takes its plugin from the latest
-- observation, and the site counts once in the feed and in search learning.
UPDATE observation SET plugin = 'noodlemagazine' WHERE plugin IN ('ukdevilz', 'mat6tube');
UPDATE stream SET plugin_name = 'noodlemagazine' WHERE plugin_name IN ('ukdevilz', 'mat6tube');
UPDATE search_event SET plugin = 'noodlemagazine' WHERE plugin IN ('ukdevilz', 'mat6tube');
UPDATE OR IGNORE feed_discovery SET entity = 'noodlemagazine'
  WHERE source = 'catalog' AND entity IN ('ukdevilz', 'mat6tube');
DELETE FROM feed_discovery WHERE source = 'catalog' AND entity IN ('ukdevilz', 'mat6tube');

-- The merged plugin lists every video under its noodlemagazine.com page and
-- external id. A stub stored from a copy keeps its own, and also answers to
-- those (unless another record already has them), so listing the video again
-- finds it instead of storing it twice.
INSERT OR IGNORE INTO entity_identity (kind, value, record)
SELECT 'source_url', 'https://noodlemagazine.com' || substr(source_url, instr(source_url, '/watch/')), id
FROM scene
WHERE (source_url LIKE 'https://ukdevilz.com/watch/%' OR source_url LIKE 'https://mat6tube.com/watch/%')
  AND NOT EXISTS (SELECT 1 FROM scene AS other
    WHERE other.source_url = 'https://noodlemagazine.com' || substr(scene.source_url, instr(scene.source_url, '/watch/')));

INSERT OR IGNORE INTO entity_identity (kind, value, record)
SELECT 'external_id', 'noodlemagazine-' || substr(external_id, instr(external_id, '-') + 1), id
FROM scene
WHERE (external_id LIKE 'ukdevilz-%' OR external_id LIKE 'mat6tube-%')
  AND NOT EXISTS (SELECT 1 FROM scene AS other
    WHERE other.external_id = 'noodlemagazine-' || substr(scene.external_id, instr(scene.external_id, '-') + 1));
