package services

import (
	"sync"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

func TestInitStatusReporter(t *testing.T) {
	tests := []struct {
		name                   string
		deploymentType         string
		expectedDeploymentType string
	}{
		{
			name:                   "init production",
			deploymentType:         "production",
			expectedDeploymentType: "production",
		},
		{
			name:                   "init development",
			deploymentType:         "development",
			expectedDeploymentType: "development",
		},
		{
			name:                   "init empty",
			deploymentType:         "",
			expectedDeploymentType: "development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configOnce = sync.Once{}
			config = statusConfig{}
			InitStatusReporter(tt.deploymentType)

			if !config.initialized {
				t.Errorf("init failed, got initialized=%v", config.initialized)
			}

			if config.deploymentType != tt.expectedDeploymentType {
				t.Errorf("init failed, got deployment=%s, want=%s", config.deploymentType, tt.expectedDeploymentType)
			}
		})
	}
}

func TestNewStatusReporter(t *testing.T) {
	tests := []struct {
		name         string
		initialized  bool
		expectedStub bool
	}{
		{
			name:         "initialized",
			initialized:  true,
			expectedStub: false,
		},
		{
			name:         "not initialized",
			initialized:  false,
			expectedStub: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config = statusConfig{}
			config.initialized = tt.initialized

			reporter := NewStatusReporter("test", 0, 0)
			_, isStub := reporter.(*stubStatusReporter)

			if isStub != tt.expectedStub {
				t.Errorf("got stub=%v, want=%v", isStub, tt.expectedStub)
			}
		})
	}
}

func TestGetRequiredEvents(t *testing.T) {
	tests := []struct {
		name           string
		deploymentType string
		rocketpool     bool
		pubkeyTags     bool
		expectedEvents int
	}{
		{
			name:           "development",
			deploymentType: "development",
			rocketpool:     false,
			pubkeyTags:     false,
			expectedEvents: len(constants.RequiredEvents),
		},
		{
			name:           "production no extras",
			deploymentType: "production",
			rocketpool:     false,
			pubkeyTags:     false,
			expectedEvents: len(constants.RequiredEvents) + len(constants.ProductionRequiredEvents),
		},
		{
			name:           "production with rocketpool",
			deploymentType: "production",
			rocketpool:     true,
			pubkeyTags:     false,
			expectedEvents: len(constants.RequiredEvents) + len(constants.ProductionRequiredEvents) + 1,
		},
		{
			name:           "production with rocketpool and pubkeyTags",
			deploymentType: "production",
			rocketpool:     true,
			pubkeyTags:     true,
			expectedEvents: len(constants.RequiredEvents) + len(constants.ProductionRequiredEvents) + 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := GetRequiredEvents(tt.deploymentType, tt.rocketpool, tt.pubkeyTags)
			if len(events) < tt.expectedEvents {
				t.Errorf("got %d events, want %d", len(events), tt.expectedEvents)
			}
		})
	}
}
