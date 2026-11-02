-- Generated migration — 1 operation(s).

ALTER TABLE Video ADD COLUMN previewFrames TEXT CHECK (json_valid(previewFrames));
