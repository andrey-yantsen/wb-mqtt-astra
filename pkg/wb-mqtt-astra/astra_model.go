package wb_mqtt_astra

import (
	"fmt"

	"sync"

	"github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
	"time"
)

type AstraModel struct {
	wbgo.ModelBase
	astra                  *astra_l.Driver
	devices                map[uint16]*AstraDevice
	addresses              AddressList
	started                bool
	mutex                  *sync.Mutex
	processTestEvents      bool
	sendLastEventOnRIM     bool
	lastSuccessfulResponse time.Time
}

const ALIVE_TIMEOUT = time.Minute

func (a *AstraModel) Start() error {
	if a.started {
		panic("Model is already started")
	}
	defer a.mutex.Unlock()
	a.mutex.Lock()
	a.lastSuccessfulResponse = time.Now()
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
			ready:         false,
		}
		a.devices[address] = ad

		if f, err := ad.device.FindDevice(); err == nil {
			ad.DevTitle = fmt.Sprintf("%s [%d]", f.DeviceType.Name, address)
			a.lastSuccessfulResponse = time.Now()
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
	defer a.mutex.Unlock()
	a.mutex.Lock()
	a.started = false
}

func (a *AstraModel) Poll() {
	if !a.started {
		return
	}
	for _, dev := range a.devices {
		if dev.ready {
			if dev.Poll() {
				a.lastSuccessfulResponse = time.Now()
			}
		}
	}
	if time.Now().After(a.lastSuccessfulResponse.Add(ALIVE_TIMEOUT)) {
		wbgo.Warn.Printf("No responses from Astra for last %d seconds, trying to reconnect to serial port...\n", ALIVE_TIMEOUT/time.Second)
		a.Stop()
		if err := a.astra.Reconnect(); err != nil {
			panic(err)
		} else {
			a.Start()
		}
	}
}
