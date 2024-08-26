package services

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

type FlakyTestService struct {
	ServiceBase
}

func (s *FlakyTestService) Start() {
	if !s.running.CompareAndSwap(false, true) {
		// already running, return error
		return
	}
	s.wg.Add(1)
	go s.internalProcess()
}

func (s *FlakyTestService) internalProcess() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(10 * time.Second):
			err := fmt.Errorf("random error")
			NewStatusReport("flaky_test", constants.Default, constants.Default)(constants.Failure, map[string]string{"error": err.Error()})
		}
	}
}
