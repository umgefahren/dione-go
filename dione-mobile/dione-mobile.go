package dione_mobile

import (
	"fmt"
	"github.com/Dione-Software/dione-go"
)

type DioneMobileHost struct {
	internal dione.DioneHost
}

func New() *DioneMobileHost {
	ret := new(DioneMobileHost)
	dioneHost := dione.NewDioneHost(uint32(0))
	ret.internal = dioneHost
	return ret
}

func (dmh *DioneMobileHost) Closest(key string) string {
	ids := dmh.internal.Closest(key)
	ret := make([]string, 0)
	for _, id := range ids {
		ret = append(ret, id.ShortString())
	}
	return ret[0]
}

func Greetings(name string) string {
	return fmt.Sprintf("Hello, %v!", name)
}
