package utils

import "database/sql"

func CalculateTotalEfficiency(attestationEff, proposalEff, syncEff sql.NullFloat64) float64 {
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
