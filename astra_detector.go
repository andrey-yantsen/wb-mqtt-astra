package main

import (
	"reflect"
	"strconv"
	"time"

	"github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
)

type AstraDetector struct {
	wbgo.DeviceBase
	sensorInfo        astra_l.SensorInfo
	astraAddress      uint16
	fieldsInitialized bool
	d                 *AstraDevice
}

func (a *AstraDetector) IsVirtual() bool {
	return false
}

func (a *AstraDetector) AcceptValue(name, value string) {
}

func (a *AstraDetector) AcceptOnValue(name, value string) bool {
	switch name {
	case "delete":
		if value == "1" {
			a.remove()
		}
	}
	return false
}

func (a *AstraDetector) remove() {
	a.d.device.DeleteLevel2Device(a.sensorInfo.Id)
	a.d.modelObserver.RemoveDevice(a)
}

func (a *AstraDetector) handleEvent(e interface{}) {
	wbgo.Debug.Printf("Inspecting %T %+v\n", e, e)
	a.Observer.OnValue(a, "Last event time", time.Now().Format(time.UnixDate))
	v := reflect.ValueOf(e)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fieldName := v.Type().Field(i).Name
		switch fi := f.Interface().(type) {
		case astra_l.EventSStateOtherBase, astra_l.EventSStateBase, astra_l.EventNoLink:
			a.handleEvent(f.Interface())
		case astra_l.EParameterCode:
			if !a.fieldsInitialized {
				a.Observer.OnNewControl(a, wbgo.Control{
					Name:        fieldName,
					Type:        "alarm",
					Writability: wbgo.ForceReadOnly,
				})
				a.Observer.OnNewControl(a, wbgo.Control{
					Name:        fieldName + "_confirmed",
					Type:        "alarm",
					Writability: wbgo.ForceReadOnly,
				})
			}
			switch f.Interface().(astra_l.EParameterCode) {
			case astra_l.PcNorm:
				a.Observer.OnValue(a, fieldName, "0")
			case astra_l.PcFault:
				a.Observer.OnValue(a, fieldName, "1")
			case astra_l.PcNormConfirmed:
				a.Observer.OnValue(a, fieldName+"_confirmed", "0")
			case astra_l.PcFaultConfirmed:
				a.Observer.OnValue(a, fieldName+"_confirmed", "1")
			}
		case bool:
			if !a.fieldsInitialized {
				a.Observer.OnNewControl(a, wbgo.Control{
					Name:        fieldName,
					Type:        "alarm",
					Value:       "0",
					Writability: wbgo.ForceReadOnly,
				})
			}
			if f.Interface().(bool) {
				a.Observer.OnValue(a, fieldName, "1")
			} else {
				a.Observer.OnValue(a, fieldName, "0")
			}
		case int:
			if !a.fieldsInitialized {
				control := wbgo.Control{
					Name:        fieldName,
					Value:       "0",
					Writability: wbgo.ForceReadOnly,
				}
				if fieldName == "Temperature" {
					control.Type = "temperature"
				}
				a.Observer.OnNewControl(a, control)
			}
			a.Observer.OnValue(a, fieldName, strconv.Itoa(f.Interface().(int)))
		case uint8:
			if !a.fieldsInitialized {
				control := wbgo.Control{
					Name:        fieldName,
					Value:       "0",
					Writability: wbgo.ForceReadOnly,
				}
				if fieldName == "Power" || fieldName == "Smoke" {
					control.Units = "%"
				}
				a.Observer.OnNewControl(a, control)
			}
			a.Observer.OnValue(a, fieldName, strconv.Itoa(int(f.Interface().(uint8))))
		case astra_l.SensorInfo:
			// do nothing
		default:
			wbgo.Error.Printf("Received unexpected field type %T %+v\n", fi, f)
		}
	}
}

func (a *AstraDetector) Publish() {
	a.Observer.OnNewControl(a, wbgo.Control{
		Name:  "delete",
		Title: "Delete sensor",
		Type:  "switch",
		Value: "0",
	})
	a.Observer.OnNewControl(a, wbgo.Control{
		Name:        "sensor id",
		Title:       "Sensor Id",
		Value:       strconv.Itoa(int(a.sensorInfo.Id)),
		Writability: wbgo.ForceReadOnly,
	})
}
