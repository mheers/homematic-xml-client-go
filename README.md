# HomeMatic XML-API Go Client

A comprehensive Go client library for the HomeMatic XML-API, providing easy access to HomeMatic Central Control Unit (CCU) devices, programs, system variables, and more.

## Features

- 🏠 **Complete Device Management** - List, query, and control HomeMatic devices
- 📊 **State Management** - Read and write device states and data points
- 🔧 **Program Control** - Execute and manage HomeMatic programs
- 📋 **System Variables** - Access and modify system variables
- 🏢 **Room & Function Organization** - Work with rooms and functional groups
- 🔐 **Secure Authentication** - Token-based authentication support
- 🌐 **Encoding Support** - Handles ISO-8859-1 and UTF-8 character encodings
- 🛡️ **TLS Support** - Works with self-signed certificates (InsecureSkipVerify)

## Installation

```bash
go get github.com/mheers/homematic-xml-client-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/mheers/homematic-xml-client-go/homematic"
)

func main() {
    // Create a new client
    client := homematic.NewClient("https://192.168.1.100", "your-security-token")
    
    // Get API version
    version, err := client.GetVersion()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("XML-API Version: %s\n", version)
    
    // List all devices
    devices, err := client.GetDeviceList(nil, false, false)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, device := range devices {
        fmt.Printf("Device: %s (ID: %s, Type: %s)\n", 
            device.Name, device.IseID, device.DeviceType)
    }
}
```

## API Overview

### Client Creation

```go
client := homematic.NewClient("https://your-ccu-ip", "your-token")
```

### Device Operations

```go
// Get all devices
devices, err := client.GetDeviceList(nil, false, false)

// Get device states
states, err := client.GetStateList("", false, false)

// Change device state (e.g., dimmer to 50%)
err = client.ChangeState([]string{"device-id"}, []string{"0.5"})

// Get specific device state
deviceStates, err := client.GetState([]string{"device-id"}, nil, nil)
```

### Program Management

```go
// List all programs
programs, err := client.GetProgramList()

// Execute a program
err = client.RunProgram("program-id", false)

// Modify program status
active := true
visible := false
err = client.ChangeProgramActions("program-id", &active, &visible)
```

### System Variables

```go
// Get all system variables
sysVars, err := client.GetSystemVariableList(true)

// Get specific system variable
sysVar, err := client.GetSystemVariable("variable-id", true)
```

### Rooms and Functions

```go
// Get all rooms
rooms, err := client.GetRoomList()

// Get all functions
functions, err := client.GetFunctionList()
```

## Data Structures

### Device
Represents a HomeMatic device with channels and data points:

```go
type Device struct {
    Name        string
    Address     string
    IseID       string
    DeviceType  string
    Channels    []Channel
    // ... other fields
}
```

### Channel
Represents a device channel with data points:

```go
type Channel struct {
    Name       string
    Type       string
    Address    string
    IseID      string
    DataPoints []DataPoint
    // ... other fields
}
```

### DataPoint
Represents a channel data point:

```go
type DataPoint struct {
    Name      string
    Type      string
    IseID     string
    Value     string
    ValueType int
    Timestamp int64
    // ... other fields
}
```

## Authentication

The HomeMatic XML-API requires authentication via security tokens. You can manage tokens using:

```go
// Register a new token
err = client.RegisterToken("My Go Application")

// Revoke an existing token
err = client.RevokeToken("token-to-revoke")
```

## Error Handling

The library provides detailed error information:

```go
devices, err := client.GetDeviceList(nil, false, false)
if err != nil {
    // Handle specific error types
    switch {
    case strings.Contains(err.Error(), "HTTP error"):
        // Handle HTTP errors
    case strings.Contains(err.Error(), "failed to parse XML"):
        // Handle XML parsing errors
    default:
        // Handle other errors
    }
}
```

## Character Encoding

The library automatically handles different character encodings commonly used by HomeMatic systems:

- UTF-8 (preferred)
- ISO-8859-1 / Latin1
- Windows-1252

XML responses are automatically converted to UTF-8 for consistent handling.

## TLS Configuration

The client is configured to work with self-signed certificates by default, which is common in HomeMatic installations:

```go
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    },
    Timeout: 30 * time.Second,
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built for the HomeMatic XML-API ecosystem
- Inspired by the need for a robust Go client for home automation
- Community-driven development

## Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/mheers/homematic-xml-client-go/issues) page
2. Create a new issue with detailed information about your problem
3. Include your HomeMatic CCU version and Go version

---

**Note**: This is a third-party library and is not officially affiliated with eQ-3 or HomeMatic.