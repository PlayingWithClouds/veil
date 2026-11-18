-- Generated migration — 1 operation(s).

ALTER TABLE Source ADD COLUMN uploader TEXT CHECK (json_valid(uploader));
