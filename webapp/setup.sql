ALTER TABLE posts DROP COLUMN imgdata;

ALTER TABLE posts ADD INDEX (user_id);
ALTER TABLE posts ADD INDEX (created_at DESC);

ALTER TABLE users ADD INDEX (created_at DESC);

ALTER TABLE comments ADD INDEX (post_id);
ALTER TABLE comments ADD INDEX (user_id);
ALTER TABLE comments ADD INDEX (created_at DESC);

