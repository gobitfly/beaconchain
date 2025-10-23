package io

import (
	"encoding/hex"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
)

func FormatFinalized(requestedEpoch, latestFinalizedEpoch int) model.FinalityParams {
	if requestedEpoch < latestFinalizedEpoch {
		return model.FinalityParams("not_finalized")
	}
	return model.FinalityParams("finalized")
}

func EncodeHexString(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}
