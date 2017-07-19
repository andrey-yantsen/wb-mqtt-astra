package main

import (
	"fmt"
	"strconv"

	"time"

	"github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
)

type AstraDevice struct {
	wbgo.DeviceBase
	astra         *astra_l.Driver
	device        *astra_l.Device
	address       uint16
	exists        bool
	modelObserver wbgo.ModelObserver
	sensors       map[uint16]*AstraDetector
	model         *AstraModel
}

func (a *AstraDevice) IsVirtual() bool {
	return false
}

func (a *AstraDevice) AcceptValue(name, value string) {
	// ignore retained values
}

func (a *AstraDevice) deleteSensor(id uint16) {
	a.device.DeleteLevel2Device(id)
	a.modelObserver.RemoveDevice(a)
	delete(a.sensors, id)
}

func (a *AstraDevice) getSensor(s astra_l.SensorInfo) *AstraDetector {
	ret := a.getSensorById(s.Id)
	ret.DeviceBase = wbgo.DeviceBase{
		DevName:  fmt.Sprintf("astra_%d_sensor_%d", a.address, s.Id),
		DevTitle: fmt.Sprintf("%s [%d-%d]", s.Type.Name, a.address, s.Id),
	}
	ret.sensorInfo = s
	return ret
}

func (a *AstraDevice) getSensorById(id uint16) *AstraDetector {
	if _, ok := a.sensors[id]; !ok {
		devName := fmt.Sprintf("astra_%d_sensor_%d", a.address, id)
		a.sensors[id] = &AstraDetector{
			DeviceBase: wbgo.DeviceBase{
				DevName: devName,
			},
			d:                 a,
			fieldsInitialized: make(map[string]bool),
		}
	}
	return a.sensors[id]
}

func (a *AstraDevice) AcceptOnValue(name, value string) bool {
	wbgo.Debug.Printf("[%s] AcceptOn %s = %s\n", a.Name(), name, value)
	switch name {
	case "register":
		a.model.lock()
		go func() {
			defer a.model.unlock()
			bcd := a.astra.GetDevice(0xFF)
			if f, err := bcd.FindDevice(); err != nil {
				wbgo.Error.Println("Got error while finding device to register ", err)
				a.Observer.OnValue(a, "register", "0")
			} else {
				if err := a.device.RegisterDevice(f.EUI.GetShortDeviceEUI()); err != nil {
					wbgo.Error.Println("Got error while registering device ", err)
					a.Observer.OnValue(a, "register", "0")
				} else {
					a.device.CreateLevel2Net(1) // init device with channel 1
					a.SetTitle(f.DeviceType.Name)
					a.modelObserver.OnNewDevice(a)
					a.Publish()
				}
			}
		}()
	case "register_l2":
		a.model.lock()
		go func() {
			defer a.model.unlock()
			defer a.Observer.OnValue(a, "register_l2", "0")
			if s, err := a.device.RegisterLevel2Device(0); err != nil {
				wbgo.Error.Println("Got error while registering L2 device ", err)
			} else {
				a.ensureSensor(*s)
			}
		}()
	case "delete_l2_all":
		a.model.lock()
		go func() {
			defer a.model.unlock()
			defer a.Observer.OnValue(a, "delete_l2_all", "0")
			if err := a.device.DeleteAllLevel2Devices(); err != nil {
				wbgo.Error.Println("Got error while deleting all L2 devices ", err)
			} else {
				for _, d := range a.sensors {
					a.modelObserver.RemoveDevice(d)
				}
			}
		}()
	case "l2_channel":
		a.model.lock()
		go func() {
			defer a.model.unlock()
			defer func() {
				if cfg, err := a.device.GetNetLevel2Config(); err != nil {
					wbgo.Error.Println("Got error while checking channel ", err)
				} else {
					wbgo.Info.Println("Channel set to", cfg.ChannelNo)
					a.Observer.OnValue(a, "l2_channel", strconv.Itoa(int(cfg.ChannelNo)))
				}
			}()
			if channelNo, err := strconv.Atoi(value); err != nil {
				wbgo.Error.Println("Unable to convert channelNo to int ", err)
			} else if channelNo == 0 {
				wbgo.Error.Println("Unable to set channel to 0")
			} else {
				if err := a.device.CreateLevel2Net(uint8(channelNo)); err != nil {
					wbgo.Error.Println("Got error while changing channel ", err)
				} else {
					for _, d := range a.sensors {
						a.modelObserver.RemoveDevice(d)
					}
				}
			}
		}()
	case "control_time":
		a.model.lock()
		go func() {
			defer a.model.unlock()
			defer func() {
				if cfg, err := a.device.GetNetLevel2Config(); err != nil {
					wbgo.Error.Println("Got error while checking control time ", err)
				} else {
					wbgo.Info.Println("Control time set to", cfg.ControlTime)
					a.Observer.OnValue(a, "control_time", strconv.Itoa(int(cfg.ControlTime)))
				}
			}()
			if controlTime, err := strconv.Atoi(value); err != nil {
				wbgo.Error.Println("Unable to convert control time to int ", err)
			} else if controlTime == 0 {
				wbgo.Error.Println("Unable to set control time to 0")
			} else if controlTime > 240 {
				wbgo.Error.Println("Control time should be less than 241")
			} else {
				if cfg, err := a.device.GetNetLevel2Config(); err != nil {
					wbgo.Error.Println("Got error while checking current config", err)
				} else {
					cfgW := astra_l.RfNetParametersToSet{
						ChannelNo:   cfg.ChannelNo,
						ControlTime: uint8(controlTime),
					}
					if err := a.device.SetNetLevel2Config(cfgW); err != nil {
						wbgo.Error.Println("Got error while changing control time ", err)
					}
				}
			}
		}()
	case "new_radio_mode":
		a.model.lock()
		go func() {
			defer a.model.unlock()
			defer func() {
				if cfg, err := a.device.GetNetLevel2Config(); err != nil {
					wbgo.Error.Println("Got error while checking net config", err)
				} else {
					wbgo.Info.Println("Radio mode set to", cfg.RfType)
					a.Observer.OnValue(a, "new_radio_mode", strconv.Itoa(int(cfg.RfType)))
				}
			}()
			mode := astra_l.RadioModeOld
			if value == "1" {
				mode = astra_l.RadioModeNew
			}
			if err := a.device.SetRadioMode(mode); err != nil {
				wbgo.Error.Println("Got error while setting radio mode", err)
			}
		}()
	}
	return false
}

func (a *AstraDevice) Poll() {
	if a.model.locked {
		return
	}
	a.model.lock()
	defer a.model.unlock()
	if events, err := a.device.GetEvents(); err != nil {
		if err.Error() != "Read timeout" {
			wbgo.Error.Println("Unable to get events ", err)
		}
	} else {
		a.Observer.OnValue(a, "Last event time", time.Now().Format(time.UnixDate))
		for _, e := range events {
			wbgo.Info.Printf("Received event %T %+v\n", e, e)
			switch e := e.(type) {
			case astra_l.EventNoLink:
				as := a.ensureSensor(e.GetSensor())
				if !e.IsTestEvent() || a.model.processTestEvents {
					as.Observer.OnValue(as, "Last event time", time.Now().Format(time.UnixDate))
					as.handleEvent(e)
				}
			case astra_l.EventSStateOtherWithNoData, astra_l.EventSStateOtherWithSmoke,
				astra_l.EventSStateOtherWithTemperature, astra_l.EventSStateOtherWithTemperature2,
				astra_l.EventSStateOtherWithPower, astra_l.EventSStateRimRtr, astra_l.EventSStateRtmLC,
				astra_l.EventSStateBrr, astra_l.EventSStateKeychain:
				ev := e.(astra_l.SensorEvent)
				as := a.ensureSensor(ev.GetSensor())
				if !ev.IsTestEvent() || a.model.processTestEvents {
					as.Observer.OnValue(as, "Last event time", time.Now().Format(time.UnixDate))
					as.handleEvent(e)
				}
			case astra_l.EventRRStateTamperNorm:
				a.Observer.OnValue(a, "tamper", "0")
			case astra_l.EventRRStateTamperFault:
				a.Observer.OnValue(a, "tamper", "1")
			case astra_l.EventRRStateMainPsuFault:
				a.Observer.OnValue(a, "main_power_fault", "1")
			case astra_l.EventRRStateMainPsuNorm:
				a.Observer.OnValue(a, "main_power_fault", "0")
			case astra_l.EventRRStateReservePsuFault:
				a.Observer.OnValue(a, "reserve_power_fault", "1")
			case astra_l.EventRRStateReservePsuNorm:
				a.Observer.OnValue(a, "reserve_power_fault", "0")
			case astra_l.EventRRStatePsuFault:
				a.Observer.OnValue(a, "power_fault", "1")
			case astra_l.EventRRStatePsuNorm:
				a.Observer.OnValue(a, "power_fault", "0")
			case astra_l.EventRadioBlocked:
				a.Observer.OnValue(a, "rf_blocked", "1")
			case astra_l.EventRadioOk:
				a.Observer.OnValue(a, "rf_blocked", "0")
			default:
				wbgo.Error.Printf("Received unexpected event %T %+v\n", e, e)
			}
		}
	}
}

var sharedSwitches = map[string]string{
	"register":       "Register new Astra-RI-M",
	"register_l2":    "Register new Level2 detector",
	"delete_l2_all":  "Unregister all Level2 detectors",
	"new_radio_mode": "Use new radio mode",
}

var sharedAlarms = map[string]string{
	"rf_blocked":          "RF blocked",
	"rf_failure":          "RF failure",
	"tamper":              "Tamper",
	"power_fault":         "Power fault",
	"main_power_fault":    "Main power fault",
	"reserve_power_fault": "Reserve power fault",
}

func btosi(b bool) string {
	if b {
		return "1"
	} else {
		return "0"
	}
}

func (a *AstraDevice) initExistsDevice() {
	if a.exists {
		return
	}

	s, err := a.device.GetState()
	if err != nil {
		return
	}

	for alias := range sharedSwitches {
		value := "0"
		if alias == "register" {
			a.Observer.OnNewControl(a, wbgo.Control{
				Name:        alias,
				Title:       sharedSwitches[alias],
				Type:        "switch",
				Value:       value,
				Writability: wbgo.ForceReadOnly,
				Order:       1,
			})
			value = "1"
		}
		a.Observer.OnValue(a, alias, value)
	}

	a.Observer.OnValue(a, "rf_blocked", btosi(s.IsRfBlocked))
	a.Observer.OnValue(a, "rf_failure", btosi(s.IsRfError))
	a.Observer.OnValue(a, "tamper", btosi(s.IsTamper))
	a.Observer.OnValue(a, "power_fault", btosi(s.IsPowerFault))
	a.Observer.OnValue(a, "main_power_fault", btosi(s.IsMainPowerFault))
	a.Observer.OnValue(a, "reserve_power_fault", btosi(s.IsReservePowerFault))

	n, err := a.device.GetNetLevel2Config()
	if err != nil {
		return
	}

	a.Observer.OnValue(a, "l2_channel", strconv.Itoa(int(n.ChannelNo)))
	a.Observer.OnValue(a, "new_radio_mode", strconv.Itoa(int(n.RfType)))
	a.Observer.OnValue(a, "control_time", strconv.Itoa(int(n.ControlTime)))

	a.exists = true
}

func (a *AstraDevice) Publish() {
	for alias, title := range sharedSwitches {
		value := "0"
		a.Observer.OnNewControl(a, wbgo.Control{
			Name:  alias,
			Title: title,
			Type:  "switch",
			Value: value,
		})
	}

	a.Observer.OnNewControl(a, wbgo.Control{
		Name:   "l2_channel",
		Title:  "Level2 net channel No",
		Type:   "range",
		Value:  "0",
		HasMax: true,
		Max:    3,
	})

	a.Observer.OnNewControl(a, wbgo.Control{
		Name:   "control_time",
		Title:  "Radio channel control time",
		Type:   "range",
		Value:  "0",
		HasMax: true,
		Max:    240,
	})

	for alias, title := range sharedAlarms {
		a.Observer.OnNewControl(a, wbgo.Control{
			Name:        alias,
			Title:       title,
			Type:        "alarm",
			Value:       "0",
			Writability: wbgo.ForceReadOnly,
		})
	}

	if _, err := a.device.FindDevice(); err == nil {
		a.initExistsDevice()
	}
}

func (a *AstraDevice) ensureSensor(s astra_l.SensorInfo) *AstraDetector {
	as := a.getSensor(s)
	a.modelObserver.OnNewDevice(as)
	if _, ok := as.fieldsInitialized["publish"]; !ok {
		as.Publish()
		as.fieldsInitialized["publish"] = true
	}
	return as
}
