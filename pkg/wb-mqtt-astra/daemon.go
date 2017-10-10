package wb_mqtt_astra

import (
	"sync"
	"time"

	"errors"
	"strconv"

	"github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
)

const driverClientId = "astra"

type AddressList []uint16

func (i *AddressList) String() string {
	return ""
}

func (i *AddressList) Set(value string) error {
	if a, err := strconv.Atoi(value); err != nil {
		return err
	} else {
		if a < 1 {
			return errors.New("Address must be greater than 0")
		} else if a > 0xFA {
			return errors.New("Address must be less than 250")
		}
		*i = append(*i, uint16(a))
	}
	return nil
}

func StartDaemon(astra *astra_l.Driver, addresses AddressList, brokerAddress string, processTestEvents bool, sendLastEventOnRIM bool) {
	model := &AstraModel{
		astra:              astra,
		addresses:          addresses,
		mutex:              &sync.Mutex{},
		processTestEvents:  processTestEvents,
		sendLastEventOnRIM: sendLastEventOnRIM,
	}
	mqtt := wbgo.NewPahoMQTTClient(brokerAddress, driverClientId, true)
	mqtt.Publish(wbgo.MQTTMessage{
		Topic:   "/devices/wb_mqtt_astra/status",
		Payload: "1",
		QoS:     1,
	})
	wDriver := wbgo.NewDriver(model, mqtt)
	wDriver.SetPollInterval(100 * time.Millisecond)
	if err := wDriver.Start(); err != nil {
		panic(err)
	}
}
