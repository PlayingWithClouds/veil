-- Generated migration — 1 operation(s).
-- Backfill completed by hand: existing categories match no terms.

-- This migration rebuilds tables. Foreign keys must be off while it runs; PRAGMA foreign_keys
-- has no effect inside a transaction, so a runner that wraps migrations must turn them off first.
PRAGMA foreign_keys = OFF;

-- Rebuild "Category": adds required "matchTerms".
-- Indexes and triggers on "Category" that the schema does not declare must be recreated by hand.
CREATE TABLE Category__new (
  id TEXT NOT NULL PRIMARY KEY,
  createdAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  updatedAt TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  deletedAt TEXT,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  parent TEXT REFERENCES Category(id),
  matchTags TEXT NOT NULL CHECK (json_valid(matchTags)),
  cover TEXT CHECK (json_valid(cover)),
  position INTEGER NOT NULL,
  matchTerms TEXT NOT NULL CHECK (json_valid(matchTerms))
);
INSERT INTO Category__new (id, createdAt, updatedAt, deletedAt, name, slug, parent, matchTags, cover, position, matchTerms)
  SELECT id, createdAt, updatedAt, deletedAt, name, slug, parent, matchTags, cover, position, '[]' FROM Category;
DROP TABLE Category;
ALTER TABLE Category__new RENAME TO Category;

PRAGMA foreign_key_check;
PRAGMA foreign_keys = ON;
