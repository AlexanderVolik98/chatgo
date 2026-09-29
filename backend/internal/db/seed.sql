-- Dev/test seed data. Safe to run multiple times (idempotent) against the
-- same database — existing rows are left untouched, nothing is duplicated.
--
-- Run with:
--   docker compose exec -T postgres psql -U chatgo -d chatgo -f - < backend/internal/db/seed.sql
-- or:
--   make seed
--
-- All seeded users share the password: password123

INSERT INTO users (username, email, password_hash) VALUES
	('alice', 'alice@example.com', '$2a$10$u5TMed38isiHokAODP7c5OBk32C8S1YNcfTX6WB/s0r4f19KdupBy'),
	('bob',   'bob@example.com',   '$2a$10$u5TMed38isiHokAODP7c5OBk32C8S1YNcfTX6WB/s0r4f19KdupBy'),
	('carol', 'carol@example.com', '$2a$10$u5TMed38isiHokAODP7c5OBk32C8S1YNcfTX6WB/s0r4f19KdupBy'),
	('dave',  'dave@example.com',  '$2a$10$u5TMed38isiHokAODP7c5OBk32C8S1YNcfTX6WB/s0r4f19KdupBy'),
	('erin',  'erin@example.com',  '$2a$10$u5TMed38isiHokAODP7c5OBk32C8S1YNcfTX6WB/s0r4f19KdupBy'),
	('frank', 'frank@example.com', '$2a$10$u5TMed38isiHokAODP7c5OBk32C8S1YNcfTX6WB/s0r4f19KdupBy')
ON CONFLICT (email) DO NOTHING;

-- Room 1: General — public, owner alice, 2 admins, 3 plain members.
INSERT INTO rooms (name, description, owner_id, is_private)
SELECT 'General', 'Общий чат для всех', id, false FROM users WHERE email = 'alice@example.com'
ON CONFLICT DO NOTHING;

INSERT INTO room_members (room_id, user_id, role)
SELECT r.id, u.id, v.role::room_role
FROM (VALUES
	('alice@example.com', 'owner'),
	('bob@example.com',   'admin'),
	('carol@example.com', 'admin'),
	('dave@example.com',  'member'),
	('erin@example.com',  'member'),
	('frank@example.com', 'member')
) AS v(email, role)
JOIN users u ON u.email = v.email
JOIN rooms r ON r.name = 'General'
ON CONFLICT (room_id, user_id) DO NOTHING;

-- Room 2: Random — public, owner bob, 1 admin, 2 members.
INSERT INTO rooms (name, description, owner_id, is_private)
SELECT 'Random', 'Оффтоп и всякое', id, false FROM users WHERE email = 'bob@example.com'
ON CONFLICT DO NOTHING;

INSERT INTO room_members (room_id, user_id, role)
SELECT r.id, u.id, v.role::room_role
FROM (VALUES
	('bob@example.com',   'owner'),
	('alice@example.com', 'admin'),
	('carol@example.com', 'member'),
	('dave@example.com',  'member')
) AS v(email, role)
JOIN users u ON u.email = v.email
JOIN rooms r ON r.name = 'Random'
ON CONFLICT (room_id, user_id) DO NOTHING;

-- Room 3: Core Team — private, owner carol, 1 admin, 1 member. frank is
-- deliberately NOT a member — useful for testing the 403 private-room path.
INSERT INTO rooms (name, description, owner_id, is_private)
SELECT 'Core Team', 'Приватная комната', id, true FROM users WHERE email = 'carol@example.com'
ON CONFLICT DO NOTHING;

INSERT INTO room_members (room_id, user_id, role)
SELECT r.id, u.id, v.role::room_role
FROM (VALUES
	('carol@example.com', 'owner'),
	('dave@example.com',  'admin'),
	('erin@example.com',  'member')
) AS v(email, role)
JOIN users u ON u.email = v.email
JOIN rooms r ON r.name = 'Core Team'
ON CONFLICT (room_id, user_id) DO NOTHING;

-- A handful of messages in General, only if that room has none yet.
INSERT INTO messages (room_id, user_id, content)
SELECT r.id, u.id, m.content
FROM (VALUES
	('alice@example.com', 'Всем привет!'),
	('bob@example.com',   'Привет, Алиса'),
	('carol@example.com', 'Как дела у всех?'),
	('dave@example.com',  'Норм, работаю над проектом'),
	('erin@example.com',  'Го обсудим завтрашний созвон'),
	('frank@example.com', 'Плюс один')
) AS m(email, content)
JOIN users u ON u.email = m.email
JOIN rooms r ON r.name = 'General'
WHERE NOT EXISTS (SELECT 1 FROM messages WHERE room_id = r.id);
