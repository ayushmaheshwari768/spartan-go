package spartan_go

import (
	"math"
	"math/rand"
	"time"

	eventemitter "github.com/vansante/go-event-emitter"
)

type FakeNet struct {
	clients            map[string]*Client
	chanceMessageFails uint
	messageDelayMax    uint
}

func NewFakeNet(cfg *FakeNet) *FakeNet {
	// chanceMessageFails and messageDelayMax default to 0 if they aren't specified
	fakeNet := &FakeNet{
		clients:            make(map[string]*Client),
		chanceMessageFails: cfg.chanceMessageFails,
		messageDelayMax:    cfg.messageDelayMax,
	}
	return fakeNet
}

func (f *FakeNet) register(clients ...*Client) {
	for _, client := range clients {
		f.clients[client.Address] = client
	}
}

func (f *FakeNet) broadcast(msg string, o ...interface{}) {
	for addr := range f.clients {
		f.sendMessage(addr, msg, o)
	}
}

func (f *FakeNet) sendMessage(addr string, msg string, o ...interface{}) {
	client := f.clients[addr]
	delay := math.Floor(rand.Float64() * float64(f.messageDelayMax))

	if rand.Float64() > float64(f.messageDelayMax) {
		time.AfterFunc(time.Duration(delay)*time.Second, func() {
			client.EmitEvent(eventemitter.EventType(msg), o)
		})
	}
}
