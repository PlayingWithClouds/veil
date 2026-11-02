-- Generated migration — 2 operation(s).

ALTER TABLE StoryChapter ADD COLUMN parent TEXT REFERENCES StoryChapter(id);

ALTER TABLE StoryChapter ADD COLUMN choice TEXT;
