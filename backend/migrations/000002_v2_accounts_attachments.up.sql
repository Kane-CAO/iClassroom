SET NAMES utf8mb4;

-- V2 account and session foundation -----------------------------------------
CREATE TABLE admin_users (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username      VARCHAR(64)     NOT NULL,
    password_hash VARCHAR(255)    NOT NULL,
    display_name  VARCHAR(64)     NOT NULL,
    status        ENUM('active','disabled') NOT NULL DEFAULT 'active',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_admin_users_username (username),
    KEY idx_admin_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE teacher_accounts (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username            VARCHAR(64)     NOT NULL,
    password_hash       VARCHAR(255)    NOT NULL,
    display_name        VARCHAR(64)     NOT NULL,
    status              ENUM('active','disabled') NOT NULL DEFAULT 'active',
    created_by_admin_id BIGINT UNSIGNED NULL,
    last_login_at       DATETIME        NULL DEFAULT NULL,
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_teacher_accounts_username (username),
    KEY idx_teacher_accounts_status (status),
    KEY idx_teacher_accounts_created_by (created_by_admin_id),
    CONSTRAINT fk_teacher_accounts_created_by
        FOREIGN KEY (created_by_admin_id) REFERENCES admin_users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE auth_sessions (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_type   ENUM('admin','teacher') NOT NULL,
    user_id     BIGINT UNSIGNED NOT NULL,
    token_hash  VARCHAR(255)    NOT NULL,
    expires_at  DATETIME        NOT NULL,
    revoked_at  DATETIME        NULL DEFAULT NULL,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_auth_sessions_token_hash (token_hash),
    KEY idx_auth_sessions_user (user_type, user_id),
    KEY idx_auth_sessions_expires_at (expires_at),
    KEY idx_auth_sessions_revoked_at (revoked_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Room ownership migration. Existing V1 rooms may stay NULL until migrated.
ALTER TABLE rooms
    ADD COLUMN teacher_id BIGINT UNSIGNED NULL AFTER id,
    ADD KEY idx_rooms_teacher_id (teacher_id),
    ADD CONSTRAINT fk_rooms_teacher
        FOREIGN KEY (teacher_id) REFERENCES teacher_accounts(id) ON DELETE SET NULL;

-- Unified V2 attachment model. Keep submission_images for migration
-- compatibility until all callers are moved.
CREATE TABLE submission_attachments (
    id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    submission_id      BIGINT UNSIGNED NOT NULL,
    kind               ENUM('image','file') NOT NULL,
    file_url           VARCHAR(1024)   NOT NULL,
    file_path          VARCHAR(1024)   NOT NULL,
    original_file_name VARCHAR(255)    NOT NULL,
    stored_file_name   VARCHAR(255)    NOT NULL,
    file_size          BIGINT UNSIGNED NOT NULL,
    mime_type          VARCHAR(128)    NOT NULL,
    created_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_submission_attachments_submission_id (submission_id),
    KEY idx_submission_attachments_kind (kind),
    CONSTRAINT fk_submission_attachments_submission
        FOREIGN KEY (submission_id) REFERENCES submissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO submission_attachments (
    submission_id,
    kind,
    file_url,
    file_path,
    original_file_name,
    stored_file_name,
    file_size,
    mime_type,
    created_at
)
SELECT
    submission_id,
    'image',
    file_url,
    file_path,
    file_name,
    file_name,
    file_size,
    mime_type,
    created_at
FROM submission_images;
