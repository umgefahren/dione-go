package main

import (
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

func forwardStream(target network.Stream, sourceReader io.Reader, sourceWriter io.Writer) {
	targetReader := io.Reader(target)
	targetWriter := io.Writer(target)
	go forwardWriter(targetReader, sourceWriter)
	forwardWriter(sourceReader, targetWriter)
}

func forwardWriter(reader io.Reader, writer io.Writer) {
	for {
		buffer := make([]byte, network.MessageSizeMax)
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
		n, err = writer.Write(buffer[:n])
		if err != nil {
			panic(err)
		}
	}
}

func (h *Handler) readMessage(stream io.ReadWriter, m protoreflect.ProtoMessage) error {
	readBuffer := make([]byte, network.MessageSizeMax)
	// Read message from stream
	n, err := stream.Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading message from stream:", err)
		return err
	}
	err = proto.Unmarshal(readBuffer[:n], m)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return err
	}
	return nil
}

func (h *Handler) writeMessage(stream io.ReadWriter, m protoreflect.ProtoMessage) {
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

	m := new(GeneralRequest)
	err := h.readMessage(stream, m)
	// fmt.Printf("Got Message %v\n", m)
	if err != nil {

		panic(err)
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

		fmt.Println("Formulated response")

		h.writeMessage(stream, generalResp)

		fmt.Println("Wrote Message")
	} else if m.GetInitTunnelRequest() != nil {
		initTnlReq := m.GetInitTunnelRequest()
		targetId, err := peer.Decode(initTnlReq.Target)
		if err != nil {
			panic(err)
		}

		pk := kyber1024.PublicKey{}
		pk.Unpack(initTnlReq.Pk)

		ct, ss := make([]byte, kyber1024.CiphertextSize), make([]byte, kyber1024.SharedKeySize)
		pk.EncapsulateTo(ct, ss, nil)

		targetStream, err := h.h.NewStream(ctx.TODO(), targetId, requestId)
		if err != nil {
			panic(err)
		}

		generalResp := new(GeneralResponse)
		generalTunnelResp := new(GeneralResponse_InitTunnelResponse)
		generalTunnel := new(InitTunnelResponse)
		generalTunnel.Ciphertext = ct
		generalTunnel.Status = ConnectionStatus_SUCCESS
		generalTunnelResp.InitTunnelResponse = generalTunnel
		generalResp.GeneralResponseKind = generalTunnelResp

		h.writeMessage(stream, generalResp)

		block, err := aes.NewCipher(ss)
		if err != nil {
			panic(err)
		}

		var iv [aes.BlockSize]byte
		enc := cipher.NewOFB(block, iv[:])

		encReader := &cipher.StreamReader{S: enc, R: stream}
		encWriter := &cipher.StreamWriter{S: enc, W: stream}

		// decrypted := make([]byte, len([]byte("Hello World")))
		// _, err = encReader.Read(decrypted)
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Printf("Decrypted %v\n", string(decrypted))

		plaintext := []byte("Hello World")
		_, err = encWriter.Write(plaintext)
		if err != nil {
			panic(err)
		}
		fmt.Println("Forwarding stream")
		go forwardStream(targetStream, encReader, encWriter)

	} else {
		fmt.Println(m.String())
	}
}
