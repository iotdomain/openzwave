// Package internal for communication with the openzwave library. Uses goopenzwave library.
package internal

import (
	"os"
	"strings"

	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// OzwAPI with openzwave API properties and methods
type OzwAPI struct {
	address string // OZW device address, eg /dev/ttyUSB0 or /dev/ttyACM0
	//sentinitialQueryComplete bool           // flag, the initial query has completed
	isRunning           bool //
	notificationHandler func(*OzwAPI, *goopenzwave.Notification)
	networkKey          string // zwave network key

	//initialQueryComplete chan bool                      // channel to publish init query has completed
	notificationChan chan *goopenzwave.Notification // notification handling channel
	homeID           uint32                         // controller home id set during discovery
	nodeID           uint8                          // controller node id set during discovery
}

// Connect the API to the device
//
func (ozwAPI *OzwAPI) Connect(
	address string,
	logLevel string,
	ozwConfigFolder string,
	enableSIS bool,
	notificationHandler func(*OzwAPI, *goopenzwave.Notification)) error {

	logrus.Warningf("OzwAPI.Connect: Connect to the OpenZwave library at %s and listen for notifications", address)
	ozwAPI.address = address
	ozwAPI.notificationHandler = notificationHandler

	// Setup the OpenZWave library.
	// Todo, figure out the path after bundling
	//configPath := "../../vendor/github.com/jimjibone/goopenzwave/.lib/etc/openzwave"
	configPath := ozwConfigFolder
	controllerPath := ozwAPI.address

	ozwLogLevel := goopenzwave.LogLevelError
	switch strings.ToLower(logLevel) {
	case "error":
		ozwLogLevel = goopenzwave.LogLevelError
	case "warn":
		ozwLogLevel = goopenzwave.LogLevelWarning
	case "info":
		ozwLogLevel = goopenzwave.LogLevelInfo
	case "debug":
		ozwLogLevel = goopenzwave.LogLevelDebug
	case "none":
		ozwLogLevel = goopenzwave.LogLevelNone
	}

	options := goopenzwave.CreateOptions(configPath, "", "")
	//options.AddOptionBool("Associate", true)  // auto associate controller with new nodes
	//options.AddOptionInt("DumpTrigger", 4)
	//options.AddOptionInt("PollInterval", 600)
	//options.AddOptionBool("IntervalBetweenPolls", true)
	//options.AddOptionBool("SaveConfiguration", true)
	options.AddOptionBool("SuppressValueRefresh", false) // tell us the device is alive

	//options.AddOptionBool("ValidateValueChanges", true)
	options.AddOptionBool("NotifyTransactions", true) // track progress
	options.AddOptionLogLevel("QueueLogLevel", ozwLogLevel)
	options.AddOptionLogLevel("SaveLogLevel", ozwLogLevel)
	options.AddOptionBool("EnableSIS", enableSIS)
	options.AddOptionString("NetworkKey", ozwAPI.networkKey, false)
	options.Lock()

	// Start the library and listen for notifications.
	err := goopenzwave.Start(
		// NOTE: Stopping on breakpoints in this callback hangs the app. Pipe notifications through a channel, breakpoints
		// in the channel handler work fine.
		func(notification *goopenzwave.Notification) {
			ozwAPI.notificationChan <- notification
		})

	if err != nil {
		logrus.Errorf("OzwAPI.Connect: ERROR: failed to start goopenzwave library: %v", err)
		return err
	}

	// Add a driver using the supplied controller path.
	err = goopenzwave.AddDriver(controllerPath)
	if err != nil {
		logrus.Errorf("OzwAPI.Connect: ERROR: failed to add goopenzwave driver: %v", err)
		return err
	}

	// Separate process to handle notifications
	ozwAPI.isRunning = true
	go ozwAPI.handleNotificationLoop()

	//// Wait here until the initial node query has completed. This can take a long time.
	//<-ozwAPI.initialQueryComplete
	//close(ozwAPI.initialQueryComplete)
	//
	//// if ozwAPI is no longer running then startup failed
	//if ozwAPI.isRunning {
	//	ozwAPI.log.Warningf("OzwAPI.Connect: Initial scan complete. Started polling for updates...")
	//} else {
	//	err = errors.New("OzwAPI.Connect: OpenZwave driver startup failed")
	//}

	return err
}

// Disconnect from the OpenZwave controller
func (ozwAPI *OzwAPI) Disconnect() {
	ozwAPI.isRunning = false
	err := goopenzwave.Stop()
	if err != nil {
		logrus.Errorf("OzwAPI.Disconnect Stopping goopenzwave error: %v", err)
	}
	goopenzwave.DestroyOptions()
	close(ozwAPI.notificationChan)
	logrus.Warningf("OzwAPI.Disconnect Stopping goopenzwave completed")
}

// listen for notifications from the channel
func (ozwAPI *OzwAPI) handleNotificationLoop() {
	logrus.Warnf("OzwAPI.handleNotificationLoop: starting listening for notifications")

	for ozwAPI.isRunning {
		notification := <-ozwAPI.notificationChan
		if notification == nil {
			break
		}
		if notification.Type == goopenzwave.NotificationTypeDriverReady {
			ozwAPI.homeID = notification.HomeID
			ozwAPI.nodeID = notification.NodeID
		}
		//if !ozwAPI.sentinitialQueryComplete {
		if notification.Type == goopenzwave.NotificationTypeAwakeNodesQueried ||
			notification.Type == goopenzwave.NotificationTypeAllNodesQueried ||
			notification.Type == goopenzwave.NotificationTypeAllNodesQueriedSomeDead {
			// Finish the connect phase as the initial node query has completed or failed.
			//ozwAPI.sentinitialQueryComplete = true
			//ozwAPI.initialQueryComplete <- true
		} else if notification.Type == goopenzwave.NotificationTypeDriverFailed {
			logrus.Errorf("OzwAPI.handleNotificationLoop.OpenZwave Driver failed (missing device?)")
			//ozwAPI.sentinitialQueryComplete = true
			//ozwAPI.initialQueryComplete <- true
			ozwAPI.isRunning = false
		}
		//}
		// always handle the notification if there is one
		ozwAPI.notificationHandler(ozwAPI, notification)
	}
	logrus.Warnf("OzwAPI.handleNotificationLoop: Exiting notification listener")
}

// GetSendQueueCount returns the nr of messages queued for sending
func (ozwAPI *OzwAPI) GetSendQueueCount() int32 {
	count := goopenzwave.GetSendQueueCount(ozwAPI.homeID)
	logrus.Infof("OzwAPI.Send queue holds %d messages", count)
	return count
}

// AddControllerToOtherNetwork receives network configuration from primary controller
// This is the same as 'set learn mode' and adds this controller to another network. The other network must have
// add node activated for this controller to be included in its network.
func (ozwAPI *OzwAPI) AddControllerToOtherNetwork() bool {
	success := goopenzwave.ReceiveConfiguration(ozwAPI.homeID)
	logrus.Infof("OzwAPI.Start AddControllerToOtherNetwork (set learn mode): %v", success)
	return success
}

// GetSucNodeID returns the SUC node ID
// The SUC manages the list of nodes in the network and can reassign the primary controller device
func (ozwAPI *OzwAPI) GetSucNodeID() uint8 {
	sucNodeID := goopenzwave.GetSUCNodeID(ozwAPI.homeID)
	logrus.Infof("OzwAPI.GetSicMpdeOd: SUC node ID: %d ", sucNodeID)
	return sucNodeID
}

// IsAlive check if openzwave connection is still alive
// Return true if initializing or running. False if stopped or controller is no longer communicating
// (this might not be fool proof)
func (ozwAPI *OzwAPI) IsAlive() bool {
	if ozwAPI.homeID == 0 {
		// initializing
		return true
	}
	// check if the device still exists
	_, err := os.Stat(ozwAPI.address)
	//qstage := goopenzwave.GetNodeQueryStage(ozwAPI.homeId, ozwAPI.nodeId)
	isFailed := goopenzwave.IsNodeFailed(ozwAPI.homeID, ozwAPI.nodeID)
	return !isFailed && err == nil
}

// NewOzwAPI creates a new instance of the OpenZwave interface
func NewOzwAPI() *OzwAPI {
	ozwAPI := new(OzwAPI)
	//ozwAPI.initialQueryComplete = make(chan bool)
	ozwAPI.notificationChan = make(chan *goopenzwave.Notification) // notification handling channel
	return ozwAPI
}
