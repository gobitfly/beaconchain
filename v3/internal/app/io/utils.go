package io

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimePtrToProtoTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
