package modules

import (
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	consmocks "github.com/gobitfly/beaconchain/pkg/consapi/mocks"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"

	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

func TestStartAll(t *testing.T) {
	mockChainHeadResponse := &types.ChainHead{
		HeadEpoch:      12345678,
		FinalizedEpoch: 12345678,
	}

	utils.Config = &types.Config{
		DeploymentType: "development",
	}

	t.Run("valid data export start", func(t *testing.T) {
		mockClient := new(consmocks.ClientInt)
		mockConsClient := new(mocks.ConsClient)
		moduleCtx := ModuleContext{
			CL:         mockClient,
			ConsClient: mockConsClient,
		}

		moduleInterface := []ModuleInterface{
			stubModuleInterface{},
		}
		events := make(chan *constypes.EventResponse, 1)
		go func() {
			events <- &constypes.EventResponse{
				Event: constypes.EventFinalizedCheckpoint,
				Data:  []byte(`{"epoch":"1"}`),
			}
			events <- &constypes.EventResponse{
				Event: constypes.EventHead,
				Data:  []byte(`{"slot":"1"}`),
			}
			events <- &constypes.EventResponse{
				Event: constypes.EventChainReorg,
				Data:  []byte(`{"slot":"1"}`),
			}
			close(events)
		}()

		mockConsClient.On("GetChainHead").Return(mockChainHeadResponse, nil)
		mockClient.On("GetEvents", []constypes.EventTopic{
			constypes.EventHead,
			constypes.EventFinalizedCheckpoint,
			constypes.EventChainReorg,
		}).Return(events)

		err := StartAll(moduleCtx, moduleInterface, true)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		mockConsClient.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("empty modules list", func(t *testing.T) {
		mockClient := new(consmocks.ClientInt)
		mockConsClient := new(mocks.ConsClient)
		moduleCtx := ModuleContext{
			CL:         mockClient,
			ConsClient: mockConsClient,
		}

		moduleInterface := []ModuleInterface{}

		mockConsClient.On("GetChainHead").Return(mockChainHeadResponse, nil)

		err := StartAll(moduleCtx, moduleInterface, true)
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		mockConsClient.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})
}

type stubModuleInterface struct{}

func (sm stubModuleInterface) Init() error {
	return nil
}

func (sm stubModuleInterface) GetName() string {
	return "stubModule"
}

func (sm stubModuleInterface) GetMonitoringEventId() constants.Event {
	return constants.Event_ExporterModuleSlotExporter
}

func (sm stubModuleInterface) OnHead(*constypes.StandardEventHeadResponse) error {
	return nil
}

func (sm stubModuleInterface) OnFinalizedCheckpoint(*constypes.StandardFinalizedCheckpointResponse) error {
	return nil
}

func (sm stubModuleInterface) OnChainReorg(*constypes.StandardEventChainReorg) error {
	return nil
}
