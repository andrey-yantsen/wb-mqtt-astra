package wb_mqtt_astra

import (
	"reflect"
	"strconv"

	"github.com/andrey-yantsen/teko-astra-go"
	"github.com/contactless/wbgo"
)

type AstraDetector struct {
	wbgo.DeviceBase
	sensorInfo        astra_l.SensorInfo
	astraAddress      uint16
	fieldsInitialized map[string]bool
	d                 *AstraDevice
}

func (a *AstraDetector) IsVirtual() bool {
	return true
}

func (a *AstraDetector) AcceptValue(name, value string) {
}

func (a *AstraDetector) AcceptOnValue(name, value string) bool {
	switch name {
	case "delete_sensor":
		if value == "1" {
			a.d.deleteSensor(a.sensorInfo.Id)
		}
	}
	return false
}

func (a *AstraDetector) handleEvent(e interface{}) {
	wbgo.Debug.Printf("Inspecting %T %+v\n", e, e)
	v := reflect.ValueOf(e)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanInterface() {
			continue
		}
		fieldName := v.Type().Field(i).Name
		switch fi := f.Interface().(type) {
		case astra_l.EventSStateOtherBase, astra_l.EventSStateBase, astra_l.EventNoLink:
			a.handleEvent(f.Interface())
		case astra_l.EParameterCode:
			pc := f.Interface().(astra_l.EParameterCode)
			switch pc {
			case astra_l.PcNorm, astra_l.PcFault:
				value := "0"
				if pc == astra_l.PcFault {
					value = "1"
				}
				if _, ok := a.fieldsInitialized[fieldName]; !ok {
					a.Observer.OnNewControl(a, wbgo.Control{
						Name:        fieldName,
						Type:        "switch",
						Value:       value,
						Writability: wbgo.ForceReadOnly,
					})
					a.fieldsInitialized[fieldName] = true
				} else {
					a.Observer.OnValue(a, fieldName, value)
				}
			case astra_l.PcNormConfirmed, astra_l.PcFaultConfirmed:
				value := "0"
				if pc == astra_l.PcFaultConfirmed {
					value = "1"
				}
				if _, ok := a.fieldsInitialized[fieldName+"_confirmed"]; !ok {
					a.Observer.OnNewControl(a, wbgo.Control{
						Name:        fieldName + "_confirmed",
						Type:        "switch",
						Value:       value,
						Writability: wbgo.ForceReadOnly,
					})
					a.fieldsInitialized[fieldName+"_confirmed"] = true
				} else {
					a.Observer.OnValue(a, fieldName+"_confirmed", value)
				}
			}
		case bool:
			value := "0"
			if f.Interface().(bool) {
				value = "1"
			}
			if _, ok := a.fieldsInitialized[fieldName]; !ok {
				a.Observer.OnNewControl(a, wbgo.Control{
					Name:        fieldName,
					Type:        "switch",
					Value:       value,
					Writability: wbgo.ForceReadOnly,
				})
				a.fieldsInitialized[fieldName] = true
			} else {
				a.Observer.OnValue(a, fieldName, value)
			}
		case int, int8:
			if _, ok := a.fieldsInitialized[fieldName]; !ok {
				control := wbgo.Control{
					Name:        fieldName,
					Value:       strconv.Itoa(int(f.Int())),
					Writability: wbgo.ForceReadOnly,
				}
				if fieldName == "Temperature" || fieldName == "Temperature2" {
					control.Type = "temperature"
				}
				a.Observer.OnNewControl(a, control)
				a.fieldsInitialized[fieldName] = true
			} else {
				a.Observer.OnValue(a, fieldName, strconv.Itoa(int(f.Int())))
			}
		case uint, uint8:
			if _, ok := a.fieldsInitialized[fieldName]; !ok {
				control := wbgo.Control{
					Name:        fieldName,
					Value:       strconv.Itoa(int(f.Uint())),
					Writability: wbgo.ForceReadOnly,
				}
				if fieldName == "Power" || fieldName == "Smoke" || fieldName == "Dust" {
					control.Units = "%"
				}
				a.Observer.OnNewControl(a, control)
				a.fieldsInitialized[fieldName] = true
			} else {
				a.Observer.OnValue(a, fieldName, strconv.Itoa(int(f.Uint())))
			}
		case astra_l.SensorInfo:
			// do nothing
		default:
			wbgo.Error.Printf("Received unexpected field type %T %+v\n", fi, f)
		}
	}
}

func (a *AstraDetector) Publish() {
	a.Observer.OnNewControl(a, wbgo.Control{
		Name:  "delete_sensor",
		Title: "Delete sensor",
		Type:  "switch",
		Value: "0",
	})
}
