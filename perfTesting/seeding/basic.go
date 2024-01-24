package seeding

import (
	"fmt"
	"perftesting/db"
)

type Seeder struct {
	TableName      string
	ValidatorsInDB int
	EpochsInDB     int
	BatchSize      int
	schemer        SeederScheme
	filler         SeederFiller
	ColumnEngine   bool
}

type SeederScheme interface {
	CreateSchema(s *Seeder) error
}

type SeederFiller interface {
	FillTable(s *Seeder) error
}

func GetSeeder(tableName string, columnarEngine bool, schemer SeederScheme, filler SeederFiller) *Seeder {
	temp := &Seeder{}
	temp.TableName = tableName
	temp.BatchSize = 100000
	temp.ColumnEngine = columnarEngine
	temp.schemer = schemer
	temp.filler = filler
	return temp
}

func (s *Seeder) CreateSchema() error {
	return s.schemer.CreateSchema(s)
}

func (s *Seeder) FillTable() error {
	return s.filler.FillTable(s)
}

func (s *Seeder) AddToColumnEngine(table, columns string) error {
	_, err := db.DB.Exec(fmt.Sprintf(`
		SELECT google_columnar_engine_add(
			relation => '%s',
			columns => '%s'
		);
		`, table))
	return err
}
