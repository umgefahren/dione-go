package main

import (
	ctx "context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
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
	)
	if err != nil {
		panic(err)
	}
	return ret
}
