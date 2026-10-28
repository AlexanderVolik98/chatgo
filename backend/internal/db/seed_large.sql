-- Large dev dataset for EXPLAIN / index practice (~500k messages).
-- Idempotent: skipped if already loaded. Remove with:
--   DELETE FROM users WHERE email LIKE '%@bulk.example.com';  (cascades to rooms, members, messages)
--
-- Run with: make seed-large
-- All bulk users share the password: password123

DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM users WHERE email = 'bulk1@bulk.example.com') THEN
		RAISE NOTICE 'bulk seed already loaded, skipping';
		RETURN;
	END IF;

	INSERT INTO users (username, email, password_hash)
	SELECT 'bulk_user_' || g, 'bulk' || g || '@bulk.example.com',
		'$2a$10$u5TMed38isiHokAODP7c5OBk32C8S1YNcfTX6WB/s0r4f19KdupBy'
	FROM generate_series(1, 1000) g;

	-- 50 public rooms, owned by the first 50 bulk users.
	INSERT INTO rooms (name, description, owner_id, is_private)
	SELECT 'bulk_room_' || g, 'bulk data room', u.id, false
	FROM generate_series(1, 50) g
	JOIN users u ON u.email = 'bulk' || g || '@bulk.example.com';

	INSERT INTO room_members (room_id, user_id, role)
	SELECT r.id, r.owner_id, 'owner'::room_role
	FROM rooms r
	WHERE r.name LIKE 'bulk_room_%';

	-- ~10% of bulk users (about 100) join each room as plain members.
	INSERT INTO room_members (room_id, user_id, role)
	SELECT r.id, u.id, 'member'::room_role
	FROM rooms r
	JOIN users u ON u.email LIKE '%@bulk.example.com' AND u.id <> r.owner_id
	WHERE r.name LIKE 'bulk_room_%' AND random() < 0.1
	ON CONFLICT (room_id, user_id) DO NOTHING;

	-- 500k messages spread over random rooms/users. created_at grows with the
	-- row number, so id order and time order agree like in a real chat.
	WITH room_ids AS (
		SELECT array_agg(id ORDER BY id) AS ids FROM rooms WHERE name LIKE 'bulk_room_%'
	), user_ids AS (
		SELECT array_agg(id ORDER BY id) AS ids FROM users WHERE email LIKE '%@bulk.example.com'
	)
	INSERT INTO messages (room_id, user_id, content, created_at)
	SELECT ra.ids[1 + floor(random() * array_length(ra.ids, 1))::int],
		ua.ids[1 + floor(random() * array_length(ua.ids, 1))::int],
		'bulk message ' || g || ' ' || md5(g::text),
		now() - interval '90 days' + g * (interval '90 days' / 500000)
	FROM generate_series(1, 500000) g, room_ids ra, user_ids ua;
END
$$;

-- Fresh planner statistics, otherwise EXPLAIN works from stale estimates.
ANALYZE users;
ANALYZE rooms;
ANALYZE room_members;
ANALYZE messages;
