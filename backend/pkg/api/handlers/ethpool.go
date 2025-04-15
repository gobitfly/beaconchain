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

package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/types"
)

func (i *inputGetEthpool) Validate(params map[string]string, body io.ReadCloser) error {
	var v validationError
	type request struct {
		Validators []intOrString `json:"validators"`
		Day        time.Time     `json:"day"`
		Signature  string        `json:"signature"`
	}
	req := request{}
	if err := v.checkBody(&req, body); err != nil {
		return err
	}

	if len(req.Validators) == 0 {
		return newBadRequestErr("no validators provided")
	}

	indices, _ := v.checkValidators(req.Validators, forbidEmpty)
	if err := v.AsError(); err != nil {
		return err
	}

	signatureBytes, err := hex.DecodeString(req.Signature)
	if err != nil || len(signatureBytes) != 32 {
		return newBadRequestErr("invalid signature")
	}

	i.Validators = indices
	i.Day = req.Day
	i.Signature = signatureBytes
	return v.AsError()
}

type inputGetEthpool struct {
	Validators []types.VDBValidator
	Day        time.Time
	Signature  []byte
}

func (h *HandlerService) InternalGetEthpool(ctx context.Context, req inputGetEthpool) (types.InternalGetEthpoolResponse, error) {
	var r types.InternalGetEthpoolResponse
	dataAccessor := h.getDataAccessor(ctx)

	secretBytes, err := hex.DecodeString(h.cfg.Frontend.BeaconchainETHPoolBridgeSecret)
	if err != nil {
		return r, newInternalServerErr("invalid secret")
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(fmt.Sprintf("%s%s", JoinValidators(req.Validators, ","), req.Day)))
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(req.Signature, expectedMAC) {
		return r, newBadRequestErr("invalid signature")
	}

	data, err := dataAccessor.GetEthpool(ctx, req.Day, req.Validators)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, newNotFoundErr("no data found for the given day")
		}
		return r, err
	}

	r.Data = data
	return r, nil
}
