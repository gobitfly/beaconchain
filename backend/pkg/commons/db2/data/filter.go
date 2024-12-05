package data

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type filter interface {
	get(address common.Address) string
	limit(prefix string) string
}

func toHex(b []byte) string {
	return fmt.Sprintf("%x", b)
}

type queryFilter struct {
	query    string
	timeFrom *timestamp.Timestamp
	timeTo   *timestamp.Timestamp
}

func newQueryFilter(options options) (*queryFilter, error) {
	query := []string{"all"}
	params := []string{"<address>"}
	if options.onlySent && options.onlyReceived {
		options.onlySent = false
		options.onlyReceived = false
	}
	if options.asset != nil && options.method != nil {
		return nil, fmt.Errorf("cannot filter by method and by asset together")
	}
	if options.method != nil {
		options.onlyTxs = true
	}
	if options.asset != nil {
		options.onlyTransfers = true
	}
	if options.onlyTxs && options.onlyTransfers {
		options.onlyTxs = false
		options.onlyTransfers = false
	}
	if options.onlyReceived {
		query[0] = "in"
		params[0] = "<address>"
	}
	if options.onlySent {
		query[0] = "out"
		params[0] = "<address>"
	}
	if options.with != nil {
		query = append(query, "with")
		params = append(params, toHex(options.with.Bytes()))
	}
	if options.chainID != nil {
		query = append(query, "chainID")
		params = append(params, *options.chainID)
	}
	if options.onlyTxs {
		query = append(query, "TX")
	}
	if options.onlyTransfers {
		query = append(query, "ERC20")
	}
	if options.method != nil {
		query = append(query, "method")
		params = append(params, *options.method)
	}
	if options.asset != nil {
		query = append(query, "asset")
		params = append(params, toHex(options.asset.Bytes()))
	}
	if options.from == nil || options.to == nil {
		options.from = nil
		options.to = nil
	}

	return &queryFilter{
		query:    strings.Join(append(query, params...), ":"),
		timeFrom: options.from,
		timeTo:   options.to,
	}, nil
}

func (f queryFilter) get(address common.Address) string {
	query := strings.Replace(f.query, "<address>", toHex(address.Bytes()), 1)
	if f.timeTo != nil {
		query = fmt.Sprintf("%s:%s", query, reversePaddedTimestamp(f.timeTo))
	}
	return query
}

func (f queryFilter) limit(prefix string) string {
	if f.timeFrom != nil {
		index := strings.LastIndex(prefix, ":")
		return toSuccessor(fmt.Sprintf("%s:%s", prefix[:index], reversePaddedTimestamp(f.timeFrom)))
	}
	return toSuccessor(prefix)
}

// toSuccessor add suffix ";" has it comes after ":" in the ascii order
// this is a simple way to have an infinite bound limit
// prefix must be a real prefix and not a key
func toSuccessor(prefix string) string {
	return prefix + ";"
}
