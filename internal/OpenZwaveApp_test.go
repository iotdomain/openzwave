package internal_test

import (
	// "myzone/adapters/openzwave"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/openzwave/internal"
	"github.com/stretchr/testify/assert"
)

const TestConfigFolder = "../test"

var messengerConfig = &messaging.MessengerConfig{Domain: "test"}
var appConfig = &internal.OpenZwaveAppConfig{}

func TestLoadConfig1(t *testing.T) {
	pub, err := publisher.NewAppPublisher(internal.AppID, TestConfigFolder, appConfig, true)
	app := internal.NewOpenZwaveApp(appConfig, pub)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	//assert.Equal(t, "/dev/ttyACM0", adapter.Ozw.Address)
}

// Test connecting to OZW. This needs an adapter on /dev/ttyACM0 (as per test config)
func TestStartStop(t *testing.T) {
	pub, err := publisher.NewAppPublisher(internal.AppID, TestConfigFolder, appConfig, true)
	app := internal.NewOpenZwaveApp(appConfig, pub)
	assert.NoError(t, err)
	err = app.Start()
	assert.NoError(t, err)
	if err == nil {
		t.Log("Sleep 660")
		time.Sleep(25 * 10 * time.Second)
	}
	t.Log("Stopping openzwave")
	app.Stop()
}
