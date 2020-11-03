package apple

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/widuu/apple/plist"
)

const (
	profileListUrl  = "https://developerservices2.apple.com/services/v1/profiles"
	profileRegenUrl = "https://developerservices2.apple.com/services/QH65B2/ios/regenProvisioningProfile.action?clientId=XABBG36SBA"
	profileAddUrl   = "https://developerservices2.apple.com/services/QH65B2/ios/createProvisioningProfile.action?clientId=XABBG36SBA"
	profileDownUrl  = "https://developerservices2.apple.com/services/QH65B2/ios/downloadTeamProvisioningProfile.action?clientId=XABBG36SBA"
)

type ProfileData struct {
	Stype      string            `json:"type"`
	Id         string            `json:"id"`
	Attributes map[string]string `json:"attributes"`
}

type profileListResponse struct {
	// Error []map[string]string `json:"errors"`
	Error []struct {
		Id     string `json:"id"`
		Status string `json:"status"`
		Detail string `json:"detail"`
	} `json:"errors"`
	Data []ProfileData `json:"data"`
}

type createProfileRequest struct {
	TeamID                  string   `plist:"teamId"`
	BundleId                string   `plist:"appIdId"`
	Devices                 []string `plist:"deviceIds"`
	Certs                   []string `plist:"certificateIds"`
	DistributionType        string   `plist:"distributionType"`
	ProvisioningProfileName string   `plist:"provisioningProfileName"`
	ClientID                string   `plist:"clientId"`
	Myacinfo                string   `plist:"myacinfo"`
	ProtocolVersion         string   `plist:"protocolVersion"`
	UserLocale              []string `plist:"userLocale"`
}

type ProvisioningProfile struct {
	ProvisioningProfileId string `plist:"provisioningProfileId"`
	Name                  string `plist:"name"`
	Status                string `plist:"status"`
	ProvisioningType      string `plist:"type"`
	DistributionMethod    string `plist:"distributionMethod"`
	ProProPlatform        string `plist:"proProPlatform"`
	UUID                  string `plist:"UUID"`
	Filename              string `plist:"filename"`
	ProfileContent        string `plist:"encodedProfile"`
}

type createProfileResponse struct {
	UserString string              `plist:"userString"`
	Code       int                 `plist:"resultCode"`
	Profile    ProvisioningProfile `plist:"provisioningProfile"`
	ResponseID string              `plist:"responseId"`
}

// CreateProfile
func CreateProfile(provisioningProfileName, bundleId, distributionType string, certs, devices []string, isRegen bool, teamId, myacinfo string) (ProvisioningProfile, error) {
	requestParams := createProfileRequest{
		TeamID:                  teamId,
		BundleId:                bundleId,
		Devices:                 devices,
		Certs:                   certs,
		DistributionType:        distributionType,
		ProvisioningProfileName: provisioningProfileName,
		ClientID:                ClientID,
		Myacinfo:                myacinfo,
		ProtocolVersion:         ProtocolVersion,
		UserLocale:              strings.Split(UserLocale, "#"),
	}

	encoder, err := plist.MarshalIndent(requestParams, "   ")
	if err != nil {
		return ProvisioningProfile{}, err
	}

	// api url
	var apiUrl string
	if isRegen {
		apiUrl = profileRegenUrl
	} else {
		apiUrl = profileAddUrl
	}

	request := NewClientRequest(apiUrl, "POST")
	RequestHeader["Cookie"] = "myacinfo=" + myacinfo
	body, _, err := request.SetHeader(RequestHeader).SetBody(encoder).GetBody()

	var data createProfileResponse
	err = plist.Unmarshal(body, &data)

	if err != nil {
		if data.Code == 0 {
			return data.Profile, nil
		}
		return ProvisioningProfile{}, err
	}

	if data.Code != 0 {
		return ProvisioningProfile{}, errors.New(data.UserString)
	}

	return data.Profile, nil
}

// ProfileLists get file
func ProfileLists(customSearch map[string]string, teamId, myacinfo string) ([]ProfileData, error) {
	search := BuildSearchQueryString(teamId, customSearch)

	responseParams := struct {
		UrlEncodedQueryParams string `json:"urlEncodedQueryParams"`
	}{
		UrlEncodedQueryParams: search,
	}

	postJson, err := json.Marshal(responseParams)
	if err != nil {
		return []ProfileData{}, err
	}

	// request
	request := NewClientRequest(profileListUrl, "POST")
	JSONRequestHeader["Cookie"] = "myacinfo=" + myacinfo
	JSONRequestHeader["X-HTTP-Method-Override"] = "GET"
	body, _, err := request.SetHeader(JSONRequestHeader).SetBody(postJson).GetBody()
	if err != nil {
		return []ProfileData{}, err
	}

	var data profileListResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return []ProfileData{}, err
	}
	if len(data.Error) > 0 {
		return []ProfileData{}, errors.New(data.Error[0].Detail)
	}

	if len(data.Data) <= 0 {
		return []ProfileData{}, errors.New("Profile does not exist")
	}

	return data.Data, nil
}

// GetProfileContent return base64 encode profileContent
func GetProfileContent(profileId, teamId, myacinfo string) (string, error) {
	profileList, err := ProfileLists(map[string]string{
		"id":    profileId,
		"limit": "1",
	}, teamId, myacinfo)

	if err != nil {
		return "", err
	}

	if len(profileList) <= 0 {
		return "", errors.New("Profile does not exists")
	}

	profileContent := profileList[0].Attributes["profileContent"]

	return profileContent, nil
}
