// Package internal with main publisher
// Based on the excellent goopenzwave library at https://github.com/jimjibone/goopenzwave
package internal

import (
	"os"
	"syscall"

	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// OpenZwaveAppConfig contains the openzwave publisher configuration
type OpenZwaveAppConfig struct {
	Gateway         string          `yaml:"gateway"`       // Gateway device
	IncludeZwInfo   bool            `yaml:"includeZWInfo"` // Include ZWave attributes in device and sensor info
	IgnoreList      map[string]bool // Noisy OpenZWave outputs to ignore
	OzwLogLevel     string          `yaml:"ozwLogLevel"` // default is warn
	OzwConfigFolder string          `yaml:"ozwConfigFolder"`
	OzwEnableSIS    bool            `yaml:"ozwEnableSIS"` // Controller is Static ID Server
}

// OpenZWaveApp main class
type OpenZWaveApp struct {
	config *OpenZwaveAppConfig // App configuration
	pub    *publisher.Publisher
	gwHWID string // the gateway node HWID to use

	ozwAPI            *OzwAPI
	ozwHomeID         uint32                    // OZW Node ID
	attrNameByValueID map[uint64]types.NodeAttr // identify attr and config from OZW value IDs
	inputIDByValueID  map[uint64]string         // input ID by zw value ID. For actuator update from OZW
	outputIDByValueID map[uint64]string         // output ID by zw valueID
	valueIDByInputID  map[string]uint64         // zw value ID by input ID. For switches updates from mqtt bus
	// valueIDByConfigID map[string]uint64         // zw value ID by node [hwAddress.ConfigAttr]
	// zwValueByAttr   map[*nodes.ConfigAttr]*goopenzwave.ValueID // For config updates from mqtt bus

	//controllerCommandSensor *nodes.NodeInOutput // virtual button running the current controller command
}

// Application constants
const (
	AppID                    = "openzwave"                                    // This publisher ID
	DefaultNetworkPassword   = "My name is groot"                             // Default password used to generate network key
	DefaultOzwConfigFolder   = "/usr/local/etc/openzwave"                     // Default path to installed openzwave configuration
	DefaultIgnoreNoisyValues = "Exporting, Color, Previous Reading, Interval" // Zwave reported values to ignore
	CheckAliveInterval       = 10
)

// configuration attributes
const (
	OzwAttrNameNetworkPassword = "password"
	OzwAttrNameIgnoreValues    = "ignore"
	OzwAttrNameIncludeZWInfo   = "includezwinfo"
)

// GateWayAddresses list of USB controller addresses to check
var GateWayAddresses = []string{
	"/dev/ttyACM0",
	"/dev/ttyACM1",
	"/dev/ttyUSB0",
	"/dev/ttyUSB1",
}

// CheckAlive checks if openzwave controller can still be reached
func (app *OpenZWaveApp) CheckAlive() {
	isAlive := app.ozwAPI.IsAlive()
	if !isAlive {
		logrus.Error("OpenZWaveAdapter.CheckAlive. ZWave Controller Connection Lost. Stopping...")
		// Gateway connection dropped. Restart needed. (when used with daemon)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		//adapter.Stop()
	}
}

// GetUsbStickAddress determines the most likely gateway (USB) address
func (app *OpenZWaveApp) GetUsbStickAddress() string {
	for index := 0; index < len(GateWayAddresses); index++ {
		addr := GateWayAddresses[index]
		if _, err := os.Stat(addr); err == nil {
			logrus.Infof("OpenZWaveAdapter.GetUsbStickAddress. Scanning for possible USB port %s. Found!", addr)
			return addr
		}
		logrus.Infof("OpenZWaveAdapter.GetUsbStickAddress. Scanning for possible USB port %s. Not found", addr)
	}
	logrus.Fatal("OpenZWaveAdapter.GetUsbStickAddress. No USB port found. Unable to proceed.")
	return ""
}

// LoadConfiguration and update logging and mqtt base URL from configuration
// Set config defaults
// func (app *OpenZWaveApp) LoadConfiguration(configFolder string) error {
// 	// set defaults
// 	//adapter.Ozw.ConfigFolder = DefaultOzwConfigFolder
// 	//adapter.Ozw.LogLevel = "warning"

// 	err := app.Load(AppID, configFolder, true)
// 	if err != nil {
// 		return err
// 	}
// 	gateway := app.GatewayNode()

// 	// Gateway configuration
// 	gateway.SetConfigDefault(OzwAttrNameIncludeZWInfo, false, nodes.DataTypeBool, "Include detailed ZWave information in node attributes")
// 	// network key is required for secure pairing of nodes. Generate from friendly password string
// 	attr := gateway.SetConfigDefault(OzwAttrNameNetworkPassword, DefaultNetworkPassword, nodes.DataTypeString, "Password to generate the openzwave network key. Changing this will disconnect all added devices.")
// 	attr.Secret = true // don't publish
// 	gateway.SetConfigDefault(OzwAttrNameIgnoreValues, DefaultIgnoreNoisyValues, nodes.DataTypeString, "Noisy ZWave node values to ignore")
// 	gateway.SetConfigDefault(OzwAttrNameConfigFolder, DefaultOzwConfigFolder, nodes.DataTypeString, "Location of the Openzwave library configuration folder")
// 	gateway.SetConfigDefault(OzwAttrNameEnableSis, true, nodes.DataTypeBool, "Enable controller to be the ZWave network SIS mode")
// 	gateway.SetConfigDefault(OzwAttrNameLogLevel, "warning", nodes.DataTypeString, "Openzwave library log level for OZW_Log.txt")

// 	// Publish gateway device information
// 	//adapter.GatewayNode.UpdateDeviceInfo(nodes.CA_Address, adapter.Ozw.Id)
// 	app.ozwAPI = NewOzwAPI(logrus)
// 	app.inputByValueID = make(map[uint64]*nodes.NodeInOutput)
// 	app.outputByValueID = make(map[uint64]*nodes.NodeInOutput)
// 	//adapter.attrByValueId = make(map[uint64]*nodes.ConfigAttr)
// 	app.configByValueID = make(map[uint64]string)
// 	app.zwValueByAttr = make(map[*nodes.ConfigAttr]*goopenzwave.ValueID)
// 	app.valueIDByInput = make(map[*nodes.NodeInOutput]uint64)

// 	networkPassword := gateway.GetConfigString(OzwAttrNameNetworkPassword)

// 	data := []byte(networkPassword)
// 	key := md5.Sum(data)
// 	keyStr := ""
// 	for i, val := range key {
// 		if i < 15 {
// 			keyStr = keyStr + fmt.Sprintf("0x%02X, ", val)
// 		} else {
// 			keyStr = keyStr + fmt.Sprintf("0x%02X", val)
// 		}
// 	}
// 	logrus.Warnf("OpenZWaveAdapter.LoadingConfiguration: Adding network key: %s", keyStr)
// 	app.ozwAPI.networkKey = keyStr
// 	app.includeZwInfo, _ = gateway.GetConfigBool(OzwAttrNameIncludeZWInfo)

// 	ignoreList := gateway.GetConfigString(OzwAttrNameIgnoreValues)
// 	for _, s := range strings.Split(ignoreList, ",") {
// 		app.ignoreList[s] = true
// 	}
// 	return nil
// }

// Start the adapter
// This loads the configuration and connect to the zwave controller
func (app *OpenZWaveApp) Start() error {
	logrus.Warningf("OpenZWaveApp.Start: Starting adapter openzwave")

	// configuration allows to select a USB device /dev/ttyACM0 or other. Default is search.
	gateWayAddress := app.config.Gateway
	if gateWayAddress == "" {
		gateWayAddress = app.GetUsbStickAddress()
	}
	ozwLogLevel := app.config.OzwLogLevel
	if ozwLogLevel == "" {
		ozwLogLevel = "warning"
	}
	ozwConfigFolder := app.config.OzwConfigFolder
	if ozwConfigFolder == "" {
		ozwConfigFolder = DefaultOzwConfigFolder
	}
	ozwEnableSIS := app.config.OzwEnableSIS
	logrus.Infof("OpenZWaveApp> Configuring openzwave. Address=%s, loglevel=%s, configfolder=%s, enableSIS=%v",
		gateWayAddress, ozwLogLevel, ozwConfigFolder, ozwEnableSIS)

	// Start publishing and listening
	app.pub.Start()

	// app.pub.UpdateNodeStatus(gwID, types.PublisherStateInitializing)
	app.pub.SetPublisherStatus(types.PublisherStateInitializing)
	//
	err := app.ozwAPI.Connect(
		gateWayAddress,
		ozwLogLevel,
		ozwConfigFolder,
		ozwEnableSIS,
		app.handleNotification)

	if err != nil {
		logrus.Error("OpenZWaveApp.Start: Failed starting openzwave")
		app.pub.SetPublisherStatus(types.PublisherStateFailed)
		//adapter.UpdateLastSeen(adapter.GatewayNode)
		//adapter.publisher.PublishDeviceStatus(myzone.PublisherNode)
	} else {
		app.pub.SetPublisherStatus(types.PublisherStateConnected)
	}
	return err
}

// Stop adapter and close connections
func (app *OpenZWaveApp) Stop() {
	logrus.Warningf("OpenZWaveApp.Stop: Stopping openzwave")

	app.ozwAPI.Disconnect()
	app.pub.SetPublisherStatus(types.PublisherStateDisconnected)
	app.pub.Stop()
}

// SetupGatewayNode creates the gateway (USB controller) device node
// This set the default gateway address in its configuration
func (app *OpenZWaveApp) SetupGatewayNode() *types.NodeDiscoveryMessage {
	logrus.Info("SetupGatewayNode")
	pub := app.pub
	// FIXME, discover the gateway nodeHWID
	gwID := types.NodeIDGateway
	// gwAddr := pub.MakeNodeDiscoveryAddress(deviceID)
	// app.gatewayNodeAddr = gwAddr

	// Create new or use existing instance
	gatewayNode := pub.CreateNode(gwID, types.NodeTypeGateway)
	return gatewayNode
}

// NewOpenZwaveApp returns a new uninitialized instance of the publisher
func NewOpenZwaveApp(config *OpenZwaveAppConfig, pub *publisher.Publisher) *OpenZWaveApp {
	ozwAPI := NewOzwAPI()
	app := &OpenZWaveApp{
		config: config,
		// ignoreList: make(map[string]bool),
		pub:               pub,
		ozwAPI:            ozwAPI,
		attrNameByValueID: map[uint64]types.NodeAttr{}, // identify attr and config from OZW value IDs
		inputIDByValueID:  map[uint64]string{},         // input ID by zw value ID. For actuator update from OZW
		outputIDByValueID: map[uint64]string{},         // output ID by zw valueID
		valueIDByInputID:  map[string]uint64{},         // zw value ID by input ID. For switches updates from mqtt bus
	}

	pub.SetNodeConfigHandler(app.HandleConfigCommand)

	app.SetupGatewayNode()
	return app
}

// Run the publisher until the SIGTERM  or SIGINT signal is received
func Run() {
	appConfig := &OpenZwaveAppConfig{}
	// Load the appConfig from <AppID>.yaml from the default config folder (~/.config/iotdomain)
	pub, _ := publisher.NewAppPublisher(AppID, "", appConfig, true)
	NewOpenZwaveApp(appConfig, pub)

	pub.Start()
	pub.WaitForSignal()
	pub.Stop()
}
