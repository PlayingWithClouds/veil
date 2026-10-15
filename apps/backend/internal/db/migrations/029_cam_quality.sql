-- Recording quality of live cam rooms. record_quality is what auto-record asks
-- for: "source" (the best rendition) or a target height such as "480". A
-- recording's quality is the rendition it captured, labelled like "480p".
ALTER TABLE cam_model ADD COLUMN record_quality TEXT NOT NULL DEFAULT 'source';
ALTER TABLE cam_recording ADD COLUMN quality TEXT;
