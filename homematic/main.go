package homematic

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client represents a HomeMatic XML-API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new HomeMatic XML-API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

	var result APIResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return &result, nil
}

// GetVersion returns the XML-API version
func (c *Client) GetVersion() (string, error) {
	resp, err := c.makeRequest("version.cgi", nil)
	if err != nil {
		return "", err
	}
	return resp.Version, nil
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

	resp, err := c.makeRequest("devicelist.cgi", params)
	if err != nil {
		return nil, err
	}
	return resp.Devices, nil
}

// GetDeviceTypes returns all possible device types
func (c *Client) GetDeviceTypes() ([]DeviceType, error) {
	resp, err := c.makeRequest("devicetypelist.cgi", nil)
	if err != nil {
		return nil, err
	}
	return resp.DeviceTypes, nil
}

// GetStateList returns all devices with their current values
func (c *Client) GetStateList(deviceID string, showInternal, showRemote bool) ([]Device, error) {
	params := make(map[string]string)

	if deviceID != "" {
		params["ise_id"] = deviceID
	}
	if showInternal {
		params["show_internal"] = "1"
	}
	if showRemote {
		params["show_remote"] = "1"
	}

	resp, err := c.makeRequest("statelist.cgi", params)
	if err != nil {
		return nil, err
	}
	return resp.Devices, nil
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

	resp, err := c.makeRequest("state.cgi", params)
	if err != nil {
		return nil, err
	}
	return resp.Devices, nil
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
	resp, err := c.makeRequest("programlist.cgi", nil)
	if err != nil {
		return nil, err
	}
	return resp.Programs, nil
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
	resp, err := c.makeRequest("roomlist.cgi", nil)
	if err != nil {
		return nil, err
	}
	return resp.Rooms, nil
}

// GetFunctionList returns all functions including channels
func (c *Client) GetFunctionList() ([]Function, error) {
	resp, err := c.makeRequest("functionlist.cgi", nil)
	if err != nil {
		return nil, err
	}
	return resp.Functions, nil
}

// GetSystemVariableList returns all system variables
func (c *Client) GetSystemVariableList(showText bool) ([]SystemVariable, error) {
	params := make(map[string]string)
	if showText {
		params["text"] = "true"
	} else {
		params["text"] = "false"
	}

	resp, err := c.makeRequest("sysvarlist.cgi", params)
	if err != nil {
		return nil, err
	}
	return resp.SystemVariables, nil
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

	resp, err := c.makeRequest("sysvar.cgi", params)
	if err != nil {
		return nil, err
	}

	if len(resp.SystemVariables) == 0 {
		return nil, fmt.Errorf("system variable not found")
	}

	return &resp.SystemVariables[0], nil
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

	resp, err := c.makeRequest("mastervalue.cgi", params)
	if err != nil {
		return nil, err
	}
	return resp.Devices, nil
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
