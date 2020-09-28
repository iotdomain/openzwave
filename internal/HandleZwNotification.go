// Package internal handle ozw notifications and publish changes onto the MQTT bus
package internal

import (
	"fmt"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// handleNotification from the openzwave controller and update the nodes and sensors
func (app *OpenZWaveApp) handleNotification(ozwAPI *OzwAPI, notification *goopenzwave.Notification) {
	pub := app.pub
	//adapter.log.Debugf("handleNotification: Received notification: %v", notification)
	notificationName := notification.String()
	nodeHWID := fmt.Sprint(notification.NodeID)
	device := pub.GetNodeByHWID(nodeHWID)
	if notification.ValueID != nil {
		valueName := notification.ValueID.GetLabel()
		// ignore certain noisy values
		if _, ignoreValue := app.config.IgnoreList[valueName]; ignoreValue {
			return
		}
		value := notification.ValueID.GetAsString()
		logrus.Infof("OpenZWaveApp.handleNotification: type=%v, node=%s, label=%s, value: %v",
			notification.Type, nodeHWID, valueName, value)
	} else {
		logrus.Infof("OpenZWaveApp.handleNotification: type=%v, node=%s", notification.Type, nodeHWID)
		// FIXME: missing notification when device state changes (UPDATE_STATE_NODE_INFO_RECEIVED)
	}
	// Switch based on notification type.
	switch notification.Type {

	case goopenzwave.NotificationTypeButtonOff, goopenzwave.NotificationTypeButtonOn:
		//adapter.Logger().Infof("handleNotification: Controller Command. Event=%v, notification=%v", notification.Event, notification.Notification)
		app.updateValue(notification.ValueID)

	case goopenzwave.NotificationTypeControllerCommand:
		logrus.Infof("OpenZWaveApp.handleNotification: Controller Command. Event=%v, notification=%v",
			notification.Event, notification.Notification)
		if *notification.Notification == goopenzwave.NotificationCodeMsgComplete {
			// How to associate this with the request?

		} else if *notification.Notification == goopenzwave.NotificationCodeTimeout {
			pub.UpdateNodeErrorStatus(nodeHWID, types.NodeStateError, fmt.Sprint(notification.Notification))
		}

	case goopenzwave.NotificationTypeCreateButton:
		app.discoverZwaveNode(notification)

	case goopenzwave.NotificationTypeDeleteButton:
		app.removeDevice(notification)

	case goopenzwave.NotificationTypeDriverReady:
		app.discoverZWaveController(notification)

	case goopenzwave.NotificationTypeAwakeNodesQueried,
		goopenzwave.NotificationTypeAllNodesQueried,
		goopenzwave.NotificationTypeAllNodesQueriedSomeDead:
		logrus.Info("OpenZWaveApp.handleNotification: Nodes Queried")
		app.discoverZWaveController(notification)

	case goopenzwave.NotificationTypeGroup:
		// group association updated
		//grpLabel := goopenzwave.GetGroupLabel(notification.HomeID, notification.NodeID, *notification.GroupIDX)
		//adapter.log.Debugf(" Node %s: group %d = %s", nodeID, *notification.GroupIDX, grpLabel)
		// add properties for group and assocations for this node
		numGroups := int(goopenzwave.GetNumGroups(notification.HomeID, notification.NodeID))
		if numGroups > 0 {
			groups := make([]string, 0)
			for groupIdx := 1; groupIdx <= numGroups; groupIdx++ {
				label := goopenzwave.GetGroupLabel(notification.HomeID, notification.NodeID, uint8(groupIdx))
				groups = append(groups, label)
			}

			if device != nil {
				//groupProp := node.NewProperty(OZWNodeTypeGroup, device.DataTypeString, false)
				//groupValue := strings.Join(groups, ",")
				//node.SetPropertyValue(groupProp.Name, groupValue)
				//adapter.myzoneMqtt.PublishProperty(node, groupProp)
			}
		}

	case goopenzwave.NotificationTypeNodeAdded: // A previously seen device is added after CC is known, eg after restart
		app.discoverZwaveNode(notification)

	case goopenzwave.NotificationTypeNodeEvent:
		// This is commonly caused when a node sends a Basic_Set command to the controller.
		// There is no ValueId in the notification so not much we can do here
		logrus.Infof("OpenZWaveApp.handleNotification: Node %s NotificationTypeNodeEvent. Ignored", nodeHWID)

	case goopenzwave.NotificationTypeNodeNew: // A new device previously unseen is added
		app.discoverZwaveNode(notification)

	case goopenzwave.NotificationTypeNodeQueriesComplete:
		app.discoverZwaveNode(notification)

	case goopenzwave.NotificationTypeNodeRemoved: // Removed from network or because the app is closing?
		// ignored until we can distinguish between removal and app closing
		// TODO: remove node. Note its values are removed first
		app.removeDevice(notification)

	case goopenzwave.NotificationTypeNodeNaming:
		//One of the node names has changed (name, manufacturer, product).
		app.DiscoverZwaveNodeAttr(notification)

	case goopenzwave.NotificationTypeNodeProtocolInfo:
		// Basic node information has been received, such as whether the node is a listening device, a routing device and
		// its baud rate and basic, generic and specific types.
		// It is after this notification that you can call Manager::GetNodeType to obtain a label containing the device description.
		app.discoverZwaveNode(notification)

	case goopenzwave.NotificationTypeValueAdded:
		// A sensor, info or configuration value has been added. Could be from cache.
		app.discoverZWaveValue(notification)

	case goopenzwave.NotificationTypeValueChanged:
		// A sensor, info or configuration value has changed value
		app.updateValue(notification.ValueID)

	case goopenzwave.NotificationTypeValueRefreshed:
		// A device/sensor value is updated, not neccesarily changed
		app.updateValue(notification.ValueID)

	case goopenzwave.NotificationTypeValueRemoved:
		// Result of a removed node. Just handle the node removal and remove its sensors
		// TODO: remove sensor. Note its values are removed before the node is removed

	case goopenzwave.NotificationTypeNotification:
		// Some error occurred
		notificationCode := *notification.Notification
		logrus.Warningf("OpenZWaveApp.handleNotification: Node %s: notification: %v", nodeHWID, notificationCode)
		if device != nil {
			if notificationCode == goopenzwave.NotificationCodeTimeout {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeStateError, fmt.Sprint(notificationCode))
			} else if notificationCode == goopenzwave.NotificationCodeDead {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeStateLost, fmt.Sprint(notificationCode))
			} else if notificationCode == goopenzwave.NotificationCodeMsgComplete {
				// complete transaction
			} else if notificationCode == goopenzwave.NotificationCodeSleep {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeStateSleeping, "")
			} else if notificationCode == goopenzwave.NotificationCodeAwake {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeStateReady, "")
			} else if notificationCode == goopenzwave.NotificationCodeAlive {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeStateReady, "")
			}
		}
	default:
		logrus.Debugf("OpenZWaveApp.handleNotification: Node %s: event %s. Ignored.", nodeHWID, notificationName)
	}
}

func (app *OpenZWaveApp) removeDevice(notification *goopenzwave.Notification) {
	nodeHWID := fmt.Sprint(notification.NodeID)
	app.pub.DeleteNode(nodeHWID)
	logrus.Warningf("OpenZWaveAdapter.removeDevice. Node %s removed", nodeHWID)
}
