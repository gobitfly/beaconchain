package modules

import (
	"context"
	"testing"
	"time"

	dbmocks "github.com/gobitfly/beaconchain/pkg/commons/db2/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

func TestMEVBoostRelaysExporter(t *testing.T) {
	tests := []struct {
		name                         string
		mockRelays                   []types.Relay
		mockLastRelayBlock           types.RelayBlock
		mockFirstRelayBlock          types.RelayBlock
		mockNoRelaysInDB             bool
		mockEmptyPayload             bool
		mockShouldTryToExportIsFalse bool
	}{
		{
			name: "successful export to db",
			mockRelays: []types.Relay{
				{
					ID:                 "ID",
					Endpoint:           "http://localhost",
					ExportFailureCount: 0,
				},
			},
			mockLastRelayBlock: types.RelayBlock{
				BlockSlot: 2,
			},
			mockFirstRelayBlock: types.RelayBlock{
				BlockSlot: 1,
			},
		},
		{
			name:             "no relays in db",
			mockRelays:       []types.Relay{},
			mockNoRelaysInDB: true,
		},
		{
			name: "empty payload",
			mockRelays: []types.Relay{
				{
					ID:                 "ID",
					Endpoint:           "http://localhost",
					ExportFailureCount: 0,
				},
			},
			mockLastRelayBlock: types.RelayBlock{
				BlockSlot: 2,
			},
			mockEmptyPayload: true,
		},
		{
			name: "shouldTryToExportRelay is false",
			mockRelays: []types.Relay{
				{
					ID:                 "ID",
					Endpoint:           "http://localhost",
					ExportFailureCount: 2,
					LastExportTryTs:    time.Now().Add(time.Minute * 10),
				},
			},
			mockShouldTryToExportIsFalse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConsDBClient := new(dbmocks.ConsensusRepository)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			exporter := relaysExporter{
				db:    mockConsDBClient,
				delay: 0,
				ctx:   ctx,
			}

			if tt.mockNoRelaysInDB {
				exporter.relayClient = new(mockRelayClient)
				mockConsDBClient.On("GetRelays").Return(tt.mockRelays, nil)
				exporter.MEVBoostRelaysExporter()
				mockConsDBClient.AssertCalled(t, "GetRelays")
				return
			}

			if tt.mockEmptyPayload {
				exporter.relayClient = new(mockRelayClientWithEmptyPayload)
				mockConsDBClient.On("GetRelays").Return(tt.mockRelays, nil)
				mockConsDBClient.On("UpdateRelay", tt.mockRelays[0].ID, tt.mockRelays[0].Endpoint).Return(nil)
				mockConsDBClient.On("GetLastRelayBlock", tt.mockRelays[0].ID).Return(tt.mockLastRelayBlock, nil)
				mockConsDBClient.On("GetFirstRelayBlock", tt.mockRelays[0].ID).Return(tt.mockFirstRelayBlock, nil)

				exporter.MEVBoostRelaysExporter()

				mockConsDBClient.AssertCalled(t, "GetRelays")
				mockConsDBClient.AssertCalled(t, "UpdateRelay", tt.mockRelays[0].ID, tt.mockRelays[0].Endpoint)
				mockConsDBClient.AssertCalled(t, "GetLastRelayBlock", tt.mockRelays[0].ID)
				mockConsDBClient.AssertCalled(t, "GetFirstRelayBlock", tt.mockRelays[0].ID)
				return
			}

			if tt.mockShouldTryToExportIsFalse {
				exporter.relayClient = new(mockRelayClient)
				mockConsDBClient.On("GetRelays").Return(tt.mockRelays, nil)

				exporter.MEVBoostRelaysExporter()

				mockConsDBClient.AssertCalled(t, "GetRelays")
				return
			}

			exporter.relayClient = new(mockRelayClient)
			mockConsDBClient.On("GetRelays").Return(tt.mockRelays, nil)
			mockConsDBClient.On("GetLastRelayBlock", tt.mockRelays[0].ID).Return(tt.mockLastRelayBlock, nil)
			mockConsDBClient.On("SaveBlockTagsAndRelays", tt.mockRelays[0].ID, payload).Return(nil)
			mockConsDBClient.On("GetFirstRelayBlock", tt.mockRelays[0].ID).Return(tt.mockFirstRelayBlock, nil)
			mockConsDBClient.On("UpdateRelay", tt.mockRelays[0].ID, tt.mockRelays[0].Endpoint).Return(nil)

			exporter.MEVBoostRelaysExporter()

			mockConsDBClient.AssertCalled(t, "GetRelays")
			mockConsDBClient.AssertCalled(t, "GetLastRelayBlock", tt.mockRelays[0].ID)
			mockConsDBClient.AssertCalled(t, "SaveBlockTagsAndRelays", tt.mockRelays[0].ID, payload)
			mockConsDBClient.AssertCalled(t, "GetFirstRelayBlock", tt.mockRelays[0].ID)
			mockConsDBClient.AssertCalled(t, "UpdateRelay", tt.mockRelays[0].ID, tt.mockRelays[0].Endpoint)
		})
	}
}

var payload = []types.BidTrace{
	{
		Slot:                 12345,
		ParentHash:           "0xaaabbb",
		BlockHash:            "0xabcdef",
		BuilderPubkey:        "0x123456",
		ProposerPubkey:       "0x234567",
		ProposerFeeRecipient: "0x345678",
		GasLimit:             1000,
		GasUsed:              100,
		Value:                toWeiString("1000"),
	},
}

func toWeiString(value string) types.WeiString {
	weiValue := types.WeiString{}
	err := weiValue.Set(value)
	if err != nil {
		panic(err)
	}
	return weiValue
}

type mockRelayClient struct{}

func (m mockRelayClient) fetchDeliveredPayloads(endpoint string, id string, offset uint64) ([]types.BidTrace, error) {
	return payload, nil
}

type mockRelayClientWithEmptyPayload struct{}

func (m mockRelayClientWithEmptyPayload) fetchDeliveredPayloads(endpoint string, id string, offset uint64) ([]types.BidTrace, error) {
	return nil, nil
}
