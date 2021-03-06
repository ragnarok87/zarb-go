package execution

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zarbchain/zarb-go/account"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/logger"
	"github.com/zarbchain/zarb-go/sandbox"
	"github.com/zarbchain/zarb-go/sortition"
	"github.com/zarbchain/zarb-go/tx"
	"github.com/zarbchain/zarb-go/validator"
)

var tExec *Execution
var tVal1 *validator.Validator
var tAddr1, tAddr2 crypto.Address
var tPriv1, tPriv2 crypto.PrivateKey
var tPub1, tPub2 crypto.PublicKey
var tSandbox *sandbox.MockSandbox
var tTotalCoin int64

func setup(t *testing.T) {
	loggerConfig := logger.TestConfig()
	logger.InitLogger(loggerConfig)

	tSandbox = sandbox.NewMockSandbox()

	acc1, priv1 := account.GenerateTestAccount(0)
	tPriv1 = priv1
	tPub1 = tPriv1.PublicKey()
	tAddr1 = tPub1.Address()
	acc1.SubtractFromBalance(acc1.Balance()) // make balance zero
	acc1.AddToBalance(3000)
	tSandbox.UpdateAccount(acc1)

	acc2, priv2 := account.GenerateTestAccount(1)
	tPriv2 = priv2
	tPub2 = tPriv2.PublicKey()
	tAddr2 = tPub2.Address()
	acc2.SubtractFromBalance(acc2.Balance()) // make balance zero
	acc2.AddToBalance(10000000000000000)
	tSandbox.UpdateAccount(acc2)

	tVal1 = validator.NewValidator(tPub1, 0, 0)
	tSandbox.UpdateValidator(tVal1)

	tExec = NewExecution(tSandbox)
	tTotalCoin = 10000000000000000 + 3000
}

func checkTotalCoin(t *testing.T) {
	total := int64(0)
	for _, acc := range tSandbox.Accounts {
		total += acc.Balance()
	}
	for _, val := range tSandbox.Validators {
		total += val.Stake()
	}
	assert.Equal(t, total+tExec.accumulatedFee, tTotalCoin)
}

func TestExecuteSendTx(t *testing.T) {
	setup(t)

	rcvAddr, recPub, rcvPriv := crypto.GenerateTestKeyPair()
	stamp := crypto.GenerateTestHash()
	tSandbox.AppendStampAndUpdateHeight(100, stamp)

	trx1 := tx.NewSendTx(stamp, 1, rcvAddr, rcvAddr, 100, 1000, "invalid sender", &recPub, nil)
	trx1.SetSignature(rcvPriv.Sign(trx1.SignBytes()))
	assert.Error(t, tExec.Execute(trx1))

	trx2 := tx.NewSendTx(stamp, tSandbox.AccSeq(tAddr1)+2, tAddr1, rcvAddr, 1000, 1000, "invalid sequence", &tPub1, nil)
	trx2.SetSignature(tPriv1.Sign(trx2.SignBytes()))
	assert.Error(t, tExec.Execute(trx2))

	trx3 := tx.NewSendTx(stamp, tSandbox.AccSeq(tAddr1)+1, tAddr1, rcvAddr, 2001, 1000, "insufficient balance", &tPub1, nil)
	trx3.SetSignature(tPriv1.Sign(trx3.SignBytes()))
	assert.Error(t, tExec.Execute(trx3))

	trx4 := tx.NewSendTx(stamp, tSandbox.AccSeq(tAddr1)+1, tAddr1, rcvAddr, 1000, 999, "invalid fee", &tPub1, nil)
	trx4.SetSignature(tPriv1.Sign(trx4.SignBytes()))
	assert.Error(t, tExec.Execute(trx4))

	trx5 := tx.NewSendTx(stamp, tSandbox.AccSeq(tAddr1)+1, tAddr1, rcvAddr, 1000, 1000, "ok", &tPub1, nil)
	trx5.SetSignature(tPriv1.Sign(trx5.SignBytes()))
	assert.NoError(t, tExec.Execute(trx5))
	assert.Equal(t, tSandbox.Account(tAddr1).Balance(), int64(1000))
	assert.Equal(t, tSandbox.Account(rcvAddr).Balance(), int64(1000))

	// Duplicated. Invalid sequence
	assert.Error(t, tExec.Execute(trx5))

	trx6 := tx.NewSendTx(stamp, tSandbox.AccSeq(tAddr1)+1, tAddr1, rcvAddr, 1, 1000, "insufficient balance", &tPub1, nil)
	trx6.SetSignature(tPriv1.Sign(trx6.SignBytes()))
	assert.Error(t, tExec.Execute(trx6))

	trx7 := tx.NewSendTx(stamp, tSandbox.AccSeq(tAddr2)+1, tAddr2, rcvAddr, 5000000, 5000, "ok", &tPub2, nil)
	trx7.SetSignature(tPriv2.Sign(trx7.SignBytes()))
	assert.NoError(t, tExec.Execute(trx7))
	assert.Equal(t, tExec.AccumulatedFee(), int64(6000))

	checkTotalCoin(t)
}

func TestExecuteBondTx(t *testing.T) {
	setup(t)

	valAddr, valPub, valPriv := crypto.GenerateTestKeyPair()
	stamp := crypto.GenerateTestHash()
	tSandbox.AppendStampAndUpdateHeight(100, stamp)

	trx1 := tx.NewBondTx(stamp, 1, valAddr, valPub, 1000, "invalid boner", &valPub, nil)
	trx1.SetSignature(valPriv.Sign(trx1.SignBytes()))
	assert.Error(t, tExec.Execute(trx1))

	trx2 := tx.NewBondTx(stamp, tSandbox.AccSeq(tAddr1)+2, tAddr1, valPub, 1000, "invalid sequence", &tPub1, nil)
	trx2.SetSignature(tPriv1.Sign(trx2.SignBytes()))
	assert.Error(t, tExec.Execute(trx2))

	trx3 := tx.NewBondTx(stamp, tSandbox.AccSeq(tAddr1)+1, tAddr1, valPub, 3001, "insufficient balance", &tPub1, nil)
	trx3.SetSignature(tPriv1.Sign(trx3.SignBytes()))
	assert.Error(t, tExec.Execute(trx3))

	trx4 := tx.NewBondTx(stamp, tSandbox.AccSeq(tAddr1)+1, tAddr1, valPub, 1000, "ok", &tPub1, nil)
	trx4.SetSignature(tPriv1.Sign(trx4.SignBytes()))
	assert.NoError(t, tExec.Execute(trx4))

	// Duplicated. Invalid sequence
	assert.Error(t, tExec.Execute(trx4))

	assert.Equal(t, tSandbox.Account(tAddr1).Balance(), int64(2000))
	assert.Equal(t, tSandbox.Validator(valAddr).Stake(), int64(1000))
	assert.Equal(t, tExec.AccumulatedFee(), int64(0))

	checkTotalCoin(t)
}

func TestExecuteSortitionTx(t *testing.T) {
	setup(t)

	valAddr, valPub, valPriv := crypto.GenerateTestKeyPair()
	stamp := crypto.GenerateTestHash()
	tSandbox.AppendStampAndUpdateHeight(100, stamp)
	proof := [48]byte{}

	trx1 := tx.NewSortitionTx(stamp, 1, valAddr, proof[:], "invalid validator", &valPub, nil)
	trx1.SetSignature(valPriv.Sign(trx1.SignBytes()))
	assert.Error(t, tExec.Execute(trx1))

	val := validator.NewValidator(valPub, 0, 0)
	tSandbox.UpdateValidator(val)

	trx2 := tx.NewSortitionTx(stamp, 1, valAddr, proof[:], "invalid proof", &valPub, nil)
	trx2.SetSignature(valPriv.Sign(trx2.SignBytes()))
	assert.Error(t, tExec.Execute(trx2))

	sortition := sortition.NewSortition(crypto.NewSigner(valPriv))
	trx3 := sortition.EvaluateTransaction(stamp, val)
	assert.NotNil(t, trx3)
	assert.NoError(t, tExec.Execute(trx3))

	assert.Equal(t, tExec.AccumulatedFee(), int64(0))

	checkTotalCoin(t)
}

func TestSendToSelf(t *testing.T) {
	setup(t)

	stamp := crypto.GenerateTestHash()
	tSandbox.AppendStampAndUpdateHeight(100, stamp)

	seq := tSandbox.AccSeq(tAddr1)
	trx := tx.NewSendTx(stamp, seq+1, tAddr1, tAddr1, 1000, 1000, "ok", &tPub1, nil)
	trx.SetSignature(tPriv1.Sign(trx.SignBytes()))
	assert.NoError(t, tExec.Execute(trx))

	assert.Equal(t, tSandbox.Account(tAddr1).Balance(), int64(2000)) // Crazy guy just want to pay fee!

	acc := tSandbox.Account(tAddr1)
	assert.Equal(t, acc.Sequence(), seq+1)
}
