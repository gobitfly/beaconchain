package benchmarks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

var reportMap map[string]Report
var mutex sync.Mutex

var initOffset = 0

func (b *Benchmarker) Run() {
	reportMap = map[string]Report{}

	b.do.RunBenchmark(*b)

	fmt.Println("\n== Benchmark finished ==")

	printResult(b.Duration)
}

func printResult(duration time.Duration) {
	for _, value := range sortReportMapByID(reportMap) {
		fmt.Printf("Trace Name: %s\n", value.TraceName)
		fmt.Printf("Max: %s\n", value.Max)
		fmt.Printf("Min: %s\n", value.Min)
		fmt.Printf("Avg: %s\n", value.Avg())
		fmt.Printf("IterationCount: %d\n", value.IterationCount)
		fmt.Printf("Req/s: %.3f\n", float64(value.IterationCount)/duration.Seconds())
		fmt.Println()
	}
}

func sortReportMapByID(reportMap map[string]Report) []Report {
	// Convert the map to a slice of Report structs
	reports := make([]Report, 0, len(reportMap))
	for _, value := range reportMap {
		reports = append(reports, value)
	}

	// Define a custom sorting function
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ID < reports[j].ID
	})

	return reports
}

func (b *Benchmarker) GetContext() *RunContext {
	endTime := time.Now().Add(b.Duration)

	return &RunContext{
		Wg:      &sync.WaitGroup{},
		EndTime: endTime,
	}
}

func (c *RunContext) RunSingle(traceName string, sleep time.Duration, f func()) *RunContext {
	c.Wg.Add(1)
	ctx, cancel := context.WithDeadline(context.Background(), c.EndTime)

	go func() {
		// initialize, make so that requests are not all executed at the same time
		mutex.Lock()
		delay := time.Duration(initOffset) * time.Millisecond
		initOffset += 40
		mutex.Unlock()

		time.Sleep(delay) // random delayed start

		for time.Now().Before(c.EndTime) {
			took := Trace(traceName, f)
			time.Sleep(sleep)
			if took < sleep {
				time.Sleep(sleep - took)
			}
		}
		cancel()
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.Wg.Done()
				return
			default:
				// Do nothing
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return c
}

func Trace(traceName string, f func()) time.Duration {
	start := time.Now()
	f()
	elapsed := time.Since(start)
	fmt.Printf("[%s] %stook %s\n", traceName, gap(len(traceName), 25), elapsed)

	mutex.Lock()
	if report, ok := reportMap[traceName]; ok {
		if elapsed > report.Max {
			report.Max = elapsed
		}
		if elapsed < report.Min {
			report.Min = elapsed
		}
		report.All += elapsed
		report.IterationCount++

		reportMap[traceName] = report
	} else {
		reportMap[traceName] = Report{
			TraceName:      traceName,
			ID:             len(reportMap),
			Max:            elapsed,
			Min:            elapsed,
			All:            elapsed,
			IterationCount: 1,
		}
	}
	mutex.Unlock()
	return elapsed
}

func gap(is, target int) string {
	erg := ""
	for i := 0; i < target-is; i++ {
		erg += " "
	}
	return erg
}
