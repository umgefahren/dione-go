package main

import (
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
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
