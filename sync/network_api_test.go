package sync

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zarbchain/zarb-go/logger"
	"github.com/zarbchain/zarb-go/message"
)

type mockNetworkAPI struct {
	ch chan *message.Message
}

func mockingNetworkAPI() *mockNetworkAPI {
	return &mockNetworkAPI{
		ch: make(chan *message.Message, 10),
	}
}
func (mock *mockNetworkAPI) Start() error {
	return nil
}
func (mock *mockNetworkAPI) Stop() {
}
func (mock *mockNetworkAPI) PublishMessage(msg *message.Message) error {
	mock.ch <- msg
	return nil
}

func (mock *mockNetworkAPI) waitingForMessage(t *testing.T, msg *message.Message) {
	timeout := time.NewTimer(1 * time.Second)

	for {
		select {
		case <-timeout.C:
			assert.NoError(t, fmt.Errorf("Timeout"))
			return
		case apiMsg := <-mock.ch:
			logger.Info("comparing messages", "apiMsg", apiMsg, "msg", msg)
			b1, _ := msg.MarshalCBOR()
			b2, _ := apiMsg.MarshalCBOR()

			tSync.ParsMessage(b2, tOurID)
			if reflect.DeepEqual(b1, b2) {
				return
			}
		}
	}
}
func (mock *mockNetworkAPI) shouldReceiveMessageWithThisType(t *testing.T, payloadType message.PayloadType) {
	timeout := time.NewTimer(1 * time.Second)

	for {
		select {
		case <-timeout.C:
			assert.NoError(t, fmt.Errorf("Timeout"))
			return
		case apiMsg := <-mock.ch:
			logger.Info("comparing messages", "apiMsg", apiMsg)
			b, _ := apiMsg.MarshalCBOR()

			tSync.ParsMessage(b, tOurID)
			if apiMsg.PayloadType() == payloadType {
				return
			}
		}
	}
}

func (mock *mockNetworkAPI) shouldNotReceiveAnyMessageWithThisType(t *testing.T, payloadType message.PayloadType) {
	timeout := time.NewTimer(1 * time.Second)

	for {
		select {
		case <-timeout.C:
			return
		case apiMsg := <-mock.ch:
			logger.Info("comparing messages", "apiMsg", apiMsg)
			b, _ := apiMsg.MarshalCBOR()

			tSync.ParsMessage(b, tOurID)
			assert.NotEqual(t, apiMsg.PayloadType(), payloadType)
		}
	}
}
