-- Progress updates used to replace a download job's payload and drop its
-- media link, so the library never found the downloaded scene. Relink the
-- jobs through the stream they were started from.
UPDATE job
   SET payload = json_set(payload, '$.media', (
         SELECT stream.media FROM stream WHERE stream.url = json_extract(job.payload, '$.url') LIMIT 1))
 WHERE kind = 'download'
   AND json_extract(payload, '$.media') IS NULL
   AND EXISTS (SELECT 1 FROM stream WHERE stream.url = json_extract(job.payload, '$.url'));
