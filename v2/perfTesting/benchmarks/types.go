package benchmarks

import (
	"sync"
	"time"
)

type Benchmarker struct {
	TableName       string
	ValidatorsInDB  int
	EpochsInDB      int
	UseLatestEpochs bool // when true it will not randomly select epochs but use the latest epoch as base for the requests
	LatestEpoch     int
	EpochDepth      int
}

type RunContext struct {
	wg      *sync.WaitGroup
	endTime time.Time
}

type Report struct {
	TraceName      string
	ID             int
	Max            time.Duration
	Min            time.Duration
	All            time.Duration
	IterationCount int
}

func (r *Report) Avg() time.Duration {
	return r.All / time.Duration(r.IterationCount)
}
