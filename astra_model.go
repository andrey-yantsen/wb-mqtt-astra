package main

import (
	"fmt"

	"sync"

	"github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
)

type AstraModel struct {
	wbgo.ModelBase
	astra             *astra_l.Driver
	devices           map[uint16]*AstraDevice
	addresses         multipleAddress
	started           bool
	mutex             *sync.Mutex
	locked            bool
	processTestEvents bool
}

func (a *AstraModel) lock() {
	a.mutex.Lock()
	a.locked = true
}

func (a *AstraModel) unlock() {
	a.mutex.Unlock()
	a.locked = false
}

func (a *AstraModel) Start() error {
	if a.started {
		panic("Model is already started")
	}
	a.started = true
	a.devices = make(map[uint16]*AstraDevice)
	for _, address := range a.addresses {
		devName := fmt.Sprintf("astra_%d", address)

		ad := &AstraDevice{
			DeviceBase: wbgo.DeviceBase{
				DevName: devName,
			},
			astra:         a.astra,
			device:        a.astra.GetDevice(address),
			address:       address,
			modelObserver: a.Observer,
			sensors:       make(map[uint16]*AstraDetector),
			model:         a,
		}
		a.devices[address] = ad

		if f, err := ad.device.FindDevice(); err == nil {
			ad.DevTitle = fmt.Sprintf("%s [%d]", f.DeviceType.Name, address)
		}

		a.Observer.OnNewDevice(ad)
		ad.Publish()
	}
	return nil
}

func (a *AstraModel) Stop() {
	if !a.started {
		panic("Model is not started")
	}
	defer a.unlock()
	a.lock()
	a.started = false
}

func (a *AstraModel) Poll() {
	if !a.started {
		return
	}
	for _, dev := range a.devices {
		dev.Poll()
	}
}
