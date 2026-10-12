-- Studios had the same exact-name matching as performers: "ADULT TIME" and
-- "Adult Time" were two studios, as were "JulesJordan" and "Jules Jordan", and
-- the same channel seen through different sites' channel pages never met.
-- name_key is the comparison form of the name (names.CompactKey), filled by
-- ingest and by the startup repair for rows that predate it.
ALTER TABLE studio ADD COLUMN name_key TEXT;
CREATE INDEX studio_name_key ON studio (name_key);
