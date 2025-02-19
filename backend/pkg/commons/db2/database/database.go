package database

type Database interface {
	BulkAdd(itemsByKey map[string][]Item, opts ...Option) error
	Read(prefix string) ([]Row, error)
	GetRow(key string) (*Row, error)
	GetRowsWithKeys(keys []string) ([]Row, error)
	GetRowsRange(high, low string, opts ...Option) ([]Row, error)

	Close() error
	Clear() error
}

var (
	_ Database = (*TableWrapper)(nil)
	_ Database = (*RemoteClient)(nil)
)
