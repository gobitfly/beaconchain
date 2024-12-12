package database

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"cloud.google.com/go/bigtable"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrNotFound = fmt.Errorf("not found")

const (
	timeout = time.Minute // Timeout duration for Bigtable operations
)

type Item struct {
	Family string
	Column string
	Data   []byte
}

type Row struct {
	Key    string
	Values map[string][]byte
}

type TableWrapper struct {
	*BigTable
	table string
}

func Wrap(db *BigTable, table string) TableWrapper {
	return TableWrapper{
		BigTable: db,
		table:    table,
	}
}

func (w TableWrapper) Add(key string, item Item, allowDuplicate bool) error {
	return w.BigTable.Add(w.table, key, item, allowDuplicate)
}

func (w TableWrapper) Read(prefix string) ([]Row, error) {
	return w.BigTable.Read(w.table, prefix)
}

func (w TableWrapper) GetLatestValue(key string) (*Row, error) {
	return w.BigTable.GetLatestValue(w.table, key)
}

func (w TableWrapper) GetRow(key string) (*Row, error) {
	return w.BigTable.GetRow(w.table, key)
}

func (w TableWrapper) GetRowKeys(prefix string, opts ...Option) ([]string, error) {
	return w.BigTable.GetRowKeys(w.table, prefix, opts...)
}

func (w TableWrapper) BulkAdd(itemsByKey map[string][]Item, opts ...Option) error {
	return w.BigTable.BulkAdd(w.table, itemsByKey, opts...)
}

func (w TableWrapper) GetRowsRange(high, low string, opts ...Option) ([]Row, error) {
	return w.BigTable.GetRowsRange(w.table, high, low, opts...)
}

func (w TableWrapper) GetRowsWithKeys(keys []string) ([]Row, error) {
	return w.BigTable.GetRowsWithKeys(w.table, keys)
}

// BigTable is a wrapper around Google Cloud Bigtable for storing and retrieving data
type BigTable struct {
	client *bigtable.Client
	admin  *bigtable.AdminClient
}

func NewBigTableWithClient(ctx context.Context, client *bigtable.Client, adminClient *bigtable.AdminClient, tablesAndFamilies map[string][]string) (*BigTable, error) {
	// Initialize the Bigtable table and column family
	if err := initTable(ctx, adminClient, tablesAndFamilies); err != nil {
		return nil, err
	}

	return &BigTable{client: client, admin: adminClient}, nil
}

// NewBigTable initializes a new BigTable
// It returns a BigTable and an error if any part of the setup fails
func NewBigTable(project, instance string, tablesAndFamilies map[string][]string, options ...option.ClientOption) (*BigTable, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create an admin client to manage Bigtable tables
	adminClient, err := bigtable.NewAdminClient(ctx, project, instance, options...)
	if err != nil {
		return nil, fmt.Errorf("could not create admin client: %v", err)
	}

	// Create a Bigtable client for performing data operations
	client, err := bigtable.NewClient(ctx, project, instance, options...)
	if err != nil {
		return nil, fmt.Errorf("could not create data operations client: %v", err)
	}

	return NewBigTableWithClient(ctx, client, adminClient, tablesAndFamilies)
}

// initTable creates the tables and column family in the Bigtable
func initTable(ctx context.Context, adminClient *bigtable.AdminClient, tablesAndFamilies map[string][]string) error {
	for table, families := range tablesAndFamilies {
		if err := createTableAndFamilies(ctx, adminClient, table, families...); err != nil {
			return err
		}
	}
	return nil
}

func createTableAndFamilies(ctx context.Context, admin *bigtable.AdminClient, tableName string, familyNames ...string) error {
	// Get the list of existing tables
	tables, err := admin.Tables(ctx)
	if err != nil {
		return fmt.Errorf("could not fetch table list: %v", err)
	}

	// Create the table if it doesn't exist
	if !slices.Contains(tables, tableName) {
		if err := admin.CreateTable(ctx, tableName); err != nil {
			return fmt.Errorf("could not create table %s: %v", tableName, err)
		}
	}

	// Retrieve information about the table
	tblInfo, err := admin.TableInfo(ctx, tableName)
	if err != nil {
		return fmt.Errorf("could not read info for table %s: %v", tableName, err)
	}

	for _, familyName := range familyNames {
		// Create the column family if it doesn't exist
		if !slices.Contains(tblInfo.Families, familyName) {
			if err := admin.CreateColumnFamily(ctx, tableName, familyName); err != nil {
				return fmt.Errorf("could not create column family %s: %v", familyName, err)
			}
		}
	}
	return nil
}

func (b BigTable) BulkAdd(table string, itemsByKey map[string][]Item, opts ...Option) error {
	options := apply(opts)

	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var keys []string
	var muts []*bigtable.Mutation
	for key, items := range itemsByKey {
		mut := bigtable.NewMutation()
		for _, item := range items {
			mut.Set(item.Family, item.Column, bigtable.Timestamp(0), item.Data)
		}
		keys = append(keys, key)
		muts = append(muts, mut)
	}
	bulk := &bulkMutations{
		Keys: keys,
		Muts: muts,
	}
	sort.Sort(bulk)
	for i := int64(0); i < int64(bulk.Len()); i = i + options.BatchSize {
		from, to := i, i+options.BatchSize
		if to > int64(bulk.Len()) {
			to = int64(bulk.Len())
		}
		errs, err := tbl.ApplyBulk(ctx, bulk.Keys[from:to], bulk.Muts[from:to])
		if err != nil {
			return fmt.Errorf("cannot ApplyBulk err: %w", err)
		}
		var bulkErrs []string
		for _, err := range errs {
			bulkErrs = append(bulkErrs, err.Error())
		}
		if len(bulkErrs) > 0 {
			return fmt.Errorf("cannot BulkAdd errors: %v", bulkErrs)
		}
	}
	return nil
}

// Add inserts a new row with the given key, column, and data into the Bigtable
// It applies a mutation that stores data in the receiver column family
// It returns error if the operation fails
func (b BigTable) Add(table, key string, item Item, allowDuplicate bool) error {
	// Open the transfer table for data operations
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a new mutation to store data in the given column
	mut := bigtable.NewMutation()
	mut.Set(item.Family, item.Column, bigtable.Now(), item.Data)

	if !allowDuplicate {
		mut = bigtable.NewCondMutation(bigtable.RowKeyFilter(key), nil, mut)
	}
	// Apply the mutation to the table using the given key
	if err := tbl.Apply(ctx, key, mut); err != nil {
		return fmt.Errorf("could not apply row mutation: %v", err)
	}
	return nil
}

// Read retrieves all rows from the Bigtable's receiver column family
// It returns the data in the form of a 2D byte slice and an error if the operation fails
func (b BigTable) Read(table, prefix string) ([]Row, error) {
	// Open the transfer table for reading
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var rows []Row
	// Read all rows from the table and collect values from the receiver column family
	err := tbl.ReadRows(ctx, bigtable.PrefixRange(prefix), func(row bigtable.Row) bool {
		values := make(map[string][]byte)
		for _, family := range row {
			for _, item := range family {
				values[item.Column] = item.Value
			}
		}
		rows = append(rows, Row{
			Key:    row.Key(),
			Values: values,
		})
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}

	return rows, nil
}

func (b BigTable) GetLatestValue(table, key string) (*Row, error) {
	// Open the transfer table for reading
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var data Row
	err := tbl.ReadRows(ctx, bigtable.PrefixRange(key), func(row bigtable.Row) bool {
		values := make(map[string][]byte)
		for _, family := range row {
			for _, item := range family {
				values[item.Column] = item.Value
			}
		}
		data = Row{
			Key:    row.Key(),
			Values: values,
		}
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}

	return &data, nil
}

func (b BigTable) GetRow(table, key string) (*Row, error) {
	// Open the transfer table for reading
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var data *Row
	row, err := tbl.ReadRow(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("could not read row: %v", err)
	}
	if row == nil {
		return nil, ErrNotFound
	}
	values := make(map[string][]byte)
	for _, family := range row {
		for _, item := range family {
			values[item.Column] = item.Value
		}
	}
	data = &Row{
		Key:    row.Key(),
		Values: values,
	}

	return data, nil
}

func (b BigTable) GetRowsRange(table, high, low string, opts ...Option) ([]Row, error) {
	options := apply(opts)

	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rowRange := bigtable.NewClosedRange(low, high)
	if options.OpenRange {
		rowRange = bigtable.NewOpenRange(low, high)
	}
	if options.OpenCloseRange {
		rowRange = bigtable.NewOpenClosedRange(low, high)
	}
	if options.ClosedOpenRange {
		rowRange = bigtable.NewClosedOpenRange(low, high)
	}
	readOptions := bigtableReadOptions(options, rowRange)
	if options.RowKeyFilter != "" {
		readOptions = append(readOptions, bigtable.RowFilter(bigtable.RowKeyFilter(options.RowKeyFilter)))
	}
	var data []Row
	err := tbl.ReadRows(ctx, rowRange, func(row bigtable.Row) bool {
		values := make(map[string][]byte)
		for _, family := range row {
			for _, item := range family {
				values[item.Column] = item.Value
			}
		}
		data = append(data, Row{
			Key:    row.Key(),
			Values: values,
		})
		return true
	}, readOptions...)
	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}
	if len(data) == 0 {
		return nil, ErrNotFound
	}

	return data, nil
}

func (b BigTable) GetRowKeys(table, prefix string, opts ...Option) ([]string, error) {
	options := apply(opts)

	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rowRange := bigtable.PrefixRange(prefix)
	readOptions := bigtableReadOptions(options, bigtable.PrefixRange(prefix))
	var data []string
	err := tbl.ReadRows(ctx, rowRange, func(row bigtable.Row) bool {
		data = append(data, row.Key())
		return true
	}, readOptions...)
	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}

	return data, nil
}

func (b BigTable) GetRowsWithKeys(table string, keys []string) ([]Row, error) {
	tbl := b.client.Open(table)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var data []Row
	err := tbl.ReadRows(ctx, bigtable.RowList(keys), func(row bigtable.Row) bool {
		values := make(map[string][]byte)
		for _, family := range row {
			for _, item := range family {
				values[item.Column] = item.Value
			}
		}
		data = append(data, Row{
			Key:    row.Key(),
			Values: values,
		})
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("could not read rows: %v", err)
	}
	if len(data) == 0 {
		return nil, ErrNotFound
	}

	return data, nil
}

func (b BigTable) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tables, err := b.admin.Tables(ctx)
	if err != nil {
		return err
	}
	for _, table := range tables {
		if err := b.admin.DropAllRows(ctx, table); err != nil {
			return fmt.Errorf("could not drop all rows: %v", err)
		}
	}
	return nil
}

// Close shuts down the BigTable by closing the Bigtable client connection
// It returns an error if the operation fails
func (b BigTable) Close() error {
	if b.client == nil {
		return fmt.Errorf("cannot close client: bigtable client is nil")
	}
	if err := b.client.Close(); err != nil && status.Code(err) != codes.Canceled {
		return fmt.Errorf("cannot close client: %v", err)
	}
	if b.admin != nil {
		if err := b.admin.Close(); err != nil && status.Code(err) != codes.Canceled {
			return fmt.Errorf("cannot close admin client: %v", err)
		}
	}
	return nil
}

type bulkMutations struct {
	Keys []string
	Muts []*bigtable.Mutation
}

func (bulkMutations *bulkMutations) Len() int {
	return len(bulkMutations.Keys)
}

func (bulkMutations *bulkMutations) Less(i, j int) bool {
	return bulkMutations.Keys[i] < bulkMutations.Keys[j]
}

func (bulkMutations *bulkMutations) Swap(i, j int) {
	bulkMutations.Keys[i], bulkMutations.Keys[j] = bulkMutations.Keys[j], bulkMutations.Keys[i]
	bulkMutations.Muts[i], bulkMutations.Muts[j] = bulkMutations.Muts[j], bulkMutations.Muts[i]
}

const (
	KeyStatRange        = "range"
	KeyStatRowsSeen     = "rowsSeen"
	KeyStatRowsReturned = "rowsReturned"
	KeyStatEfficiency   = "efficiency"
)

func bigtableReadOptions(options options, rowRange bigtable.RowRange) []bigtable.ReadOption {
	readOptions := []bigtable.ReadOption{bigtable.LimitRows(options.Limit)}
	if options.StatsReporter != nil {
		readOptions = append(readOptions, bigtable.WithFullReadStats(func(stats *bigtable.FullReadStats) {
			efficiency := int64(1)
			if stats.ReadIterationStats.RowsSeenCount != 0 {
				efficiency = stats.ReadIterationStats.RowsReturnedCount / stats.ReadIterationStats.RowsSeenCount
			}
			options.StatsReporter(
				"query stats",
				KeyStatRange, rowRange.String(),
				KeyStatRowsSeen, stats.ReadIterationStats.RowsSeenCount,
				KeyStatRowsReturned, stats.ReadIterationStats.RowsReturnedCount,
				KeyStatEfficiency, efficiency,
			)
		}))
	}
	return readOptions
}
