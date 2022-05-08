package spartan_go

import (
	"math"
	"math/rand"
	"time"

	eventemitter "github.com/vansante/go-event-emitter"
)

type FakeNet struct {
	clients            map[string]*Client
	miners             map[string]*Miner
	chanceMessageFails uint
	messageDelayMax    uint
}

func NewFakeNet(cfg *FakeNet) *FakeNet {
	// chanceMessageFails and messageDelayMax default to 0 if they aren't specified
	fakeNet := &FakeNet{
		clients:            make(map[string]*Client),
		miners:             make(map[string]*Miner),
		chanceMessageFails: cfg.chanceMessageFails,
		messageDelayMax:    cfg.messageDelayMax,
	}
	return fakeNet
}

func (f *FakeNet) RegisterClients(clients ...*Client) {
	for _, client := range clients {
		f.clients[client.Address] = client
	}
}

func (f *FakeNet) RegisterMiners(miners ...*Miner) {
	for _, miner := range miners {
		f.miners[miner.Client.Address] = miner
	}
}

func (f *FakeNet) Broadcast(msg string, o ...interface{}) {
	for addr := range f.clients {
		f.SendMessage(addr, msg, o...)
	}
	for addr := range f.miners {
		f.SendMessage(addr, msg, o...)
	}
}

func (f *FakeNet) SendMessage(addr string, msg string, o ...interface{}) {
	delay := math.Floor(rand.Float64() * float64(f.messageDelayMax))
	if client, ok := f.clients[addr]; ok {
		if rand.Float64() > float64(f.messageDelayMax) {
			time.AfterFunc(time.Duration(delay)*time.Second, func() {
				client.EmitEvent(eventemitter.EventType(msg), o...)
			})
		}
	} else {
		miner := f.miners[addr]
		if rand.Float64() > float64(f.messageDelayMax) {
			time.AfterFunc(time.Duration(delay)*time.Second, func() {
				miner.EmitEvent(eventemitter.EventType(msg), o...)
			})
		}
	}
}
