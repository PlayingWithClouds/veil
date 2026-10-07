-- Subscriptions follow a studio (channel), performer or tag as well as a saved
-- search. kind is 'search' | 'studio' | 'performer' | 'tag'; entity rows keep
-- the target's name in query so they still read sensibly if the target goes.
ALTER TABLE search_subscription ADD COLUMN kind TEXT NOT NULL DEFAULT 'search'
  CHECK (kind IN ('search', 'studio', 'performer', 'tag'));
-- followed record id ("studio:…", "performer:…", "tag:…"); NULL for searches
ALTER TABLE search_subscription ADD COLUMN entity TEXT;

-- One subscription per search query and one per followed record.
DROP INDEX search_subscription_query;
CREATE UNIQUE INDEX search_subscription_query ON search_subscription (normalized_query) WHERE kind = 'search';
CREATE UNIQUE INDEX search_subscription_entity ON search_subscription (entity) WHERE entity IS NOT NULL;
