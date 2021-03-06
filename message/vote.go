package message

import (
	"github.com/zarbchain/zarb-go/errors"
	"github.com/zarbchain/zarb-go/vote"
)

type VotePayload struct {
	Vote *vote.Vote `cbor:"1,keyasint"`
}

func NewVoteMessage(vote *vote.Vote) *Message {
	return &Message{
		Type: PayloadTypeVote,
		Payload: &VotePayload{
			Vote: vote,
		},
	}
}

func (p *VotePayload) SanityCheck() error {
	if err := p.Vote.SanityCheck(); err != nil {
		return errors.Errorf(errors.ErrInvalidMessage, err.Error())
	}
	return nil
}

func (p *VotePayload) Type() PayloadType {
	return PayloadTypeVote
}

func (p *VotePayload) Fingerprint() string {
	return p.Vote.Fingerprint()
}
