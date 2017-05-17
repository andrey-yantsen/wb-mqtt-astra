package main

import (
	"errors"
	"flag"
	"strconv"
	"time"

	"fmt"

	astra_l "github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
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
			return errors.New("Address must be greater than 0")
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
	debug := flag.Bool("debug", false, "Enable debug output")
	processTestEvents := flag.Bool("process-test-events", false, "Do not ignore test events emitted by detectors")
	deleteDevice := flag.Bool("delete-device", false, "Reset registration status of the device with given address")
	flag.Parse()
	wbgo.SetDebuggingEnabled(*debug)
	if len(addresses) == 0 {
		panic("You should specify at least one address")
	}
	if driver, err := astra_l.Connect(*serial, wbgo.Debug); err != nil {
		panic(err)
	} else {
		driver.Start()
		if *deleteDevice {
			for _, addr := range addresses {
				fmt.Printf("Removing device %d ... ", addr)
				if err := driver.GetDevice(addr).DeleteDevice(); err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("ok")
				}
			}
		} else {
			startDaemon(driver, addresses, *broker, *processTestEvents)
			for {
				time.Sleep(1 * time.Second)
			}
		}
	}
}
