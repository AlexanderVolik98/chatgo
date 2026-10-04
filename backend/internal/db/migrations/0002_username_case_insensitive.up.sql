-- users.username UNIQUE is case-sensitive; this makes "Alex" and "alex" collide.
-- Violations still raise 23505, so existing error handling keeps working.
CREATE UNIQUE INDEX users_username_lower_idx ON users (lower(username));
