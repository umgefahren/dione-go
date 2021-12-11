package dione

import (
	ctx "context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	record_pb "github.com/libp2p/go-libp2p-record/pb"
)

type CustomValidator struct{}

func (v CustomValidator) Validate(key string, _ []byte) error {
	fmt.Printf("Validating %v\n", key)
	return nil
}
func (v CustomValidator) Select(_ string, values [][]byte) (int, error) {
	if len(values) == 0 {
		return 0, errors.New("zero values, can't unwrap when only zero values are found")
	} else {
		return 0, nil
	}
}

func newDht(h host.Host) *dht.IpfsDHT {
	ret, err := dht.New(ctx.TODO(), h,
		dht.NamespacedValidator("kad-m", CustomValidator{}),
		dht.ProtocolPrefix("dione"),
		dht.ProtocolExtension("kad"),
		dht.Mode(dht.ModeServer),
	)
	if err != nil {
		panic(err)
	}
	return ret
}

type GeneralCommand struct {
	closest *ClosestCommand
	put     *PutCommand
	get     *GetCommand
}

type ClosestCommand struct {
	key      string
	response chan<- []peer.ID
}

func newClosestCommand(key string) (ClosestCommand, chan []peer.ID) {
	ret := new(ClosestCommand)
	ret.key = key
	response := make(chan []peer.ID, 1)
	ret.response = response
	return *ret, response
}

type PutCommand struct {
	key   string
	value []byte
}

func newPutCommand(key string, value []byte) PutCommand {
	ret := new(PutCommand)
	ret.key = key
	ret.value = value
	return *ret
}

type GetCommand struct {
	key      string
	response chan<- []byte
}

func newGetCommand(key string) (GetCommand, chan []byte) {
	ret := new(GetCommand)
	ret.key = key
	response := make(chan []byte, 1)
	ret.response = response
	return *ret, response
}

type dhtHandler struct {
	d      *dht.IpfsDHT
	inputs <-chan GeneralCommand
}

func newDhtHandler(d *dht.IpfsDHT) (dhtHandler, chan<- GeneralCommand) {
	ret := dhtHandler{}
	ret.d = d
	channel := make(chan GeneralCommand, 10)
	ret.inputs = channel
	return ret, channel
}

func (dhtHand dhtHandler) handle() {
	for command := range dhtHand.inputs {
		if command.closest != nil {
			responseChannel := command.closest.response
			keyString := command.closest.key
			peers, err := dhtHand.d.GetClosestPeers(ctx.TODO(), keyString)
			if err != nil {
				response := make([]peer.ID, 0)
				responseChannel <- response
			} else {
				responseChannel <- peers
			}
		}
		if command.put != nil {
			rawKey := command.put.key
			key := fmt.Sprintf("/kad-m/%v", rawKey)
			rec := record.MakePutRecord(key, command.put.value)
			data, err := rec.Marshal()
			if err != nil {
				panic(err)
			}
			err = dhtHand.d.PutValue(ctx.TODO(), key, data)
			if err != nil {
				panic(err)
			}
		}
		if command.get != nil {
			rawKey := command.get.key
			key := fmt.Sprintf("/kad-m/%v", rawKey)
			rawData, err := dhtHand.d.GetValue(ctx.TODO(), key)
			if err != nil {
				panic(err)
			}
			rec := new(record_pb.Record)
			err = rec.Unmarshal(rawData)
			if err != nil {
				panic(err)
			}
			command.get.response <- rec.GetValue()
		}
	}
}
