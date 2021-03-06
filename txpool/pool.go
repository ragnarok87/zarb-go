package txpool

import (
	"fmt"
	"time"

	"github.com/zarbchain/zarb-go/execution"

	"github.com/zarbchain/zarb-go/sandbox"

	"github.com/sasha-s/go-deadlock"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/errors"
	"github.com/zarbchain/zarb-go/libs/linkedmap"
	"github.com/zarbchain/zarb-go/logger"
	"github.com/zarbchain/zarb-go/message"
	"github.com/zarbchain/zarb-go/tx"
)

type txPool struct {
	lk deadlock.RWMutex

	config      *Config
	checker     *execution.Execution
	pendings    *linkedmap.LinkedMap
	appendTxCh  chan *tx.Tx
	broadcastCh chan *message.Message
	logger      *logger.Logger
}

func NewTxPool(
	conf *Config,
	broadcastCh chan *message.Message) (TxPool, error) {
	pool := &txPool{
		config:      conf,
		pendings:    linkedmap.NewLinkedMap(conf.MaxSize),
		appendTxCh:  make(chan *tx.Tx, 5),
		broadcastCh: broadcastCh,
	}

	pool.logger = logger.NewLogger("_pool", pool)
	return pool, nil
}

func (pool *txPool) SetSandbox(sb sandbox.Sandbox) {
	pool.checker = execution.NewExecution(sb)
}

func (pool *txPool) AppendTx(trx *tx.Tx) error {
	pool.lk.Lock()
	defer pool.lk.Unlock()

	if err := pool.appendTx(trx); err != nil {
		return err
	}

	pool.appendTxCh <- trx

	return nil
}

func (pool *txPool) AppendTxAndBroadcast(trx *tx.Tx) error {
	pool.lk.Lock()
	defer pool.lk.Unlock()

	if err := pool.appendTx(trx); err != nil {
		return err
	}

	msg := message.NewTxsMessage([]*tx.Tx{trx})
	pool.broadcastCh <- msg

	return nil
}

func (pool *txPool) appendTx(trx *tx.Tx) error {
	if pool.pendings.Has(trx.ID()) {
		return errors.Errorf(errors.ErrInvalidTx, "Transaction is already in pool. id: %v", trx.ID())
	}

	if err := pool.checker.Execute(trx); err != nil {
		pool.logger.Error("Invalid transaction", "tx", trx, "err", err)
		return err
	}

	pool.pendings.PushBack(trx.ID(), trx)

	return nil
}

func (pool *txPool) RemoveTx(id crypto.Hash) {
	pool.lk.Lock()
	defer pool.lk.Unlock()

	pool.pendings.Remove(id)
}

func (pool *txPool) PendingTx(id crypto.Hash) *tx.Tx {
	pool.lk.RLock()

	val, found := pool.pendings.Get(id)
	if found {
		trx := val.(*tx.Tx)
		pool.lk.RUnlock()
		return trx
	}

	pool.logger.Debug("Request transaction from peers", "id", id)
	pool.lk.RUnlock()

	msg := message.NewTxsReqMessage([]crypto.Hash{id})
	pool.broadcastCh <- msg

	timeout := time.NewTimer(pool.config.WaitingTimeout)

	for {
		select {
		case <-timeout.C:
			pool.logger.Warn("Transaction not received", "id", id, "timeout", pool.config.WaitingTimeout)
			return nil
		case trx := <-pool.appendTxCh:
			pool.logger.Debug("Transaction found", "id", id)
			if trx.ID().EqualsTo(id) {
				return trx
			}
		}
	}
}

func (pool *txPool) AllTransactions() []*tx.Tx {
	pool.lk.RLock()
	defer pool.lk.RUnlock()

	trxs := make([]*tx.Tx, 0, pool.pendings.Size())
	for e := pool.pendings.FirstElement(); e != nil; e = e.Next() {
		trx := e.Value.(*linkedmap.Pair).Second.(*tx.Tx)
		trxs = append(trxs, trx)
	}

	return trxs
}

func (pool *txPool) HasTx(id crypto.Hash) bool {
	pool.lk.RLock()
	defer pool.lk.RUnlock()

	return pool.pendings.Has(id)
}

func (pool *txPool) Size() int {
	pool.lk.RLock()
	defer pool.lk.RUnlock()

	return pool.pendings.Size()
}

func (pool *txPool) Fingerprint() string {
	return fmt.Sprintf("{%v}", pool.pendings.Size())
}
