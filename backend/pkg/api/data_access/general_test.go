package dataaccess

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func SetupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	return sqlx.NewDb(mockDB, "sqlmock"), mock
}

func TestSelectContext(t *testing.T) {
	sqlxDB, mock := SetupTestDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	ds := goqu.Dialect("postgres").From(goqu.I("users")).Select(goqu.C("id"), goqu.C("name"))

	mockRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Rami").
		AddRow(2, "Lucca").
		AddRow(3, "Sasha").
		AddRow(4, "Invis")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id", "name" FROM "users"`)).WillReturnRows(mockRows)

	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var expectedUsers = []User{
		{ID: 1, Name: "Rami"},
		{ID: 2, Name: "Lucca"},
		{ID: 3, Name: "Sasha"},
		{ID: 4, Name: "Invis"},
	}

	result, err := runQueryRows[[]User](ctx, sqlxDB, ds)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, result)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Invalid expression
	// T is a slice, this will fail to compile
	//_, err = runQueryRows[User](ctx, sqlxDB, ds)
}

func TestGetContext(t *testing.T) {
	sqlxDB, mock := SetupTestDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	ds := goqu.Dialect("postgres").From(goqu.I("users")).Select(goqu.C("id"), goqu.C("name")).Where(goqu.Ex{"id": 1})

	mockRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Axel")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id", "name" FROM "users" WHERE ("id" = $1)`)).WithArgs(1).WillReturnRows(mockRows)

	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var expectedUser = User{ID: 1, Name: "Axel"}

	result, err := runQuery[User](ctx, sqlxDB, ds)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
