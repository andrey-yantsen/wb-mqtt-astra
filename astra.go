package main

import (
	"errors"
	"flag"
	"strconv"
	"time"

	astra_l "github.com/andrey-yantsen/teko-astra-go"
)

type multipleAddress []uint16

func (i *multipleAddress) String() string {
	return ""
}

func (i *multipleAddress) Set(value string) error {
	if a, err := strconv.Atoi(value); err != nil {
		return err
	} else {
		if a < 1 {
			return errors.New("Address must be greather than 0")
		} else if a > 0xFA {
			return errors.New("Address must be less than 250")
		}
		*i = append(*i, uint16(a))
	}
	return nil
}

func main() {
	var addresses multipleAddress
	serial := flag.String("serial", "/dev/ttyAPP4", "serial port address (/dev/...)")
	flag.Var(&addresses, "address", "device address")
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker url")
	flag.Parse()
	if len(addresses) == 0 {
		panic("You should specify at least one address")
	}
	if driver, err := astra_l.Connect(*serial); err != nil {
		panic(err)
	} else {
		driver.Start()
		startDaemon(driver, addresses, *broker)
		for {
			time.Sleep(1 * time.Second)
		}
	}
}
