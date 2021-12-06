package main

import (
	ctx "context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	libp2p_crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libp2p/go-libp2p-kad-dht"
	noise "github.com/libp2p/go-libp2p-noise"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
)

type keyPair struct {
	PrivateKey libp2p_crypto.PrivKey
	PublicKey  libp2p_crypto.PubKey
}

func newKeyPair() (keyPair, error) {
	priv, pub, err := libp2p_crypto.GenerateKeyPair(
		libp2p_crypto.Ed25519,
		256,
	)
	ret := new(keyPair)
	ret.PrivateKey = priv
	ret.PublicKey = pub
	return *ret, err
}

func NewHost(port int) host.Host {
	var internalDht *dht.IpfsDHT
	kP, err := newKeyPair()
	if err != nil {
		panic(err)
	}
	tcpString := fmt.Sprintf("/ip4/0.0.0.0/tcp/%v", port)
	quicString := fmt.Sprintf("/ip4/0.0.0.0/udp/%v/quic", port)
	host, err := libp2p.New(
		libp2p.Identity(kP.PrivateKey),
		libp2p.ListenAddrStrings(
			tcpString,
			quicString),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,

		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			internalDht, err = dht.New(ctx.TODO(), h)
			return internalDht, err
		}),
		libp2p.EnableAutoRelay(),
	)
	if err != nil {
		panic(err)
	}
	return host
}
