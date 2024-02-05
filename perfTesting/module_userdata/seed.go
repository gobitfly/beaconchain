package module_userdata

import (
	"perftesting/db"
	"perftesting/seeding"
)

type Schemav1 struct{}

func Get(tableName string, columnarEngine bool, data SeederData) *seeding.Seeder {
	return getSeeder(tableName, columnarEngine, &Schemav1{}, data)
}

type Network int

const NetworkMainnet Network = 0
const NetworkTestnet Network = 1

func CreateValDashboard(user_id int64, network Network) error {
	_, err := db.DB.Exec(`
		INSERT INTO users_val_dashboards (user_id, network) VALUES ($1, $2)
	`, user_id, network)
	return err
}

func CreateValDashboardGroup(id, dashboard_id int64, name string) error {
	_, err := db.DB.Exec(`
		INSERT INTO users_val_dashboards_groups (id, dashboard_id, name) VALUES ($1, $2, $3)
	`, id, dashboard_id, name)
	return err
}

func CreateValDashboardValidator(dashboard_id, group_id, validator_index int64) error {
	_, err := db.DB.Exec(`
		INSERT INTO users_val_dashboards_validators (dashboard_id, group_id, validator_index, validator_index_version) VALUES ($1, $2, $3, 1)
	`, dashboard_id, group_id, validator_index)
	return err
}

func CreateValDashboardSharing(dashboard_id int64, shared_groups bool) error {
	_, err := db.DB.Exec(`	
		INSERT INTO users_val_dashboards_sharing (dashboard_id, shared_groups) VALUES ($1, $2)
	`, dashboard_id, shared_groups)
	return err
}

func CreateAccDashboard(user_id int64) error {
	_, err := db.DB.Exec(`
		INSERT INTO users_acc_dashboards (user_id) VALUES ($1)
	`, user_id)
	return err
}

func CreateAccDashboardGroup(id, dashboard_id int64, name string) error {
	_, err := db.DB.Exec(`
		INSERT INTO users_acc_dashboards_groups (id, dashboard_id, name) VALUES ($1, $2, $3)
	`, id, dashboard_id, name)
	return err
}

func CreateAccDashboardAccount(dashboard_id, group_id int64, address []byte) error {
	_, err := db.DB.Exec(`
		INSERT INTO users_acc_dashboards_accounts (dashboard_id, group_id, address) VALUES ($1, $2, $3)
	`, dashboard_id, group_id, address)
	return err
}

func CreateAccDashboardSharing(dashboard_id int64, shared_groups, tx_notes_shared bool, userData string) error {
	_, err := db.DB.Exec(`
		INSERT INTO users_acc_dashboards_sharing (dashboard_id, shared_groups, tx_notes_shared, user_settings) VALUES ($1, $2, $3, $4)
	`, dashboard_id, shared_groups, tx_notes_shared, userData)
	return err
}

func (*Schemav1) CreateSchema(s *seeding.Seeder) error {
	_, err := db.DB.Exec(`
		-- Validator Dashboard

		DROP TABLE IF EXISTS users_val_dashboards;
		CREATE TABLE IF NOT EXISTS users_val_dashboards (
			id 			BIGSERIAL 		NOT NULL,
			user_id 	BIGINT 			NOT NULL,
			network 	SMALLINT 		NOT NULL, -- indicate gnosis/eth mainnet and potentially testnets
			created_at  TIMESTAMP 		DEFAULT(NOW()),
			primary key (id)
		);

		DROP TABLE IF EXISTS users_val_dashboards_groups;
		CREATE TABLE IF NOT EXISTS users_val_dashboards_groups (
			id 				SMALLINT 		DEFAULT(0),
			dashboard_id 	BIGINT 			NOT NULL,
			name 			VARCHAR(50) 	NOT NULL,
			primary key (dashboard_id, id)
		);

		DROP TABLE IF EXISTS users_val_dashboards_validators;
		CREATE TABLE IF NOT EXISTS users_val_dashboards_validators ( -- a validator must not be in multiple groups
			dashboard_id 				BIGINT 		NOT NULL,
			group_id 					SMALLINT 	NOT NULL,
			validator_index     		BIGINT      NOT NULL,
    		validator_index_version     SMALLINT    NOT NULL, -- version will be 0 for pending validators, 1 for normal one and 1-x for reused one (where x is the times the index was reused)
			primary key (dashboard_id, validator_index, validator_index_version)
		);

		DROP TABLE IF EXISTS users_val_dashboards_sharing;
		CREATE TABLE IF NOT EXISTS users_val_dashboards_sharing (
			dashboard_id 		BIGINT 		NOT NULL,
			public_id	 		CHAR(38) 	DEFAULT ('v-' || gen_random_uuid()::text) UNIQUE, -- prefix with "v" for validator dashboards. Public ID to dashboard
			shared_groups 		bool	 	NOT NULL, -- all groups or default 0
			primary key (public_id)
		);

		DROP TABLE IF EXISTS validators;
		CREATE TABLE IF NOT EXISTS validators ( -- minimal only, columns missing
			validator_index SERIAL NOT NULL, -- only serial if version is 0 (pending validator), otherwise will be the beaconchain validator index
			validator_index_version SMALLINT NOT NULL,
			pubkey bytea NOT NULL,
			PRIMARY KEY (validator_index, validator_index_version)
		);

		-- Account Dashboard

		DROP TABLE IF EXISTS users_acc_dashboards;
		CREATE TABLE IF NOT EXISTS users_acc_dashboards (
			id 				BIGSERIAL 	NOT NULL,
			user_id 		BIGINT 		NOT NULL,
			user_settings 	JSONB		DEFAULT '{}'::jsonb, -- or do we want to use a separate kv table for this?
			created_at 		TIMESTAMP 	DEFAULT(NOW()),
			primary key (id)
		);

		DROP TABLE IF EXISTS users_acc_dashboards_groups;
		CREATE TABLE IF NOT EXISTS users_acc_dashboards_groups (
			id 				INT 		NOT NULL,
			dashboard_id 	BIGINT 		NOT NULL,
			name 			VARCHAR(50) NOT NULL,
			primary key (dashboard_id, id)
		);

		DROP TABLE IF EXISTS users_acc_dashboards_accounts;
		CREATE TABLE IF NOT EXISTS users_acc_dashboards_accounts ( -- an account must not be in multiple groups
			dashboard_id 		BIGINT 		NOT NULL,
			group_id 			SMALLINT 	NOT NULL,
			address 			BYTEA 		NOT NULL,
			primary key (dashboard_id, address)
		);

		DROP TABLE IF EXISTS users_acc_dashboards_sharing;
		CREATE TABLE IF NOT EXISTS users_acc_dashboards_sharing (
			dashboard_id 		BIGINT 		NOT NULL,
			public_id 			CHAR(38) 	DEFAULT('a-' || gen_random_uuid()::text) UNIQUE, -- prefix with "a" for validator dashboards
			user_settings 		JSONB		DEFAULT '{}'::jsonb, -- snapshots users_dashboards.user_settings at the time of creating the share
			shared_groups 		bool	 	NOT NULL, -- all groups or default 0
			tx_notes_shared 	BOOLEAN 	NOT NULL, -- not snapshoted
			primary key (public_id)
		);

		-- todo notes

		-- Notification Dashboard (wip)

		DROP TABLE IF EXISTS users_not_dashboards;
		CREATE TABLE IF NOT EXISTS users_not_dashboards (
			id BIGINT,
			user_id BIGINT,
			created_at timestamp,
			primary key (id)
		);

		
	`)
	if err != nil {
		return err
	}

	return nil
}
