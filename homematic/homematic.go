package homematic

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Client represents a HomeMatic XML-API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new HomeMatic XML-API client
func NewClient(baseURL, token string) *Client {
	// a http client that uses insecure TLS settings
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}

	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: client,
	}
}

// Device represents a HomeMatic device
type Device struct {
	XMLName     xml.Name  `xml:"device"`
	Name        string    `xml:"name,attr"`
	Address     string    `xml:"address,attr"`
	IseID       string    `xml:"ise_id,attr"`
	Unreach     bool      `xml:"unreach,attr"`
	Config      bool      `xml:"config,attr"`
	DeviceType  string    `xml:"device_type,attr"`
	InterfaceID string    `xml:"interface_id,attr"`
	Channels    []Channel `xml:"channel"`
}

// Channel represents a device channel
type Channel struct {
	XMLName      xml.Name    `xml:"channel"`
	Name         string      `xml:"name,attr"`
	Type         string      `xml:"type,attr"`
	Address      string      `xml:"address,attr"`
	IseID        string      `xml:"ise_id,attr"`
	Direction    string      `xml:"direction,attr"`
	ParentType   string      `xml:"parent_type,attr"`
	Index        int         `xml:"index,attr"`
	GroupPartner string      `xml:"group_partner,attr"`
	AESAvailable bool        `xml:"aes_available,attr"`
	Transmission string      `xml:"transmission_mode,attr"`
	Visible      bool        `xml:"visible,attr"`
	Ready        bool        `xml:"ready_config,attr"`
	Operate      bool        `xml:"operate,attr"`
	DataPoints   []DataPoint `xml:"datapoint"`
}

// DataPoint represents a channel data point
type DataPoint struct {
	XMLName   xml.Name `xml:"datapoint"`
	Name      string   `xml:"name,attr"`
	Type      string   `xml:"type,attr"`
	IseID     string   `xml:"ise_id,attr"`
	Value     string   `xml:"value,attr"`
	ValueType int      `xml:"valuetype,attr"`
	ValueUnit string   `xml:"valueunit,attr"`
	Timestamp int64    `xml:"timestamp,attr"`
}

// Program represents a HomeMatic program
type Program struct {
	XMLName     xml.Name `xml:"program"`
	ID          string   `xml:"id,attr"`
	Name        string   `xml:"name,attr"`
	Description string   `xml:"description,attr"`
	Info        string   `xml:"info,attr"`
	Visible     bool     `xml:"visible,attr"`
	Active      bool     `xml:"active,attr"`
	Timestamp   int64    `xml:"timestamp,attr"`
}

// Room represents a HomeMatic room
type Room struct {
	XMLName  xml.Name  `xml:"room"`
	Name     string    `xml:"name,attr"`
	IseID    string    `xml:"ise_id,attr"`
	Channels []Channel `xml:"channel"`
}

// Function represents a HomeMatic function
type Function struct {
	XMLName  xml.Name  `xml:"function"`
	Name     string    `xml:"name,attr"`
	IseID    string    `xml:"ise_id,attr"`
	Channels []Channel `xml:"channel"`
}

// SystemVariable represents a HomeMatic system variable
type SystemVariable struct {
	XMLName    xml.Name `xml:"systemVariable"`
	Name       string   `xml:"name,attr"`
	Variable   string   `xml:"variable,attr"`
	Value      string   `xml:"value,attr"`
	ValueType  int      `xml:"valuetype,attr"`
	IseID      string   `xml:"ise_id,attr"`
	Min        string   `xml:"min,attr"`
	Max        string   `xml:"max,attr"`
	Unit       string   `xml:"unit,attr"`
	Type       string   `xml:"type,attr"`
	Subtype    string   `xml:"subtype,attr"`
	Logged     bool     `xml:"logged,attr"`
	Visible    bool     `xml:"visible,attr"`
	Timestamp  int64    `xml:"timestamp,attr"`
	ValueName0 string   `xml:"value_name_0,attr"`
	ValueName1 string   `xml:"value_name_1,attr"`
	ValueText  string   `xml:"value_text,attr"`
}

// DeviceType represents a HomeMatic device type
type DeviceType struct {
	XMLName xml.Name `xml:"deviceType"`
	Name    string   `xml:"name,attr"`
	ID      string   `xml:"id,attr"`
}

// APIResponse represents the common XML response structure
type APIResponse struct {
	XMLName         xml.Name         `xml:"result"`
	Devices         []Device         `xml:"device"`
	Programs        []Program        `xml:"program"`
	Rooms           []Room           `xml:"room"`
	Functions       []Function       `xml:"function"`
	SystemVariables []SystemVariable `xml:"systemVariable"`
	DeviceTypes     []DeviceType     `xml:"deviceType"`
	Version         string           `xml:"version"`
}

// DeviceListResponse represents the devicelist.cgi response
type DeviceListResponse struct {
	XMLName xml.Name `xml:"deviceList"`
	Devices []Device `xml:"device"`
}

// StateListResponse represents the statelist.cgi response
type StateListResponse struct {
	XMLName xml.Name `xml:"stateList"`
	Devices []Device `xml:"device"`
}

// ProgramListResponse represents the programlist.cgi response
type ProgramListResponse struct {
	XMLName  xml.Name  `xml:"programList"`
	Programs []Program `xml:"program"`
}

// RoomListResponse represents the roomlist.cgi response
type RoomListResponse struct {
	XMLName xml.Name `xml:"roomList"`
	Rooms   []Room   `xml:"room"`
}

// FunctionListResponse represents the functionlist.cgi response
type FunctionListResponse struct {
	XMLName   xml.Name   `xml:"functionList"`
	Functions []Function `xml:"function"`
}

// SystemVariableListResponse represents the sysvarlist.cgi response
type SystemVariableListResponse struct {
	XMLName         xml.Name         `xml:"systemVariables"`
	SystemVariables []SystemVariable `xml:"systemVariable"`
}

// DeviceTypeListResponse represents the devicetypelist.cgi response
type DeviceTypeListResponse struct {
	XMLName     xml.Name     `xml:"deviceTypes"`
	DeviceTypes []DeviceType `xml:"deviceType"`
}

// VersionResponse represents the version endpoint response
type VersionResponse struct {
	XMLName xml.Name `xml:"version"`
	Value   string   `xml:",chardata"`
}

// charsetReader handles different character encodings for XML parsing
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "iso-8859-1", "latin1":
		return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
	case "windows-1252", "cp1252":
		return transform.NewReader(input, charmap.Windows1252.NewDecoder()), nil
	default:
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}
}

// convertToUTF8 converts XML content to UTF-8 if needed
func convertToUTF8(data []byte) ([]byte, error) {
	// Check if already UTF-8
	if utf8.Valid(data) {
		return data, nil
	}

	// Try to detect encoding from XML declaration
	xmlDecl := string(data[:min(len(data), 200)])
	if strings.Contains(strings.ToLower(xmlDecl), "iso-8859-1") ||
		strings.Contains(strings.ToLower(xmlDecl), "latin1") {
		decoder := charmap.ISO8859_1.NewDecoder()
		utf8Data, err := decoder.Bytes(data)
		if err != nil {
			return nil, fmt.Errorf("failed to convert from ISO-8859-1: %w", err)
		}
		// Replace encoding declaration with UTF-8
		utf8String := string(utf8Data)
		utf8String = strings.ReplaceAll(utf8String, "ISO-8859-1", "UTF-8")
		utf8String = strings.ReplaceAll(utf8String, "iso-8859-1", "utf-8")
		return []byte(utf8String), nil
	}

	return data, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// makeRequest performs an HTTP request to the XML-API
func (c *Client) makeRequest(endpoint string, params map[string]string) (*APIResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/addons/xmlapi/%s", c.BaseURL, endpoint))
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("sid", c.Token)

	for key, value := range params {
		q.Set(key, value)
	}

	u.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Convert to UTF-8 if needed
	utf8Body, err := convertToUTF8(body)
	if err != nil {
		return nil, fmt.Errorf("failed to convert encoding: %w", err)
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(utf8Body))
	decoder.CharsetReader = charsetReader

	var result APIResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return &result, nil
}

// makeRawRequest performs an HTTP request and returns raw XML bytes
func (c *Client) makeRawRequest(endpoint string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(fmt.Sprintf("%s/addons/xmlapi/%s", c.BaseURL, endpoint))
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("sid", c.Token)

	for key, value := range params {
		q.Set(key, value)
	}

	u.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Convert to UTF-8 if needed
	utf8Body, err := convertToUTF8(body)
	if err != nil {
		return nil, fmt.Errorf("failed to convert encoding: %w", err)
	}

	return utf8Body, nil
}

// GetVersion returns the XML-API version
func (c *Client) GetVersion() (string, error) {
	body, err := c.makeRawRequest("version.cgi", nil)
	if err != nil {
		return "", err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var versionResp VersionResponse
	if err := decoder.Decode(&versionResp); err != nil {
		return "", fmt.Errorf("failed to parse version XML: %w", err)
	}

	return strings.TrimSpace(versionResp.Value), nil
}

// GetDeviceList returns all devices with their channels
func (c *Client) GetDeviceList(deviceIDs []string, showInternal, showRemote bool) ([]Device, error) {
	params := make(map[string]string)

	if len(deviceIDs) > 0 {
		params["device_id"] = strings.Join(deviceIDs, ",")
	}
	if showInternal {
		params["show_internal"] = "1"
	}
	if showRemote {
		params["show_remote"] = "1"
	}

	body, err := c.makeRawRequest("devicelist.cgi", params)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result DeviceListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.Devices, nil
}

// GetDeviceTypes returns all possible device types
func (c *Client) GetDeviceTypes() ([]DeviceType, error) {
	body, err := c.makeRawRequest("devicetypelist.cgi", nil)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result DeviceTypeListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.DeviceTypes, nil
}

// GetStateList returns all devices with their current values
func (c *Client) GetStateList(deviceID string, showInternal, showRemote bool) ([]Device, error) {
	params := make(map[string]string)

	if showInternal {
		params["show_internal"] = "1"
	}
	if showRemote {
		params["show_remote"] = "1"
	}

	body, err := c.makeRawRequest("statelist.cgi", params)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result StateListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if deviceID != "" {
		// Filter devices by deviceID if provided
		filteredDevices := make([]Device, 0)
		for _, device := range result.Devices {
			if device.IseID == deviceID {
				filteredDevices = append(filteredDevices, device)
			}
		}
		return filteredDevices, nil
	}

	return result.Devices, nil
}

// GetState returns specific devices/channels with their current values
func (c *Client) GetState(deviceIDs, channelIDs, datapointIDs []string) ([]Device, error) {
	params := make(map[string]string)

	if len(deviceIDs) > 0 {
		params["device_id"] = strings.Join(deviceIDs, ",")
	}
	if len(channelIDs) > 0 {
		params["channel_id"] = strings.Join(channelIDs, ",")
	}
	if len(datapointIDs) > 0 {
		params["datapoint_id"] = strings.Join(datapointIDs, ",")
	}

	body, err := c.makeRawRequest("state.cgi", params)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result StateListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.Devices, nil
}

// ChangeState changes the state of one or more devices
func (c *Client) ChangeState(deviceIDs, newValues []string) error {
	if len(deviceIDs) != len(newValues) {
		return fmt.Errorf("device IDs and new values must have the same length")
	}

	params := map[string]string{
		"ise_id":    strings.Join(deviceIDs, ","),
		"new_value": strings.Join(newValues, ","),
	}

	_, err := c.makeRequest("statechange.cgi", params)
	return err
}

// GetProgramList returns all programs
func (c *Client) GetProgramList() ([]Program, error) {
	body, err := c.makeRawRequest("programlist.cgi", nil)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result ProgramListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.Programs, nil
}

// RunProgram starts a program with the specified ID
func (c *Client) RunProgram(programID string, condCheck bool) error {
	params := map[string]string{
		"program_id": programID,
	}
	if condCheck {
		params["cond_check"] = "1"
	}

	_, err := c.makeRequest("runprogram.cgi", params)
	return err
}

// ChangeProgramActions modifies program active/visible status
func (c *Client) ChangeProgramActions(programID string, active, visible *bool) error {
	params := map[string]string{
		"program_id": programID,
	}

	if active != nil {
		params["active"] = strconv.FormatBool(*active)
	}
	if visible != nil {
		params["visible"] = strconv.FormatBool(*visible)
	}

	_, err := c.makeRequest("programactions.cgi", params)
	return err
}

// GetRoomList returns all configured rooms including channels
func (c *Client) GetRoomList() ([]Room, error) {
	body, err := c.makeRawRequest("roomlist.cgi", nil)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result RoomListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.Rooms, nil
}

// GetFunctionList returns all functions including channels
func (c *Client) GetFunctionList() ([]Function, error) {
	body, err := c.makeRawRequest("functionlist.cgi", nil)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result FunctionListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.Functions, nil
}

// GetSystemVariableList returns all system variables
func (c *Client) GetSystemVariableList(showText bool) ([]SystemVariable, error) {
	params := make(map[string]string)
	if showText {
		params["text"] = "true"
	} else {
		params["text"] = "false"
	}

	body, err := c.makeRawRequest("sysvarlist.cgi", params)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result SystemVariableListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.SystemVariables, nil
}

// GetSystemVariable returns a single system variable
func (c *Client) GetSystemVariable(iseID string, showText bool) (*SystemVariable, error) {
	params := map[string]string{
		"ise_id": iseID,
	}
	if showText {
		params["text"] = "true"
	} else {
		params["text"] = "false"
	}

	body, err := c.makeRawRequest("sysvar.cgi", params)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result SystemVariableListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	if len(result.SystemVariables) == 0 {
		return nil, fmt.Errorf("system variable not found")
	}

	return &result.SystemVariables[0], nil
}

// RegisterToken registers a new security access token
func (c *Client) RegisterToken(description string) error {
	params := map[string]string{
		"desc": description,
	}

	_, err := c.makeRequest("tokenregister.cgi", params)
	return err
}

// RevokeToken revokes an existing security access token
func (c *Client) RevokeToken(tokenID string) error {
	params := map[string]string{
		"sid": tokenID,
	}

	_, err := c.makeRequest("tokenrevoke.cgi", params)
	return err
}

// GetMasterValue outputs devices with their master values
func (c *Client) GetMasterValue(deviceIDs, requestedNames []string) ([]Device, error) {
	params := make(map[string]string)

	if len(deviceIDs) > 0 {
		params["device_id"] = strings.Join(deviceIDs, ",")
	}
	if len(requestedNames) > 0 {
		params["requested_names"] = strings.Join(requestedNames, ",")
	}

	body, err := c.makeRawRequest("mastervalue.cgi", params)
	if err != nil {
		return nil, err
	}

	// Create XML decoder with charset reader support
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charsetReader

	var result DeviceListResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return result.Devices, nil
}

// ChangeMasterValue sets master values for devices
func (c *Client) ChangeMasterValue(deviceIDs, names, values []string) error {
	if len(deviceIDs) != len(names) || len(names) != len(values) {
		return fmt.Errorf("device IDs, names, and values must have the same length")
	}

	params := map[string]string{
		"device_id": strings.Join(deviceIDs, ","),
		"name":      strings.Join(names, ","),
		"value":     strings.Join(values, ","),
	}

	_, err := c.makeRequest("mastervaluechange.cgi", params)
	return err
}

// Example usage function
func ExampleUsage() {
	// Create a new client
	client := NewClient("https://192.168.1.100", "your-token-here")

	// Get API version
	version, err := client.GetVersion()
	if err != nil {
		fmt.Printf("Error getting version: %v\n", err)
		return
	}
	fmt.Printf("XML-API Version: %s\n", version)

	// Get all devices
	devices, err := client.GetDeviceList(nil, false, false)
	if err != nil {
		fmt.Printf("Error getting devices: %v\n", err)
		return
	}

	for _, device := range devices {
		fmt.Printf("Device: %s (ID: %s, Type: %s)\n",
			device.Name, device.IseID, device.DeviceType)
	}

	// Change device state (example: set dimmer to 20%)
	err = client.ChangeState([]string{"12345"}, []string{"0.20"})
	if err != nil {
		fmt.Printf("Error changing state: %v\n", err)
		return
	}

	// Run a program
	err = client.RunProgram("1234", false)
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return
	}

	// Get system variables
	sysVars, err := client.GetSystemVariableList(true)
	if err != nil {
		fmt.Printf("Error getting system variables: %v\n", err)
		return
	}

	for _, sysVar := range sysVars {
		fmt.Printf("System Variable: %s = %s\n", sysVar.Name, sysVar.Value)
	}
}

// Installation instructions:
// To use this client, you need to install the required dependency:
// go get golang.org/x/text
