package modules

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func TestSSVExport(t *testing.T) {
	mockConsDBClient := new(mocks.ConsensusDBI)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	utils.Config = &types.Config{
		SSVExporter: types.SSVExporterConfig{
			Address: "ws://localhost:8080",
		},
	}

	t.Run("websocket works, data is saved and deleted from db", func(t *testing.T) {
		mockExporterResponse := &types.SSVExporterResponse{
			Data: []types.SSVExporterData{
				{
					Publickey: "0xabcd",
				},
			},
		}

		mockConn := NewMockWebSocketConn(mockExporterResponse)
		exporter := ssvExporter{
			db:     mockConsDBClient,
			dialer: &StubDialer{MockConn: mockConn},
		}

		mockConsDBClient.On("DeleteInvalidTags").Return(nil)
		mockConsDBClient.On("SaveValidatorTags", mockExporterResponse.Data).Return(nil)
		mockConsDBClient.On("DeleteValidatorTags").Return(nil)

		err := exporter.exportSSV(ctx)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		mockConsDBClient.AssertCalled(t, "DeleteInvalidTags")
		mockConsDBClient.AssertCalled(t, "SaveValidatorTags", mockExporterResponse.Data)
		mockConsDBClient.AssertCalled(t, "DeleteValidatorTags")
	})

	t.Run("websocket error", func(t *testing.T) {
		exporter := ssvExporter{
			db:     mockConsDBClient,
			dialer: &StubDialerWebsocketError{},
		}

		err := exporter.exportSSV(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

type MockWebSocketConn struct {
	response *types.SSVExporterResponse
}

func NewMockWebSocketConn(response *types.SSVExporterResponse) *MockWebSocketConn {
	return &MockWebSocketConn{
		response: response,
	}
}

func (m *MockWebSocketConn) ReadMessage() (int, []byte, error) {
	jsonData, err := json.Marshal(m.response)
	if err != nil {
		return 0, nil, err
	}
	return websocket.TextMessage, jsonData, nil
}

func (m *MockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	return nil
}

func (m *MockWebSocketConn) Close() error {
	return nil
}

type StubDialer struct {
	MockConn WebSocketConnInterface
}

func (s *StubDialer) Dial(url string, requestHeader http.Header) (WebSocketConnInterface, error) {
	return s.MockConn, nil
}

type StubDialerWebsocketError struct{}

func (s *StubDialerWebsocketError) Dial(url string, requestHeader http.Header) (WebSocketConnInterface, error) {
	return nil, errors.New("error")
}
