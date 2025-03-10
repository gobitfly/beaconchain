package database

import (
	"context"
	"errors"
	"testing"

	"github.com/coocood/freecache"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
)

func TestCache(t *testing.T) {
	tests := []struct {
		name        string
		constructor func(t *testing.T) RemoteCache
	}{
		{
			name:        "mem",
			constructor: func(t *testing.T) RemoteCache { return &MemCache{} },
		},
		{
			name:        "redis",
			constructor: func(t *testing.T) RemoteCache { return Redis{Client: databasetest.NewRedis(t)} },
		},
		{
			name:        "freecache",
			constructor: func(t *testing.T) RemoteCache { return FreeCache{Cache: freecache.NewCache(0)} },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.constructor(t)
			t.Run("set and get", func(t *testing.T) {
				if err := cache.Set(context.Background(), "key", []byte("value"), 0); err != nil {
					t.Errorf("Set() error = %v", err)
				}
				val, err := cache.Get(context.Background(), "key")
				if err != nil {
					t.Errorf("Get() error = %v", err)
				}
				if got, want := string(val), "value"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			})
			t.Run("get return ErrNotFound", func(t *testing.T) {
				_, err := cache.Get(context.Background(), "not found")
				if got, want := err, ErrNotFound; !errors.Is(got, want) {
					t.Errorf("got %v want %v", got, want)
				}
			})
		})
	}
}
