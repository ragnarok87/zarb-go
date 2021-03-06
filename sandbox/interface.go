package sandbox

import (
	"github.com/zarbchain/zarb-go/account"
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/validator"
)

type Sandbox interface {
	Account(crypto.Address) *account.Account
	MakeNewAccount(crypto.Address) *account.Account
	UpdateAccount(*account.Account)

	Validator(crypto.Address) *validator.Validator
	MakeNewValidator(crypto.PublicKey) *validator.Validator
	UpdateValidator(*validator.Validator)

	VerifySortition(blockHash crypto.Hash, proof []byte, val *validator.Validator) bool
	AddToSet(crypto.Hash, crypto.Address) error

	CurrentHeight() int
	RecentBlockHeight(crypto.Hash) int
	TransactionToLiveInterval() int
	MaxMemoLength() int
	FeeFraction() float64
	MinFee() int64
}
