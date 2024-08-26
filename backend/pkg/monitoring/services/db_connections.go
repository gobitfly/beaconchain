package services

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-redis/redis/v8"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/jmoiron/sqlx"
)

// create db connection service that checks for the status of db connections

type ServerDbConnections struct {
	ServiceBase
}

func (s *ServerDbConnections) internalProcess() {
	defer s.wg.Done()
	s.checkDBConnections()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(10 * time.Second):
			s.checkDBConnections()
		}
	}
}

func (s *ServerDbConnections) Start() {
	if !s.running.CompareAndSwap(false, true) {
		return
	}
	s.wg.Add(1)
	go s.internalProcess()
}

type Entry struct {
	ID string
	DB any
}

func n[T interface{}](id string, db T) *Entry {
	// use reflect to check if db is nil. use reflect. do not simply compare to nil
	if v := reflect.ValueOf(db); !v.IsValid() || v.IsNil() {
		return nil
	}

	return &Entry{id, db}
}

func (s *ServerDbConnections) checkDBConnections() {
	entries := []*Entry{
		n("db_conn_reader_db", db.ReaderDb),
		n("db_conn_writer_db", db.WriterDb),
		n("db_conn_user_reader", db.UserReader),
		n("db_conn_user_writer", db.UserWriter),
		n("db_conn_alloy_reader", db.AlloyReader),
		n("db_conn_alloy_writer", db.AlloyWriter),
		n("db_conn_frontend_reader_db", db.FrontendReaderDB),
		n("db_conn_frontend_writer_db", db.FrontendWriterDB),
		n("db_conn_clickhouse_reader", db.ClickHouseReader),
		n("db_conn_clickhouse_writer", db.ClickHouseWriter),
		n("db_conn_clickhouse_native_writer", db.ClickHouseNativeWriter),
		n("db_conn_persistent_redis_db_client", db.PersistentRedisDbClient),
	}
	if cache.TieredCache != nil {
		entries = append(entries, n("db_conn_tiered_cache", cache.TieredCache))
	}
	wg := sync.WaitGroup{}
	for _, entry := range entries {
		if entry == nil {
			// ignore
			continue
		}
		wg.Add(1)
		go func(entry *Entry) {
			defer wg.Done()
			log.Debugf("checking db connection for %s", entry.ID)
			// context with deadline
			ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
			defer cancel()
			switch edb := entry.DB.(type) {
			case *sqlx.DB:
				err := edb.PingContext(ctx)
				ReportStatus(s.ctx, entry.ID, err, nil, nil)
			case *redis.Client:
				err := edb.Ping(ctx).Err()
				ReportStatus(s.ctx, entry.ID, err, nil, nil)
			case *cache.TieredCacheBase:
				// have to use reflection cause nothing is public. this is a hack. but it works
				val := reflect.ValueOf(edb).Elem().FieldByName("remoteCache")
				if !val.IsValid() {
					log.Error(fmt.Errorf("failed to get remoteCache"), "failed to get remoteCache", 0)
					return
				}
				// its a pointer to a pointer that is cache.RemoteCache compliant. convert it so we can call Get() on it
				rf := reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
				vals := rf.MethodByName("GetBool").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf("test")})
				err := vals[1].Interface().(error)
				// check if its redis nil, if yes ignore
				if err != nil && errors.Is(err, redis.Nil) {
					err = nil
				}
				ReportStatus(s.ctx, entry.ID, err, nil, nil)
			case ch.Conn: // its an interface
				err := edb.Ping(ctx)
				ReportStatus(s.ctx, entry.ID, err, nil, nil)
			default:
				log.Error(fmt.Errorf("unknown db type"), "unknown db type", 0, map[string]interface{}{"entry": entry})
			}
		}(entry)
	}
	wg.Wait()
}
