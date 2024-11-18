package data

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type formatType string

const (
	typeTx       = formatType("tx")
	typeTransfer = formatType("transfer")
)

type filterType string

const (
	byMethod        = filterType("byMethod")
	bySent          = filterType("bySent")
	byReceived      = filterType("byReceived")
	byAsset         = filterType("byAsset")
	byAssetSent     = filterType("byAssetSent")
	byAssetReceived = filterType("byAssetReceived")
)

type filter interface {
	get(chainID string, address common.Address) string
	limit(prefix string) string
}

type chainFilter interface {
	addByMethod(method string) error
	addBySent() error
	addByReceived() error
	addByAsset(asset common.Address) error
	addTimeRange(from *timestamp.Timestamp, to *timestamp.Timestamp) error
	valid() error
	filterType() filterType
	filter
}

type chainFilterTx struct {
	base     string
	filtered filterType

	method *string
	from   *timestamp.Timestamp
	to     *timestamp.Timestamp
}

func newChainFilterTx() *chainFilterTx {
	return &chainFilterTx{
		base: "<chainID>:I:TX:<address>:TIME",
	}
}

func (c *chainFilterTx) addByMethod(method string) error {
	if c.filtered != "" {
		return fmt.Errorf("filter tx already filtered by %s", c.filtered)
	}
	c.base = "<chainID>:I:TX:<address>:METHOD:<method>"
	c.method = &method
	c.filtered = byMethod
	return nil
}

func (c *chainFilterTx) addBySent() error {
	if c.filtered != "" {
		return fmt.Errorf("filter tx already filtered by %s", c.filtered)
	}
	c.base = "<chainID>:I:TX:<address>:TO"
	c.filtered = bySent
	return nil
}

func (c *chainFilterTx) addByReceived() error {
	if c.filtered != "" {
		return fmt.Errorf("filter tx already filtered by %s", c.filtered)
	}
	c.base = "<chainID>:I:TX:<address>:FROM"
	c.filtered = byReceived
	return nil
}

func (c *chainFilterTx) addByAsset(common.Address) error {
	return fmt.Errorf("cannot filter tx by asset")
}

func (c *chainFilterTx) addTimeRange(from *timestamp.Timestamp, to *timestamp.Timestamp) error {
	if from == nil || to == nil {
		return fmt.Errorf("invalid time range: empty border")
	}
	c.from = from
	c.to = to
	return nil
}

func (c *chainFilterTx) filterType() filterType {
	return c.filtered
}

func (c *chainFilterTx) valid() error {
	return nil
}

func (c *chainFilterTx) get(chainID string, address common.Address) string {
	query := strings.Replace(c.base, "<chainID>", chainID, 1)
	query = strings.Replace(query, "<address>", fmt.Sprintf("%x", address.Bytes()), 1)
	if c.method != nil {
		query = strings.Replace(query, "<method>", *c.method, 1)
	}
	if c.to != nil {
		query = fmt.Sprintf("%s:%s", query, reversePaddedTimestamp(c.to))
	}
	return query
}

func (c *chainFilterTx) limit(prefix string) string {
	if c.from != nil {
		index := strings.LastIndex(prefix, ":")
		return toSuccessor(fmt.Sprintf("%s:%s", prefix[:index], reversePaddedTimestamp(c.from)))
	}
	return toSuccessor(prefix)
}

type chainFilterTransfer struct {
	base     string
	filtered filterType

	asset *common.Address
	from  *timestamp.Timestamp
	to    *timestamp.Timestamp
}

func newChainFilterTransfer() *chainFilterTransfer {
	return &chainFilterTransfer{
		base: "<chainID>:I:ERC20:<address>:TIME",
	}
}

func (c *chainFilterTransfer) addByMethod(string) error {
	return fmt.Errorf("cannot filter transfer by method")
}

func (c *chainFilterTransfer) addBySent() error {
	if c.filtered != "" {
		if c.filtered != byAsset {
			return fmt.Errorf("filter transfer already filtered by %s", c.filtered)
		}
		return c.addByAssetSent()
	}
	c.base = "<chainID>:I:ERC20:<address>:TOKEN_SENT"
	c.filtered = bySent
	return nil
}

func (c *chainFilterTransfer) addByReceived() error {
	if c.filtered != "" {
		if c.filtered != byAsset {
			return fmt.Errorf("filter transfer already filtered by %s", c.filtered)
		}
		return c.addByAssetReceived()
	}
	c.base = "<chainID>:I:ERC20:<address>:TOKEN_RECEIVED"
	c.filtered = byReceived
	return nil
}

func (c *chainFilterTransfer) addByAssetReceived() error {
	c.base = "<chainID>:I:ERC20:<address>:TOKEN_RECEIVED:<asset>"
	c.filtered = byAssetReceived
	return nil
}

func (c *chainFilterTransfer) addByAssetSent() error {
	c.base = "<chainID>:I:ERC20:<address>:TOKEN_SENT:<asset>"
	c.filtered = byAssetSent
	return nil
}

func (c *chainFilterTransfer) addByAsset(asset common.Address) error {
	if c.filtered != "" {
		if c.filtered != byReceived && c.filtered != bySent {
			return fmt.Errorf("filter transfer already filtered by %s", c.filtered)
		}
	}
	c.asset = &asset
	if c.filtered == byReceived {
		return c.addByAssetReceived()
	}
	if c.filtered == bySent {
		return c.addByAssetSent()
	}
	c.base = "<chainID>:I:ERC20:<asset>:<address>:TIME"
	c.filtered = byAsset
	return nil
}

func (c *chainFilterTransfer) addTimeRange(from, to *timestamp.Timestamp) error {
	if from == nil || to == nil {
		return fmt.Errorf("invalid time range: empty border")
	}
	if c.filtered == byReceived || c.filtered == bySent {
		return fmt.Errorf("cannot apply range over filter by %s", c.filtered)
	}
	c.from = from
	c.to = to
	return nil
}

func (c *chainFilterTransfer) filterType() filterType {
	return c.filtered
}

func (c *chainFilterTransfer) valid() error {
	if (c.from != nil || c.to != nil) && (c.filtered == bySent || c.filtered == byReceived) {
		return fmt.Errorf("cannot apply range over filter by %s", c.filtered)
	}
	return nil
}

func (c *chainFilterTransfer) get(chainID string, address common.Address) string {
	query := strings.Replace(c.base, "<chainID>", chainID, 1)
	query = strings.Replace(query, "<address>", fmt.Sprintf("%x", address.Bytes()), 1)
	if c.asset != nil {
		query = strings.Replace(query, "<asset>", fmt.Sprintf("%x", c.asset.Bytes()), 1)
	}
	if c.to != nil {
		query = fmt.Sprintf("%s:%s", query, reversePaddedTimestamp(c.to))
	}
	return query
}

func (c *chainFilterTransfer) limit(prefix string) string {
	if c.from != nil {
		index := strings.LastIndex(prefix, ":")
		return toSuccessor(fmt.Sprintf("%s:%s", prefix[:index], reversePaddedTimestamp(c.from)))
	}
	return toSuccessor(prefix)
}

// toSuccessor add suffix ";" has it comes after ":" in the ascii order
// this is a simple way to have an infinite bound limit
// prefix must be a real prefix and not a key
func toSuccessor(prefix string) string {
	return prefix + ";"
}
