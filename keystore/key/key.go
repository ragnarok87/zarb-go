package key

import (
	"fmt"

	"github.com/zarbchain/zarb-go/crypto"
)

type Key struct {
	data keyData
}

type keyData struct {
	Address    crypto.Address
	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey
}

func GenKey() *Key {
	addr, pk, pv := crypto.RandomKeyPair()
	return &Key{
		data: keyData{
			PrivateKey: pv,
			PublicKey:  pk,
			Address:    addr,
		},
	}
}

func NewKey(addr crypto.Address, pv crypto.PrivateKey) (*Key, error) {
	/// Check if the address is derived from given private key
	if !addr.Verify(pv.PublicKey()) {
		return nil, fmt.Errorf("This address doesn't belong to this privatekey")
	}

	return &Key{
		data: keyData{
			PrivateKey: pv,
			PublicKey:  pv.PublicKey(),
			Address:    addr,
		},
	}, nil
}

func (k *Key) Address() crypto.Address {
	return k.data.Address
}

func (k *Key) PublicKey() crypto.PublicKey {
	return k.data.PublicKey
}

func (k *Key) PrivateKey() crypto.PrivateKey {
	return k.data.PrivateKey
}

func (k *Key) ToSigner() crypto.Signer {
	return crypto.NewSigner(k.data.PrivateKey)
}
