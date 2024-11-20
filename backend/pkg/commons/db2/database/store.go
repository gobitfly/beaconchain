package database

type Database interface {
	Add(key, column string, data []byte, allowDuplicate bool) error
	BulkAdd(itemsByKey map[string][]Item) error
	Read(prefix string) ([][]byte, error)
	GetRow(key string) (map[string][]byte, error)
	GetRowKeys(prefix string) ([]string, error)
	GetLatestValue(key string) ([]byte, error)
	GetRowsRange(high, low string) ([]Row, error)
	Close() error
	Clear() error
}

var (
	_ Database = (*TableWrapper)(nil)
	_ Database = (*RemoteClient)(nil)
)
