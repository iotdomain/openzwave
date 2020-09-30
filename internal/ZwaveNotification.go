// Package internal handle ozw notifications and publish changes onto the MQTT bus
package internal

import (
	"fmt"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZWaveNotification is the main entry point for all OZW notifications
// This dispatches the notification to the function to handle it.
func (app *OpenZWaveApp) ZWaveNotification(ozwAPI *OzwAPI, notification *goopenzwave.Notification) {
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
		logrus.Infof("ZWaveNotification: type=%v, node=%s, label=%s, value: %v",
			notification.Type, nodeHWID, valueName, value)
	} else {
		logrus.Infof("ZWaveNotification: type=%v, node=%s", notification.Type, nodeHWID)
		// FIXME: missing notification when device state changes (UPDATE_STATE_NODE_INFO_RECEIVED)
	}
	// Switch based on notification type.
	switch notification.Type {

	case goopenzwave.NotificationTypeButtonOff, goopenzwave.NotificationTypeButtonOn:
		//adapter.Logger().Infof("handleNotification: Controller Command. Event=%v, notification=%v", notification.Event, notification.Notification)
		app.ZWaveUpdateOutputValue(notification.ValueID)

	case goopenzwave.NotificationTypeControllerCommand:
		logrus.Infof("ZWaveNotification: Controller Command. Event=%v, notification=%v",
			notification.Event, notification.Notification)
		if *notification.Notification == goopenzwave.NotificationCodeMsgComplete {
			// How to associate this with the request?

		} else if *notification.Notification == goopenzwave.NotificationCodeTimeout {
			pub.UpdateNodeErrorStatus(nodeHWID, types.NodeRunStateError, fmt.Sprint(notification.Notification))
		}

	case goopenzwave.NotificationTypeCreateButton:
		app.ZwaveDiscoverNode(notification)

	case goopenzwave.NotificationTypeDeleteButton:
		app.ZWaveRemoveNode(notification)

	case goopenzwave.NotificationTypeDriverReady:
		app.ZWaveDiscoverController(notification)

	case goopenzwave.NotificationTypeAwakeNodesQueried,
		goopenzwave.NotificationTypeAllNodesQueried,
		goopenzwave.NotificationTypeAllNodesQueriedSomeDead:
		logrus.Info("ZWaveNotification: Nodes Queried")
		app.ZWaveDiscoverController(notification)

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
		app.ZwaveDiscoverNode(notification)

	case goopenzwave.NotificationTypeNodeEvent:
		// This is commonly caused when a node sends a Basic_Set command to the controller.
		// There is no ValueId in the notification so not much we can do here
		logrus.Infof("ZWaveNotification: Node %s NotificationTypeNodeEvent. Ignored", nodeHWID)

	case goopenzwave.NotificationTypeNodeNew: // A new device previously unseen is added
		app.ZwaveDiscoverNode(notification)

	case goopenzwave.NotificationTypeNodeQueriesComplete:
		app.ZwaveDiscoverNode(notification)

	case goopenzwave.NotificationTypeNodeRemoved: // Removed from network or because the app is closing?
		// ignored until we can distinguish between removal and app closing
		// TODO: remove node. Note its values are removed first
		app.ZWaveRemoveNode(notification)

	case goopenzwave.NotificationTypeNodeNaming:
		//One of the node names has changed (name, manufacturer, product).
		app.ZWaveUpdateNode(notification)

	case goopenzwave.NotificationTypeNodeProtocolInfo:
		// Basic node information has been received, such as whether the node is a listening device, a routing device and
		// its baud rate and basic, generic and specific types.
		// It is after this notification that you can call Manager::GetNodeType to obtain a label containing the device description.
		app.ZwaveDiscoverNode(notification)

	case goopenzwave.NotificationTypeValueAdded:
		{
			// An output, attribute or configuration value has been added. Could be from cache.
			// Determine which device and output this is about
			// nodeHWID := fmt.Sprint(notification.NodeID)

			zwValue := notification.ValueID
			zwGenre := zwValue.Genre
			zwValueLabel := zwValue.GetLabel()
			zwValueWritable := !zwValue.IsReadOnly()

			// try to map the zwaveLabel to its node outputType so we know if it is a known output
			outputType := zwaveToOutputTypeMap[zwValueLabel]

			//if zwGenre == goopenzwave.ValueIDGenreUser && sensorType != "" {
			if outputType != "" {
				// UserGenres are either sensors/actuators, or config for a CC
				// Unfortunately determining the difference is non-deterministic.
				// Sooo, use a list of known labels to determine what is an actual output.
				// Note 1: the zwValueLabel can be modified so there is no guarantee this follows a standard naming
				// Note 2: in case of a real but unknown sensor or actuator, it shows as a config value
				app.ZWaveDiscoverOutput(nodeHWID, outputType, zwValue)
			} else if zwGenre == goopenzwave.ValueIDGenreUser {
				// Assume user genre's are attributes
				app.ZWaveUpdateNodeAttr(nodeHWID, zwValue)
			} else if zwGenre == goopenzwave.ValueIDGenreConfig || zwValueWritable {
				// Anything else that is writable is configuration
				app.ZWaveDiscoverNodeConfigAttr(zwValue)
			} else {
				// what remains are info values
				app.ZWaveUpdateNodeAttr(nodeHWID, zwValue)
			}
		}
	case goopenzwave.NotificationTypeValueChanged, goopenzwave.NotificationTypeValueRefreshed:
		// A sensor, info or configuration value has changed value
		app.ZWaveUpdateValue(notification.ValueID)

		// case :goopenzwave.NotificationTypeValueRefreshed
		// A device/output value is updated, not neccesarily changed
		// app.ZWaveUpdateValue(notification.ValueID)

	case goopenzwave.NotificationTypeValueRemoved:
		// Result of a removed node. Just handle the node removal and remove its sensors
		// TODO: remove sensor. Note its values are removed before the node is removed

	case goopenzwave.NotificationTypeNotification:
		// Some error occurred
		notificationCode := *notification.Notification
		logrus.Warningf("ZWaveNotification: Node %s: notification: %v", nodeHWID, notificationCode)
		if device != nil {
			if notificationCode == goopenzwave.NotificationCodeTimeout {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeRunStateError, fmt.Sprint(notificationCode))
			} else if notificationCode == goopenzwave.NotificationCodeDead {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeRunStateLost, fmt.Sprint(notificationCode))
			} else if notificationCode == goopenzwave.NotificationCodeMsgComplete {
				// complete transaction
			} else if notificationCode == goopenzwave.NotificationCodeSleep {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeRunStateSleeping, "")
			} else if notificationCode == goopenzwave.NotificationCodeAwake {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeRunStateReady, "")
			} else if notificationCode == goopenzwave.NotificationCodeAlive {
				pub.UpdateNodeErrorStatus(nodeHWID, types.NodeRunStateReady, "")
			}
		}
	default:
		logrus.Debugf("ZWaveNotification: Node %s: event %s. Ignored.", nodeHWID, notificationName)
	}
}
