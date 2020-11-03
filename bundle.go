package apple

import (
	// "fmt"
	"encoding/json"
	"errors"
	"strings"

	"github.com/widuu/apple/plist"
)

const (
	bundleListURL = "https://developerservices2.apple.com/services/QH65B2/ios/listAppIds.action?clientId=XABBG36SBA"
	bundleAddURL  = "https://developerservices2.apple.com/services/QH65B2/ios/addAppId.action?clientId=XABBG36SBA"
	bundleAPIURL  = "https://developerservices2.apple.com/services/v1/bundleIdCapabilities"
)

type bundleListParams struct {
	ClientID        string   `plist:"clientId"`
	Myacinfo        string   `plist:"myacinfo"`
	ProtocolVersion string   `plist:"protocolVersion"`
	UserLocale      []string `plist:"userLocale"`
	TeamID          string   `plist:"teamId"`
	PageNumber      int      `plist:"pageNumber"`
	PageSize        int      `plist:"pageSize"`
	RequestID       string   `plist:"requestId"`
}

type addBundleParams struct {
	ClientID        string   `plist:"clientId"`
	Myacinfo        string   `plist:"myacinfo"`
	ProtocolVersion string   `plist:"protocolVersion"`
	UserLocale      []string `plist:"userLocale"`
	TeamID          string   `plist:"teamId"`
	Identifier      string   `plist:"identifier"`
	Name            string   `plist:"name"`
	Platform        string   `plist:"appIdPlatform"`
	RequestID       string   `plist:"requestId"`
}

// AppId
type AppId struct {
	AppIDID    string `plist:"appIdId"`
	NAME       string `plist:"name"`
	Platform   string `plist:"appIdPlatform"`
	Prefix     string `plist:"prefix"`
	Identifier string `plist:"identifier"`
}

// AppIds
type AppIds struct {
	AppIds       []AppId `plist:"appIds"`
	UserString   string  `plist:"userString"`
	Code         int     `plist:"resultCode"`
	PageNumber   int     `plist:"pageNumber"`
	PageSize     int     `plist:"pageSize"`
	TotalRecords int     `plist:"totalRecords"`
}

// AppIds
type AddAppId struct {
	AppIds       AppId  `plist:"appId"`
	UserString   string `plist:"userString"`
	Code         int    `plist:"resultCode"`
	PageNumber   int    `plist:"pageNumber"`
	PageSize     int    `plist:"pageSize"`
	TotalRecords int    `plist:"totalRecords"`
}

type relationShipsData struct {
	Data map[string]string `json:"data"`
}

type postRelationShips struct {
	BundleID   relationShipsData `json:"bundleId"`
	Capability relationShipsData `json:"capability"`
}

type postData struct {
	Stype         string            `json:"type"`
	Attributes    map[string]string `json:"attributes"`
	Relationships postRelationShips `json:"relationships"`
}

type postParams struct {
	Data postData `json:"data"`
}

type apiData struct {
	Stype         string            `json:"type"`
	ID            string            `json:"id"`
	Relationships postRelationShips `json:"relationships"`
}

type apiRequest struct {
	Error []map[string]string `json:"errors"`
	Data  apiData             `json:"data"`
}

// Capabilities
func Capabilities(teamID, bundleID, capabilityType, myacinfo string) ([]map[string]string, error) {
	params := postParams{
		Data: postData{
			Stype: "bundleIdCapabilities",
			Attributes: map[string]string{
				"teamId":         teamID,
				"capabilityType": capabilityType,
			},
			Relationships: postRelationShips{
				BundleID: relationShipsData{
					Data: map[string]string{
						"type": "bundleIds",
						"id":   bundleID,
					},
				},
				Capability: relationShipsData{
					Data: map[string]string{
						"type": "capabilities",
						"id":   capabilityType,
					},
				},
			},
		},
	}

	postJson, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	// request
	request := NewClientRequest(bundleAPIURL, "POST")
	JSONRequestHeader["Cookie"] = "myacinfo=" + myacinfo
	if _, ok := JSONRequestHeader["X-HTTP-Method-Override"]; ok {
		delete(JSONRequestHeader, "X-HTTP-Method-Override")
	}
	body, _, err := request.SetHeader(JSONRequestHeader).SetBody(postJson).GetBody()
	if err != nil {
		return nil, err
	}
	var data apiRequest
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	if len(data.Error) > 0 {
		return nil, errors.New(data.Error[0]["detail"])
	}
	res := []map[string]string{
		map[string]string{"id": data.Data.ID, "type": data.Data.Stype},
		data.Data.Relationships.BundleID.Data,
		data.Data.Relationships.Capability.Data,
	}
	return res, nil
}

// addBundleID
func AddBundleID(name, bundleId, platform, myacinfo, teamId string) (AddAppId, error) {
	requestId := GenerateUDID()
	params := addBundleParams{ClientID, myacinfo, ProtocolVersion, strings.Split(UserLocale, "#"), teamId, bundleId, name, platform, requestId}
	encoder, _ := plist.MarshalIndent(params, "   ")
	request := NewClientRequest(bundleAddURL, "POST")
	RequestHeader["Cookie"] = "myacinfo=" + myacinfo
	body, _, err := request.SetHeader(RequestHeader).SetBody(encoder).GetBody()
	if err != nil {
		return AddAppId{}, err
	}
	var app AddAppId
	err = plist.Unmarshal(body, &app)
	if err != nil {
		return AddAppId{}, err
	}
	if app.UserString != "" && app.Code != 0 {
		return AddAppId{}, errors.New(app.UserString)
	}
	return app, nil
}

// GetBundleLists lists
func GetBundleLists(myacinfo, teamId string, pageNumber, pageSize int) (AppIds, error) {
	requestId := GenerateUDID()
	params := bundleListParams{ClientID, myacinfo, ProtocolVersion, strings.Split(UserLocale, "#"), teamId, pageNumber, pageSize, requestId}
	encoder, _ := plist.MarshalIndent(params, "   ")
	request := NewClientRequest(bundleListURL, "POST")
	RequestHeader["Cookie"] = "myacinfo=" + myacinfo
	body, _, err := request.SetHeader(RequestHeader).SetBody(encoder).GetBody()
	if err != nil {
		return AppIds{}, err
	}
	var list AppIds
	err = plist.Unmarshal(body, &list)
	if err != nil {
		return AppIds{}, err
	}
	if list.UserString != "" && list.Code != 0 {
		return AppIds{}, errors.New(list.UserString)
	}
	return list, nil
}
