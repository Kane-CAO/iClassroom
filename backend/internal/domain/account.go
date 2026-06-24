package domain

import "time"

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleTeacher UserRole = "teacher"
)

func (r UserRole) Valid() bool {
	switch r {
	case RoleAdmin, RoleTeacher:
		return true
	default:
		return false
	}
}

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusDisabled AccountStatus = "disabled"
)

func (s AccountStatus) Valid() bool {
	switch s {
	case AccountStatusActive, AccountStatusDisabled:
		return true
	default:
		return false
	}
}

type AdminUser struct {
	ID           int64         `json:"userId"`
	Username     string        `json:"username"`
	PasswordHash string        `json:"-"`
	DisplayName  string        `json:"displayName"`
	Status       AccountStatus `json:"status"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type TeacherAccount struct {
	ID               int64         `json:"teacherId"`
	Username         string        `json:"username"`
	PasswordHash     string        `json:"-"`
	DisplayName      string        `json:"displayName"`
	Status           AccountStatus `json:"status"`
	CreatedByAdminID *int64        `json:"createdByAdminId,omitempty"`
	LastLoginAt      *time.Time    `json:"lastLoginAt,omitempty"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
}

type AuthSession struct {
	ID        int64      `json:"-"`
	UserType  UserRole   `json:"role"`
	UserID    int64      `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"-"`
	CreatedAt time.Time  `json:"createdAt"`
}

type AuthUser struct {
	UserID      int64    `json:"userId"`
	Role        UserRole `json:"role"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName,omitempty"`
}
