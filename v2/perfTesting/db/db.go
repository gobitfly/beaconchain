package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitWithDSN(dsn string) error {
	realDb, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return err
	}

	err = realDb.Ping()
	if err != nil {
		return err
	}

	DB = realDb

	return nil
}

func Init(host, port, user, password, dbname string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	realDb, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}

	err = realDb.Ping()
	if err != nil {
		return err
	}

	DB = realDb

	return nil
}
