package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"iclassroom/backend/internal/apperr"
	"iclassroom/backend/internal/domain"
	"iclassroom/backend/internal/repository"
)

const (
	minPasswordLen       = 8
	maxUsernameLen       = 64
	maxDisplayNameLen    = 64
	defaultSessionTTLDays = 7
)

type AccountStore interface {
	CountAdmins(ctx context.Context) (int, error)
	CreateAdmin(ctx context.Context, admin *domain.AdminUser) error
	GetAdminByUsername(ctx context.Context, username string) (*domain.AdminUser, error)
	GetAdminByID(ctx context.Context, id int64) (*domain.AdminUser, error)
	CreateTeacher(ctx context.Context, teacher *domain.TeacherAccount) error
	GetTeacherByUsername(ctx context.Context, username string) (*domain.TeacherAccount, error)
	GetTeacherByID(ctx context.Context, id int64) (*domain.TeacherAccount, error)
	ListTeachers(ctx context.Context) ([]domain.TeacherAccount, error)
	UpdateTeacherStatus(ctx context.Context, teacherID int64, status domain.AccountStatus) (*domain.TeacherAccount, error)
	ResetTeacherPassword(ctx context.Context, teacherID int64, passwordHash string) (*domain.TeacherAccount, error)
	DeleteTeacher(ctx context.Context, teacherID int64) error
	TouchTeacherLogin(ctx context.Context, teacherID int64, t time.Time) error
	CreateSession(ctx context.Context, session *domain.AuthSession) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthSession, error)
	RevokeSession(ctx context.Context, tokenHash string, revokedAt time.Time) error
}

type AuthService struct {
	accounts   AccountStore
	sessionTTL time.Duration
}

func NewAuthService(accounts AccountStore, sessionTTL time.Duration) *AuthService {
	if sessionTTL <= 0 {
		sessionTTL = defaultSessionTTLDays * 24 * time.Hour
	}
	return &AuthService{accounts: accounts, sessionTTL: sessionTTL}
}

type LoginResult struct {
	Token string
	User  domain.AuthUser
}

func (s *AuthService) EnsureInitialAdmin(ctx context.Context, username, password, displayName string) error {
	count, err := s.accounts.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	username = strings.TrimSpace(username)
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = username
	}
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	return s.accounts.CreateAdmin(ctx, &domain.AdminUser{
		Username:     username,
		PasswordHash: hash,
		DisplayName:  displayName,
		Status:       domain.AccountStatusActive,
	})
}

func (s *AuthService) LoginAdmin(ctx context.Context, username, password string) (*LoginResult, error) {
	admin, err := s.accounts.GetAdminByUsername(ctx, strings.TrimSpace(username))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.InvalidCredentials()
	}
	if err != nil {
		return nil, err
	}
	if admin.Status != domain.AccountStatusActive {
		return nil, apperr.Forbidden()
	}
	if !passwordMatches(admin.PasswordHash, password) {
		return nil, apperr.InvalidCredentials()
	}
	return s.createLogin(ctx, domain.RoleAdmin, admin.ID, admin.Username, admin.DisplayName)
}

func (s *AuthService) LoginTeacher(ctx context.Context, username, password string) (*LoginResult, error) {
	teacher, err := s.accounts.GetTeacherByUsername(ctx, strings.TrimSpace(username))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.InvalidCredentials()
	}
	if err != nil {
		return nil, err
	}
	if teacher.Status != domain.AccountStatusActive {
		return nil, apperr.TeacherDisabled()
	}
	if !passwordMatches(teacher.PasswordHash, password) {
		return nil, apperr.InvalidCredentials()
	}
	now := time.Now().UTC()
	if err := s.accounts.TouchTeacherLogin(ctx, teacher.ID, now); err != nil {
		return nil, err
	}
	return s.createLogin(ctx, domain.RoleTeacher, teacher.ID, teacher.Username, teacher.DisplayName)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return apperr.Unauthorized()
	}
	return s.accounts.RevokeSession(ctx, tokenHash(token), time.Now().UTC())
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (*domain.AuthUser, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, apperr.Unauthorized()
	}

	session, err := s.accounts.GetSessionByTokenHash(ctx, tokenHash(token))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.Unauthorized()
	}
	if err != nil {
		return nil, err
	}
	if session.RevokedAt != nil || !time.Now().UTC().Before(session.ExpiresAt.UTC()) {
		return nil, apperr.Unauthorized()
	}

	switch session.UserType {
	case domain.RoleAdmin:
		admin, err := s.accounts.GetAdminByID(ctx, session.UserID)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.Unauthorized()
		}
		if err != nil {
			return nil, err
		}
		if admin.Status != domain.AccountStatusActive {
			return nil, apperr.Forbidden()
		}
		return &domain.AuthUser{UserID: admin.ID, Role: domain.RoleAdmin, Username: admin.Username, DisplayName: admin.DisplayName}, nil
	case domain.RoleTeacher:
		teacher, err := s.accounts.GetTeacherByID(ctx, session.UserID)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperr.Unauthorized()
		}
		if err != nil {
			return nil, err
		}
		if teacher.Status != domain.AccountStatusActive {
			return nil, apperr.TeacherDisabled()
		}
		return &domain.AuthUser{UserID: teacher.ID, Role: domain.RoleTeacher, Username: teacher.Username, DisplayName: teacher.DisplayName}, nil
	default:
		return nil, apperr.Unauthorized()
	}
}

type CreateTeacherInput struct {
	Username        string
	DisplayName     string
	InitialPassword string
	CreatedByAdminID int64
}

func (s *AuthService) CreateTeacher(ctx context.Context, in CreateTeacherInput) (*domain.TeacherAccount, error) {
	username := strings.TrimSpace(in.Username)
	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" {
		displayName = username
	}
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if len([]rune(displayName)) > maxDisplayNameLen {
		return nil, apperr.InvalidRequest("displayName must be at most 64 characters")
	}
	if err := validatePassword(in.InitialPassword); err != nil {
		return nil, err
	}
	hash, err := hashPassword(in.InitialPassword)
	if err != nil {
		return nil, err
	}
	creatorID := in.CreatedByAdminID
	teacher := &domain.TeacherAccount{
		Username:         username,
		PasswordHash:     hash,
		DisplayName:      displayName,
		Status:           domain.AccountStatusActive,
		CreatedByAdminID: &creatorID,
	}
	if err := s.accounts.CreateTeacher(ctx, teacher); errors.Is(err, repository.ErrDuplicate) {
		return nil, apperr.TeacherUsernameDuplicated()
	} else if err != nil {
		return nil, err
	}
	return s.accounts.GetTeacherByID(ctx, teacher.ID)
}

func (s *AuthService) ListTeachers(ctx context.Context) ([]domain.TeacherAccount, error) {
	return s.accounts.ListTeachers(ctx)
}

func (s *AuthService) UpdateTeacherStatus(ctx context.Context, teacherID int64, status domain.AccountStatus) (*domain.TeacherAccount, error) {
	if teacherID <= 0 {
		return nil, apperr.TeacherNotFound()
	}
	if !status.Valid() {
		return nil, apperr.InvalidRequest("status must be active or disabled")
	}
	teacher, err := s.accounts.UpdateTeacherStatus(ctx, teacherID, status)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperr.TeacherNotFound()
	}
	return teacher, err
}

func (s *AuthService) ResetTeacherPassword(ctx context.Context, teacherID int64, newPassword string) (*domain.TeacherAccount, string, error) {
	if teacherID <= 0 {
		return nil, "", apperr.TeacherNotFound()
	}
	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		generated, err := newToken("pwd")
		if err != nil {
			return nil, "", err
		}
		newPassword = generated[:16]
	}
	if err := validatePassword(newPassword); err != nil {
		return nil, "", err
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return nil, "", err
	}
	teacher, err := s.accounts.ResetTeacherPassword(ctx, teacherID, hash)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, "", apperr.TeacherNotFound()
	}
	return teacher, newPassword, err
}

func (s *AuthService) DeleteTeacher(ctx context.Context, teacherID int64) error {
	if teacherID <= 0 {
		return apperr.TeacherNotFound()
	}
	if err := s.accounts.DeleteTeacher(ctx, teacherID); errors.Is(err, repository.ErrNotFound) {
		return apperr.TeacherNotFound()
	} else {
		return err
	}
}

func (s *AuthService) createLogin(ctx context.Context, role domain.UserRole, userID int64, username, displayName string) (*LoginResult, error) {
	token, err := newToken(string(role) + "_session")
	if err != nil {
		return nil, err
	}
	session := &domain.AuthSession{
		UserType:  role,
		UserID:    userID,
		TokenHash: tokenHash(token),
		ExpiresAt: time.Now().UTC().Add(s.sessionTTL),
	}
	if err := s.accounts.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return &LoginResult{
		Token: token,
		User: domain.AuthUser{
			UserID:      userID,
			Role:        role,
			Username:    username,
			DisplayName: displayName,
		},
	}, nil
}

func validateUsername(username string) error {
	if username == "" || len([]rune(username)) > maxUsernameLen {
		return apperr.InvalidRequest("username is required and must be at most 64 characters")
	}
	return nil
}

func validatePassword(password string) error {
	if len([]rune(password)) < minPasswordLen {
		return apperr.InvalidRequest("password must be at least 8 characters")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func passwordMatches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
