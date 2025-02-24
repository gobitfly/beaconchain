package modules

import (
	"fmt"
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

	utils.Config = &types.Config{
		DeploymentType: "test",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := initializeModules(tt.modules)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestGetEvents(t *testing.T) {
	mockClient := new(consmocks.Client)
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
	utils.Config = &types.Config{
		DeploymentType: "test",
	}

	events := make(chan *constypes.EventResponse, 1)
	go func() {
		events <- &constypes.EventResponse{
			Event: constypes.EventFinalizedCheckpoint,
			Data:  []byte(`{"epoch":"1"}`),
		}
		close(events)
	}()

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
				Error: fmt.Errorf("event error"),
			},
			expectedError: true,
		},
	}

	utils.Config = &types.Config{
		DeploymentType: "test",
		Indexer: types.IndexerConfig{
			Node: types.NodeConfig{
				Host: "",
				Port: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(consmocks.Client)
			context := ModuleContext{
				CL: mockClient,
				ConsClient: &rpc.LighthouseClient{
					CL: &consapi.NodeClient{
						Endpoint: "",
					},
				},
			}

			modules := []ModuleInterface{
				NewSlotExporter(context),
			}

			err := handleEvent(tt.event, eventPool, modules)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}

			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
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

	utils.Config = &types.Config{
		DeploymentType: "test",
		Indexer: types.IndexerConfig{
			Node: types.NodeConfig{
				Host: "",
				Port: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(consmocks.Client)
			context := ModuleContext{
				CL: mockClient,
				ConsClient: &rpc.LighthouseClient{
					CL: &consapi.NodeClient{
						Endpoint: "",
					},
				},
			}

			modules := []ModuleInterface{
				NewSlotExporter(context),
			}

			err := handleHeadEvent(tt.event, eventPool, modules)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}

			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
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

	utils.Config = &types.Config{
		DeploymentType: "test",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleFinalizedCheckpointEvent(tt.event, eventPool, modules)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}

			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
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

	utils.Config = &types.Config{
		DeploymentType: "test",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleChainReorgEvent(tt.event, eventPool, modules)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}

			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
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

	utils.Config = &types.Config{
		DeploymentType: "test",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifyAllModules(eventPool, modules, tt.moduleFunc)
			err := eventPool.Wait()

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}

			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestCreateLighthouseClient(t *testing.T) {
	tests := []struct {
		name          string
		client        consapi.Client
		expectedError bool
	}{
		{
			name:          "successful lighthouse client creation",
			client:        &consapi.NodeClient{},
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
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if client == nil {
					t.Error("expected client to be non-nil")
				}
			}
		})
	}
}

func TestCreateClient(t *testing.T) {
	tests := []struct {
		name          string
		mockSpecResp  *constypes.StandardSpecResponse
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful client creation",
			mockSpecResp:  &constypes.StandardSpecResponse{},
			expectedError: false,
		},
		{
			name:          "error getting spec",
			mockSpecResp:  &constypes.StandardSpecResponse{},
			mockError:     errors.New("spec error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				Indexer: types.IndexerConfig{
					Node: types.NodeConfig{
						Host: "localhost",
						Port: "8080",
					},
				},
			}

			mockClient := new(consmocks.Client)
			mockClient.On("GetSpec").Return(tt.mockSpecResp, tt.mockError)

			mockClientCreator := new(consmocks.ClientCreator)
			mockClientCreator.On("NewClient", "http://localhost:8080").Return(mockClient)

			client, err := createClient(mockClientCreator)
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if client == nil {
					t.Error("expected client to be non-nil")
				}
			}
		})
	}
}
