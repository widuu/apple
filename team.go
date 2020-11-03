package apple

import (
	"errors"
	"strings"

	"github.com/widuu/apple/plist"
)

const teamURL = "https://developerservices2.apple.com/services/QH65B2/listTeams.action?clientId=XABBG36SBA"

type requestParams struct {
	ClientID        string   `plist:"clientId"`
	Myacinfo        string   `plist:"myacinfo"`
	ProtocolVersion string   `plist:"protocolVersion"`
	UserLocale      []string `plist:"userLocale"`
	RequestID       string   `plist:"requestId"`
}

type team struct {
	Status string `plist:"status"`
	Name   string `plist:"name"`
	Type   string `plist:"type"`
	TeamID string `plist:"teamId"`
}

type teamResp struct {
	UserString string `plist:"userString"`
	Code       int    `plist:"resultCode"`
	Teams      []team `plist:"teams"`
	ResponseID string `plist:"responseId"`
}

// GetTeam team info
func GetTeam(myacinfo string) ([]byte, error) {
	requestId := GenerateUDID()
	params := requestParams{ClientID, myacinfo, ProtocolVersion, strings.Split(UserLocale, "#"), requestId}
	encoder, _ := plist.MarshalIndent(params, "   ")
	request := NewClientRequest(teamURL, "POST")
	RequestHeader["Cookie"] = "myacinfo=" + myacinfo
	body, _, err := request.SetHeader(RequestHeader).SetBody(encoder).GetBody()
	if err != nil {
		return nil, err
	}
	return body, nil
}

// GetTeamID - Get the Apple Developer ID of the current account
func GetTeamID(myacinfo string) (string, error) {
	body, err := GetTeam(myacinfo)
	if err != nil {
		return "", err
	}
	var teamlist teamResp
	err = plist.Unmarshal(body, &teamlist)
	if err != nil {
		return "", err
	}

	if teamlist.Code != 0 {
		return "", errors.New(teamlist.UserString)
	}

	// get team id
	if len(teamlist.Teams) > 0 {
		var t []team
		t = teamlist.Teams
		for _, v := range t {
			if v.Status == "active" {
				return v.TeamID, nil
			}
		}
	}
	return "", errors.New("team id not found")
}
