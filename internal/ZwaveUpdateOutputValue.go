// Package internal for discovery of node input and output values
package internal

import (
	"github.com/jimjibone/goopenzwave"
	"github.com/sirupsen/logrus"
)

// ZWaveUpdateOutputValue  is invoked by OZW when an output's value is updated
// This updates the corresponding IoTDomain output, which will publish the new output value if it has
// changed. Output value updates are not published immediately but during the publication cyle.
// Unknown output types and blacklisted output types are ignored.
func (app *OpenZWaveApp) ZWaveUpdateOutputValue(zwValue *goopenzwave.ValueID) {
	// Does updateValue get called with cached values?
	zwValueString := zwValue.GetAsString()

	// unknown and blacklisted types don't exist in this table and are ignored
	outputID := app.outputIDByValueID[zwValue.ID]
	if outputID != "" {
		output := app.pub.GetOutputByID(outputID)
		if output != nil {
			// this is an update of an output value
			oldValue := app.pub.GetOutputValueByID(outputID)
			if oldValue != nil {
				//adapter.Logger().Infof("updateValue: NodeInOutput value of %s, old value=%s, new value = %v", zwValueLabel, oldValue, zwValueString)
				logrus.Infof("ZWaveUpdateOutputValue: Output value update. outputID='%s' old value=%s, new value=%s",
					outputID, oldValue.Value, zwValueString)
			}
			app.pub.UpdateOutputValue(output.NodeHWID, output.OutputType, output.Instance, zwValueString)
		} else {
			logrus.Errorf("ZWaveUpdateOutputValue: Output for outputID '%s' not found. This should never happen", outputID)
		}
	} else {
		logrus.Infof("ZWaveUpdateOutputValue: Ignored unknown output value. zwValueID='%d'. Value=%s",
			zwValue.ID, zwValueString)
		// the debugger stops here even though there is an outputID???
		if outputID != "" {
			panic("OutputID exists, should never get here")
		}
		_ = outputID
	}
}
