package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	h host.Host
}

func (h *Handler) readMessage(stream network.Stream, readBuffer []byte) *GeneralRequest {
	// Read message from stream
	n, err := stream.Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading message from stream:", err)
		return nil
	}
	ret := new(GeneralRequest)
	err = proto.Unmarshal(readBuffer[:n], ret)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return nil
	}
	return ret
}

func (h *Handler) writeMessage(stream network.Stream, m *GeneralRequest) {
	// Write Message to Stream
	data, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	_, err = stream.Write(data)
	if err != nil {
		panic(err)
	}
}

func (h *Handler) handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")
	readBuffer := make([]byte, network.MessageSizeMax)

	for {
		m := h.readMessage(stream, readBuffer)
		fmt.Printf("Got Message %v\n", m)
		if m == nil {
			panic("Can't stop read message")
		}
	}
}
