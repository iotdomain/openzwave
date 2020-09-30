// Package internal with discovery functions
package internal

import (
	"fmt"

	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZWaveRemoveNode is invoked by OZW when it removes a node from its network.
// This removes the node from the IoTDomain.
func (app *OpenZWaveApp) ZWaveRemoveNode(notification *goopenzwave.Notification) {
	nodeHWID := fmt.Sprint(notification.NodeID)
	app.pub.DeleteNode(nodeHWID)
	logrus.Warningf("ZWaveRemoveNode. Node %s removed", nodeHWID)
}
