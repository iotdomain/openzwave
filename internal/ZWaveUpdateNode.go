// Package internal with discovery functions
package internal

import (
	"fmt"
	"strings"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZWaveUpdateNode is invoked by OZW when it update node attributes.
// This is called when a node is first discovered and when names or info is updated
// Openzwave performs discovery in stages. During first discovery not all attributes are
// known yet, so this function will be called when additional node information is discovered.
func (app *OpenZWaveApp) ZWaveUpdateNode(notification *goopenzwave.Notification) {
	pub := app.pub
	homeID := notification.HomeID
	zwNodeID := notification.NodeID
	hwID := fmt.Sprint(zwNodeID)

	//--- These are the known mapped attributes
	manuID := goopenzwave.GetNodeManufacturerID(homeID, zwNodeID)
	_ = manuID
	manufacturer := goopenzwave.GetNodeManufacturerName(homeID, zwNodeID)
	zwBasicType := goopenzwave.GetNodeBasicType(homeID, zwNodeID)
	zwControllerNodeID := goopenzwave.GetControllerNodeID(homeID)
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

	// Ensure that the node exists
	node := app.pub.GetNodeByHWID(hwID)
	if node == nil {
		app.pub.CreateNode(hwID, types.NodeTypeUnknown)
	}
	nodeType := zwDeviceTypeStr
	pub.UpdateNodeAttr(hwID, map[types.NodeAttr]string{
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

	nodeStatus := types.NodeRunStateLost
	if zwIsNodeFailed {
		nodeStatus = types.NodeRunStateError
	} else if !goopenzwave.IsNodeAwake(homeID, zwNodeID) {
		nodeStatus = types.NodeRunStateReady
	} else {
		queryStage := goopenzwave.GetNodeQueryStage(homeID, zwNodeID)
		nodeStatus = queryStage
	}
	pub.UpdateNodeStatus(hwID, map[types.NodeStatus]string{
		types.NodeStatusRunState: nodeStatus,
	})

	//--- ZWave Specific detailed parameters
	if app.config.IncludeZwInfo {
		zwMaxBaudrate := goopenzwave.GetNodeMaxBaudRate(homeID, zwNodeID)

		pub.UpdateNodeAttr(hwID, types.NodeAttrMap{
			"zwIsNodeFailed":     fmt.Sprint(zwIsNodeFailed),
			"zwBasicType":        fmt.Sprint(zwBasicType),
			"zwDeviceType":       fmt.Sprintf("%v (0x%04X)", zwDeviceType, zwDeviceType),
			"zwControllerNodeID": fmt.Sprint(zwControllerNodeID),
			"zwHomeID":           fmt.Sprintf("%x", notification.HomeID),

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
			pub.UpdateNodeAttr(hwID, types.NodeAttrMap{
				"zwGroups": strings.Join(groups, ","),
			})
			//groupProp := device.SetAttr(OZWNodeTypeGroup, device.DataTypeString,false)

		} //node.SetPropertyValue(groupProp.Name, strings.Join(groups, ","))
	}
	//adapter.Logger().Infof("OpenZWaveAdapter.SetNodeAttr: Discovered device exit. NodeId=%d", nodeId)
	logrus.Infof("ZWaveUpdateNode: HWAddress=%s, manufacturer=%s, nodeType=%v, ...",
		hwID, manufacturer, zwNodeType)
}
