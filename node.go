package main

import (
	ctx "context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/manifoldco/promptui"
)

func validate(input string) error {
	switch input {
	case
		"info",
		"closest",
		"put",
		"get",
		"put-provider",
		"get-provider",
		"refresh-rt":
		return nil
	}
	return errors.New("Invalid input")
}

func handleInput(h host.Host, dht *dht.IpfsDHT, input string) {
	switch input {
	case "info":
		fmt.Printf("Current ID is %v\n", h.ID())
		fmt.Printf("Current Network is %v\n", h.Network())
	case "closest":
		prompt := promptui.Prompt{
			Label: "Key",
		}
		result, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Chose key \"%v\"\n", result)
		dht.RoutingTable().Print()
		closestPeers, err := dht.GetClosestPeers(ctx.TODO(), result)
		for _, closestPeer := range closestPeers {
			fmt.Printf("Closest peer %v\n", closestPeer)
		}
	case "refresh-rt":
		fmt.Printf("Current routing table")
		dht.RoutingTable().Print()
		fmt.Printf("Refreshing routing table\n")
		dht.RefreshRoutingTable()
		fmt.Printf("New routing table \n")
		dht.RoutingTable().Print()
	case "put":
		prompt := promptui.Prompt{
			Label: "Key",
		}
		keyraw, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		key := fmt.Sprintf("/kad-m/%v", keyraw)
		fmt.Printf("Key %v\n", key)
		prompt = promptui.Prompt{
			Label: "Value",
		}
		value, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Value %v\n", value)
		valueData := []byte(value)
		rec := record.MakePutRecord(key, valueData)
		data, err := rec.Marshal()
		if err != nil {
			panic(err)
		}
		err = dht.PutValue(ctx.TODO(), key, data)

		if err != nil {
			panic(err)
		}
	case "get":
		prompt := promptui.Prompt{
			Label: "Key",
		}
		keyraw, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		key := fmt.Sprintf("/kad-m/%v", keyraw)
		fmt.Printf("Key %v\n", key)
		valueData, err := dht.GetValue(ctx.TODO(), key)
		valueString := string(valueData)
		fmt.Printf("Got value %#v for key %v\n", valueString, key)
	case "put-provider":
		prompt := promptui.Prompt{
			Label: "Key",
		}
		keyRaw, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		key := fmt.Sprintf("%v", keyRaw)
		fmt.Printf("Key %v\n", key)
		c, err := cid.Decode(key)
		if err != nil {
			panic(err)
		}
		err = dht.Provide(ctx.TODO(), c, true)
		if err != nil {
			panic(err)
		}
	case "get-provider":
		prompt := promptui.Prompt{
			Label: "Key",
		}
		keyRaw, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		key := fmt.Sprintf("%v", keyRaw)
		fmt.Printf("Key %v\n", key)
		c, err := cid.Decode(key)
		if err != nil {
			panic(err)
		}
		peers, err := dht.FindProviders(ctx.TODO(), c)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Found peers %v\n", peers)
	}

}

func main() {
	h, DHT := NewHost(0)
	defer func(h host.Host) {
		fmt.Println("Closing")
		err := h.Close()
		if err != nil {
			panic(err)
		}
	}(h)

	hostId := h.ID()
	fmt.Printf("We have a new host with id %v\n", hostId)
	address := h.Addrs()
	fmt.Printf("We have the addresses %v\n", address)
	fmt.Printf("DHT mode %v\n", DHT.Mode())

	mdnsFound := initMdns(h, "dione")
	go connectMdns(h, DHT, mdnsFound)
	prompt := promptui.Prompt{
		Label:    "",
		Validate: validate,
	}

	for true {
		result, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		fmt.Printf("You chose %v\n", result)
		handleInput(h, DHT, result)
	}
}
