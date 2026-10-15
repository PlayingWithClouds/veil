-- The video a gallery's photo set comes from, picked by the user from
-- findGalleryVideos' matches: one scene per gallery.
CREATE TABLE gallery_video (
  gallery    TEXT PRIMARY KEY,
  scene      TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX gallery_video_scene ON gallery_video (scene);
