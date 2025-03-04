package auth

type CtxKey string

const CtxUserKey CtxKey = "user"

type Header string

const (
	ApiKeyHeader Header = "apikey"
)
