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
	rowKeyFilter(prefix string) string
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

func (f queryFilter) rowKeyFilter(prefix string) string {
	return ""
}

// toSuccessor add suffix ";" has it comes after ":" in the ascii order
// this is a simple way to have an infinite bound limit
// prefix must be a real prefix and not a key
func toSuccessor(prefix string) string {
	return prefix + ";"
}

type queryFilterV3 struct {
	query    string
	timeFrom *timestamp.Timestamp
	timeTo   *timestamp.Timestamp
}

func newQueryFilterV3(options options) (*queryFilterV3, error) {
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
	time, side, with, chainID, interactionType, extra := "*", "*", "*", "*", "*", "*"
	if options.onlyReceived {
		side = "in"
	}
	if options.onlySent {
		side = "out"
	}
	if options.with != nil {
		with = toHex(options.with.Bytes())
	}
	if options.chainID != nil {
		chainID = *options.chainID
	}
	if options.onlyTxs {
		interactionType = "TX"
	}
	if options.onlyTransfers {
		interactionType = "ERC20"
	}
	if options.method != nil {
		extra = *options.method
	}
	if options.asset != nil {
		extra = toHex(options.asset.Bytes())
	}
	if options.to != nil {
		time = reversePaddedTimestamp(options.to)
	}
	replacer := strings.NewReplacer(
		"<side>", side,
		"<with>", with,
		"<chainID>", chainID,
		"<type>", interactionType,
		"<extra_type>", extra,
		"<time>", time,
		"<id>", "*",
	)

	return &queryFilterV3{
		query:    replacer.Replace(base),
		timeFrom: options.from,
		timeTo:   options.to,
	}, nil
}

func (f queryFilterV3) get(address common.Address) string {
	query := strings.Replace(f.query, "<address>", toHex(address.Bytes()), 1)
	if f.timeTo != nil {
		query = strings.Replace(query, "<time>", reversePaddedTimestamp(f.timeTo), 1)
	}
	return query
}

func (f queryFilterV3) limit(prefix string) string {
	if f.timeFrom != nil {
		parts := strings.Split(prefix, ":")
		parts[1] = reversePaddedTimestamp(f.timeFrom) + ";"
		prefix2 := strings.Join(parts, ":")
		prefix2 = strings.Replace(prefix2, ";:", ";", 1)
		return prefix2
	}
	/*parts := strings.Split(prefix, ":")
	hasFilter := false
	for i := len(parts) - 1; i != 0 && !hasFilter; i-- {
		if parts[i] != "*" {
			hasFilter = true
			hasRune := []rune(parts[i])
			hasRune[len(hasRune)-1] = hasRune[len(hasRune)-1] + 1
			parts[i] = string(hasRune)
		}
	}
	if hasFilter {
		return strings.Join(parts, ":")
	}*/

	return toSuccessor(strings.Split(prefix, ":")[0]) //strings.Replace(prefix, ":", ";", 1)
}

func (f queryFilterV3) rowKeyFilter(prefix string) string {
	parts := strings.Split(prefix, ":")
	parts[1] = "*"
	prefix = strings.ReplaceAll(strings.Join(parts, ":"), "*", ".*")
	return fmt.Sprintf("^(%s)", prefix)
}
