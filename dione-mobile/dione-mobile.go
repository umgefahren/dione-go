package dione_mobile

import (
	"fmt"
	"github.com/Dione-Software/dione-go"
	"github.com/libp2p/go-libp2p-core/peer"
)

type DioneMobileHost struct {
	internal dione.DioneHost
}

type DioneMobileTunnel struct {
	internal dione.DioneTunnel
}

func (d *DioneMobileTunnel) Closest(key string) string {
	ids := d.internal.Closest(key)
	retString := ""
	for _, id := range ids {
		retString += peer.Encode(id)
		retString += "||"
	}
	return retString
}

func (d *DioneMobileTunnel) PutKad(key string, value []byte) {
	d.internal.Put(key, value)
}

func (d *DioneMobileTunnel) GetKad(key string) []byte {
	return d.internal.Get(key)
}

func (d *DioneMobileTunnel) Connect(p string) DioneMobileInterface {
	targetPeer, err := peer.Decode(p)
	if err != nil {
		panic(err)
	}
	tun := d.internal.Connect(targetPeer)
	mobileTun := new(DioneMobileTunnel)
	mobileTun.internal = *tun
	return mobileTun
}

type DioneMobileInterface interface {
	Closest(key string) string
	PutKad(key string, value []byte)
	GetKad(key string) []byte
	Connect(peer string) DioneMobileInterface
}

func New() *DioneMobileHost {
	ret := new(DioneMobileHost)
	dioneHost := dione.NewDioneHost(0)
	ret.internal = dioneHost
	return ret
}

func (dmh *DioneMobileHost) Closest(key string) string {
	ids := dmh.internal.Closest(key)
	retString := ""
	for _, id := range ids {
		retString += peer.Encode(id)
		retString += "||"
	}
	return retString
}

func (dmh *DioneMobileHost) PutKad(key string, value []byte) {
	dmh.internal.Put(key, value)
}

func (dmh *DioneMobileHost) GetKad(key string) []byte {
	return dmh.internal.Get(key)
}

func (dmh *DioneMobileHost) Connect(p string) DioneMobileInterface {
	targetPeer, err := peer.Decode(p)
	if err != nil {
		panic(err)
	}
	tun := dmh.internal.Connect(targetPeer)
	mobileTunnel := new(DioneMobileTunnel)
	mobileTunnel.internal = *tun
	return mobileTunnel
}

func (dmh *DioneMobileHost) Peers() string {
	peers := dmh.internal.GetPeers()
	retstring := ""
	for _, p := range peers {
		retstring += peer.Encode(p)
		retstring += "||"
	}
	return retstring
}

func Greetings(name string) string {
	return fmt.Sprintf("Hello, %v!", name)
}
