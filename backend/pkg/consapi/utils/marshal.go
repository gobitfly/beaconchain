package utils

import (
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/google/uuid"

	//"encoding/json"
	//"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
)

const MinSpeed = 2.5 * 1024 * 1024 // 1 MB/s
// keep track of how many debug readers are active
// we will use this later to only trigger the abort if its less than 5
// it needs to be coroutine safe

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	mu sync.Mutex
	v  int
}

// Increment increments the counter by 1.
func (c *SafeCounter) Increment() {
	c.mu.Lock()
	c.v++
	c.mu.Unlock()
}

// Decrement decrements the counter by 1.
func (c *SafeCounter) Decrement() {
	c.mu.Lock()
	c.v--
	c.mu.Unlock()
}

// Value returns the current value of the counter.
func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.v
}

var debugReaders = SafeCounter{v: 0}

// DebugReadCloser is a custom ReadCloser that wraps another io.ReadCloser and logs the data being read.
type DebugReadCloser struct {
	rc          io.ReadCloser
	id          string
	startTime   time.Time
	totalBytes  int
	debugTicker *time.Ticker
	stopChan    chan struct{} // Add a channel to signal the goroutine to stop
}

// Read reads from the underlying ReadCloser, logs the data, and then returns the data.
func (drc *DebugReadCloser) Read(p []byte) (int, error) {
	if drc.id == "" {
		debugReaders.Increment()
		// Initialize
		drc.id = uuid.New().String()
		drc.startTime = time.Now()
		// Start ticker, every 3 seconds
		drc.debugTicker = time.NewTicker(time.Second * 3)
		drc.stopChan = make(chan struct{}) // Initialize the stop channel
		go func() {
			for {
				select {
				case <-drc.debugTicker.C:
					elapsed := time.Since(drc.startTime).Seconds()
					if elapsed > 0 {
						actualMinSpeed := math.Max(MinSpeed, MinSpeed*float64((elapsed-12)/10))
						// cap it at 10MiB/s
						actualMinSpeed = math.Min(actualMinSpeed, 10*1024*1024)
						averageSpeed := float64(drc.totalBytes) / elapsed
						log.Debugf("Read %d bytes in the last 3 seconds. Average speed: %s/s, Threshold %s/s (response %s)", drc.totalBytes, humanize.IBytes(uint64(averageSpeed)), humanize.IBytes(uint64(actualMinSpeed)), drc.id)
						// scale by time. at 15 seconds, min should be 1x
						// at 30 seconds at min should be 2x
						// only abort if we have less than 5 debug readers
						// check if hostname contains invis, do nothing if true. contains, not equals
						osHostname, _ := os.Hostname()
						if strings.Contains(osHostname, "invis") {
							continue
						}
						if elapsed >= 12 && averageSpeed < actualMinSpeed && debugReaders.Value() <= 5 {
							log.Warnf("Average speed below %s/s over %v, aborting reader %s", humanize.IBytes(uint64(actualMinSpeed)), time.Second*time.Duration(elapsed), drc.id)
							drc.rc.Close()
							return
						}
					}
				case <-drc.stopChan: // Listen for stop signal
					return
				}
			}
		}()
	}

	n, err := drc.rc.Read(p)
	if n > 0 {
		drc.totalBytes += n
	}
	return n, err
}

// Close closes the underlying ReadCloser.
func (drc *DebugReadCloser) Close() error {
	drc.debugTicker.Stop()
	close(drc.stopChan) // Signal the goroutine to stop
	debugReaders.Decrement()
	return drc.rc.Close()
}

func Unmarshal[T any](source_org io.ReadCloser, err error) (*T, error) {
	source := &DebugReadCloser{rc: source_org}
	defer source.Close()
	var target T
	if err != nil {
		return &target, err
	}

	if err := json.NewDecoder(source).Decode(&target); err != nil {
		return &target, errors.Wrap(err, "unmarshal json failed")
	}

	return &target, nil
}

func UnmarshalOld[T any](source []byte, err error) (*T, error) {
	var target T
	if err != nil {
		return &target, err
	}

	if err := json.Unmarshal(source, &target); err != nil {
		return &target, errors.Wrap(err, "unmarshal json failed")
	}

	return &target, nil
}
