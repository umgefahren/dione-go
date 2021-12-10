package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/cloudflare/circl/kem/kyber/kyber1024"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
)

type tunnel struct {
	reader    io.Reader
	writer    io.Writer
	netStream io.ReadWriteCloser
}

func (t tunnel) Read(p []byte) (n int, err error) {
	n, err = t.reader.Read(p)
	return n, err
}

func (t tunnel) Write(p []byte) (n int, err error) {
	n, err = t.writer.Write(p)
	return n, err
}

func (t tunnel) Close() error {
	err := t.netStream.Close()
	return err
}

func newTunnel(stream io.ReadWriteCloser, target peer.ID) tunnel {
	fmt.Printf("target %v\n", target.String())

	generalRequest := new(GeneralRequest)
	generalTunnelReq := new(GeneralRequest_InitTunnelRequest)
	tunnelReq := new(InitTunnelRequest)
	tunnelReq.Target = peer.Encode(target)

	keyPub, keyPrv, err := kyber1024.GenerateKeyPair(nil)
	if err != nil {
		panic(err)
	}
	keyPubBytes := make([]byte, kyber1024.PublicKeySize)
	keyPub.Pack(keyPubBytes)

	tunnelReq.Pk = keyPubBytes
	generalTunnelReq.InitTunnelRequest = tunnelReq
	generalRequest.GeneralRequestKind = generalTunnelReq
	handler := &Handler{}
	handler.writeMessage(stream, generalRequest)

	generalResp := new(GeneralResponse)
	err = handler.readMessage(stream, generalResp)
	if err != nil {
		panic(err)
	}

	tunnelResp := generalResp.GetInitTunnelResponse()
	ct := make([]byte, kyber1024.CiphertextSize)

	ss := make([]byte, kyber1024.SharedKeySize)
	ct = tunnelResp.Ciphertext
	keyPrv.DecapsulateTo(ss, ct)

	block, err := aes.NewCipher(ss)
	if err != nil {
		panic(err)
	}

	var iv [aes.BlockSize]byte
	enc := cipher.NewOFB(block, iv[:])

	encReader := &cipher.StreamReader{S: enc, R: stream}
	encWriter := &cipher.StreamWriter{S: enc, W: stream}

	if err != nil {
		panic(err)
	}

	tun := new(tunnel)
	tun.reader = encReader
	tun.writer = encWriter
	tun.netStream = stream

	decrypted := make([]byte, len([]byte("Hello World")))
	_, err = encReader.Read(decrypted)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("Decrypted %v\n", string(decrypted))

	return *tun
}
