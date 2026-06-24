package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"iclassroom/backend/internal/domain"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) CountAdmins(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return 0, fmt.Errorf("repository: count admins: %w", err)
	}
	return count, nil
}

func (r *AccountRepository) CreateAdmin(ctx context.Context, admin *domain.AdminUser) error {
	const q = `INSERT INTO admin_users (username, password_hash, display_name, status)
VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, admin.Username, admin.PasswordHash, admin.DisplayName, admin.Status)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("repository: create admin: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: admin last insert id: %w", err)
	}
	admin.ID = id
	return nil
}

func (r *AccountRepository) GetAdminByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	const q = `SELECT id, username, password_hash, display_name, status, created_at, updated_at
FROM admin_users WHERE username = ?`
	admin, err := scanAdmin(r.db.QueryRowContext(ctx, q, username))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get admin by username: %w", err)
	}
	return admin, nil
}

func (r *AccountRepository) GetAdminByID(ctx context.Context, id int64) (*domain.AdminUser, error) {
	const q = `SELECT id, username, password_hash, display_name, status, created_at, updated_at
FROM admin_users WHERE id = ?`
	admin, err := scanAdmin(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get admin by id: %w", err)
	}
	return admin, nil
}

func (r *AccountRepository) CreateTeacher(ctx context.Context, teacher *domain.TeacherAccount) error {
	const q = `INSERT INTO teacher_accounts
(username, password_hash, display_name, status, created_by_admin_id)
VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(
		ctx,
		q,
		teacher.Username,
		teacher.PasswordHash,
		teacher.DisplayName,
		teacher.Status,
		nullableInt64(teacher.CreatedByAdminID),
	)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("repository: create teacher: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: teacher last insert id: %w", err)
	}
	teacher.ID = id
	return nil
}

func (r *AccountRepository) GetTeacherByUsername(ctx context.Context, username string) (*domain.TeacherAccount, error) {
	const q = `SELECT id, username, password_hash, display_name, status, created_by_admin_id, last_login_at, created_at, updated_at
FROM teacher_accounts WHERE username = ?`
	teacher, err := scanTeacher(r.db.QueryRowContext(ctx, q, username))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get teacher by username: %w", err)
	}
	return teacher, nil
}

func (r *AccountRepository) GetTeacherByID(ctx context.Context, id int64) (*domain.TeacherAccount, error) {
	const q = `SELECT id, username, password_hash, display_name, status, created_by_admin_id, last_login_at, created_at, updated_at
FROM teacher_accounts WHERE id = ?`
	teacher, err := scanTeacher(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get teacher by id: %w", err)
	}
	return teacher, nil
}

func (r *AccountRepository) ListTeachers(ctx context.Context) ([]domain.TeacherAccount, error) {
	const q = `SELECT id, username, password_hash, display_name, status, created_by_admin_id, last_login_at, created_at, updated_at
FROM teacher_accounts ORDER BY created_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repository: list teachers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.TeacherAccount, 0)
	for rows.Next() {
		teacher, err := scanTeacher(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *teacher)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate teachers: %w", err)
	}
	return out, nil
}

func (r *AccountRepository) UpdateTeacherStatus(ctx context.Context, teacherID int64, status domain.AccountStatus) (*domain.TeacherAccount, error) {
	const q = `UPDATE teacher_accounts SET status = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, status, time.Now().UTC(), teacherID)
	if err != nil {
		return nil, fmt.Errorf("repository: update teacher status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("repository: teacher status rows affected: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return r.GetTeacherByID(ctx, teacherID)
}

func (r *AccountRepository) ResetTeacherPassword(ctx context.Context, teacherID int64, passwordHash string) (*domain.TeacherAccount, error) {
	const q = `UPDATE teacher_accounts SET password_hash = ?, updated_at = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, passwordHash, time.Now().UTC(), teacherID)
	if err != nil {
		return nil, fmt.Errorf("repository: reset teacher password: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("repository: reset teacher password rows affected: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return r.GetTeacherByID(ctx, teacherID)
}

func (r *AccountRepository) DeleteTeacher(ctx context.Context, teacherID int64) error {
	const q = `DELETE FROM teacher_accounts WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, teacherID)
	if err != nil {
		return fmt.Errorf("repository: delete teacher: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete teacher rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AccountRepository) TouchTeacherLogin(ctx context.Context, teacherID int64, t time.Time) error {
	const q = `UPDATE teacher_accounts SET last_login_at = ?, updated_at = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, q, t.UTC(), t.UTC(), teacherID); err != nil {
		return fmt.Errorf("repository: touch teacher login: %w", err)
	}
	return nil
}

func (r *AccountRepository) CreateSession(ctx context.Context, session *domain.AuthSession) error {
	const q = `INSERT INTO auth_sessions (user_type, user_id, token_hash, expires_at)
VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, session.UserType, session.UserID, session.TokenHash, session.ExpiresAt.UTC())
	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("repository: create auth session: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: auth session last insert id: %w", err)
	}
	session.ID = id
	return nil
}

func (r *AccountRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error) {
	const q = `SELECT id, user_type, user_id, token_hash, expires_at, revoked_at, created_at
FROM auth_sessions WHERE token_hash = ?`
	var (
		session   domain.AuthSession
		revokedAt sql.NullTime
	)
	err := r.db.QueryRowContext(ctx, q, tokenHash).Scan(
		&session.ID,
		&session.UserType,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&revokedAt,
		&session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get auth session: %w", err)
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		session.RevokedAt = &t
	}
	return &session, nil
}

func (r *AccountRepository) RevokeSession(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	const q = `UPDATE auth_sessions SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, revokedAt.UTC(), tokenHash)
	if err != nil {
		return fmt.Errorf("repository: revoke auth session: %w", err)
	}
	return nil
}

type accountScanner interface {
	Scan(dest ...any) error
}

func scanAdmin(s accountScanner) (*domain.AdminUser, error) {
	var admin domain.AdminUser
	if err := s.Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.DisplayName,
		&admin.Status,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &admin, nil
}

func scanTeacher(s accountScanner) (*domain.TeacherAccount, error) {
	var (
		teacher          domain.TeacherAccount
		createdByAdminID sql.NullInt64
		lastLoginAt      sql.NullTime
	)
	if err := s.Scan(
		&teacher.ID,
		&teacher.Username,
		&teacher.PasswordHash,
		&teacher.DisplayName,
		&teacher.Status,
		&createdByAdminID,
		&lastLoginAt,
		&teacher.CreatedAt,
		&teacher.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if createdByAdminID.Valid {
		v := createdByAdminID.Int64
		teacher.CreatedByAdminID = &v
	}
	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		teacher.LastLoginAt = &t
	}
	return &teacher, nil
}
