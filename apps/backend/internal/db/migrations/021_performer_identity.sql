-- Performers used to be matched by exact name, so "Isla Summer" and
-- "ISLA SUMMER" became two records, and a performer page (identified by its
-- source_url) never met the record made from a scene credit. name_key holds the
-- comparison form of the name (see internal/names), filled by ingest and by the
-- startup repair for rows that predate it.
ALTER TABLE performer ADD COLUMN name_key TEXT;
CREATE INDEX performer_name_key ON performer (name_key);

-- Extra identities of a record: a performer merged from several sites' pages
-- keeps one source_url and external_id in its own columns, so the others are
-- listed here and ingest resolves them to the same record.
CREATE TABLE entity_identity (
  kind   TEXT NOT NULL CHECK (kind IN ('source_url', 'external_id')),
  value  TEXT NOT NULL,
  record TEXT NOT NULL,
  PRIMARY KEY (kind, value)
);
CREATE INDEX entity_identity_record ON entity_identity (record);
