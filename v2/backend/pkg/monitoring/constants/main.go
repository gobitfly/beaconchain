package constants

import "time"

// status enum
type StatusType string

const (
	Running StatusType    = "running"
	Success StatusType    = "success"
	Failure StatusType    = "failure"
	Default time.Duration = -1 * time.Second
)
