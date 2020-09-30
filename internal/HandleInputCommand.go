// Package internal with handling of iotdomain commands for updating openzwave nodes
package internal

import (
	"strconv"
	"strings"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// InputID's of buttons to manage ZWave nodes
const (
	ButtonInstanceAddNode          = "addnode"
	ButtonInstanceRemoveNode       = "removenode"
	ButtonInstanceRemoveFailedNode = "removefailednode"
	ButtonInstanceHealNetwork      = "healnetwork"
	ButtonInstanceRefreshNodeInfo  = "refreshnodeinfo"
	ButtonInstanceRequestNodeValue = "requestnodevalue"
	ButtonInstanceUpdateNeighbors  = "updateneighbors"
)

// HandleInputCommand for openzwave node
// Currently very basic. Only switch status is supported.
func (app *OpenZWaveApp) HandleInputCommand(input *types.InputDiscoveryMessage, sender string, payloadStr string) {
	valueID := app.valueIDByInputID[input.InputID]
	if valueID == 0 {
		// This is not a known openzwave sensor, check for button commands
		if input == nil {
			logrus.Warn("HandleInputCommand: Command for input but input is not yet discovered in Zwave")
			return
		}
		if input.InputType != types.InputTypePushButton {
			return
		}
		// Virtual push buttons for the controller
		startStop, _ := strconv.ParseBool(payloadStr)
		nodeID, _ := strconv.ParseInt(payloadStr, 10, 8)
		logrus.Infof("HandleInputCommand: PushButton '%s'. Value=%v", input.Instance, payloadStr)
		if input.Instance == ButtonInstanceHealNetwork {
			app.StartHealNetwork()
		} else if input.Instance == ButtonInstanceAddNode {
			app.AddZWaveNode(startStop)
		} else if input.Instance == ButtonInstanceRemoveNode {
			app.RemoveZWaveNode(startStop)
		} else if input.Instance == ButtonInstanceRemoveFailedNode {
			goopenzwave.RemoveFailedNode(app.ozwHomeID, uint8(nodeID))
		} else if input.Instance == ButtonInstanceRefreshNodeInfo {
			goopenzwave.RefreshNodeInfo(app.ozwHomeID, uint8(nodeID))
		} else if input.Instance == ButtonInstanceRequestNodeValue {
			goopenzwave.RequestNodeAllConfigParam(app.ozwHomeID, uint8(nodeID))
		} else if input.Instance == ButtonInstanceUpdateNeighbors {
			goopenzwave.RequestNodeNeighborUpdate(app.ozwHomeID, uint8(nodeID))
		} else {
			// unknown button ignored
			logrus.Warningf("HandleInputCommand: PushButton '%s' is not a known command. Ignored.",
				input.Instance)
		}
		return
	}
	var err error

	// for now only support on/off
	dataType := types.DataType(input.DataType)
	switch dataType {
	case types.DataTypeBool:
		//adapter.UpdateSensorCommand(sensor, payloadStr)
		app.SwitchOnOff(input.NodeHWID, input, payloadStr)
	case types.DataTypeString:
		//device.UpdateSensorCommand(sensor, payloadStr)
		err = goopenzwave.SetValueString(app.ozwHomeID, valueID, payloadStr)
	case types.DataTypeNumber:
		//device.UpdateSensorCommand(sensor, payloadStr)
		//err = goopenzwave.SetValueString(adapter.ozwHomeID, valueId, payloadStr) // let the library handle conversion
		valueInt, _ := strconv.ParseInt(payloadStr, 10, 32)
		err = goopenzwave.SetValueInt32(app.ozwHomeID, valueID, int32(valueInt))
	default:
		logrus.Warningf("HandleInputCommand: Device %s: Unexpected data type %s for property %s",
			input.NodeHWID, dataType, input.InputType)
	}
	if err != nil {
		logrus.Errorf("HandleInputCommand: Device %s failed handling command for input %s: %s",
			input.NodeHWID, input.Address, err.Error())
	}
}

// AddZWaveNode Starts the inclusion process to add a node with secure mode enabled.
// Do not start this until all nodes have been discovered, eg 'ready' state
// Unfortunately there is no way to determine if this is ongoing or completed/cancelled
func (app *OpenZWaveApp) AddZWaveNode(startStop bool) {
	logrus.Infof("AddZWaveNode")
	if startStop == true {
		goopenzwave.AddNode(app.ozwHomeID, true)
	} else {
		goopenzwave.CancelControllerCommand(app.ozwHomeID)
	}
}

// RemoveZWaveNode Starts the exclusion process to remove a node
// Do not start this until all nodes have been discovered, eg 'ready' state
// Unfortunately there is no way to determine if this is ongoing or completed/cancelled
func (app *OpenZWaveApp) RemoveZWaveNode(startStop bool) {
	logrus.Infof("RemoveZWaveNode")
	if startStop == true {
		goopenzwave.RemoveNode(app.ozwHomeID)
	} else {
		goopenzwave.CancelControllerCommand(app.ozwHomeID)
	}
}

// GetNeighbors Not supported by goopenzwave
func (app *OpenZWaveApp) GetNeighbors(nodeHWID string) {
	logrus.Warningf("GetNeighbors: Not supported by goopenzwave")
	// nodeID, _ := strconv.Atoi(nodeHWID)
	// goopenzwave.GetNodeNeighbors(app.ozwHomeID, uint8(nodeID))
}

// RefreshNodeInfo Refresh the node info
func (app *OpenZWaveApp) RefreshNodeInfo(nodeHWID string) {
	zwNodeID, _ := strconv.Atoi(nodeHWID)
	goopenzwave.RefreshNodeInfo(app.ozwHomeID, uint8(zwNodeID))
}

// RemoveFailedNode This requires the node to be in a failed state.
func (app *OpenZWaveApp) RemoveFailedNode(nodeHWID string) {
	logrus.Infof("RemovefailedNode: Node %s", nodeHWID)
	zwNodeID, _ := strconv.Atoi(nodeHWID)
	goopenzwave.RemoveFailedNode(app.ozwHomeID, uint8(zwNodeID))
}

// StartHealNetwork starts the heal network process
func (app *OpenZWaveApp) StartHealNetwork() {
	logrus.Infof("StartHealNetwork")
	goopenzwave.HealNetwork(app.ozwHomeID, true)
}

// StartHealNode tells a node to rediscover its neighbors including return routes
// Unfortunately there is no way to determine if this is ongoing or completed
func (app *OpenZWaveApp) StartHealNode(nodeHWID string) {
	logrus.Infof("StartHealNode")
	zwNodeID, _ := strconv.Atoi(nodeHWID)
	goopenzwave.HealNetworkNode(app.ozwHomeID, uint8(zwNodeID), true)
}

// UpdateNeighbors to request a device to update its neighbors. Useful after device has moved.
func (app *OpenZWaveApp) UpdateNeighbors(nodeHWID string) {
	logrus.Infof("UpdateNeighbors: Node %s", nodeHWID)
	zwNodeID, _ := strconv.Atoi(nodeHWID)
	goopenzwave.RequestNodeNeighborUpdate(app.ozwHomeID, uint8(zwNodeID))
}

// SwitchOnOff enable/disable actuators
// Value can be on/off, 0/1, true/false
func (app *OpenZWaveApp) SwitchOnOff(nodeHWID string, input *types.InputDiscoveryMessage, newValue string) {
	var err error

	// any non-zero, false or off value is considered on
	onoff := !(newValue == "0" || strings.ToLower(newValue) == "off" || strings.ToLower(newValue) == "false")
	valueID := app.valueIDByInputID[input.InputID]
	currentValue := goopenzwave.GetValueAsString(app.ozwHomeID, valueID)
	logrus.Infof("SwitchOnOff. Device %s: Property %s: current value=%s. new value=%s, changing to: %t",
		nodeHWID, input.InputType, currentValue, newValue, onoff)

	err = goopenzwave.SetValueBool(app.ozwHomeID, valueID, onoff)
	if err != nil {
		logrus.Warnf("SwitchOnOff: Node %s: Property %s. Error: %v", nodeHWID, input.InputType, err)
	}
}
