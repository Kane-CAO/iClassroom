-- Rollback for 000001_init_schema. Drop in reverse dependency order so foreign
-- keys never block a DROP.

DROP TABLE IF EXISTS featured_answers;
DROP TABLE IF EXISTS submission_images;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS task_target_groups;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS `groups`;
DROP TABLE IF EXISTS rooms;
