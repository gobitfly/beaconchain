package module_userdata

import (
	crypto "crypto/rand"
	"fmt"
	"math/rand"
	"perftesting/db"
	"perftesting/seeding"
	"time"

	"github.com/lib/pq"
)

type SeederData struct {
	UsersInDB      int
	ValidatorsInDB int
}

func getSeeder(tableName string, columnarEngine bool, scheme seeding.SeederScheme, filler seeding.SeederFiller) *seeding.Seeder {
	return seeding.GetSeeder(tableName, columnarEngine, scheme, filler)
}

func (data SeederData) FillTable(s *seeding.Seeder) error {
	iterations := data.UsersInDB // users

	timeStart := time.Now()
	defer func() {
		fmt.Printf("Time taken: %v\n", time.Since(timeStart))
	}()

	// Add 20 pending
	err := insertValidatorsTable(int64(data.ValidatorsInDB), 20, int64(data.ValidatorsInDB+20), int64(data.ValidatorsInDB))
	if err != nil {
		return err
	}

	normalValidatorsInDB := data.ValidatorsInDB

	//data.ValidatorsInDB += 20

	for i := 0; i < iterations; i++ {
		dashboardID := int64(i)
		err := CreateValDashboard(dashboardID, NetworkMainnet, "")
		if err != nil {
			return err
		}

		groupLimit := 1
		medGroup := rand.Intn(4) == 0
		highGroup := rand.Intn(80) == 0
		if medGroup {
			groupLimit = 3
		} else if highGroup {
			groupLimit = 10
		}

		groupCount := rand.Intn(groupLimit) + 1
		validatorIndex := rand.Intn(data.ValidatorsInDB)
		for j := 0; j < groupCount; j++ {
			err = CreateValDashboardGroup(int64(j), dashboardID, fmt.Sprintf("Group %v", j))
			if err != nil {
				return err
			}

			limit := 1000
			highDashboard := rand.Intn(500) == 0
			medDashboard := rand.Intn(100) == 0
			smallDashboard := rand.Intn(3) == 0
			if highDashboard {
				limit = 200000
			} else if medDashboard {
				limit = 10000
			} else if smallDashboard {
				limit = 10
			}

			validatorCount := rand.Intn(limit) + 1

			err = insertValidators(dashboardID, int64(j), int64(validatorIndex), int64(validatorCount), int64(data.ValidatorsInDB), int64(normalValidatorsInDB))
			if err != nil {
				return err
			}
			validatorIndex = (validatorIndex + validatorCount) % data.ValidatorsInDB

			addPending := rand.Intn(60) == 0
			if addPending {
				numberOfPending := int64(rand.Intn(5) + 1)
				err = insertValidators(dashboardID, int64(j), int64(data.ValidatorsInDB)+numberOfPending, numberOfPending, int64(data.ValidatorsInDB+20), int64(normalValidatorsInDB))
				if err != nil {
					return err
				}

			}

		}

		share := rand.Intn(4) == 0
		shareAll := rand.Intn(3) == 0

		if share {
			err = CreateValDashboardSharing(int64(i), "", shareAll)
			if err != nil {
				return err
			}
		}

		// -- Acc --

		err = CreateAccDashboard(dashboardID, "")
		if err != nil {
			return err
		}

		for j := 0; j < groupCount; j++ {
			err = CreateAccDashboardGroup(int64(j), dashboardID, fmt.Sprintf("Group %v", j))
			if err != nil {
				return err
			}

			limit := 1
			highDashboard := rand.Intn(200) == 0
			medDashboard := rand.Intn(50) == 0
			if highDashboard {
				limit = 10
			} else if medDashboard {
				limit = 3
			}

			addressCount := rand.Intn(limit) + 1
			for k := 0; k < addressCount; k++ {
				randomBytes := make([]byte, 20)
				_, err = crypto.Read(randomBytes)
				if err != nil {
					return err
				}

				err = CreateAccDashboardAccount(dashboardID, int64(j), randomBytes)
				if err != nil {
					return err
				}
			}
		}

		share = rand.Intn(4) == 0
		shareAll = rand.Intn(3) == 0
		shareNotes := rand.Intn(2) == 0

		if share {
			err = CreateAccDashboardSharing(int64(i), "", shareAll, shareNotes, `{"test": true}`)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func insertValidators(dashboardID, groupID, start, count, maxValidatorIndex, pendingAfter int64) error {
	err := insertValidatorsTable(start, count, maxValidatorIndex, pendingAfter)
	if err != nil {
		return err
	}
	err = insertValidatorsDashboard(dashboardID, groupID, start, count, maxValidatorIndex, pendingAfter)
	if err != nil {
		return err
	}
	return nil
}

var multipleMap = map[int64]bool{}

func insertValidatorsTable(start, count, maxValidatorIndex, pendingAfter int64) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn("validators",
		"validator_index",
		"pubkey",
		"validator_index_version",
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k := start; k < start+count; k++ {
		randomBytes := make([]byte, 48)
		_, err = crypto.Read(randomBytes)
		if err != nil {
			return err
		}
		index := k % maxValidatorIndex

		pending := false
		if index > pendingAfter {
			pending = true
		}

		version := 1
		if pending {
			version = 0
		}

		_, ok := multipleMap[index]
		if !ok {
			_, err = stmt.Exec(
				k%maxValidatorIndex,
				randomBytes,
				version,
			)
			multipleMap[index] = true
		}
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func insertValidatorsDashboard(dashboard, group, start, count, maxValidatorIndex, pendingAfter int64) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(pq.CopyIn("users_val_dashboards_validators",
		"validator_index",
		"dashboard_id",
		"group_id",
		"validator_index_version",
	))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k := start; k < start+count; k++ {
		index := k % maxValidatorIndex
		pending := false
		if index > pendingAfter {
			pending = true
		}

		version := 1
		if pending {
			version = 0
		}

		_, err = stmt.Exec(
			index,
			dashboard,
			group,
			version,
		)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type Entry struct {
	ValidatorIndex        int64
	DashboardID           int64
	GroupID               int64
	ValidatorIndexVersion int64
}
