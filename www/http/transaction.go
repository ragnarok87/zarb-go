package http

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zarbchain/zarb-go/tx"
	"github.com/zarbchain/zarb-go/www/capnp"
)

func (s *Server) GetTransactionHandler(w http.ResponseWriter, r *http.Request) {
	b := s.server.GetTransaction(s.ctx, func(p capnp.ZarbServer_getTransaction_Params) error {
		vars := mux.Vars(r)
		if err := p.SetHash([]byte(vars["hash"])); err != nil {
			return err
		}
		return nil
	})

	t, err := b.Struct()
	if err != nil {
		s.writeError(w, err)
		return
	}

	res, _ := t.Result()
	trxData, _ := res.Data()
	fmt.Printf("%x\n", trxData)
	trx := new(tx.Tx)
	err = trx.Decode(trxData)
	if err != nil {
		s.writeError(w, err)
		return
	}

	rec, _ := res.Receipt()
	receiptData, _ := rec.Data()
	receipt := new(tx.Receipt)
	err = receipt.Decode(receiptData)
	if err != nil {
		s.writeError(w, err)
		return
	}

	out := new(TransactionResult)
	out.Hash = trx.ID()
	out.Tx = *trx
	out.Data = hex.EncodeToString(trxData)
	out.Receipt.Hash = receipt.Hash()
	out.Receipt.Data = hex.EncodeToString(receiptData)
	out.Receipt.Receipt = *receipt

	s.writeJSON(w, out)
}
