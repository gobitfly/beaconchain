package benchmarks

import (
	"sync"
	"time"
)

type Benchmarker struct {
	TableName string
	Duration  time.Duration
	do        BenchmarkTest
}

func NewBenchmarker(tableName string, duration time.Duration, do BenchmarkTest) *Benchmarker {
	return &Benchmarker{
		TableName: tableName,
		Duration:  duration,
		do:        do,
	}
}

type RunContext struct {
	Wg      *sync.WaitGroup
	EndTime time.Time
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

type BenchmarkTest interface {
	RunBenchmark(b Benchmarker)
}
