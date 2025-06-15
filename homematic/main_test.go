package homematic

import (
	"testing"
)

// Example usage function
func TestExampleUsage(t *testing.T) {
	// Create a new client
	client := NewClient("https://192.168.2.160", "your-token-here")

	// Get API version
	version, err := client.GetVersion()
	if err != nil {
		t.Errorf("Error getting version: %v\n", err)
		return
	}
	t.Logf("XML-API Version: %s\n", version)

	// Get all devices
	devices, err := client.GetDeviceList(nil, false, false)
	if err != nil {
		t.Errorf("Error getting devices: %v\n", err)
		return
	}

	for _, device := range devices {
		t.Logf("Device: %s (ID: %s, Type: %s)\n",
			device.Name, device.IseID, device.DeviceType)
	}

	// Change device state (example: set dimmer to 20%)
	err = client.ChangeState([]string{"12345"}, []string{"0.20"})
	if err != nil {
		t.Errorf("Error changing state: %v\n", err)
		return
	}

	// Run a program
	err = client.RunProgram("1234", false)
	if err != nil {
		t.Errorf("Error running program: %v\n", err)
		return
	}

	// Get system variables
	sysVars, err := client.GetSystemVariableList(true)
	if err != nil {
		t.Errorf("Error getting system variables: %v\n", err)
		return
	}

	for _, sysVar := range sysVars {
		t.Logf("System Variable: %s = %s\n", sysVar.Name, sysVar.Value)
	}
}
