-- Tags are matched by spelling-insensitive name like studios: "blow job",
-- "blowjob" and "Blow-Job" are one tag. name_key holds the comparison form (see
-- internal/names), filled by ingest and by the startup repair for older rows.
ALTER TABLE tag ADD COLUMN name_key TEXT;
CREATE INDEX tag_name_key ON tag (name_key);
