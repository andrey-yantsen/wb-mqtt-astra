package main

import (
	"sync"
	"time"

	"github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
)

const driverClientId = "astra"

func startDaemon(astra *astra_l.Driver, addresses multipleAddress, brokerAddress string) {
	model := &AstraModel{
		devices:   make(map[uint16]AstraDevice),
		astra:     astra,
		addresses: addresses,
		mutex:     &sync.Mutex{},
	}
	wDriver := wbgo.NewDriver(model, wbgo.NewPahoMQTTClient(brokerAddress, driverClientId, true))
	wDriver.SetPollInterval(100 * time.Millisecond)
	if err := wDriver.Start(); err != nil {
		panic(err)
	}
}
