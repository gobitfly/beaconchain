package dataaccess

import (
	"database/sql"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

//////////////////// 		Helper functions (must be used by more than one VDB endpoint!)

func (d DataAccessService) getDashboardValidators(dashboardId t.VDBId) ([]uint32, error) {
	var validatorsArray []uint32
	if len(dashboardId.Validators) == 0 {
		err := d.alloyReader.Select(&validatorsArray, `SELECT validator_index FROM users_val_dashboards_validators WHERE dashboard_id = $1 ORDER BY validator_index`, dashboardId.Id)
		if err != nil {
			return nil, err
		}
	} else {
		validatorsArray = make([]uint32, 0, len(dashboardId.Validators))
		for _, validator := range dashboardId.Validators {
			validatorsArray = append(validatorsArray, uint32(validator.Index))
		}
	}
	return validatorsArray, nil
}

func (d DataAccessService) calculateTotalEfficiency(attestationEff, proposalEff, syncEff sql.NullFloat64) float64 {
	efficiency := float64(0)

	if !attestationEff.Valid && !proposalEff.Valid && !syncEff.Valid {
		efficiency = 0
	} else if attestationEff.Valid && !proposalEff.Valid && !syncEff.Valid {
		efficiency = attestationEff.Float64 * 100.0
	} else if attestationEff.Valid && proposalEff.Valid && !syncEff.Valid {
		efficiency = ((56.0 / 64.0 * attestationEff.Float64) + (8.0 / 64.0 * proposalEff.Float64)) * 100.0
	} else if attestationEff.Valid && !proposalEff.Valid && syncEff.Valid {
		efficiency = ((62.0 / 64.0 * attestationEff.Float64) + (2.0 / 64.0 * syncEff.Float64)) * 100.0
	} else {
		efficiency = (((54.0 / 64.0) * attestationEff.Float64) + ((8.0 / 64.0) * proposalEff.Float64) + ((2.0 / 64.0) * syncEff.Float64)) * 100.0
	}

	if efficiency < 0 {
		efficiency = 0
	}

	return efficiency
}
