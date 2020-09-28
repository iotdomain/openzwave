package internal

import (
	"github.com/iotdomain/iotdomain-go/types"
)

// HandleConfigCommand handles configuration updates for openzwave nodes
func (app *OpenZWaveApp) HandleConfigCommand(nodeAddress string, changes types.NodeAttrMap) types.NodeAttrMap {

	// var err error
	// var attr *nodes.ConfigAttr
	// for attrName, configValue := range changes {
	// 	if inOutput != nil {
	// 		// in openzwave, I/O doesn't have configuration, but just in case this is added later, lets handle it
	// 		attr = inOutput.GetConfigAttr(attrName)
	// 		if attr == nil {
	// 			logrus.Warningf("OpenZWaveAdapter.HandleConfig: Device %s, InOutput %s/%s. %s is not a known configuration attribute.",
	// 				device.Id, inOutput.IOType, inOutput.Instance, attrName)
	// 			return
	// 		}
	// 	} else {
	// 		attr = device.GetConfigAttr(attrName)
	// 		if attr == nil {
	// 			logrus.Warningf("OpenZWaveAdapter.HandleConfig: Device %s. %s is not a known configuration attribute.",
	// 				device.Id, attrName)
	// 			return
	// 		}
	// 	}
	// 	zwValue := app.zwValueByAttr[attr]
	// 	if zwValue == nil {
	// 		// a non-zwave node configuration
	// 		logrus.Infof("OpenZWaveAdapter.HandleConfig: Device %s, Updating node configuration %s. Old value=%s, new value=%s",
	// 			device.Id, attrName, attr.Value, configValue)
	// 		// TODO: don't modify value directly. Take FP approach
	// 		attr = device.UpdateConfig(attrName, configValue, false)
	// 	} else {
	// 		// a zwave node configuration. Update the zwave node and wait for a change notification to update the notification
	// 		zwValueID := zwValue.ID

	// 		switch zwValue.Type {
	// 		case goopenzwave.ValueIDTypeBool:
	// 			valueBool, _ := strconv.ParseBool(configValue)
	// 			err = goopenzwave.SetValueBool(app.ozwHomeID, zwValueID, valueBool)
	// 		case goopenzwave.ValueIDTypeButton:
	// 			valueOnOff, _ := strconv.ParseBool(configValue)
	// 			err = goopenzwave.SetValueBool(app.ozwHomeID, zwValueID, valueOnOff)
	// 		case goopenzwave.ValueIDTypeString:
	// 			err = goopenzwave.SetValueString(app.ozwHomeID, zwValueID, configValue)
	// 		case goopenzwave.ValueIDTypeList:
	// 			err = goopenzwave.SetValueListSelection(app.ozwHomeID, zwValueID, configValue)
	// 		case goopenzwave.ValueIDTypeShort:
	// 			valueInt16, _ := strconv.ParseInt(configValue, 10, 16)
	// 			err = goopenzwave.SetValueInt16(app.ozwHomeID, zwValueID, int16(valueInt16))
	// 		case goopenzwave.ValueIDTypeInt:
	// 			valueInt, _ := strconv.ParseInt(configValue, 10, 32)
	// 			err = goopenzwave.SetValueInt32(app.ozwHomeID, zwValueID, int32(valueInt))
	// 		case goopenzwave.ValueIDTypeDecimal:
	// 			valueFloat, _ := strconv.ParseFloat(configValue, 32)
	// 			err = goopenzwave.SetValueFloat(app.ozwHomeID, zwValueID, float32(valueFloat))
	// 		case goopenzwave.ValueIDTypeByte:
	// 			valueByte, _ := strconv.ParseUint(configValue, 10, 8)
	// 			err = goopenzwave.SetValueUint8(app.ozwHomeID, zwValueID, uint8(valueByte))
	// 		default:
	// 			err = lib.Errorf("OpenZWaveAdapter.HandleConfig: Handling of datatype %v (%s) not supported", zwValue.Type, attr.DataType)
	// 		}
	// 		if err != nil {
	// 			logrus.Errorf("OpenZWaveAdapter.HandleConfig: Failed handling configuration update for device %s: %v", device.Id, err)
	// 		} else {
	// 			// for testing, include readback
	// 			//time.Sleep(time.Second*1)
	// 			//stuff := goopenzwave.GetValueAsString(adapter.ozwHomeID, zwValueId)
	// 			//logrus.Infof("Updating configuration for device %s.%s with value %v (readback = %v)", device.Alias, attr.Name, value, stuff)
	// 			logrus.Infof("OpenZWaveAdapter.HandleConfig: Updating configuration for device %s, config %s with value %v (type=%s)", device.Id, attrName, configValue, zwValue.Type)
	// 		}
	// 	}
	// }
	return changes
}
