package modules

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi"
	consmocks "github.com/gobitfly/beaconchain/pkg/consapi/mocks"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	eventPool = &errgroup.Group{}
	modules   = []ModuleInterface{
		NewSlotExporter(ModuleContext{}),
	}
)

func TestInitializeModules(t *testing.T) {
	tests := []struct {
		name          string
		modules       []ModuleInterface
		expectedError bool
	}{
		{
			name: "valid module initialization",
			modules: []ModuleInterface{
				NewSlotExporter(ModuleContext{}),
			},
			expectedError: false,
		},
		{
			name:          "empty modules list",
			modules:       []ModuleInterface{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				DeploymentType: "development",
			}

			err := initializeModules(tt.modules)
			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestGetEvents(t *testing.T) {
	mockClient := new(consmocks.ClientInt)
	mockClient.On("GetEvents", []constypes.EventTopic{
		constypes.EventHead,
		constypes.EventFinalizedCheckpoint,
		constypes.EventChainReorg,
	}).Return(make(chan *constypes.EventResponse, 1))

	context := &ModuleContext{
		CL: mockClient,
	}

	events := getEvents(context)
	if events == nil {
		t.Error("expected events channel to be non-nil")
	}
}

func TestHandleEvents(t *testing.T) {
	events := make(chan *constypes.EventResponse, 1)
	go func() {
		events <- &constypes.EventResponse{
			Event: constypes.EventFinalizedCheckpoint,
			Data:  []byte(`{"epoch":"1"}`),
		}
		close(events)
	}()

	utils.Config = &types.Config{
		DeploymentType: "development",
	}

	handleEvents(events, modules)
}

func TestHandleEvent(t *testing.T) {
	tests := []struct {
		name          string
		event         *constypes.EventResponse
		expectedError bool
	}{
		{
			name: "head event",
			event: &constypes.EventResponse{
				Event: constypes.EventHead,
				Data:  []byte(`{"slot":"1"}`),
			},
			expectedError: false,
		},
		{
			name: "finalized checkpoint event",
			event: &constypes.EventResponse{
				Event: constypes.EventFinalizedCheckpoint,
				Data:  []byte(`{"epoch":"1"}`),
			},
			expectedError: false,
		},
		{
			name: "chain reorg event",
			event: &constypes.EventResponse{
				Event: constypes.EventChainReorg,
				Data:  []byte(`{"slot":"1"}`),
			},
			expectedError: false,
		},
		{
			name: "event with error",
			event: &constypes.EventResponse{
				Error: fmt.Errorf("error"),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				DeploymentType: "development",
			}

			mockClient := new(consmocks.ClientInt)
			lc, err := rpc.NewLighthouseClient(&consapi.NodeClient{}, big.NewInt(1))
			if err != nil {
				t.Errorf("error creating lighthouse client: %v", err)
			}

			context := ModuleContext{
				CL:         mockClient,
				ConsClient: lc,
			}

			modules := []ModuleInterface{
				NewSlotExporter(context),
			}

			err = handleEvent(tt.event, eventPool, modules)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestHandleHeadEvent(t *testing.T) {
	tests := []struct {
		name          string
		event         *constypes.EventResponse
		expectedError bool
	}{
		{
			name: "head event",
			event: &constypes.EventResponse{
				Event: constypes.EventHead,
				Data:  []byte(`{"slot":"1"}`),
			},
			expectedError: false,
		},
		{
			name: "event with error",
			event: &constypes.EventResponse{
				Event: constypes.EventHead,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				DeploymentType: "development",
			}

			mockClient := new(consmocks.ClientInt)
			lc, err := rpc.NewLighthouseClient(&consapi.NodeClient{}, big.NewInt(1))
			if err != nil {
				t.Errorf("error creating lighthouse client: %v", err)
			}

			context := ModuleContext{
				CL:         mockClient,
				ConsClient: lc,
			}

			modules := []ModuleInterface{
				NewSlotExporter(context),
			}

			err = handleHeadEvent(tt.event, eventPool, modules)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestHandleFinalizedCheckpointEvent(t *testing.T) {
	tests := []struct {
		name          string
		event         *constypes.EventResponse
		expectedError bool
	}{
		{
			name: "valid finalized checkpoint event",
			event: &constypes.EventResponse{
				Event: constypes.EventFinalizedCheckpoint,
				Data:  []byte(`{"epoch":"1"}`),
			},
			expectedError: false,
		},
		{
			name: "event with error",
			event: &constypes.EventResponse{
				Event: constypes.EventFinalizedCheckpoint,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				DeploymentType: "development",
			}

			err := handleFinalizedCheckpointEvent(tt.event, eventPool, modules)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestHandleChainReorgEvent(t *testing.T) {
	tests := []struct {
		name          string
		event         *constypes.EventResponse
		expectedError bool
	}{
		{
			name: "chain reorg event",
			event: &constypes.EventResponse{
				Event: constypes.EventChainReorg,
				Data:  []byte(`{"slot":"1"}`),
			},
			expectedError: false,
		},
		{
			name: "event with error",
			event: &constypes.EventResponse{
				Event: constypes.EventChainReorg,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				DeploymentType: "development",
			}

			err := handleChainReorgEvent(tt.event, eventPool, modules)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestNotifyAllModules(t *testing.T) {
	tests := []struct {
		name          string
		moduleFunc    func(module ModuleInterface) error
		expectedError bool
	}{
		{
			name: "valid notify all modules",
			moduleFunc: func(module ModuleInterface) error {
				return module.Init()
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				DeploymentType: "development",
			}

			notifyAllModules(eventPool, modules, tt.moduleFunc)
			err := eventPool.Wait()

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestCreateLighthouseClient(t *testing.T) {
	tests := []struct {
		name          string
		client        consapi.ClientInt
		expectedError bool
	}{
		{
			name:          "successful lighthouse client creation",
			client:        consapi.NewClient("http://localhost:8080"),
			expectedError: false,
		},
		{
			name:          "error creating lighthouse client",
			client:        nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						DepositChainID: 1,
					},
				},
			}

			client, err := createLighthouseClient(tt.client)

			if err != nil {
				if tt.expectedError {
					return
				}
				if client == nil {
					t.Error("expected client to be non-nil")
				}

				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestCreateClient(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful client creation",
			expectedError: false,
		},
		{
			name:          "error getting spec",
			mockError:     errors.New("error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(consmocks.ClientInt)
			mockClient.On("GetSpec").Return(&constypes.StandardSpecResponse{}, tt.mockError)
			err := getClientSpec(mockClient)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
		})
	}
}
