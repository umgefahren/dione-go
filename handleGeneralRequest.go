package main

import (
	"bufio"
	ctx "context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/cloudflare/circl/kem/kyber/kyber1024"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
)

type Handler struct {
	h        host.Host
	dhtInput chan<- GeneralCommand
}

func forwardStream(target network.Stream, sourceReader *bufio.Reader, sourceWriter *bufio.Writer) {
	targetReader := bufio.NewReader(target)
	targetWriter := bufio.NewWriter(target)
	go forwardWriter(targetReader, sourceWriter)
	go forwardReader(targetWriter, sourceReader)
}

func forwardWriter(target *bufio.Reader, source *bufio.Writer) {
	for {
		_, err := target.WriteTo(source)
		if err != nil {
			panic(err)
		}
	}
}

func forwardReader(target *bufio.Writer, source *bufio.Reader) {
	for {
		_, err := target.ReadFrom(source)
		if err != nil {
			panic(err)
		}
	}
}

func (h *Handler) readMessage(stream network.Stream, m protoreflect.ProtoMessage) {
	readBuffer := make([]byte, network.MessageSizeMax)
	// Read message from stream
	n, err := stream.Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading message from stream:", err)
	}
	err = proto.Unmarshal(readBuffer[:n], m)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
	}
}

func (h *Handler) writeMessage(stream network.Stream, m protoreflect.ProtoMessage) {
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

	for {
		m := new(GeneralRequest)
		h.readMessage(stream, m)
		fmt.Printf("Got Message %v\n", m)
		if m == nil {
			panic("Can't stop read message")
		}

		if m.GetClosestProviderRequest() != nil {
			clPrvReq := m.GetClosestProviderRequest()
			general := new(GeneralCommand)
			closest, response := newClosestCommand(clPrvReq.Key)
			general.closest = &closest
			h.dhtInput <- *general
			closestPeers := <-response
			responseStrings := make([]string, 0)
			for _, closestPeer := range closestPeers {
				peerString := closestPeer.String()
				responseStrings = append(responseStrings, peerString)
			}
			generalResp := new(GeneralResponse)
			generalClosestResp := new(GeneralResponse_ClosestProviderResponse)
			closestResp := new(ClosestProviderResponse)
			closestResp.Provider = responseStrings
			generalClosestResp.ClosestProviderResponse = closestResp
			generalResp.GeneralResponseKind = generalClosestResp

			h.writeMessage(stream, generalResp)
		} else if m.GetInitTunnelRequest() != nil {
			initTnlReq := m.GetInitTunnelRequest()
			targetId := peer.ID(initTnlReq.Target)
			pk := kyber1024.PublicKey{}
			pk.Unpack(initTnlReq.Pk)
			ct, ss := make([]byte, kyber1024.CiphertextSize), make([]byte, kyber1024.SharedKeySize)
			pk.EncapsulateTo(ct, ss, nil)
			block, err := aes.NewCipher(ss)
			if err != nil {
				panic(err)
			}

			var iv [aes.BlockSize]byte
			encStream := cipher.NewOFB(block, iv[:])
			streamReader := io.Reader(stream)
			streamWriter := io.Writer(stream)
			encStreamReader := &cipher.StreamReader{
				S: encStream,
				R: streamReader,
			}
			encStreamWriter := &cipher.StreamWriter{
				S: encStream,
				W: streamWriter,
			}
			encReader := bufio.NewReader(encStreamReader)
			encWriter := bufio.NewWriter(encStreamWriter)
			targetStream, err := h.h.NewStream(ctx.TODO(), targetId)
			if err != nil {
				panic(err)
			}
			forwardStream(targetStream, encReader, encWriter)
		}
	}
}
