package apple

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/widuu/apple/plist"
)

const (
	deviceListURL = "https://developerservices2.apple.com/services/QH65B2/ios/listDevices.action?clientId=XABBG36SBA"
	deviceAddURL  = "https://developerservices2.apple.com/services/QH65B2/ios/addDevice.action"
	deviceAPIURL  = "https://developerservices2.apple.com/services/v1/devices"
)

type deviceListParams struct {
	ClientID        string   `plist:"clientId"`
	Myacinfo        string   `plist:"myacinfo"`
	ProtocolVersion string   `plist:"protocolVersion"`
	UserLocale      []string `plist:"userLocale"`
	TeamID          string   `plist:"teamId"`
	PageNumber      int      `plist:"pageNumber"`
	PageSize        int      `plist:"pageSize"`
}

// DeviceListRequest Devices is device list
type DeviceListRequest struct {
	Devices      []map[string]string `plist:"devices"`
	Error        string              `plist:"userString"`
	Code         int                 `plist:"resultCode"`
	PageNumber   int                 `plist:"pageNumber"`
	PageSize     int                 `plist:"pageSize"`
	TotalRecords int                 `plist:"totalRecords"`
}

type deviceData struct {
	Stype      string            `json:"type"`
	Id         string            `json:"id"`
	Attributes map[string]string `json:"attributes"`
}

type findDeviceRequest struct {
	Error []map[string]string `json:"errors"`
	Data  []deviceData        `json:"data"`
}

type addDeviceParams struct {
	Name            string   `plist:"name"`
	Udid            string   `plist:"deviceNumber"`
	TeamID          string   `plist:"teamId"`
	ClientID        string   `plist:"clientId"`
	Myacinfo        string   `plist:"myacinfo"`
	ProtocolVersion string   `plist:"protocolVersion"`
	UserLocale      []string `plist:"userLocale"`
}

type addDeviceBody struct {
	Code   int               `plist:"resultCode"`
	Error  string            `plist:"userString"`
	Device map[string]string `plist:"device"`
}

// AddDevice add device to developer apple
func AddDevice(deviceName, deviceNumber, teamID, myacinfo string) (map[string]string, error) {
	params := addDeviceParams{
		Name:            deviceName,
		Udid:            deviceNumber,
		TeamID:          teamID,
		ClientID:        ClientID,
		Myacinfo:        myacinfo,
		ProtocolVersion: ProtocolVersion,
		UserLocale:      strings.Split(UserLocale, "#"),
	}
	encoder, _ := plist.MarshalIndent(params, "   ")

	RequestHeader["Cookie"] = "myacinfo=" + myacinfo

	request := NewClientRequest(deviceAddURL, "POST")
	body, _, err := request.SetHeader(RequestHeader).SetBody(encoder).GetBody()

	var data addDeviceBody
	err = plist.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}

	if data.Code != 0 {
		return nil, errors.New(data.Error)
	}

	return data.Device, nil
}

// DeviceLists list devices and return DeviceListRequest
func DeviceLists(teamID, myacinfo string, pageNumber, pageSize int) (DeviceListRequest, error) {
	params := deviceListParams{ClientID, myacinfo, ProtocolVersion, strings.Split(UserLocale, "#"), teamID, pageNumber, pageSize}
	encoder, _ := plist.MarshalIndent(params, "   ")
	request := NewClientRequest(deviceListURL, "POST")
	RequestHeader["Cookie"] = "myacinfo=" + myacinfo
	body, _, err := request.SetHeader(RequestHeader).SetBody(encoder).GetBody()
	if err != nil {
		return DeviceListRequest{}, err
	}
	var data DeviceListRequest
	err = plist.Unmarshal(body, &data)
	if err != nil {
		return DeviceListRequest{}, err
	}

	if data.Code != 0 {
		return DeviceListRequest{}, errors.New(data.Error)
	}

	return data, nil
}

// GetDeviceTotal device total
func GetDeviceTotal(teamID, myacinfo string) (map[string]int, error) {
	requestData, err := DeviceLists(teamID, myacinfo, 0, 500)
	if err != nil {
		return nil, err
	}

	res := map[string]int{"iphone": 0, "ipad": 0, "ipod": 0, "tv": 0, "watch": 0}

	devices := requestData.Devices
	for _, device := range devices {
		res[device["deviceClass"]] += 1
	}

	return res, nil
}

// FindDevice use udid to find device infomation
func FindDevice(udid, teamID, myacinfo string) (map[string]string, error) {
	responseParams := struct {
		UrlEncodedQueryParams string `json:"urlEncodedQueryParams"`
	}{
		UrlEncodedQueryParams: fmt.Sprintf("teamId=%s&filter[udid]=%s&limit=1", teamID, udid),
	}
	postJson, err := json.Marshal(responseParams)
	if err != nil {
		return nil, err
	}
	// request
	request := NewClientRequest(deviceAPIURL, "POST")
	JSONRequestHeader["Cookie"] = "myacinfo=" + myacinfo
	JSONRequestHeader["X-HTTP-Method-Override"] = "GET"
	body, _, err := request.SetHeader(JSONRequestHeader).SetBody(postJson).GetBody()

	var data findDeviceRequest
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	if len(data.Error) > 0 {
		return nil, errors.New(data.Error[0]["detail"])
	}

	if len(data.Data) <= 0 {
		return map[string]string{}, errors.New("Device does not exist")
	}

	res := map[string]string{"id": data.Data[0].Id, "type": data.Data[0].Stype}
	for k, v := range data.Data[0].Attributes {
		res[k] = v
	}

	return res, nil
}
