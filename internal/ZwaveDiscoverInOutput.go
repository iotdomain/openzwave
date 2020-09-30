// Package internal for discovery of node input and output values
package internal

import (
	"fmt"
	"strings"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// Map of Zwave value labels to output types
// There doesn't seem to be another way to determine the difference between a sensor and info or configuration
// CommandClasses refer to the scope of the value but says nothing about the value itself.
// Genre user is often a sensor but not in case of Energy (ZW096) it has a different genre
// TODO: load from map file so updates don't require code changes
var zwaveToOutputTypeMap = map[string]types.OutputType{
	//"Basic":             types.SensorTypeDimmer,
	"Battery":           types.OutputTypeBattery,
	"Battery Level":     types.OutputTypeBattery,
	"Brightness":        types.OutputTypeLuminance,
	"Brightness Level":  types.OutputTypeLuminance,
	"Burglar":           types.OutputTypeMotion,
	"Current":           types.OutputTypeElectricCurrent,
	"Dimmer":            types.OutputTypeDimmer,
	"Energy":            types.OutputTypeElectricEnergy,
	"Light":             types.OutputTypeOnOffSwitch,
	"Luminance":         types.OutputTypeLuminance,
	"Motion":            types.OutputTypeMotion,
	"Uptime":            "uptime",
	"Power":             types.OutputTypeElectricPower,
	"Relative Humidity": types.OutputTypeHumidity,
	"Sensor":            types.OutputTypeMotion,
	"Switch":            types.OutputTypeOnOffSwitch,
	"Temperature":       types.OutputTypeTemperature,
	"Ultraviolet":       types.OutputTypeUltraviolet,
	"Voltage":           types.OutputTypeVoltage,
}

// Note: Basic is not an output. See https://github.com/OpenZWave/open-zwave/wiki/Basic-Command-Class
// The question is what best to do here?
//
//// Command Classes that are published sensors types
//var ccToSensorTypeMap = map[uint8]string{
//  0x20: types.SensorTypeTypeUnknown,   // COMMAND_CLASS_BASIC
//  0x25: types.SensorTypeOnOffSwitch,   // COMMAND_CLASS_SWITCH_BINARY
//  0x26: types.SensorTypeDimmer,        // COMMAND_CLASS_SWITCH_MULTILEVEL
//  0x27: types.SensorTypeOnOffSwitch,   // COMMAND_CLASS_SWITCH_ALL
//  0x30: types.SensorTypeContact,       // COMMAND_CLASS_SENSOR_BINARY
//  0x31: types.SensorTypeLevel,         // COMMAND_CLASS_SENSOR_MULTILEVEL
//  0x32: types.SensorTypeTypeUnknown,   // COMMAND_CLASS_METER
//  0x35: types.SensorTypeTypeUnknown,   // COMMAND_CLASS_METER_PULSE
//  0x50: types.SensorTypeOnOffSwitch,   // COMMAND_CLASS_BASIC_WINDOW_COVERING
//  0x62: types.SensorTypeLock,          // COMMAND_CLASS_DOOR_LOCK
//  0x71: types.SensorTypeAlarm,         // COMMAND_CLASS_ALARM
//  0x73: types.SensorTypeElectricPower, // COMMAND_CLASS_POWERLEVEL
//  0x76: types.SensorTypeLock,          // COMMAND_CLASS_LOCK
//  0x80: types.SensorTypeBattery,       // COMMAND_CLASS_BATTERY
//  0x81: types.SensorTypeTypeUnknown,   // COMMAND_CLASS_CLOCK
//  0x82: types.SensorTypeTypeUnknown,   // COMMAND_CLASS_HAIL
//  0x90: types.SensorTypeTypeUnknown,   // COMMAND_CLASS_ENERGY_PRODUCTION
//  0x9c: types.SensorTypeMotion,        // COMMAND_CLASS_SENSOR_ALARM
//}

//
//// Command Classes that are published configuration
//var ccToConfigMap = map[uint8]string{
//  0x33: types.CA_Color,         // COMMAND_CLASS_COLOR
//  // Non-standardized configuration names, use its name
//  0x40: types.CA_,   // COMMAND_CLASS_THERMOSTAT_MODE
//  0x42: types.CA_,   // COMMAND_CLASS_THERMOSTAT_OPERATING_STATE
//  0x43: types.CA_,   // COMMAND_CLASS_THERMOSTAT_SETPOINT
//  0x44: types.CA_,   // COMMAND_CLASS_THERMOSTAT_FAN_MODE
//  0x45: types.CA_,   // COMMAND_CLASS_THERMOSTAT_FAN_STATE
//  0x46: types.CA_,   // COMMAND_CLASS_CLIMATE_CONTROL_SCHEDULE
//  0x63: types.CA_,   // COMMAND_CLASS_USER_CODE
//  0x70: types.CA_,   // COMMAND_CLASS_CONFIGURATION
//  0x75: types.CA_,   // COMMAND_CLASS_PROTECTION
//  0x84: types.CA_,   // COMMAND_CLASS_WAKE_UP
//  0x85: types.CA_,   // COMMAND_CLASS_ASSOCIATION
//  0x8B: types.CA_,   // COMMAND_CLASS_TIME_PARAMETERS
//  0x9b: types.CA_,   // COMMAND_CLASS_ASSOCIATION_COMMAND_CONFIGURATION
//}
//
//// Command Classes that are published information
//var ccToInfoMap = map[uint8]string{
//  0x22: types.IA_,        // COMMAND_CLASS_APPLICATION_STATUS
//  0x5B: types.IA_,        // COMMAND_CLASS_CENTRAL_SCENE
//  0x5E: types.IA_,        // COMMAND_CLASS_ZWAVEPLUS_INFO
//  0x86: types.IA_Version, // COMMAND_CLASS_VERSION
//}

// zwave constants
//const OZWNodeTypeGroup = "Groups"

// Map zwave data types to myzone data types
var dataTypeMap = map[goopenzwave.ValueIDType]types.DataType{
	goopenzwave.ValueIDTypeBool:    types.DataTypeBool,   // Boolean, true or false
	goopenzwave.ValueIDTypeButton:  types.DataTypeBool,   // A write-only value that is the equivalent of pressing a button to send a command to a device
	goopenzwave.ValueIDTypeByte:    types.DataTypeNumber, // 8-bit unsigned value, convert to number
	goopenzwave.ValueIDTypeDecimal: types.DataTypeNumber, // Represents a non-integer value as a string, to avoid floating point accuracy issues.
	goopenzwave.ValueIDTypeInt:     types.DataTypeNumber, // 32-bit signed value
	goopenzwave.ValueIDTypeList:    types.DataTypeEnum,   // List from which one item can be selected
	goopenzwave.ValueIDTypeShort:   types.DataTypeNumber, // 16-bit signed value
	goopenzwave.ValueIDTypeRaw:     types.DataTypeBytes,  // Collection of bytes
	goopenzwave.ValueIDTypeString:  types.DataTypeString,
}

// knownUnits converts openzwave unit name (lower case)
var standardUnitsMap = map[string]types.Unit{
	"%":   types.UnitPercent,
	"a":   types.UnitAmp,
	"c":   types.UnitCelcius,
	"f":   types.UnitFahrenheit,
	"kwh": types.UnitKWH,
	"lux": types.UnitLux,
	"v":   types.UnitVolt,
	"w":   types.UnitWatt,
}

// ZWaveDiscoverOutput is invoked by OZW when it discovers a new output
// This creates the corresponding node output. If the output already exists it is updated.
func (app *OpenZWaveApp) ZWaveDiscoverOutput(nodeHWID string, outputType types.OutputType, zwValue *goopenzwave.ValueID) {

	// Discover a sensor if the value represents one
	zwValueInstanceStr := fmt.Sprint(zwValue.Instance)
	output := app.pub.GetOutputByNodeHWID(nodeHWID, outputType, zwValueInstanceStr)

	zwValueWritable := !zwValue.IsReadOnly()
	zwValueUnit := zwValue.GetUnits()
	zwValueID := zwValue.ID
	//zwValueIndexStr := fmt.Sprint(zwValue.Index) // index in the instance
	zwValueLabel := zwValue.GetLabel()
	zwValueString := zwValue.GetAsString()
	zwHelp := zwValue.GetHelp()

	if output == nil {
		// New output
		logrus.Infof("ZWaveDiscoverOutput: New output: zwNodeID=%v, zwSensorType=%s, zwInstance=%s",
			zwValue.NodeID, outputType, zwValueInstanceStr)
		output = app.pub.CreateOutput(nodeHWID, outputType, zwValueInstanceStr)
	}
	// Track the output of a ZwValueID for fast lookup
	app.outputIDByValueID[zwValueID] = output.OutputID
	// Set the output unit if it has one
	unitName, _ := standardUnitsMap[strings.ToLower(zwValueUnit)]
	if unitName != "" {
		output.Unit = unitName
		app.pub.UpdateOutput(output)
	}

	// Writable values are also inputs
	if zwValueWritable {
		inputID := app.inputIDByValueID[zwValueID]
		if inputID == "" {
			input := app.pub.CreateInput(
				nodeHWID, types.InputType(outputType), zwValueInstanceStr, app.HandleInputCommand)
			input.Unit = unitName
			inputID = input.InputID
			app.inputIDByValueID[zwValueID] = inputID
		} else {
			input := app.pub.GetInputByID(inputID)
			input.Unit = unitName
		}
		app.valueIDByInputID[inputID] = zwValueID
	}

	dataType := dataTypeMap[zwValue.Type]
	logrus.Infof("ZWaveDiscoverOutput: Node %s: discoverProperty (%d) - type='%s' (%s), info='%s', "+
		"dataType='%s' (%v), writable='%v', unit='%s' (%s), Value='%s'",
		nodeHWID, zwValueID, outputType, zwValueLabel, zwHelp,
		dataType, zwValue.Type, zwValueWritable, unitName, zwValueUnit, zwValueString)

	// during initial discovery the value comes from the cache.
	// Do not update the sensor value until it was set by the device.
	if goopenzwave.IsValueSet(zwValue.HomeID, zwValue.ID) {
		app.ZWaveUpdateValue(zwValue)
	}
}
