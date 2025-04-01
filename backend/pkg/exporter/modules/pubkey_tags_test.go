package modules

import (
	"context"
	"testing"
	"time"

	dbmocks "github.com/gobitfly/beaconchain/pkg/commons/db2/mocks"
	"github.com/pkg/errors"
)

func TestPubkeyTagsUpdate(t *testing.T) {
	tests := []struct {
		name      string
		mockError error
	}{
		{
			name:      "successful db update",
			mockError: nil,
		},
		{
			name:      "db update error",
			mockError: errors.New("error"),
		},
	}

	mockConsDBClient := new(dbmocks.ConsensusRepository)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	exporter := pubkeyTagsUpdater{
		db:             mockConsDBClient,
		delay:          0,
		ctx:            ctx,
		statusReporter: stubStatusReporter{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConsDBClient.On("UpdatePubkeyTags").Return(tt.mockError)

			exporter.Update()

			mockConsDBClient.AssertCalled(t, "UpdatePubkeyTags")
		})
	}
}
