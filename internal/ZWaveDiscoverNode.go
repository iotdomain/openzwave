// Package internal with discovery functions
package internal

import (
	"fmt"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZwaveDiscoverNode is invoked by OZW when it discovers a new node.
// This adds the node to the IoTDomain. If the node already exists, it is updated.
func (app *OpenZWaveApp) ZwaveDiscoverNode(notification *goopenzwave.Notification) {
	logrus.Infof("ZwaveDiscoverNode: NodeID=%d", notification.NodeID)
	// create the node and update its attributes from the zw notification
	app.ZWaveUpdateNode(notification)

}

// ZWaveDiscoverController is invoked by OZW when it discovers the ZWave controller.
// This adds a node for the controller to the IoTDomain with input pushbuttons to
// manage the network, such as adding, removing nodes, refresh node info, heal network, etc.
func (app *OpenZWaveApp) ZWaveDiscoverController(notification *goopenzwave.Notification) {
	// Initialization completed
	pub := app.pub
	zwNodeID := notification.NodeID
	homeID := notification.HomeID
	nodeHWID := fmt.Sprint(zwNodeID)

	logrus.Infof("ZWaveDiscoverController: HWAddress=%s", nodeHWID)

	app.ZwaveDiscoverNode(notification)

	// Add Gateway specific attributes and sensors:
	zwControllerNodeID := goopenzwave.GetControllerNodeID(homeID)
	zwLibraryTypeName := goopenzwave.GetLibraryTypeName(homeID)
	zwLibraryVersion := goopenzwave.GetLibraryVersion(homeID)
	zwPrimary := goopenzwave.IsPrimaryController(homeID)
	zwSUC := goopenzwave.IsStaticUpdateController(homeID)
	zwSucNodeID := goopenzwave.GetSUCNodeID(homeID)
	zwVersion := goopenzwave.GetVersionAsString()
	app.ozwHomeID = notification.HomeID // for some reason ozwHomeID is not available in the command callback. Store it in ozwAPI

	//version := goopenzwave.GetLibraryVersion(notification.HomeID)
	//gw.SetAddress(adapter.Ozw.Address)
	pub.UpdateNodeAttr(nodeHWID, types.NodeAttrMap{
		types.NodeAttrSoftwareVersion: zwVersion,
		types.NodeAttrName:            "OpenZwave controller", // the actual name is in node 1
	})

	// Other device descriptions
	if app.config.IncludeZwInfo {
		pub.UpdateNodeAttr(nodeHWID, map[types.NodeAttr]string{
			"zwControllerNodeID":    fmt.Sprint(zwControllerNodeID),
			"zwHomeID":              fmt.Sprintf("%x", notification.HomeID),
			"zwIsPrimaryController": fmt.Sprint(zwPrimary),
			"zwIsSUC":               fmt.Sprint(zwSUC),
			"zwLibraryTypeName":     zwLibraryTypeName,
			"zwLibraryVersion":      zwLibraryVersion,
			"zwSucNodeId":           fmt.Sprint(zwSucNodeID),
		})
	}
	// Create the control inputs on first run
	input := pub.GetInputByNodeHWID(nodeHWID, types.InputTypePushButton, ButtonInstanceAddNode)
	if input == nil {
		input := pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceAddNode, app.HandleInputCommand)
		input.Attr[types.NodeAttrDescription] = "Start the controller inclusion process to add a node"
		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceHealNetwork, app.HandleInputCommand)
		input.Attr[types.NodeAttrDescription] = "Start the controller heal network process"
		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRemoveFailedNode, app.HandleInputCommand)
		input.Attr[types.NodeAttrDescription] = "Remove failed nodes"
		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRemoveNode, app.HandleInputCommand)
		input.Attr[types.NodeAttrDescription] = "Start the controller exclusion process to remove a node"

		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRefreshNodeInfo, app.HandleInputCommand)
		input.Attr[types.NodeAttrDescription] = "Refresh the node information. Use when node information is incomplete."

		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRequestNodeValue, app.HandleInputCommand)
		input.Attr[types.NodeAttrDescription] = "Refresh the node sensor value(s)"

		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceUpdateNeighbors, app.HandleInputCommand)
		input.Attr[types.NodeAttrDescription] = "Request the node to update its neighbors. Use after network changes."
	}
	pub.UpdateNodeErrorStatus(nodeHWID, types.NodeRunStateReady, "")
}
