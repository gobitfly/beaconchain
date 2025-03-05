package notification

import (
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/google/go-cmp/cmp"
)

func TestExtractEventNameStringFromNotificationForMetrics(t *testing.T) {
	tests := map[string]struct {
		notification Notification
		want         *types.EventName
		wantsError   bool
	}{
		"Basic Notification": {
			notification: Notification{Content: "notification\": {\"body\": \"Attestation missed: 1 validator (999999)\", \"title\": \"Info for epoch 344581\"}}"},
			want:         ptr(types.ValidatorMissedAttestationEventName),
			wantsError:   false,
		},
		"Unknown EventName": {
			notification: Notification{Content: "Im Mr.Meeseeks look at me"},
			want:         nil,
			wantsError:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ExtractEventNameFromNotification(tc.notification)
			if tc.wantsError {
				if err == nil {
					t.Fatalf("expected an error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				diff := cmp.Diff(tc.want, got)
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

// Cant create a pointer to a constant, so this function helps us refer to a "nil" EventName in the case where an error is returned
func ptr(e types.EventName) *types.EventName {
	return &e
}

func TestGetEventLabelForNotification(t *testing.T) {
	tests := map[string]struct {
		notification Notification
		want         string
	}{
		"Basic Notification": {
			notification: Notification{Content: "notification\": {\"body\": \"Attestation missed: 1 validator (999999)\", \"title\": \"Info for epoch 344581\"}}"},
			want:         string(types.ValidatorMissedAttestationEventName),
		},
		"Unknown EventName": {
			notification: Notification{Content: "Im Mr.Meeseeks look at me"},
			want:         UnknownEvent,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := GetEventLabelForNotification(tc.notification)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestCountByEventType(t *testing.T) {
	tests := map[string]struct {
		notifications []Notification
		expectedCount map[string]int
	}{
		"Single event type": {
			notifications: []Notification{
				{Content: types.EventLabel[types.ValidatorMissedAttestationEventName]},
			},
			expectedCount: map[string]int{
				string(types.ValidatorMissedAttestationEventName): 1,
				UnknownEvent: 0,
			},
		},
		"Multiple event types": {
			notifications: []Notification{
				{Content: types.EventLabel[types.ValidatorMissedAttestationEventName]},
				{Content: types.EventLabel[types.ValidatorMissedAttestationEventName]},
				{Content: types.EventLabel[types.ValidatorIsOfflineEventName]},
			},
			expectedCount: map[string]int{
				string(types.ValidatorMissedAttestationEventName): 2,
				string(types.ValidatorIsOfflineEventName):         1,
				UnknownEvent: 0,
			},
		},
		"Unknown event type": {
			notifications: []Notification{
				{Content: "Im Mr.Meeseeks look at me"},
			},
			expectedCount: map[string]int{
				string(types.ValidatorMissedAttestationEventName): 0,
				UnknownEvent: 1,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := CountByEventType(tc.notifications)
			for eventType, expectedCount := range tc.expectedCount {
				if got[eventType] != expectedCount {
					t.Fatalf("expected %d for event type %s, got %d", expectedCount, eventType, got[eventType])
				}
			}
		})
	}
}

func TestCountByChannelType(t *testing.T) {
	tests := map[string]struct {
		notifications []Notification
		expectedCount map[string]int
	}{
		"Single channel type": {
			notifications: []Notification{
				{Channel: string(types.EmailNotificationChannel)},
			},
			expectedCount: map[string]int{
				string(types.EmailNotificationChannel): 1,
			},
		},
		"Multiple channel types": {
			notifications: []Notification{
				{Channel: string(types.EmailNotificationChannel)},
				{Channel: string(types.EmailNotificationChannel)},
				{Channel: string(types.PushNotificationChannel)},
			},
			expectedCount: map[string]int{
				string(types.EmailNotificationChannel): 2,
				string(types.PushNotificationChannel):  1,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := CountByChannel(tc.notifications)
			for channelType, expectedCount := range tc.expectedCount {
				if got[channelType] != expectedCount {
					t.Fatalf("expected %d for channel type %s, got %d", expectedCount, channelType, got[channelType])
				}
			}
		})
	}
}
