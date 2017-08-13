package main

import (
	"flag"
	"time"

	"fmt"

	astra_l "github.com/andrey-yantsen/teko-astra-go"
	astra "github.com/andrey-yantsen/wb-mqtt-astra/pkg/wb-mqtt-astra"
	"github.com/contactless/wbgo"
)

func main() {
	var addresses astra.AddressList
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
	if driver, err := astra_l.Connect(*serial, wbgo.Debug, 50*time.Millisecond); err != nil {
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
			astra.StartDaemon(driver, addresses, *broker, *processTestEvents)
			for {
				time.Sleep(1 * time.Second)
			}
		}
	}
}
