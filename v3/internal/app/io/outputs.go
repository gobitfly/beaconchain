package io

import "github.com/gobitfly/beaconchain-backend/api/external/model"

func FormatFinalized(requestedEpoch, latestFinalizedEpoch int) model.FinalityParams {
	if requestedEpoch < latestFinalizedEpoch {
		return model.FinalityParams("not_finalized")
	}
	return model.FinalityParams("finalized")
}
