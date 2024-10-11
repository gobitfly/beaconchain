package types

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/hexutil"
)

type GetBlockTimings struct {
	Headers  time.Duration
	Receipts time.Duration
	Traces   time.Duration
}

type Eth1RpcGetBlockResponse struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		BaseFeePerGas   hexutil.Bytes `json:"baseFeePerGas"`
		Difficulty      hexutil.Bytes `json:"difficulty"`
		ExtraData       hexutil.Bytes `json:"extraData"`
		GasLimit        hexutil.Bytes `json:"gasLimit"`
		GasUsed         hexutil.Bytes `json:"gasUsed"`
		Hash            hexutil.Bytes `json:"hash"`
		LogsBloom       hexutil.Bytes `json:"logsBloom"`
		Miner           hexutil.Bytes `json:"miner"`
		MixHash         hexutil.Bytes `json:"mixHash"`
		Nonce           hexutil.Bytes `json:"nonce"`
		Number          hexutil.Bytes `json:"number"`
		ParentHash      hexutil.Bytes `json:"parentHash"`
		ReceiptsRoot    hexutil.Bytes `json:"receiptsRoot"`
		Sha3Uncles      hexutil.Bytes `json:"sha3Uncles"`
		Size            hexutil.Bytes `json:"size"`
		StateRoot       hexutil.Bytes `json:"stateRoot"`
		Timestamp       hexutil.Bytes `json:"timestamp"`
		TotalDifficulty hexutil.Bytes `json:"totalDifficulty"`
		Transactions    []struct {
			BlockHash            hexutil.Bytes `json:"blockHash"`
			BlockNumber          hexutil.Bytes `json:"blockNumber"`
			From                 hexutil.Bytes `json:"from"`
			Gas                  hexutil.Bytes `json:"gas"`
			GasPrice             hexutil.Bytes `json:"gasPrice"`
			Hash                 hexutil.Bytes `json:"hash"`
			Input                hexutil.Bytes `json:"input"`
			Nonce                hexutil.Bytes `json:"nonce"`
			To                   hexutil.Bytes `json:"to"`
			TransactionIndex     hexutil.Bytes `json:"transactionIndex"`
			Value                hexutil.Bytes `json:"value"`
			Type                 hexutil.Bytes `json:"type"`
			V                    hexutil.Bytes `json:"v"`
			R                    hexutil.Bytes `json:"r"`
			S                    hexutil.Bytes `json:"s"`
			ChainID              hexutil.Bytes `json:"chainId"`
			MaxFeePerGas         hexutil.Bytes `json:"maxFeePerGas"`
			MaxPriorityFeePerGas hexutil.Bytes `json:"maxPriorityFeePerGas"`

			AccessList []struct {
				Address     hexutil.Bytes   `json:"address"`
				StorageKeys []hexutil.Bytes `json:"storageKeys"`
			} `json:"accessList"`

			// Optimism specific fields
			YParity    hexutil.Bytes `json:"yParity"`
			Mint       hexutil.Bytes `json:"mint"`       // The ETH value to mint on L2.
			SourceHash hexutil.Bytes `json:"sourceHash"` // the source-hash, uniquely identifies the origin of the deposit.

			// Arbitrum specific fields
			// Arbitrum Nitro
			RequestId           hexutil.Bytes `json:"requestId"`           // On L1 to L2 transactions, this field is added to indicate position in the Inbox queue
			RefundTo            hexutil.Bytes `json:"refundTo"`            //
			L1BaseFee           hexutil.Bytes `json:"l1BaseFee"`           //
			DepositValue        hexutil.Bytes `json:"depositValue"`        //
			RetryTo             hexutil.Bytes `json:"retryTo"`             // nil means contract creation
			RetryValue          hexutil.Bytes `json:"retryValue"`          // wei amount
			RetryData           hexutil.Bytes `json:"retryData"`           // contract invocation input data
			Beneficiary         hexutil.Bytes `json:"beneficiary"`         //
			MaxSubmissionFee    hexutil.Bytes `json:"maxSubmissionFee"`    //
			TicketId            hexutil.Bytes `json:"ticketId"`            //
			MaxRefund           hexutil.Bytes `json:"maxRefund"`           // the maximum refund sent to RefundTo (the rest goes to From)
			SubmissionFeeRefund hexutil.Bytes `json:"submissionFeeRefund"` // the submission fee to refund if successful (capped by MaxRefund)

			// Arbitrum Classic
			L1SequenceNumber hexutil.Bytes `json:"l1SequenceNumber"`
			ParentRequestId  hexutil.Bytes `json:"parentRequestId"`
			IndexInParent    hexutil.Bytes `json:"indexInParent"`
			ArbType          hexutil.Bytes `json:"arbType"`
			ArbSubType       hexutil.Bytes `json:"arbSubType"`
			L1BlockNumber    hexutil.Bytes `json:"l1BlockNumber"`
		} `json:"transactions"`
		TransactionsRoot hexutil.Bytes `json:"transactionsRoot"`

		Withdrawals []struct {
			Index          hexutil.Bytes `json:"index"`
			ValidatorIndex hexutil.Bytes `json:"validatorIndex"`
			Address        hexutil.Bytes `json:"address"`
			Amount         hexutil.Bytes `json:"amount"`
		} `json:"withdrawals"`
		WithdrawalsRoot hexutil.Bytes `json:"withdrawalsRoot"`

		Uncles []hexutil.Bytes `json:"uncles"`

		// Optimism specific fields

		// Arbitrum specific fields
		L1BlockNumber hexutil.Bytes `json:"l1BlockNumber"` // An approximate L1 block number that occurred before this L2 block.
		SendCount     hexutil.Bytes `json:"sendCount"`     // The number of L2 to L1 messages since Nitro genesis
		SendRoot      hexutil.Bytes `json:"sendRoot"`      // The Merkle root of the outbox tree state
	} `json:"result"`

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
