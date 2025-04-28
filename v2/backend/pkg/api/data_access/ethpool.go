// Copyright (C) 2025 Bitfly GmbH
//
// This file is part of Beaconcha.in
//
// Beaconchain Dashboard is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Beaconchain Dashboard is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Beaconchain Dashboard.  If not, see <https://www.gnu.org/licenses/>.

package dataaccess

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type EthpoolRepository interface {
	GetEthpool(context context.Context, day time.Time, validators []t.VDBValidator) ([]t.EthpoolData, error)
}

func (d *DataAccessService) GetEthpool(ctx context.Context, day time.Time, validators []t.VDBValidator) ([]t.EthpoolData, error) {
	type Data struct {
		ValidatorIndex        uint64    `db:"validator_index"`
		T                     time.Time `db:"t"`
		Reward                int64     `db:"reward"`
		ProposedBlocks        uint64    `db:"blocks_proposed"`
		ScheduledBlocks       uint64    `db:"blocks_scheduled"`
		AttestationsExecuted  uint64    `db:"attestations_observed"`
		AttestationsScheduled uint64    `db:"attestations_scheduled"`
		SyncExecuted          uint64    `db:"sync_executed"`
		SyncScheduled         uint64    `db:"sync_scheduled"`
	}

	var queryResults []Data
	ds := goqu.Dialect("postgres").
		From(goqu.T("validator_dashboard_data_daily")).
		Select(
			goqu.C("validator_index"),
			goqu.C("t"),
			goqu.L("COALESCE(attestations_reward, 0) + COALESCE(blocks_cl_reward, 0) + COALESCE(sync_reward_rewards_only, 0)").As("reward"),
			goqu.C("blocks_proposed"),
			goqu.C("blocks_scheduled"),
			goqu.C("attestations_observed"),
			goqu.C("attestations_scheduled"),
			goqu.C("sync_executed"),
			goqu.C("sync_scheduled"),
		).
		Where(
			goqu.C("t").Eq(truncateToDay(day)),
			goqu.C("validator_index").In(validators),
		)
	queryResults, err := runQueryRows[[]Data](ctx, d.clickhouseReader, ds)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %w", err)
	}

	mapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}

	ethpoolData := make([]t.EthpoolData, len(queryResults))
	for i, result := range queryResults {
		ethpoolData[i] = t.EthpoolData{
			Pubkey:               mapping.ValidatorPubkeys[result.ValidatorIndex],
			Day:                  result.T,
			Reward:               result.Reward,
			ProposedBlocks:       result.ProposedBlocks,
			MissedBlocks:         result.ScheduledBlocks - result.ProposedBlocks,
			AttestationsExecuted: result.AttestationsExecuted,
			MissedAttestations:   result.AttestationsScheduled - result.AttestationsExecuted,
			SyncExecuted:         result.SyncExecuted,
			SyncMissed:           result.SyncScheduled - result.SyncExecuted,
		}
	}

	return ethpoolData, nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
