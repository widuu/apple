package apple

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

const apiURL = "https://idmsa.apple.com/IDMSWebAuth/clientDAW.cgi"

var requestHeader map[string]string = map[string]string{
	"Accept":          "application/vnd.api+json",
	"Content-Type":    "application/x-www-form-urlencoded",
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36",
	"Accept-Language": "en-us",
}

// GetAuth is Login
func GetAuth(email, password string) (map[string]string, error) {
	params := url.Values{
		"appIdKey":        strings.Split(APIKey, "#"),
		"userLocale":      strings.Split(UserLocale, "#"),
		"protocolVersion": strings.Split(ProtocolVersion, "#"),
		"format":          []string{"json"},
	}
	params["appleId"] = strings.Split(email, "#")
	params["password"] = strings.Split(password, "#")
	request := NewClientRequest(apiURL, "POST")
	body, _, err := request.SetHeader(requestHeader).SetBody(params.Encode()).GetBody()
	if err != nil {
		return nil, err
	}
	var info map[string]string
	json.Unmarshal(body, &info)
	if mycainfo, ok := info["myacinfo"]; ok && len(mycainfo) > 0 {
		return info, nil
	}
	return nil, errors.New(info["userString"])
}
