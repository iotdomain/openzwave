// Package internal with discovery functions
package internal

import (
	"fmt"

	"github.com/iotdomain/iotdomain-go/nodes"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZWaveDiscoverNodeConfigAttr is invoked by OZW when it discovers a new zwave node configuration or attribute.
// This updates the corresponding IoTDomain node configuration/attribute
func (app *OpenZWaveApp) ZWaveDiscoverNodeConfigAttr(zwValue *goopenzwave.ValueID) {
	dataType := dataTypeMap[zwValue.Type]

	// http://www.openzwave.com/dev/classOpenZWave_1_1ValueID.html
	// "In the case of configurable parameters (handled by the configuration command class), the index is the same as the parameter ID"
	// Yet, not for user attributes that are not sensors.
	zwValueLabel := zwValue.GetLabel()
	zwValueString := zwValue.GetAsString()
	nodeHWID := fmt.Sprint(zwValue.NodeID)

	attrName := types.NodeAttr(fmt.Sprint(zwValue.Index)) // This seems not to be true in spite of documentation
	if zwValue.Genre != goopenzwave.ValueIDGenreConfig {
		// unidentified configuration label
		attrName = types.NodeAttr(zwValueLabel)
	}
	zwIsWritable := !zwValue.IsReadOnly()
	zwGenre := zwValue.Genre

	// Value Index is the parameter nr for GetConfigAttr Genres
	description := zwValueLabel
	if zwGenre == goopenzwave.ValueIDGenreConfig {
		description = fmt.Sprintf("%d: %s", zwValue.Index, zwValueLabel)
	}
	if zwIsWritable {
		// writable values are configurable
		// save the zwValue for the command to update the configuration
		configAttrID := fmt.Sprintf("%s/%s", nodeHWID, attrName)
		app.zwValueByAttrID[configAttrID] = zwValue

		configAttr := nodes.NewNodeConfig(dataType, description, "")
		app.pub.UpdateNodeConfig(nodeHWID, attrName, configAttr)
		app.pub.UpdateNodeConfigValues(nodeHWID, types.NodeAttrMap{attrName: zwValueString})
		// attr.DataType = dataType
		// attr.Description = description
		if zwValue.Type == goopenzwave.ValueIDTypeList {
			configAttr.Enum, _ = zwValue.GetListItems()
		}
		// for fast lookup of configuration by ZW value ID and by attribute instance
		// configID := deviceHwAddr + "." + attrName
		// configAttr.x := zwValue.ID
		app.attrNameByValueID[zwValue.ID] = attrName
		// app.valueIDByConfigID[configAttr.ID] = zwValue.ID
		logrus.Infof("ZWaveDiscoverNodeConfigAttr: Node %s; Added configuration %s (%s), value = %v",
			nodeHWID, attrName, zwValueLabel, zwValueString)
	} else {
		//non writable values are attributes
		// as attributes have no description, include it in the name
		app.pub.UpdateNodeAttr(nodeHWID, types.NodeAttrMap{types.NodeAttr(description): zwValueString})
		logrus.Infof("ZWaveDiscoverNodeConfigAttr: Node %s; Added attribute %s (%s), value = %v",
			nodeHWID, attrName, zwValueLabel, zwValueString)
	}
}
