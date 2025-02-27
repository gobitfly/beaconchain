package modules

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func TestGenesisDepositsExporter_Export(t *testing.T) {
	tests := []struct {
		name                 string
		mockExporterResponse *SSVExporterResponse
		mockWebsocketError   bool
	}{
		{
			name: "websocket works, data is saved and deleted from db",
			mockExporterResponse: &SSVExporterResponse{
				Data: []SSVExporterData{
					{Publickey: "0xabcd"},
				},
			},
		},
		{
			name: "websocket empty response",
			mockExporterResponse: &SSVExporterResponse{
				Data: []SSVExporterData{
					{Publickey: "0xabcd"},
				},
			},
			mockWebsocketError: true,
		},
	}

	mockConsDBClient := new(mocks.ConsensusDBI)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	exporter := ssvExporter{
		db:  mockConsDBClient,
		ctx: ctx,
	}

	utils.Config = &types.Config{
		SSVExporter: types.SSVExporterConfig{
			Address: "ws://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockWebsocketError {
				exporter.dialer = &StubDialerWebsocketError{}
				exporter.Export()
			} else {
				exporter.dialer = &StubDialer{}

				valueStrings, valueArgs := prepareBatchInsert(tt.mockExporterResponse.Data)

				mockConsDBClient.On("DeleteInvalidTags").Return(nil)
				mockConsDBClient.On("SaveValidatorTags", valueStrings, valueArgs).Return(nil)
				mockConsDBClient.On("DeleteValidatorTags").Return(nil)

				exporter.Export()
				mockConsDBClient.AssertCalled(t, "DeleteInvalidTags")
				mockConsDBClient.AssertCalled(t, "SaveValidatorTags", valueStrings, valueArgs)
				mockConsDBClient.AssertCalled(t, "DeleteValidatorTags")
			}
		})
	}
}

type MockWebSocketConn struct{}

func (m *MockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	if messageType != websocket.TextMessage {
		return errors.New("invalid message type")
	}
	return nil
}

func (m *MockWebSocketConn) ReadMessage() (int, []byte, error) {
	mockResponse := SSVExporterResponse{
		Type: "validator",
		Filter: SSVExporterFilter{
			From: 0,
			To:   10,
		},
		Data: []SSVExporterData{
			{
				Index:     1,
				Publickey: "0xabcd",
				Operators: []SSVExporterOperators{
					{
						Nodeid:    101,
						Publickey: "0x1234",
					},
				},
			},
		},
	}
	jsonData, err := json.Marshal(mockResponse)
	if err != nil {
		return websocket.TextMessage, nil, err
	}
	return websocket.TextMessage, jsonData, nil
}

func (m *MockWebSocketConn) Close() error {
	return nil
}

type StubDialer struct{}

func (s *StubDialer) Dial(url string, requestHeader http.Header) (WebSocketConn, *http.Response, error) {
	return &MockWebSocketConn{}, &http.Response{
		StatusCode: http.StatusSwitchingProtocols,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
	}, nil
}

type StubDialerWebsocketError struct{}

func (s *StubDialerWebsocketError) Dial(url string, requestHeader http.Header) (WebSocketConn, *http.Response, error) {
	return nil, nil, errors.New("empty response")
}
