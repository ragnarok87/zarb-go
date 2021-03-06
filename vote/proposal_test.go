package vote

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zarbchain/zarb-go/crypto"
)

func TestProposalMarshaling(t *testing.T) {
	p1, _ := GenerateTestProposal(10, 10)
	bz1, err := p1.MarshalCBOR()
	assert.NoError(t, err)
	var p2 Proposal
	err = p2.UnmarshalCBOR(bz1)
	assert.NoError(t, err)
	bz2, _ := p2.MarshalCBOR()

	assert.Equal(t, bz1, bz2)
	assert.Equal(t, p1.Hash(), p2.Hash())
}

func TestProposalSignature(t *testing.T) {
	_, pb0, pv0 := crypto.GenerateTestKeyPair()

	p, pv := GenerateTestProposal(5, 5)
	pb := pv.PublicKey()
	assert.NoError(t, p.Verify(pb))

	assert.Error(t, p.Verify(pb0)) // invalid public key

	sig0 := pv0.Sign(p.SignBytes())
	p.SetSignature(sig0)
	assert.Error(t, p.Verify(pb)) // invalid signature
}
