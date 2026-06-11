// Package repository owns all database reads and writes. It returns plain Go
// values and a small set of sentinel errors; it never speaks HTTP and never
// makes business decisions. The service layer maps these sentinels to
// business errors (apperr).
package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// mysqlErrDupEntry is MySQL's error number for a duplicate key violation.
const mysqlErrDupEntry = 1062

var (
	// ErrNotFound is returned when a queried row does not exist.
	ErrNotFound = errors.New("repository: not found")

	// ErrDuplicate is returned when an insert violates a UNIQUE constraint
	// (e.g. room_code, teacher_token, client_token, or room_id+nickname). The
	// service layer decides which business error this maps to based on context.
	ErrDuplicate = errors.New("repository: duplicate key")

	// ErrGroupFull is returned by StudentRepository.Join when the target group
	// is already at capacity.
	ErrGroupFull = errors.New("repository: group full")
)

// isDuplicateKey reports whether err is a MySQL duplicate-key (1062) error.
func isDuplicateKey(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == mysqlErrDupEntry
	}
	return false
}
