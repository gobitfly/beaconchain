package db


import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"
	"sync"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

var ClickHouseNativeWriter ch.Conn

func MustInitClickhouseNative(writer *types.DatabaseConfig) ch.Conn {
	if writer.MaxOpenConns == 0 {
		writer.MaxOpenConns = 50
	}
	if writer.MaxIdleConns == 0 {
		writer.MaxIdleConns = 10
	}
	if writer.MaxOpenConns < writer.MaxIdleConns {
		writer.MaxIdleConns = writer.MaxOpenConns
	}
	log.Infof("initializing clickhouse native writer db connection to %v:%v/%v with %v/%v conn limit", writer.Host, writer.Port, writer.Name, writer.MaxIdleConns, writer.MaxOpenConns)
	dbWriter, err := ch.Open(&ch.Options{
		MaxOpenConns: writer.MaxOpenConns,
		MaxIdleConns: writer.MaxIdleConns,
		// ConnMaxLifetime: time.Minute,
		// the following lowers traffic between client and server
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		Addr: []string{fmt.Sprintf("%s:%s", writer.Host, writer.Port)},
		Auth: ch.Auth{
			Username: writer.Username,
			Password: writer.Password,
			Database: writer.Name,
		},
		Debug: false,
		TLS:   &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12},
		// this gets only called when debug is true
		Debugf: func(s string, p ...interface{}) {
			log.Debugf("CH NATIVE WRITER: "+s, p...)
		},
	})
	if err != nil {
		log.Fatal(err, "Error connecting to clickhouse native writer", 0)
	}
	// verify connection
	ClickHouseTestConnection(&dbWriter, writer.Name)

	return dbWriter
}

func ClickHouseTestConnection(db *ch.Conn, dataBaseName string) {
	v, err := (*db).ServerVersion()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to ping clickhouse database %s: %w", dataBaseName, err), "", 0)
	}
	log.Debugf("connected to clickhouse database %s with version %s", dataBaseName, v)
}

func DumpToClickhouse(data interface{}, table string) error {
	start := time.Now()
	columns, err := ConvertToColumnar(data)
	if err != nil {
		return err
	}
	log.Debugf("converted to columnar in %s", time.Since(start))
	start = time.Now()
	// abort after 3 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	batch, err := ClickHouseNativeWriter.PrepareBatch(ctx, `INSERT INTO `+table)
	if err != nil {
		return err
	}
	log.Debugf("prepared batch in %s", time.Since(start))
	start = time.Now()
	defer func() {
		if batch.IsSent() {
			return
		}
		err := batch.Abort()
		if err != nil {
			log.Warnf("failed to abort batch: %v", err)
		}
	}()
	for c := 0; c < len(columns); c++ {
		// type assert to correct type
		log.Debugf("appending column %d", c)
		switch columns[c].(type) {
		case []int64:
			err = batch.Column(c).Append(columns[c].([]int64))
		case []uint64:
			err = batch.Column(c).Append(columns[c].([]uint64))
		case []time.Time:
			// appending unix timestamps as int64 to a DateTime column is actually faster than appending time.Time directly
			// tho with how many columns we have it doesn't really matter
			err = batch.Column(c).Append(columns[c].([]time.Time))
		case []float64:
			err = batch.Column(c).Append(columns[c].([]float64))
		case []bool:
			err = batch.Column(c).Append(columns[c].([]bool))
		default:
			// warning: slow path. works but try to avoid this
			cType := reflect.TypeOf(columns[c])
			log.Warnf("fallback: column %d of type %s is not natively supported, falling back to reflection", c, cType)
			startSlow := time.Now()
			cValue := reflect.ValueOf(columns[c])
			length := cValue.Len()
			cSlice := reflect.MakeSlice(reflect.SliceOf(cType.Elem()), length, length)
			for i := 0; i < length; i++ {
				cSlice.Index(i).Set(cValue.Index(i))
			}
			err = batch.Column(c).Append(cSlice.Interface())
			log.Debugf("fallback: appended column %d in %s", c, time.Since(startSlow))
		}
		if err != nil {
			return err
		}
	}
	log.Debugf("appended all columns to batch in %s", time.Since(start))
	start = time.Now()
	err = batch.Send()
	if err != nil {
		return err
	}
	log.Debugf("sent batch in %s", time.Since(start))
	return nil

}

// ConvertToColumnar efficiently converts a slice of any struct type to a slice of slices, each representing a column.
func ConvertToColumnar(data interface{}) ([]interface{}, error) {
	start := time.Now()
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("provided data is not a slice")
	}

	if v.Len() == 0 {
		return nil, fmt.Errorf("slice is empty")
	}

	elemType := v.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("slice elements are not structs")
	}

	numFields := elemType.NumField()
	columns := make([]interface{}, numFields)
	colValues := make([]reflect.Value, numFields)

	for i := 0; i < numFields; i++ {
		fieldType := elemType.Field(i).Type
		colSlice := reflect.MakeSlice(reflect.SliceOf(fieldType), v.Len(), v.Len())
		x := reflect.New(colSlice.Type())
		x.Elem().Set(colSlice)
		columns[i] = colSlice
		colValues[i] = colSlice.Slice(0, v.Len())
	}

	var wg sync.WaitGroup
	wg.Add(numFields)

	for j := 0; j < numFields; j++ {
		go func(j int) {
			defer wg.Done()
			for i := 0; i < v.Len(); i++ {
				structValue := v.Index(i)
				colValues[j].Index(i).Set(structValue.Field(j))
			}
		}(j)
	}
	wg.Wait()

	for i, col := range colValues {
		columns[i] = col.Interface()
	}
	log.Infof("columnarized %d rows with %d columns in %s", v.Len(), numFields, time.Since(start))
	return columns, nil
}

