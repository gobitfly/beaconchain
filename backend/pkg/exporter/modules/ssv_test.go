package modules

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gorilla/websocket"
)

func TestSSVExport(t *testing.T) {
	mockConsDBClient := new(mocks.ConsensusDBI)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	utils.Config = &types.Config{
		SSVExporter: types.SSVExporterConfig{
			Address: "ws://localhost:8080",
		},
	}

	server := startTestWebsocketServer()
	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Log(err)
		}
	}()
	defer server.Close()

	t.Run("websocket works, data is saved and deleted from db", func(t *testing.T) {
		mockExporterResponse := &types.SSVExporterResponse{
			Data: []types.SSVExporterData{
				{
					Publickey: "0xabcd",
				},
			},
		}
		mockConsDBClient.On("DeleteInvalidTags").Return(nil)
		mockConsDBClient.On("SaveValidatorTags", mockExporterResponse.Data).Return(nil)
		mockConsDBClient.On("DeleteValidatorTags").Return(nil)

		exporter := ssvExporter{
			db: mockConsDBClient,
		}

		err := exporter.exportSSV(ctx)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		mockConsDBClient.AssertCalled(t, "DeleteInvalidTags")
		mockConsDBClient.AssertCalled(t, "SaveValidatorTags", mockExporterResponse.Data)
		mockConsDBClient.AssertCalled(t, "DeleteValidatorTags")
	})
}

func startTestWebsocketServer() *http.Server {
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	return &http.Server{
		Addr:              "localhost:8080",
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			if err := conn.WriteJSON(&types.SSVExporterResponse{
				Data: []types.SSVExporterData{{Publickey: "0xabcd"}},
			}); err != nil {
				log.Printf("Error writing JSON: %v", err)
			}
		}),
	}
}
