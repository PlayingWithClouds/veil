-- Scenes the background feeder pulled from the sites for the recommendation
-- feed, without a search: a plugin's catalog listing (source 'catalog',
-- entity = plugin name) or the site pages of top-affinity performers, studios
-- and tags (source 'taste', entity = their record id). rank is the position
-- in that listing; a re-found scene takes the newer found_at and rank.
CREATE TABLE feed_discovery (
  scene    TEXT NOT NULL,
  source   TEXT NOT NULL,
  entity   TEXT NOT NULL,
  rank     INTEGER NOT NULL,
  found_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  PRIMARY KEY (source, entity, scene)
);
CREATE INDEX feed_discovery_found ON feed_discovery (source, found_at);
CREATE INDEX feed_discovery_scene ON feed_discovery (scene);
