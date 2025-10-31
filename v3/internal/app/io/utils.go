package io

import (
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimePtrToProtoTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func IsHealthCheckEndpoint(fullMethod string) bool {
	return strings.HasPrefix(fullMethod, "/grpc.health.v1.Health/")
}
