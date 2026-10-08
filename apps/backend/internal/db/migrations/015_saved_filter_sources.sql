-- Saved filter presets can also narrow by site: the plugin names to keep.
ALTER TABLE saved_filter ADD COLUMN sources JSON NOT NULL DEFAULT '[]';
