package dione

import (
	ctx "context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type discoveryNotifier struct {
	PeerChan chan peer.AddrInfo
}

func (notifier *discoveryNotifier) HandlePeerFound(pi peer.AddrInfo) {
	notifier.PeerChan <- pi
}

func initMdns(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	n := &discoveryNotifier{}
	n.PeerChan = make(chan peer.AddrInfo)

	ser := mdns.NewMdnsService(peerhost, rendezvous, n)
	if err := ser.Start(); err != nil {
		panic(err)
	}

	return n.PeerChan
}

func connectMdns(h host.Host, dht *dht.IpfsDHT, channel chan peer.AddrInfo) {
	for info := range channel {
		if info.ID == h.ID() {
			continue
		}
		h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
		fmt.Printf("Connecting to %v\n", info)
		err := h.Connect(ctx.TODO(), info)
		if err != nil {
			panic(err)
		}
		fmt.Println("Added to peerstore")
		err = dht.Ping(ctx.TODO(), info.ID)
		if err != nil {
			panic(err)
		}
	}
}
