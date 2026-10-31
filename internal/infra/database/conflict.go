package database

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/uptrace/bun/driver/pgdriver"
	domain "go-ddd-book-manager/internal/domain/common"
)

// ValidateNoConflict returns the domain conflict error when no rows were affected.
func ValidateNoConflict(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}

func ConflictOnDuplicate(err error) error {
	var pg pgdriver.Error
	if errors.As(err, &pg) && pg.Field('C') == pgerrcode.UniqueViolation {
		return domain.ErrConflict
	}
	return err
}
