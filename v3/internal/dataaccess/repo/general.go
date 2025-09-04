package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
)

// Generic function to execute a query
// Ensures T is a slice in compile-time
// Retrieves multiple rows and stores them in a slice
// Used when the query returns multiple results (e.g., SELECT * FROM users)
// The destination must be a slice ([]T)
type goquDataset[T any] interface {
	Prepared(bool) T
	ToSQL() (string, []interface{}, error)
}

func RunQuery[T any, dataSet goquDataset[dataSet]](ctx context.Context, db *sqlx.DB, ds dataSet) (T, error) {
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		var zero T
		return zero, fmt.Errorf("error preparing query: %w", err)
	}

	var result T
	err = db.GetContext(ctx, &result, query, args...)
	if err != nil {
		return result, fmt.Errorf("error executing query: %w", mapError(err))
	}
	return result, nil
}

func RunQueryRows[T ~[]E, E any, dataSet goquDataset[dataSet]](ctx context.Context, db *sqlx.DB, ds dataSet) (T, error) {
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		var zero T
		return zero, fmt.Errorf("error preparing query: %w", err)
	}

	var result T
	err = db.SelectContext(ctx, &result, query, args...)
	if err != nil {
		return result, fmt.Errorf("error executing query: %w", mapError(err))
	}
	return result, nil
}

func ExecAndCheckRows[dataSet goquDataset[dataSet]](ctx context.Context, db *sqlx.DB, ds dataSet) error {
	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", mapError(err))
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", mapError(err))
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var errDuplicateEntry = &pgconn.PgError{Code: "23505"}

func pgxIsDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), errDuplicateEntry.Code) // there's prob a better way
}

// MapError maps db-specific errors to domain-specific errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case pgxIsDuplicateKeyError(err):
		return errors.Join(domain.ErrDuplicate, err)
	case errors.Is(err, sql.ErrNoRows):
		return errors.Join(domain.ErrNotFound, err)
	default:
		return err
	}
}
