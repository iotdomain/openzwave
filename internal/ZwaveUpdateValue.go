// Package internal for discovery of node input and output values
package internal

import (
	"fmt"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZWaveUpdateValue updates the value of a zwave output, or the value of an attribute or that of a configuration
// there is no direct way to determine what is updated so use previous discovery to see if the valueID is an output
func (app *OpenZWaveApp) ZWaveUpdateValue(zwValue *goopenzwave.ValueID) {
	// Does updateValue get called with cached values?
	zwValueLabel := zwValue.GetLabel()
	zwValueString := zwValue.GetAsString()
	nodeHWID := fmt.Sprint(zwValue.NodeID)

	outputID := app.outputIDByValueID[zwValue.ID]
	if outputID != "" {
		app.ZWaveUpdateOutputValue(zwValue)
	} else {
		// This is an update of a device attribute or configuration
		node := app.pub.GetNodeByHWID(nodeHWID)
		attrName := app.attrNameByValueID[zwValue.ID]
		//isReadOnly := zwValue.IsReadOnly()
		//_ = isReadOnly
		if node != nil && attrName != "" {
			logrus.Infof("handleZWaveValueUpdate: Configuration update node=%s, attr='%s' zwValue=%v",
				nodeHWID, attrName, zwValueString)

			_, isConfig := node.Config[attrName]
			if isConfig {
				app.pub.UpdateNodeConfigValues(nodeHWID, types.NodeAttrMap{attrName: zwValueString})
			} else {
				// not config, default to so it is info attribute
				app.pub.UpdateNodeAttr(nodeHWID, types.NodeAttrMap{types.NodeAttr(zwValueLabel): zwValueString})
			}
		} else {
			// ???
			logrus.Debugf("handleZWaveValueUpdate: Ignored value update for node=%s, zwAttr='%s' zwValue=%v",
				nodeHWID, zwValueLabel, zwValueString)
		}

	}
}
