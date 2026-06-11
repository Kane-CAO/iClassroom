-- iClassroom MVP — initial schema (Backend Step 1).
--
-- Requirements:
--   * MySQL 8.0.16+  (CHECK constraints are enforced from this version).
--   * Store all time values in UTC. Run the server with time_zone = '+00:00'
--     (or set --default-time-zone='+00:00') so DEFAULT CURRENT_TIMESTAMP
--     columns are written in UTC, consistent with the API contract.
--
-- All id / *_id columns are BIGINT UNSIGNED so foreign keys match exactly.

SET NAMES utf8mb4;

-- rooms -----------------------------------------------------------------------
CREATE TABLE rooms (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    room_code          VARCHAR(16)     NOT NULL,
    title              VARCHAR(255)    NOT NULL,
    status             ENUM('created','active','ended') NOT NULL DEFAULT 'created',
    group_count        INT             NOT NULL DEFAULT 6,
    group_capacity     INT             NOT NULL DEFAULT 10,
    allow_choose_group TINYINT(1)      NOT NULL DEFAULT 1,
    teacher_token      VARCHAR(64)     NOT NULL,
    created_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    ended_at           DATETIME        NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_rooms_room_code (room_code),
    UNIQUE KEY uk_rooms_teacher_token (teacher_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- groups ("groups" is a reserved word, hence backticks) ------------------------
CREATE TABLE `groups` (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    room_id     BIGINT UNSIGNED NOT NULL,
    group_name  VARCHAR(64)     NOT NULL,
    capacity    INT             NOT NULL DEFAULT 10,
    score_total INT             NOT NULL DEFAULT 0,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_groups_room_id (room_id),
    CONSTRAINT fk_groups_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- students --------------------------------------------------------------------
CREATE TABLE students (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    room_id      BIGINT UNSIGNED NOT NULL,
    group_id     BIGINT UNSIGNED NOT NULL,
    nickname     VARCHAR(64)     NOT NULL,
    client_token VARCHAR(64)     NOT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_students_room_nickname (room_id, nickname),
    UNIQUE KEY uk_students_client_token (client_token),
    KEY idx_students_room_id (room_id),
    KEY idx_students_group_id (group_id),
    CONSTRAINT fk_students_room  FOREIGN KEY (room_id)  REFERENCES rooms(id)    ON DELETE CASCADE,
    CONSTRAINT fk_students_group FOREIGN KEY (group_id) REFERENCES `groups`(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- tasks -----------------------------------------------------------------------
CREATE TABLE tasks (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    room_id        BIGINT UNSIGNED NOT NULL,
    title          VARCHAR(255)    NOT NULL,
    description    TEXT            NULL,
    attachment_url VARCHAR(1024)   NULL,
    deadline_at    DATETIME        NOT NULL,
    target_type    ENUM('all','groups')                NOT NULL DEFAULT 'all',
    status         ENUM('published','paused','closed') NOT NULL DEFAULT 'published',
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_tasks_room_id (room_id),
    KEY idx_tasks_room_status (room_id, status),
    CONSTRAINT fk_tasks_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- task_target_groups (only used when tasks.target_type = 'groups') -------------
CREATE TABLE task_target_groups (
    id       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id  BIGINT UNSIGNED NOT NULL,
    group_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_ttg_task_group (task_id, group_id),
    KEY idx_ttg_task_id (task_id),
    KEY idx_ttg_group_id (group_id),
    CONSTRAINT fk_ttg_task  FOREIGN KEY (task_id)  REFERENCES tasks(id)     ON DELETE CASCADE,
    CONSTRAINT fk_ttg_group FOREIGN KEY (group_id) REFERENCES `groups`(id)  ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- submissions -----------------------------------------------------------------
CREATE TABLE submissions (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id      BIGINT UNSIGNED NOT NULL,
    student_id   BIGINT UNSIGNED NOT NULL,
    room_id      BIGINT UNSIGNED NOT NULL,
    group_id     BIGINT UNSIGNED NOT NULL,
    content_text TEXT            NULL,
    status       ENUM('submitted','graded') NOT NULL DEFAULT 'submitted',
    score        INT             NULL DEFAULT NULL,
    `comment`    VARCHAR(1024)   NULL,
    submitted_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    graded_at    DATETIME        NULL DEFAULT NULL,
    created_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_submissions_task_student (task_id, student_id),
    KEY idx_submissions_task_id (task_id),
    KEY idx_submissions_student_id (student_id),
    KEY idx_submissions_room_id (room_id),
    KEY idx_submissions_group_id (group_id),
    CONSTRAINT fk_submissions_task    FOREIGN KEY (task_id)    REFERENCES tasks(id)     ON DELETE CASCADE,
    CONSTRAINT fk_submissions_student FOREIGN KEY (student_id) REFERENCES students(id)  ON DELETE CASCADE,
    CONSTRAINT fk_submissions_room    FOREIGN KEY (room_id)    REFERENCES rooms(id)     ON DELETE CASCADE,
    CONSTRAINT fk_submissions_group   FOREIGN KEY (group_id)   REFERENCES `groups`(id)  ON DELETE CASCADE,
    CONSTRAINT chk_submissions_score  CHECK (score IS NULL OR (score BETWEEN 1 AND 10))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- submission_images -----------------------------------------------------------
CREATE TABLE submission_images (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    submission_id BIGINT UNSIGNED NOT NULL,
    file_url      VARCHAR(1024)   NOT NULL,
    file_path     VARCHAR(1024)   NOT NULL,
    file_name     VARCHAR(255)    NOT NULL,
    file_size     BIGINT UNSIGNED NOT NULL,
    mime_type     VARCHAR(64)     NOT NULL,
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_submission_images_submission_id (submission_id),
    CONSTRAINT fk_submission_images_submission FOREIGN KEY (submission_id) REFERENCES submissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- featured_answers ------------------------------------------------------------
CREATE TABLE featured_answers (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    room_id       BIGINT UNSIGNED NOT NULL,
    submission_id BIGINT UNSIGNED NOT NULL,
    display_mode  ENUM('anonymous','showGroup') NOT NULL DEFAULT 'anonymous',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_featured_submission (submission_id),
    KEY idx_featured_room_id (room_id),
    CONSTRAINT fk_featured_room       FOREIGN KEY (room_id)       REFERENCES rooms(id)       ON DELETE CASCADE,
    CONSTRAINT fk_featured_submission FOREIGN KEY (submission_id) REFERENCES submissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
