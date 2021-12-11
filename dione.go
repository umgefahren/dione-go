package dione

import (
	ctx "context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"io"
)

type DioneHost struct {
	internalHost   host.Host
	internalDht    *dht.IpfsDHT
	internalInputs chan<- GeneralCommand
}

type DioneTunnel struct {
	connection io.ReadWriteCloser
}

type DioneInterface interface {
	Closest(key string) []peer.ID
	Put(key string, value []byte)
	Get(key string) []byte
	Connect(peer peer.ID) *DioneTunnel
}

func NewDioneHost(port int) DioneHost {
	h, DHT := NewHost(port)

	mdnsFound := initMdns(h, "dione")
	go connectMdns(h, DHT, mdnsFound)

	dhtHan, inputs := newDhtHandler(DHT)

	go dhtHan.handle()

	handler := &Handler{}
	handler.h = h
	handler.dhtInput = inputs

	go h.SetStreamHandler(requestId, handler.handleStream)

	result := DioneHost{}
	result.internalHost = h
	result.internalDht = DHT
	result.internalInputs = inputs
	return result
}

func (d DioneHost) GetPeers() []peer.ID {
	return d.internalHost.Network().Peers()
}

func (dh DioneHost) Closest(key string) []peer.ID {

	resp := make(chan []peer.ID, 1)

	command := new(GeneralCommand)
	closestCommand := new(ClosestCommand)
	closestCommand.key = key
	closestCommand.response = resp
	command.closest = closestCommand

	dh.internalInputs <- *command

	closestPeers := <-resp
	return closestPeers
}

func (dh DioneHost) Put(key string, value []byte) {
	generalCommand := new(GeneralCommand)
	put := newPutCommand(key, value)
	generalCommand.put = &put
	dh.internalInputs <- *generalCommand
}

func (dh DioneHost) Get(key string) []byte {
	generalCommand := new(GeneralCommand)
	get, response := newGetCommand(key)
	generalCommand.get = &get
	dh.internalInputs <- *generalCommand
	valueData := <-response
	return valueData
}

func (dh DioneHost) Connect(peer peer.ID) *DioneTunnel {
	stream, err := dh.internalHost.NewStream(ctx.TODO(), peer, requestId)
	if err != nil {
		panic(err)
	}
	tun := newTunnel(stream, peer)

	ret := new(DioneTunnel)
	ret.connection = tun

	return ret
}

func (d *DioneTunnel) Closest(key string) []peer.ID {
	generalRequest := new(GeneralRequest)
	generalClosestRequest := new(GeneralRequest_ClosestProviderRequest)
	closestRequest := new(ClosestProviderRequest)

	closestRequest.Key = key
	generalClosestRequest.ClosestProviderRequest = closestRequest
	generalRequest.GeneralRequestKind = generalClosestRequest

	handler := &Handler{}
	handler.writeMessage(d.connection, generalRequest)

	generalResponse := new(GeneralResponse)
	err := handler.readMessage(d.connection, generalResponse)
	if err != nil {
		panic(err)
	}
	providers := generalResponse.GetClosestProviderResponse().Provider

	peers := make([]peer.ID, 0)
	for _, provider := range providers {
		id, err := peer.Decode(provider)
		if err != nil {
			panic(err)
		}
		peers = append(peers, id)
	}

	return peers
}

func (d *DioneTunnel) Put(key string, value []byte) {
	generalRequest := new(GeneralRequest)
	generalPutRequest := new(GeneralRequest_PutKadRequest)
	putRequest := new(PutKadRequest)
	putRequest.Key = key
	putRequest.Value = value
	generalPutRequest.PutKadRequest = putRequest
	generalRequest.GeneralRequestKind = generalPutRequest

	handle := &Handler{}
	handle.writeMessage(d.connection, generalRequest)

	generalResponse := new(GeneralResponse)
	err := handle.readMessage(d.connection, generalResponse)

	if err != nil {
		panic(err)
	}

	putKadResponse := generalResponse.GetPutKadResponse()
	fmt.Printf("Response put kad %v\n", putKadResponse.Status)
}

func (d *DioneTunnel) Get(key string) []byte {
	generalRequest := new(GeneralRequest)
	generalGetRequest := new(GeneralRequest_GetKadRequest)
	getRequest := new(GetKadRequest)
	getRequest.Key = key
	generalGetRequest.GetKadRequest = getRequest
	generalRequest.GeneralRequestKind = generalGetRequest

	handler := &Handler{}
	handler.writeMessage(d.connection, generalRequest)

	generalResponse := new(GeneralResponse)

	err := handler.readMessage(d.connection, generalResponse)
	if err != nil {
		panic(err)
	}

	getResponse := generalResponse.GetGetKadResponse()

	return getResponse.Value
}

func (d *DioneTunnel) Connect(peer peer.ID) *DioneTunnel {
	tun := newTunnel(d.connection, peer)

	nt := new(DioneTunnel)
	nt.connection = tun
	return nt
}
