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

func (b *Benchmarker) RunBenchmarkDBKiller(duration time.Duration) {
	reportMap = map[string]Report{}
	runContext := getContext(duration)

	// 10/s
	// TODO change to 100ms
	for i := 0; i < 10; i++ {
		runContext.RunSingle("10 Validators", 1*time.Millisecond, func() { b.RunRandomValis(10, b.EpochDepth) })
		runContext.RunSingle("100 Validators", 1*time.Millisecond, func() { b.RunRandomValis(100, b.EpochDepth) })
		runContext.RunSingle("1000 Validators", 1*time.Millisecond, func() { b.RunRandomValis(1000, b.EpochDepth) }) // 100ms

		// 5/s
		runContext.RunSingle("10.000 Validators", 1*time.Millisecond, func() { b.RunRandomValis(10000, b.EpochDepth) })

		if b.ValidatorsInDB > 100000 {
			// 1/s
			runContext.RunSingle("100.000 Validators", 1*time.Millisecond, func() { b.RunRandomValis(100000, b.EpochDepth) })

			// 0.5/s
			runContext.RunSingle("200.000 Validators", 1*time.Millisecond, func() { b.RunRandomValis(200000, b.EpochDepth) })
		} else {
			fmt.Println("!! Skipping 100.000 Validators")
			fmt.Println("!! Skipping 200.000 Validators")
		}
	}

	// 1/10m
	runContext.RunSingle("ExporterAggr 6 Epochs", 5*time.Minute, func() { b.RunGetAllForExport(b.EpochDepth) })

	runContext.RunSingle("ExporterAggr 31 Epochs", 5*time.Minute, func() { b.RunGetAllForExport(31) })

	runContext.wg.Wait()

	fmt.Println("\n== Benchmark finished ==")

	printResult(duration)
}

func (b *Benchmarker) RunBenchmarkParallel(duration time.Duration) {
	reportMap = map[string]Report{}
	runContext := getContext(duration)

	// 10/s
	// TODO change to 100ms
	runContext.RunSingle("10 Validators", 200*time.Millisecond, func() { b.RunRandomValis(10, b.EpochDepth) })
	runContext.RunSingle("100 Validators", 200*time.Millisecond, func() { b.RunRandomValis(100, b.EpochDepth) })
	runContext.RunSingle("1000 Validators", 200*time.Millisecond, func() { b.RunRandomValis(1000, b.EpochDepth) }) // 100ms

	// 5/s
	runContext.RunSingle("10.000 Validators", 200*time.Millisecond, func() { b.RunRandomValis(10000, b.EpochDepth) })

	if b.ValidatorsInDB > 100000 {
		// 1/s
		runContext.RunSingle("100.000 Validators", 1*time.Second, func() { b.RunRandomValis(100000, b.EpochDepth) })

		// 0.5/s
		runContext.RunSingle("200.000 Validators", 2*time.Second, func() { b.RunRandomValis(200000, b.EpochDepth) })
	} else {
		fmt.Println("!! Skipping 100.000 Validators")
		fmt.Println("!! Skipping 200.000 Validators")
	}

	// 1/10m
	runContext.RunSingle("ExporterAggr 6 Epochs", 10*time.Minute, func() { b.RunGetAllForExport(b.EpochDepth) })

	//runContext.RunSingle("ExporterAggr 31 Epochs", 10*time.Minute, func() { b.RunGetAllForExport(31) })

	runContext.wg.Wait()

	fmt.Println("\n== Benchmark finished ==")

	printResult(duration)
}

func (b *Benchmarker) RunBenchmarkSequential(duration time.Duration) {
	reportMap = map[string]Report{}

	getContext(duration).RunSingle("10 Validators", 10*time.Millisecond, func() { b.RunRandomValis(10, b.EpochDepth) }).wg.Wait()
	getContext(duration).RunSingle("100 Validators", 10*time.Millisecond, func() { b.RunRandomValis(100, b.EpochDepth) }).wg.Wait()
	getContext(duration).RunSingle("1000 Validators", 10*time.Millisecond, func() { b.RunRandomValis(1000, b.EpochDepth) }).wg.Wait() // 100ms
	getContext(duration).RunSingle("10.000 Validators", 10*time.Millisecond, func() { b.RunRandomValis(10000, b.EpochDepth) }).wg.Wait()

	if b.ValidatorsInDB > 100000 {
		getContext(duration).RunSingle("100.000 Validators", 10*time.Millisecond, func() { b.RunRandomValis(100000, b.EpochDepth) }).wg.Wait()
		getContext(duration).RunSingle("200.000 Validators", 10*time.Millisecond, func() { b.RunRandomValis(200000, b.EpochDepth) }).wg.Wait()
	} else {
		fmt.Println("!! Skipping 100.000 Validators")
		fmt.Println("!! Skipping 200.000 Validators")
	}

	getContext(duration).RunSingle("ExporterAggr 6 Epochs", 10*time.Millisecond, func() { b.RunGetAllForExport(b.EpochDepth) }).wg.Wait()
	getContext(duration).RunSingle("ExporterAggr 31 Epochs", 10*time.Millisecond, func() { b.RunGetAllForExport(31) }).wg.Wait()

	fmt.Println("\n== Benchmark finished ==")

	printResult(duration)
}

func printResult(duration time.Duration) {
	for _, value := range SortReportMapByID(reportMap) {
		fmt.Printf("Trace Name: %s\n", value.TraceName)
		fmt.Printf("Max: %s\n", value.Max)
		fmt.Printf("Min: %s\n", value.Min)
		fmt.Printf("Avg: %s\n", value.Avg())
		fmt.Printf("IterationCount: %d\n", value.IterationCount)
		fmt.Printf("Req/s: %.3f\n", float64(value.IterationCount)/duration.Seconds())
		fmt.Println()
	}
}

func SortReportMapByID(reportMap map[string]Report) []Report {
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

func getContext(duration time.Duration) *RunContext {
	endTime := time.Now().Add(duration)

	return &RunContext{
		wg:      &sync.WaitGroup{},
		endTime: endTime,
	}
}

func (c *RunContext) RunSingle(traceName string, sleep time.Duration, f func()) *RunContext {
	c.wg.Add(1)
	ctx, cancel := context.WithDeadline(context.Background(), c.endTime)

	go func() {
		// initialize, make so that requests are not all executed at the same time
		mutex.Lock()
		delay := time.Duration(initOffset) * time.Millisecond
		initOffset += 40
		mutex.Unlock()

		time.Sleep(delay) // random delayed start

		for time.Now().Before(c.endTime) {
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
				c.wg.Done()
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
