package http

import (
	"time"

	"github.com/zarbchain/zarb-go/block"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/tx"
)

type BlockResult struct {
	Hash  crypto.Hash
	Time  time.Time
	Data  string
	Block block.Block
}

type ReceiptResult struct {
	Hash    crypto.Hash
	Data    string
	Receipt tx.Receipt
}
type TransactionResult struct {
	Hash    crypto.Hash
	Data    string
	Tx      tx.Tx
	Receipt ReceiptResult
}
