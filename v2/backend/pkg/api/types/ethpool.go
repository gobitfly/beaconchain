// Copyright (C) 2025 Bitfly GmbH
//
// This file is part of Beaconcha.in.
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

package types

import "time"

type EthpoolData struct {
	Pubkey               string    `json:"pubkey"`
	Day                  time.Time `json:"day"`
	Reward               int64     `json:"reward"`
	ProposedBlocks       uint64    `json:"proposed_blocks"`
	MissedBlocks         uint64    `json:"missed_blocks"`
	AttestationsExecuted uint64    `json:"attestations_executed"`
	MissedAttestations   uint64    `json:"missed_attestations"`
	SyncExecuted         uint64    `json:"sync_executed"`
	SyncMissed           uint64    `json:"sync_missed"`
	BalanceEnd           uint64    `json:"balance_end"`
	BalanceStart         uint64    `json:"balance_start"`
}

type InternalGetEthpoolResponse ApiDataResponse[[]EthpoolData]
