SET NAMES utf8mb4;

DROP TABLE IF EXISTS submission_attachments;

ALTER TABLE rooms
    DROP FOREIGN KEY fk_rooms_teacher,
    DROP KEY idx_rooms_teacher_id,
    DROP COLUMN teacher_id;

DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS teacher_accounts;
DROP TABLE IF EXISTS admin_users;
