package apple

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"strings"

	uuid "github.com/satori/go.uuid"
)

const (
	pemBegin = "-----BEGIN CERTIFICATE-----"
	pemEnd   = "-----END CERTIFICATE-----"
	splitEnd = "\r\n"
)

func BuildSearchQueryString(teamId string, s map[string]string) string {
	search := strings.Builder{}

	search.WriteString("teamId=")
	search.WriteString(teamId)

	for k, v := range s {
		if k == "limit" {
			search.WriteString("&limit=")
			search.WriteString(v)
			continue
		}
		search.WriteString("&filter[")
		search.WriteString(k)
		search.WriteString("]=")
		search.WriteString(v)
	}
	return search.String()
}

func ContentToPem(body string, chunklen uint) string {
	runes, erunes, begin, end := []rune(body), []rune(splitEnd), []rune(pemBegin), []rune(pemEnd)
	l := uint(len(runes))
	if l <= 1 || l < chunklen {
		return body + splitEnd
	}
	ns := make([]rune, 0, len(begin)+len(runes)+len(erunes)+len(end))
	var i uint
	ns = append(ns, begin...)
	ns = append(ns, erunes...)
	for i = 0; i < l; i += chunklen {
		if i+chunklen > l {
			ns = append(ns, runes[i:]...)
		} else {
			ns = append(ns, runes[i:i+chunklen]...)
		}
		ns = append(ns, erunes...)
	}
	ns = append(ns, end...)
	return string(ns)
}

func GetSubject(commonName, emailAddress string) pkix.Name {
	oidEmailAddress := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}
	return pkix.Name{
		CommonName: commonName,
		ExtraNames: []pkix.AttributeTypeAndValue{
			{
				Type: oidEmailAddress,
				Value: asn1.RawValue{
					Tag:   asn1.TagIA5String,
					Bytes: []byte(emailAddress),
				},
			},
		},
	}
}

func CreateCertificateSigningRequest(commonName, emailAddress string, years int) (string, string) {
	subject := GetSubject(commonName, emailAddress)
	keys, _ := rsa.GenerateKey(rand.Reader, 2048)

	csrTemplate := x509.CertificateRequest{
		Subject:            subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
		EmailAddresses:     strings.Split(emailAddress, "#"),
	}
	csr, _ := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)

	var csrContent, privateKey bytes.Buffer

	pem.Encode(&csrContent, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr})
	pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keys)})

	return csrContent.String(), privateKey.String()
}

func GenerateUDID() string {
	UDID := uuid.NewV4()
	return UDID.String()
}
