package api

type Address string
type PubKey string
type Hash string // blocks, txs etc.

type Paging struct {
	PrevCursor string `json:"prev_cursor"`
	NextCursor string `json:"next_cursor"`
}
