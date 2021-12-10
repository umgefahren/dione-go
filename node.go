package main

import (
	ctx "context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/manifoldco/promptui"
	"github.com/multiformats/go-multihash"
)

const requestId = protocol.ID("req")

func validate(input string) error {
	switch input {
	case
		"info",
		"closest",
		"put",
		"get",
		"put-provider",
		"get-provider",
		"connect",
		"refresh-rt":
		return nil
	}
	return errors.New("Invalid input")
}

func handleInput(h host.Host, dht *dht.IpfsDHT, input string, dhtHan chan<- GeneralCommand) {
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

		resp := make(chan []peer.ID, 1)

		command := new(GeneralCommand)
		closest := new(ClosestCommand)
		closest.key = result
		closest.response = resp
		command.closest = closest

		dhtHan <- *command
		closestPeers := <-resp

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
		prompt = promptui.Prompt{
			Label: "Value",
		}
		value, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Value %v\n", value)
		valueData := []byte(value)
		generalCommand := new(GeneralCommand)
		put := newPutCommand(keyraw, valueData)
		generalCommand.put = &put
		dhtHan <- *generalCommand
	case "get":
		prompt := promptui.Prompt{
			Label: "Key",
		}
		keyraw, err := prompt.Run()
		if err != nil {
			panic(err)
		}

		generalCommand := new(GeneralCommand)
		get, response := newGetCommand(keyraw)
		generalCommand.get = &get
		dhtHan <- *generalCommand
		valueData := <-response

		valueString := string(valueData)
		fmt.Printf("Got value %#v for key %v\n", valueString, keyraw)
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
		pref := cid.Prefix{
			Version: 1,
			Codec:   cid.Raw,
			MhType:  multihash.SHA2_256,
		}
		c, err := pref.Sum([]byte(key))
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
		pref := cid.Prefix{
			Version: 1,
			Codec:   cid.Raw,
			MhType:  multihash.SHA2_256,
		}
		c, err := pref.Sum([]byte(key))
		if err != nil {
			panic(err)
		}
		peers, err := dht.FindProviders(ctx.TODO(), c)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Found peers %v\n", peers)
	case "connect":
		prompt := promptui.Prompt{
			Label: "Key",
		}
		keyraw, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		key := fmt.Sprintf("/kad-m/%v", keyraw)
		fmt.Printf("Key %v\n", key)

		closestPeers, err := dht.GetClosestPeers(ctx.TODO(), key)
		closestPeer := closestPeers[0]
		fmt.Printf("Closest Peer %v\n", closestPeer)
		stream, err := h.NewStream(ctx.TODO(), closestPeer, requestId)
		if err != nil {
			panic(err)
		}

		generalReq := new(GeneralRequest)
		generalClosestReq := new(GeneralRequest_ClosestProviderRequest)
		generalClosest := new(ClosestProviderRequest)
		generalClosest.Key = keyraw
		generalClosestReq.ClosestProviderRequest = generalClosest
		generalReq.GeneralRequestKind = generalClosestReq

		tun := newTunnel(stream, closestPeers[1])

		defer func(tun tunnel) {
			fmt.Println("closing the tunnel")
			err := tun.Close()
			if err != nil {
				panic(err)
			}
		}(tun)

		tun2 := newTunnel(tun, closestPeers[2])

		handler := &Handler{}
		handler.writeMessage(tun2, generalReq)

		generalResp := new(GeneralResponse)
		err = handler.readMessage(tun2, generalResp)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Read Message %+v\n", generalResp)
	}

}

func main() {
	h, DHT := NewHost(0)
	defer func(h host.Host) {
		fmt.Println("Closing Host")
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

	dhtHan, inputs := newDhtHandler(DHT)

	go dhtHan.handle()

	handler := &Handler{}
	handler.h = h
	handler.dhtInput = inputs
	go h.SetStreamHandler(requestId, handler.handleStream)

	prompt := promptui.Prompt{
		Label:    "",
		Validate: validate,
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			panic(err)
		}
		fmt.Printf("You chose %v\n", result)
		handleInput(h, DHT, result, inputs)
	}
}
