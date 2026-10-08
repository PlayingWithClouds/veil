-- The aylo plugin (pornhub, redtube, youporn, tube8) used to return scene
-- pages without downloads, so visited scenes got no stream and could not be
-- played. Mark those visits undone so the next one fetches the page again
-- and stores its stream.
UPDATE scene
   SET detail_fetched_at = NULL
 WHERE detail_fetched_at IS NOT NULL
   AND (source_url LIKE 'https://www.pornhub.com/%'
     OR source_url LIKE 'https://www.redtube.com/%'
     OR source_url LIKE 'https://www.youporn.com/%'
     OR source_url LIKE 'https://www.tube8.com/%')
   AND NOT EXISTS (SELECT 1 FROM stream WHERE stream.media = scene.id);
