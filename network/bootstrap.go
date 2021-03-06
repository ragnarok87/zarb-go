package network

import (
	"context"
	"math/rand"
	"sync"
	"time"

	host "github.com/libp2p/go-libp2p-core/host"
	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/zarbchain/zarb-go/logger"
)

// Bootstrapper attempts to keep the p2p host connected to the network
// by keeping a minimum threshold of connections. If the threshold isn't met it
// connects to a random subset of the bootstrap peers. It does not use peer routing
// to discover new peers. To stop a Bootstrapper cancel the context passed in Start()
// or call Stop().
type Bootstrapper struct {
	// Config
	// MinPeerThreshold is the number of connections it attempts to maintain.
	minPeerThreshold int
	// Peers to connect to if we fall below the threshold.
	bootstrapPeers []peer.AddrInfo
	// Period is the interval at which it periodically checks to see
	// if the threshold is maintained.
	period time.Duration
	// ConnectionTimeout is how long to wait before timing out a connection attempt.
	connectionTimeout time.Duration

	// Dependencies
	h host.Host
	d inet.Dialer
	r routing.Routing
	// Does the work. Usually Bootstrapper.bootstrap. Argument is a slice of
	// currently-connected peers (so it won't attempt to reconnect).
	Bootstrap func([]peer.ID)

	// Bookkeeping
	ticker         *time.Ticker
	ctx            context.Context
	cancel         context.CancelFunc
	dhtBootStarted bool
}

// NewBootstrapper returns a new Bootstrapper that will attempt to keep connected
// to the network by connecting to the given bootstrap peers.
func NewBootstrapper(ctx context.Context, bootstrapPeers []peer.AddrInfo, h host.Host, d inet.Dialer, r routing.Routing, minPeer int, period time.Duration) *Bootstrapper {
	b := &Bootstrapper{
		minPeerThreshold:  minPeer,
		bootstrapPeers:    bootstrapPeers,
		period:            period,
		connectionTimeout: 20 * time.Second,
		ctx:               ctx,

		h: h,
		d: d,
		r: r,
	}
	b.Bootstrap = b.bootstrap

	b.Bootstrap(b.d.Peers())

	return b
}

// Start starts the Bootstrapper bootstrapping. Cancel `ctx` or call Stop() to stop it.
func (b *Bootstrapper) Start() {
	b.ctx, b.cancel = context.WithCancel(b.ctx)
	b.ticker = time.NewTicker(b.period)

	go func() {
		defer b.ticker.Stop()

		for {
			select {
			case <-b.ctx.Done():
				return
			case <-b.ticker.C:
				b.Bootstrap(b.d.Peers())
			}
		}
	}()
}

// Stop stops the Bootstrapper.
func (b *Bootstrapper) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

// bootstrap does the actual work. If the number of connected peers
// has fallen below b.MinPeerThreshold it will attempt to connect to
// a random subset of its bootstrap peers.
func (b *Bootstrapper) bootstrap(currentPeers []peer.ID) {
	peersNeeded := b.minPeerThreshold - len(currentPeers)
	if peersNeeded < 1 {
		return
	}

	ctx, cancel := context.WithTimeout(b.ctx, b.connectionTimeout)
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		// After connecting to bootstrap peers, bootstrap the DHT.
		// DHT Bootstrap is a persistent process so only do this once.
		if !b.dhtBootStarted {
			b.dhtBootStarted = true
			err := b.bootstrapIpfsRouting()
			if err != nil {
				logger.Warn("got error trying to bootstrap Routing. Peer discovery may suffer.", "err", err)
			}
		}
		cancel()
	}()

	peersAttempted := 0
	for _, i := range rand.Perm(len(b.bootstrapPeers)) {
		pinfo := b.bootstrapPeers[i]
		// Don't try to connect to an already connected peer.
		if hasPID(currentPeers, pinfo.ID) {
			continue
		}

		wg.Add(1)
		go func() {
			if err := b.h.Connect(ctx, pinfo); err != nil {
				logger.Error("got error trying to connect to bootstrap node ", "info", pinfo, "err", err.Error())
			}
			wg.Done()
		}()
		peersAttempted++
		if peersAttempted == peersNeeded {
			return
		}
	}
	logger.Warn("not enough bootstrap nodes to maintain connections", "threshold", b.minPeerThreshold, "current", len(currentPeers))
}

func hasPID(pids []peer.ID, pid peer.ID) bool {
	for _, p := range pids {
		if p == pid {
			return true
		}
	}
	return false
}

func (b *Bootstrapper) bootstrapIpfsRouting() error {
	dht, ok := b.r.(*dht.IpfsDHT)
	if !ok {
		// No bootstrapping to do exit quietly.
		return nil
	}

	return dht.Bootstrap(b.ctx)
}
