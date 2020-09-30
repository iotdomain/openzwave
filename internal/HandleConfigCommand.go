package internal

import (
	"fmt"
	"strconv"

	"github.com/iotdomain/iotdomain-go/lib"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// HandleConfigCommand handles configuration updates for openzwave nodes and
// returns configuration attributes that can be applied immediately. ZWave node configuration
// settings are not applied until the controller sends a notification with the configuration
// changes.
func (app *OpenZWaveApp) HandleConfigCommand(nodeAddress string, changes types.NodeAttrMap) {

	var err error
	var attr *types.ConfigAttr
	var applyChanges = types.NodeAttrMap{}

	// After the zwave node accepts the configuration the controller will send a notification which
	// in turn will update the node configuration.
	// through a value notification. The downside of this approach is that there is no confirmation of the
	// update.
	node := app.pub.GetNodeByAddress(nodeAddress)
	if node == nil {
		logrus.Warningf("HandleConfigCommand: Unknown node with address '%s'", nodeAddress)
		return // nothing to apply
	}
	for attrName, configValue := range changes {
		// ZWave node config attribute IDs are set during discovery of the config value
		// See handleZWaveConfigAttrDiscovery()
		attrID := fmt.Sprintf("%s/%s", node.HWID, attrName)
		zwValue := app.zwValueByAttrID[attrID]
		if zwValue == nil {
			// a non-zwave node configuration is applied immediately
			applyChanges[attrName] = configValue
			oldValue := node.Attr[attrName]
			logrus.Infof("HandleConfigCommand: Node address '%s' (HWID=%s); Configuration %s: Old value=%s, new value=%s",
				nodeAddress, node.HWID, attrName, oldValue, configValue)
		} else {
			// a zwave node configuration. Update the zwave node and wait for a change notification to update the notification
			zwValueID := zwValue.ID

			switch zwValue.Type {
			case goopenzwave.ValueIDTypeBool:
				valueBool, _ := strconv.ParseBool(configValue)
				err = goopenzwave.SetValueBool(app.ozwHomeID, zwValueID, valueBool)
			case goopenzwave.ValueIDTypeButton:
				valueOnOff, _ := strconv.ParseBool(configValue)
				err = goopenzwave.SetValueBool(app.ozwHomeID, zwValueID, valueOnOff)
			case goopenzwave.ValueIDTypeString:
				err = goopenzwave.SetValueString(app.ozwHomeID, zwValueID, configValue)
			case goopenzwave.ValueIDTypeList:
				err = goopenzwave.SetValueListSelection(app.ozwHomeID, zwValueID, configValue)
			case goopenzwave.ValueIDTypeShort:
				valueInt16, _ := strconv.ParseInt(configValue, 10, 16)
				err = goopenzwave.SetValueInt16(app.ozwHomeID, zwValueID, int16(valueInt16))
			case goopenzwave.ValueIDTypeInt:
				valueInt, _ := strconv.ParseInt(configValue, 10, 32)
				err = goopenzwave.SetValueInt32(app.ozwHomeID, zwValueID, int32(valueInt))
			case goopenzwave.ValueIDTypeDecimal:
				valueFloat, _ := strconv.ParseFloat(configValue, 32)
				err = goopenzwave.SetValueFloat(app.ozwHomeID, zwValueID, float32(valueFloat))
			case goopenzwave.ValueIDTypeByte:
				valueByte, _ := strconv.ParseUint(configValue, 10, 8)
				err = goopenzwave.SetValueUint8(app.ozwHomeID, zwValueID, uint8(valueByte))
			default:
				err = lib.MakeErrorf("HandleConfigCommand: Handling of datatype %v (%s) not supported", zwValue.Type, attr.DataType)
			}
			if err != nil {
				logrus.Errorf("HandleConfigCommand: Failed handling configuration update for node %s: %v", node.HWID, err)
			} else {
				// for testing, include readback
				//time.Sleep(time.Second*1)
				//stuff := goopenzwave.GetValueAsString(adapter.ozwHomeID, zwValueId)
				//logrus.Infof("Updating configuration for device %s.%s with value %v (readback = %v)", device.Alias, attr.Name, value, stuff)
				logrus.Infof("HandleConfigCommand: Updating configuration for node %s, config %s with value %v (type=%s)", node.HWID, attrName, configValue, zwValue.Type)
			}
		}
	}
	// apply configurations that are not zwave device configs
	if len(applyChanges) > 0 {
		app.pub.UpdateNodeConfigValues(node.HWID, applyChanges)
	}
}
