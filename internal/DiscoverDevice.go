// Package internal with discovery functions
package internal

import (
	"fmt"
	"strings"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// DiscoverZwaveNodeAttr updates the node attributes from the zwave notification
// Call when node is first discovered and when names or info is updated
func (app *OpenZWaveApp) DiscoverZwaveNodeAttr(notification *goopenzwave.Notification) {
	pub := app.pub
	homeID := notification.HomeID
	zwNodeID := notification.NodeID
	hwAddress := fmt.Sprint(zwNodeID)

	//--- These are the myzone known attributes
	manuID := goopenzwave.GetNodeManufacturerID(homeID, zwNodeID)
	_ = manuID
	manufacturer := goopenzwave.GetNodeManufacturerName(homeID, zwNodeID)
	zwBasicType := goopenzwave.GetNodeBasicType(homeID, zwNodeID)
	zwDeviceType := goopenzwave.GetNodeDeviceType(homeID, zwNodeID)
	zwDeviceTypeStr := goopenzwave.GetNodeDeviceTypeString(homeID, zwNodeID)
	zwGenericType := goopenzwave.GetNodeGenericType(homeID, zwNodeID)
	zwIsNodeFailed := goopenzwave.IsNodeFailed(homeID, zwNodeID)
	zwIsSecurityDevice := goopenzwave.IsNodeSecurityDevice(homeID, zwNodeID)
	zwIsZwavePlus := goopenzwave.IsNodeZWavePlus(homeID, zwNodeID)
	zwLocation := goopenzwave.GetNodeLocation(homeID, zwNodeID)
	zwPlusType := goopenzwave.GetNodePlusType(homeID, zwNodeID)
	zwPlusTypeStr := goopenzwave.GetNodePlusTypeString(homeID, zwNodeID)
	zwNodeName := goopenzwave.GetNodeName(homeID, zwNodeID)
	zwNodeType := goopenzwave.GetNodeType(homeID, zwNodeID) // based on genericType or basicType
	zwProductName := goopenzwave.GetNodeProductName(homeID, zwNodeID)
	zwQueryStage := goopenzwave.GetNodeQueryStage(homeID, zwNodeID)
	zwSpecificType := goopenzwave.GetNodeSpecificType(homeID, zwNodeID)
	zwVersion := goopenzwave.GetVersionAsString()

	// make sure the node exists
	node := app.pub.GetNodeByHWID(hwAddress)
	if node == nil {
		app.pub.CreateNode(hwAddress, types.NodeTypeUnknown)
	}
	nodeType := zwDeviceTypeStr
	pub.UpdateNodeAttr(hwAddress, map[types.NodeAttr]string{
		types.NodeAttrManufacturer:    manufacturer,
		types.NodeAttrModel:           zwProductName,
		types.NodeAttrName:            zwNodeName,
		types.NodeAttrType:            nodeType,
		types.NodeAttrDescription:     fmt.Sprint(zwNodeType),
		types.NodeAttrSoftwareVersion: zwVersion,
		// TODO: support for SetNodeLocation() via config
		types.NodeAttrLocationName:      zwLocation,
		types.NodeAttr("Security Node"): fmt.Sprint(zwIsSecurityDevice),
	})

	nodeStatus := types.NodeStateLost
	if zwIsNodeFailed {
		nodeStatus = types.NodeStateError
	} else if !goopenzwave.IsNodeAwake(homeID, zwNodeID) {
		nodeStatus = types.NodeStateReady
	} else {
		queryStage := goopenzwave.GetNodeQueryStage(homeID, zwNodeID)
		nodeStatus = queryStage
	}
	pub.UpdateNodeStatus(hwAddress, map[types.NodeStatusAttr]string{
		types.NodeStatusAttrState: nodeStatus,
	})

	//--- ZWave Specific detailed parameters
	if app.config.IncludeZwInfo {
		zwMaxBaudrate := goopenzwave.GetNodeMaxBaudRate(homeID, zwNodeID)

		pub.UpdateNodeAttr(hwAddress, types.NodeAttrMap{
			"zwIsNodeFailed": fmt.Sprint(zwIsNodeFailed),
			"zwBasicType":    fmt.Sprint(zwBasicType),
			"zwDeviceType":   fmt.Sprintf("%v (0x%04X)", zwDeviceType, zwDeviceType),
			//"zwHasNodeFailed": fmt.Sprint(zwHasNodeFailed),
			"zwIsAwake":                   fmt.Sprint(goopenzwave.IsNodeAwake(homeID, zwNodeID)),
			"zwIsBeamingDevice":           fmt.Sprint(goopenzwave.IsNodeBeamingDevice(homeID, zwNodeID)),
			"zwIsFrequentListeningDevice": fmt.Sprint(goopenzwave.IsNodeFrequentListeningDevice(homeID, zwNodeID)),
			"zwIsInfoReceived":            fmt.Sprint(goopenzwave.IsNodeInfoReceived(homeID, zwNodeID)),
			"zwIsRoutingDevice":           fmt.Sprint(goopenzwave.IsNodeRoutingDevice(homeID, zwNodeID)),
			"zwGenericType":               fmt.Sprintf("%v", zwGenericType),
			"zwSpecificType":              fmt.Sprint(zwSpecificType),
			"ZWwave+ type":                fmt.Sprintf("%s (%v)", zwPlusTypeStr, zwPlusType),
			"ZWave+":                      fmt.Sprint(zwIsZwavePlus),
			"zwQueryStage":                fmt.Sprint(zwQueryStage),
			"zwVersion":                   fmt.Sprint(zwVersion),
			// "zwPlusType": fmt.Sprintf("%s (%v)", zwPlusTypeStr, zwPlusType)),

			"zwMaxBaudRate": fmt.Sprint(zwMaxBaudrate),
		})

		// TODO: Add neighbor support
		//  neighbours := goopenzwave.getNodeNeighbors(homeId, nodeId)
		//  _ = neighbours
		// TODO: Add association group support
		//  groups := goopenzwave.getNodeGroups(homeId, nodeId)
		//  _=groups
		//  assoc := goopenzwave.getNodeAssociations(homeId, nodeId)
		//_ = assoc
		// add properties for group and assocations for this node
		numGroups := int(goopenzwave.GetNumGroups(homeID, zwNodeID))
		if numGroups > 0 {
			groups := make([]string, 0)
			for groupIdx := 1; groupIdx < numGroups; groupIdx++ {
				label := goopenzwave.GetGroupLabel(homeID, zwNodeID, uint8(groupIdx))
				groups = append(groups, label)
			}
			pub.UpdateNodeAttr(hwAddress, types.NodeAttrMap{
				"zwGroups": strings.Join(groups, ","),
			})
			//groupProp := device.SetAttr(OZWNodeTypeGroup, device.DataTypeString,false)

		} //node.SetPropertyValue(groupProp.Name, strings.Join(groups, ","))
	}
	//adapter.Logger().Infof("OpenZWaveAdapter.SetNodeAttr: Discovered device exit. NodeId=%d", nodeId)
	logrus.Infof("OpenZWaveAdapter.SetNodeAttr: HWAddress=%s, manufacturer=%s, nodeType=%v",
		hwAddress, manufacturer, zwNodeType)
}

// discoverZwaveNode adds a new ? to the device or updates a pre-configured sensor
func (app *OpenZWaveApp) discoverZwaveNode(notification *goopenzwave.Notification) {
	// update the node with attributes from the zw notification
	app.DiscoverZwaveNodeAttr(notification)

}

// discoverZWaveController
// Update the gateway node with controller info.
// This adds inputs to the gateway for:
//    add, remove a node
//    refresh a node info, update neighbors, request value
//    heal the network
func (app *OpenZWaveApp) discoverZWaveController(notification *goopenzwave.Notification) {
	// Initialization completed
	pub := app.pub
	zwNodeID := notification.NodeID
	homeID := notification.HomeID
	nodeHWID := fmt.Sprint(zwNodeID)

	app.DiscoverZwaveNodeAttr(notification)

	// Add Gateway specific attributes and sensors:
	zwBasicType := goopenzwave.GetNodeBasicType(homeID, zwNodeID)
	zwControllerNodeID := goopenzwave.GetControllerNodeID(homeID)
	zwIsNodeFailed := goopenzwave.IsNodeFailed(homeID, zwNodeID)
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
			"zwBasicType":           fmt.Sprint(zwBasicType),
			"zwControllerNodeID":    fmt.Sprint(zwControllerNodeID),
			"zwHomeID":              fmt.Sprintf("%x", notification.HomeID),
			"zwIsNodeFailed":        fmt.Sprint(zwIsNodeFailed),
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
		input := pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceAddNode, app.HandleCommand)
		input.Attr[types.NodeAttrDescription] = "Start the controller inclusion process to add a node"
		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceHealNetwork, app.HandleCommand)
		input.Attr[types.NodeAttrDescription] = "Start the controller heal network process"
		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRemoveFailedNode, app.HandleCommand)
		input.Attr[types.NodeAttrDescription] = "Remove failed nodes"
		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRemoveNode, app.HandleCommand)
		input.Attr[types.NodeAttrDescription] = "Start the controller exclusion process to remove a node"

		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRefreshNodeInfo, app.HandleCommand)
		input.Attr[types.NodeAttrDescription] = "Refresh the node information. Use when node information is incomplete."

		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceRequestNodeValue, app.HandleCommand)
		input.Attr[types.NodeAttrDescription] = "Refresh the node sensor value(s)"

		input = pub.CreateInput(nodeHWID, types.InputTypePushButton, ButtonInstanceUpdateNeighbors, app.HandleCommand)
		input.Attr[types.NodeAttrDescription] = "Request the node to update its neighbors. Use after network changes."
	}
	pub.UpdateNodeErrorStatus(nodeHWID, types.NodeStateReady, "")
}
