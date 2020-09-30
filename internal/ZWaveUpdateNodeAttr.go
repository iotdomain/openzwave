// Package internal with discovery functions
package internal

import (
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZWaveUpdateNodeAttr is invoked by OZW when a node attribute is updated
// This updates the corresponding IoTDomain node attribute if the value is known.
// Only values of System Genre are apparently node attributes.
func (app *OpenZWaveApp) ZWaveUpdateNodeAttr(nodeHWID string, zwValue *goopenzwave.ValueID) {
	zwGenre := zwValue.Genre
	zwValueLabel := zwValue.GetLabel()
	zwValueString := zwValue.GetAsString()

	if zwGenre == goopenzwave.ValueIDGenreSystem {
		// Values of significance only to users who understand the Z-Wave protocol, eg info attribute
		logrus.Infof("ZWaveUpdateNodeAttr: Node %s; System attribute %s, value = %v",
			nodeHWID, zwValueLabel, zwValueString)
		if app.config.IncludeZwInfo {
			app.pub.UpdateNodeAttr(nodeHWID, types.NodeAttrMap{
				types.NodeAttr(zwValueLabel): zwValueString,
			})
		}
	} else {
		// value not a known attribute or ignored sensor
		logrus.Warningf("ZWaveUpdateNodeAttr: Node %s; Ignored unknown attribute %s, value = %v",
			nodeHWID, zwValueLabel, zwValueString)
	}
}
