package apple

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/widuu/apple/pkcs12"
)

const certApiUrl = "https://developerservices2.apple.com/services/v1/certificates"

type CertificateData struct {
	CertType   string            `json:"type"`
	Id         string            `json:"id"`
	Attributes map[string]string `json:"attributes"`
}

type certificateListRequest struct {
	Error []map[string]string `json:"errors"`
	Data  []CertificateData   `json:"data"`
}

type certificatDataRequest struct {
	CertType   string            `json:"type"`
	Attributes map[string]string `json:"attributes"`
}

type createCertificatRequest struct {
	Data certificatDataRequest `json:"data"`
}

type createCertificatResponse struct {
	Error []map[string]string `json:"errors"`
	Data  CertificateData     `json:"data"`
}

var pemCSRPrefix = []byte("-----BEGIN")

// CreateCertificate info
func CreateCertificate(certificateType, csrContent, teamId, myacinfo string) (CertificateData, error) {
	requestParams := createCertificatRequest{
		certificatDataRequest{
			CertType: "certificates",
			Attributes: map[string]string{
				"teamId":          teamId,
				"certificateType": certificateType,
				"csrContent":      csrContent,
			},
		},
	}

	postJson, err := json.Marshal(requestParams)
	if err != nil {
		return CertificateData{}, err
	}

	// request
	request := NewClientRequest(certApiUrl, "POST")
	JSONRequestHeader["Cookie"] = "myacinfo=" + myacinfo
	if _, ok := JSONRequestHeader["X-HTTP-Method-Override"]; ok {
		delete(JSONRequestHeader, "X-HTTP-Method-Override")
	}

	body, _, err := request.SetHeader(JSONRequestHeader).SetBody(postJson).GetBody()
	var data createCertificatResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return CertificateData{}, err
	}

	if len(data.Error) > 0 {
		return CertificateData{}, errors.New(data.Error[0]["detail"])
	}

	return data.Data, nil
}

// CertLists filter[id] certificateType
func CertLists(customSearch map[string]string, teamId, myacinfo string) ([]CertificateData, error) {
	search := BuildSearchQueryString(teamId, customSearch)

	requestParams := struct {
		UrlEncodedQueryParams string `json:"urlEncodedQueryParams"`
	}{
		UrlEncodedQueryParams: search,
	}

	postJson, err := json.Marshal(requestParams)
	if err != nil {
		return []CertificateData{}, err
	}

	// request
	request := NewClientRequest(certApiUrl, "POST")
	JSONRequestHeader["Cookie"] = "myacinfo=" + myacinfo
	JSONRequestHeader["X-HTTP-Method-Override"] = "GET"
	body, _, err := request.SetHeader(JSONRequestHeader).SetBody(postJson).GetBody()

	var data certificateListRequest

	err = json.Unmarshal(body, &data)
	if err != nil {
		return []CertificateData{}, err
	}

	if len(data.Error) > 0 {
		return []CertificateData{}, errors.New(data.Error[0]["detail"])
	}

	if len(data.Data) <= 0 {
		return []CertificateData{}, errors.New("Certificate does not exist")
	}

	return data.Data, nil
}

// DeleteCertficate
func DeleteCertficate(Id, teamId, myacinfo string) (bool, error) {

	requestParams := struct {
		UrlEncodedQueryParams string `json:"urlEncodedQueryParams"`
	}{
		UrlEncodedQueryParams: fmt.Sprintf("teamId=%s", teamId),
	}

	// decode json
	postJson, err := json.Marshal(requestParams)
	if err != nil {
		return false, err
	}

	// delete api url
	deleteApiUrl := certApiUrl + "/" + Id

	// header
	JSONRequestHeader["Cookie"] = "myacinfo=" + myacinfo
	JSONRequestHeader["X-HTTP-Method-Override"] = "DELETE"

	// request
	request := NewClientRequest(deleteApiUrl, "POST")
	body, code, err := request.SetHeader(JSONRequestHeader).SetBody(postJson).GetBody()

	if code == 204 {
		return true, nil
	}

	var data certificateListRequest
	err = json.Unmarshal(body, &data)
	if err != nil {
		return false, err
	}

	if len(data.Error) > 0 {
		return false, errors.New(data.Error[0]["detail"])
	}

	return false, errors.New("delete certficate fail")
}

// ExportCertficate csrContent
func ExportCertficate(csrContent, priKey, password string) ([]byte, error) {
	// format cert
	if !bytes.HasPrefix([]byte(csrContent), pemCSRPrefix) {
		csrContent = ContentToPem(csrContent, 64)
	}
	// cert
	certBlock, _ := pem.Decode([]byte(csrContent))
	if certBlock == nil {
		return nil, errors.New("error decoding certificate")
	}
	// pri
	priKeyBlock, _ := pem.Decode([]byte(priKey))
	if priKeyBlock == nil {
		return nil, errors.New("error decoding private key")
	}
	// parse private key
	var parsedKey interface{}
	var err error
	if priKeyBlock.Type == "EC PRIVATE KEY" {
		parsedKey, err = x509.ParseECPrivateKey(priKeyBlock.Bytes)
		if err != nil {
			return nil, errors.New("unable to parse private key to native object")
		}
	} else {
		parsedKey, err = x509.ParsePKCS1PrivateKey(priKeyBlock.Bytes)
		if err != nil {
			parsedKey, err = x509.ParsePKCS8PrivateKey(priKeyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("private key should be a PEM or plain PKCS1 or PKCS8; parse error: %v", err)
			}
		}
	}
	// export p12
	p12, err := pkcs12.Encode(certBlock.Bytes, parsedKey, password)
	if err != nil {
		return nil, err
	}

	return p12, nil
}
