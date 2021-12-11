package main

import (
	dione "dione-go"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/manifoldco/promptui"
)

type commandRound struct {
	connection dione.DioneInterface
}

func (c commandRound) closest() []peer.ID {
	prompt := promptui.Prompt{
		Label: "Key",
	}

	result, err := prompt.Run()

	if err != nil {
		panic(err)
	}

	return c.connection.Closest(result)
}

func (c commandRound) put() {
	prompt := promptui.Prompt{
		Label: "Key",
	}
	key, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	prompt = promptui.Prompt{
		Label: "Value",
	}
	value, err := prompt.Run()
	valuedata := []byte(value)
	c.connection.Put(key, valuedata)
}

func (c commandRound) get() {
	prompt := promptui.Prompt{
		Label: "Key",
	}
	key, err := prompt.Run()

	if err != nil {
		panic(err)
	}
	valueData := c.connection.Get(key)
	value := string(valueData)
	fmt.Printf("Value: %v\n", value)
}

func (c commandRound) connect() {
	closests := c.closest()
	target := closests[0]
	tun := c.connection.Connect(target)

	nc := new(commandRound)
	nc.connection = tun

	for {
		res := nc.newRound()
		if res {
			break
		}
	}
}

func (c commandRound) newRound() bool {
	prompt := promptui.Select{
		Label: "Action",
		Items: []string{"Connect", "Closest", "Put", "Get", "Close"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		panic(err)
	}

	switch result {
	case "Closest":
		closest := c.closest()
		fmt.Printf("Closest peers %v\n", closest)
	case "Connect":
		c.connect()
	case "Put":
		c.put()
	case "Get":
		c.get()
	case "Close":
		return true
	}

	return false
}

func main() {
	h := dione.NewDioneHost(0)
	cR := new(commandRound)
	cR.connection = h
	for {
		res := cR.newRound()
		if res {
			break
		}
	}
}
