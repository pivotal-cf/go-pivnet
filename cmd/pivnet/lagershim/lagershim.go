package lagershim

import (
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
	"github.com/pivotal-golang/lager"
)

type LagerShim interface {
	Debug(action string, data ...logger.Data)
}

type lagerShim struct {
	l lager.Logger
}

func NewLagerShim(l lager.Logger) LagerShim {
	return &lagerShim{
		l: l,
	}
}

func (l lagerShim) Debug(action string, data ...logger.Data) {
	allLagerData := make([]lager.Data, len(data))

	for i, d := range data {
		allLagerData[i] = lager.Data(d)
	}

	l.l.Debug(action, allLagerData...)
}
