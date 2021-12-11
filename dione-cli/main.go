package dione_cli

import (
	"github.com/Dione-Software/dione-go"
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
	c.connection.Get(key)
}

func main() {

}
