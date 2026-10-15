-- The built-in Favorites collection: a user collection every library has,
-- which the favorite buttons add to. The backend refuses to rename or delete it.
INSERT OR IGNORE INTO collection (id, name, user_created) VALUES ('collection:favorites', 'Favorites', 1);
