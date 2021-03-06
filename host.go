package dione

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	libp2p_crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libp2p/go-libp2p-kad-dht"
	noise "github.com/libp2p/go-libp2p-noise"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
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

func NewHost(port int) (host.Host, *dht.IpfsDHT) {
	var internalDht *dht.IpfsDHT
	kP, err := newKeyPair()
	if err != nil {
		panic(err)
	}
	tcpString := fmt.Sprintf("/ip4/0.0.0.0/tcp/%v", port)
	quicString := fmt.Sprintf("/ip4/0.0.0.0/udp/%v/quic", port)
	h, err := libp2p.New(
		libp2p.Identity(kP.PrivateKey),
		libp2p.ListenAddrStrings(
			tcpString,
			quicString),
		libp2p.Security(noise.ID, noise.New),

		libp2p.DefaultTransports,

		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			internalDht = newDht(h)
			return internalDht, err
		}),
		libp2p.EnableAutoRelay(),
	)
	if err != nil {
		panic(err)
	}
	routedHost := routedhost.Wrap(h, internalDht)

	return routedHost, internalDht
}
