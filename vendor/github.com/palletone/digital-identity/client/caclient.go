/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package client

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"

	"net/http"
	"crypto/tls"
	"github.com/cloudflare/cfssl/csr"

	"github.com/palletone/digital-identity/config"
)

type PalletCAClient struct {
	// Uri is access point for palletone-ca server. Port number and scheme must be provided.
	// for example http://127.0.0.1:7054
	Url string
	// SkipTLSVerification define how connection must handle invalid TLC certificates.
	// if true, all verifications are skipped. This value is overwritten by Transport property, if provided
	SkipTLSVerification bool
	// Crypto is CryptSuite implementation used to sign request for palletone-ca server
	Crypto CryptoSuite
	// Transport define transport rules for communication with palletone-ca server. If nil, default Go setting will be used
	// It is responsibility of the user to provide proper TLS/certificate setting in TLS communication.
	Transport *http.Transport
	// MspId value will be added to Identity in Enrollment and ReEnrollment invocations.
	// This value is not used anywhere in CA implementation, but is need in every call to palletone and is added here
	// for convenience, because (in general case) palletoneCA is serving one MSP
	// User can overwrite this value at any time.
	MspId string

	FilePath string

	ServerInfo ServerInfo

	Admin   string
	Adminpw string
}

var CA *PalletCAClient

type ServerInfo struct {
	CAName string
	CACert *x509.Certificate
}

// RegistrationRequest holds all data needed for new registration of new user in Certificate Authority
type CARegistrationRequest struct {
	// EnrolmentId is unique name that identifies identity
	EnrolmentId string `json:"id"`
	// Type defines type of this identity (user,client, auditor etc...)
	Type string `json:"type"`
	// Secret is password that will be used for enrollment. If not provided random password will be generated
	Secert string `json:"secert,omitempty" mask:"password" help:"The enrollment secret for the identity being registered"`
	// MaxEnrollments define maximum number of times that identity can enroll. If not provided or is 0 there is no limit
	MaxEnrollments int `json:"max_enrollments,omitempty"`
	// Affiliation associates identity with particular organisation.
	Affiliation string `json:"affiliation" help:"The identity's affiliation"`
	// Attrs are attributes associated with this identity
	Attrs []CaRegisterAttribute `json:"attrs,omitempty"`
	// CAName is the name of the CA that should be used.
	CAName string `json:"caname,omitempty" skip:"true"`
}

// CaRegisterAttribute holds user attribute used for registration
// for example user may have attr `accountType` with value `premium`
// this attributes can be accessed in chainCode and build business logic on top of them
type CaRegisterAttribute struct {
	// Name is the name of the attribute.
	Name string `json:"name"`
	// Value is the value of the attribute. Can be empty string
	Value string `json:"value"`
	// ECert define how this attribute will be included in ECert. If this value is true this attribute will be
	// added to ECert automatically on Enrollment if no attributes are requested on Enrollment request.
	// ECert : certificate of registration
	ECert bool `json:"ecert,omitempty"`
}

type certificateRequest struct {
	CaEnrollmentRequest
	CR string `json:"certificate_request"`
}

// enrollmentResponse is response from fabric-ca server for enrolment that contains created Ecert
type enrollmentResponse struct {
	caResponse
	Result enrollmentResponseResult `json:"result"`
}

type enrollmentResponseResult struct {
	Cert       string
	ServerInfo enrollmentResponseServerInfo
	Version    string
}

type enrollmentResponseServerInfo struct {
	CAName  string
	CAChain string
}

// CAResponse represents response message from fabric-ca server
type caResponse struct {
	Success  bool            `json:"success"`
	Errors   []caResponseErr `json:"errors"`
	Messages []string        `json:"messages"`
}

type caResponseErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CaEnrollAttribute struct {
	// Name is the name of the attribute
	Name string `json:"name"`
	// Optional define behaviour when required attribute is not available to user. If `true` then request will continue,
	// but attribute will not be included in ECert. If `false` and attribute is missing, request will fail.
	// If false and attribute is available, request will continue and attribute will be added in ECert
	Optional bool `json:"optional,omitempty"`
}

// CaEnrollmentRequest holds data needed for getting ECert (enrollment) from CA server
type CaEnrollmentRequest struct {
	EnrollmentId string `json:"name" skip:"true"`
	// Secret is the password for this identity
	Secret  string `json:"secret,omitempty" skip:"true" mask:"password"`
	Profile string `json:"profile,omitempty"`
	//Label is used for hardware secure modules.
	Label  string `json:"label,omitempty"`
	CAName string `json:"caname,omitempty" skip:"true"`
	// Host is the list of valid host names for this certificate. If empty default hosts will be used
	Hosts []string `json:"hosts"`
	// Attrs are the attributes that must be included in ECert. This is subset of the attributes used in registration.
	Attrs []CaEnrollAttribute `json:"attr_reqs,omitempty"`
	// CSR is Certificate Signing Request info
	CSR *CSRInfo `json:"csr,omitempty" help:"Certificate Signing Request info"`
}

type CSRInfo struct {
	CN           string           `json:"CN"`
	Names        []csr.Name       `json:"names,omitempty"`
	Hosts        []string         `json:"hosts,omitempty"`
	KeyRequest   *BasicKeyRequest `json:"key,omitempty"`
	CA           *csr.CAConfig    `json:"ca,omitempty"`
	SerialNumber string           `json:"serial_number,omitempty"`
}

// BasicKeyRequest encapsulates size and algorithm for the key to be generated
type BasicKeyRequest struct {
	Algo string `json:"algo" yaml:"algo"`
	Size int    `json:"size" yaml:"size"`
}

// CARegisterCredentialResponse credentials from palletone-ca server registration request
type caRegisterCredentialResponse struct {
	Secret string `json:"secret"`
}

type caRegisterResponse struct {
	caResponse
	Result caRegisterCredentialResponse `json:"result"`
}

// CAGetCertsResponse holds response from `GetCaCertificateChain`
type CAGetCertResponse struct {
	// RootCertificates is list of pem encoded certificates
	RootCertificates []*x509.Certificate
	// IntermediateCertificates is list of pem encoded intermediate certificates
	IntermediateCertificates []*pem.Block
	// CAName is the name of the CA server that returns this certificates
	CAName string
	// Version is the version of server that returns this certificates
	Version string
}

type caInfoRequest struct {
	CaName string `json:"caname,omitempty"`
}

type caInfoResponse struct {
	caResponse
	Result caResponseResult `json:"result"`
}

type caResponseResult struct {
	CAName  string `json:"CAName"`
	CAChain string `json:"CAChain"`
	Version string `json:"Version"`
}

type CARevocationRequest struct {
	EnrollmentId string `json:"id,omitempty"`
	Serial       string `json:"serial,omitempty"`
	AKI          string `json:"aki,omitempty"`
	Reason       string `json:"reason,omitempty"`
	CAName       string `json:"caname,omitempty"`
	// GenCRL specifies whether to generate a CRL. CRL will be returned only when AKI and Serial are provided.
	GenCRL bool `json:"gencrl,omitempty"`
}

// CaRevokeResultCertificate identify revoked certificate
type CaRevokeResultCertificate struct {
	Serial string `json:"Serial"`
	AKI    string `json:"AKI"`
}

type CARevokeResult struct {
	RevokedCertificates []CaRevokeResultCertificate `json:"RevokedCerts"`
	CRL                 string                      `json:"CRL"`
}

type caRevokeResponse struct {
	caResponse
	Result CARevokeResult
}

type CaIdentityResponse struct {
	ID             string                `json:"id"`
	Type           string                `json:"type"`
	Affiliation    string                `json:"affiliation"`
	Attributes     []CaRegisterAttribute `json:"attrs" mapstructure:"attrs"`
	MaxEnrollments int                   `json:"max_enrollments" mapstructure:"max_enrollments"`
}

type CAGetIdentityResponse struct {
	CaIdentityResponse
	CAName string `json:"caname"`
}

type caGetIdentity struct {
	caResponse
	Result CAGetIdentityResponse `json:"result"`
}

type CAListAllIdentitesResponse struct {
	CAName     string               `json:"caname"`
	Identities []CaIdentityResponse `json:"identities,omitempty"`
}

type caListAllIdentities struct {
	caResponse
	Result CAListAllIdentitesResponse `json:"result"`
}

func NewCAClient(path string, transport *http.Transport) (*PalletCAClient, error) {
	config, err := config.NewCAConfig(path)
	if err != nil {
		return nil, err
	}
	return NewCaClientFromConfig(*config, transport)
}

func concatErrors(errs []caResponseErr) error {
	errors := ""
	for _, e := range errs {
		errors += e.Message + ":"
	}
	return fmt.Errorf(errors)
}

func NewCaClientFromConfig(conf config.CAConfig, transport *http.Transport) (*PalletCAClient, error) {
	var crypt CryptoSuite
	var err error

	switch conf.CryptoConfig.Family {
	case "ecdsa":
		crypt, err = NewECCryptoSuiteFromConfig(conf.CryptoConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, config.ErrInvalidAlgorithmFamily
	}
	CA = &PalletCAClient{SkipTLSVerification: conf.SkipTLSValidation,
		Url: conf.Url,
		Crypto: crypt,
		Transport: transport,
		MspId: conf.MspId,
		FilePath: conf.FilePath,
		Admin: conf.Admin,
		Adminpw: conf.Adminpw,
	}

	return CA, nil
}

//Enroll execute enrollment request for registered user in FabricCA server.
//On success new Identity with ECert and generated csr are returned.
func (f *PalletCAClient) Enroll(request CaEnrollmentRequest,key interface{}) (*Identity, []byte, error) {
	var hosts []string
	if len(request.Hosts) == 0 {
		parsedUrl, err := url.Parse(f.Url)
		if err != nil {
			return nil, nil, err
		}
		hosts = []string{parsedUrl.Host}
	} else {
		hosts = request.Hosts
	}

	csr, err := f.Crypto.CreateCertificateRequest(request.EnrollmentId, key, hosts)
	if err != nil {
		return nil, nil, err
	}

	crm, err := json.Marshal(certificateRequest{CR: string(csr), CaEnrollmentRequest: request})
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/enroll", f.Url), bytes.NewBuffer(crm))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(request.EnrollmentId, request.Secret)

	httpClient := &http.Client{Transport: f.getTransport()}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		enrResp := new(enrollmentResponse)
		if err := json.Unmarshal(body, enrResp); err != nil {
			return nil, nil, err
		}
		if !enrResp.Success {
			return nil, nil, concatErrors(enrResp.Errors)
		}

		cabyte, err := base64.StdEncoding.DecodeString(enrResp.Result.ServerInfo.CAChain)
		if err != nil {
			return nil, nil, err
		}
		cablock, _ := pem.Decode(cabyte)
		cacert, err := x509.ParseCertificate(cablock.Bytes)
		f.ServerInfo.CAName = enrResp.Result.ServerInfo.CAName
		f.ServerInfo.CACert = cacert

		rawCert, err := base64.StdEncoding.DecodeString(enrResp.Result.Cert)
		if err != nil {
			return nil, nil, err
		}
		a, _ := pem.Decode(rawCert)
		cert, err := x509.ParseCertificate(a.Bytes)

		if err != nil {
			return nil, nil, err
		}
		if ID == nil {
			ID = &Identity{Certificate: cert, PrivateKey: key, MspId: f.MspId}
		}
		return &Identity{Certificate: cert, PrivateKey: key, MspId: f.MspId}, csr, nil
	}
	return nil, nil, fmt.Errorf("non 200 response: %v message is: %s", resp.StatusCode, string(body))
}

// Register registers new user in fabric-ca server. In registration request attributes, affiliation and
// max enrolments must be set.
// It is responsibility of the SDK user to ensure passwords are with big entropy.
// Identity parameter is certificate for user that makes registration and this user MUST have the role for
// registering new users.
func (f *PalletCAClient) Register(identity *Identity, req *CARegistrationRequest) (string, error) {
	if req.EnrolmentId == "" {
		return "", config.ErrEnrollmentIdMissing
	}
	if req.Affiliation == "" {
		return "", config.ErrAffiliationMissing
	}
	if req.Type == "" {
		return "", config.ErrTypeMissing
	}
	if identity == nil {
		return "", config.ErrCertificateEmpty
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/register", f.Url), bytes.NewBuffer(reqJson))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	token, err := f.createAuthToken(identity, reqJson)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("authorization", token)

	httpClient := &http.Client{Transport: f.getTransport()}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		result := new(caRegisterResponse)
		if err = json.Unmarshal(body, result); err != nil {
			return "", err
		}
		if !result.Success {
			return "", concatErrors(result.Errors)
		}
		if len(result.Errors) > 0 {
			return "", concatErrors(result.Errors)
		}
		return result.Result.Secret, nil
	}
	return "", fmt.Errorf("non 200 response: %v message is: %s", resp.StatusCode, string(body))
}

// GetCaCertificateChain gets root and intermediate certificates used by PalletoneCA server.
func (f *PalletCAClient) GetCaCertificateChain(caName string) (*CAGetCertResponse, error) {
	reqJson, err := json.Marshal(caInfoRequest{CaName: caName})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cainfo", f.Url), bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Transport: f.getTransport()}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		result := new(caInfoResponse)
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
		if !result.Success {
			return nil, err
		}
		certs, err := base64.StdEncoding.DecodeString(result.Result.CAChain)
		if err != nil {
			return nil, err
		}

		var root []*x509.Certificate
		var intermediate []*pem.Block

		for len(certs) > 0 {
			var block *pem.Block
			block, certs = pem.Decode(certs)
			if block == nil {
				break
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("error parsing certificate from ca chain")
			}
			if !cert.IsCA {
				return nil, fmt.Errorf("invalid certificate in ca chain")
			}
			// If authority key id is not present or if it is present and equal to subject key id,
			// then it is a root certificate
			if len(cert.AuthorityKeyId) == 0 || bytes.Equal(cert.AuthorityKeyId, cert.SubjectKeyId) {
				root = append(root, cert)
			} else {
				intermediate = append(intermediate, block)
			}
		}
		return &CAGetCertResponse{
			RootCertificates:         root,
			IntermediateCertificates: intermediate,
			Version:                  result.Result.Version,
			CAName:                   result.Result.CAName,
		}, nil
	}
	return nil, fmt.Errorf("non 200 response: %v message is: %s", resp.StatusCode, string(body))
}

// createAuthToken creates http authorization header token to verify the request.
// it is composed by base64 encoded Cert concatenated by base64 encoded request signed with Cert private key
func (f *PalletCAClient) createAuthToken(identity *Identity, request []byte) (string, error) {
	encPm := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: identity.Certificate.Raw,
		},
	)
	encCert := base64.StdEncoding.EncodeToString(encPm)
	body := base64.StdEncoding.EncodeToString(request)
	sigString := body + "." + encCert
	sig, err := f.Crypto.Sign([]byte(sigString), identity.PrivateKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", encCert, base64.StdEncoding.EncodeToString(sig)), nil
}

func (f *PalletCAClient) getTransport() *http.Transport {
	var tr *http.Transport
	if f.Transport == nil {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: f.SkipTLSVerification},
		}
	} else {
		tr = f.Transport
	}
	return tr
}

// Revoke revokes ECert
func (f *PalletCAClient) Revoke(identity *Identity, request *CARevocationRequest) (*CARevokeResult, error) {
	reqJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/revoke", f.Url), bytes.NewBuffer(reqJson))
	httpReq.Header.Set("Content-Type", "application/json")

	token, err := f.createAuthToken(identity, reqJson)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("authorization", token)

	httpClient := &http.Client{Transport: f.getTransport()}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		result := new(caRevokeResponse)
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
		fmt.Println(result)
		if !result.Success {
			return nil, concatErrors(result.Errors)
		}
		return &result.Result, nil
	}
	return nil, fmt.Errorf("non 200 response: %v message is: %s", resp.StatusCode, string(body))
}

func (f *PalletCAClient) GetIndentity(identity *Identity, id string, caName string) (*CAGetIdentityResponse, error) {
	if identity == nil {
		return nil, config.ErrCertificateEmpty
	}
	if len(id) == 0 {
		return nil, config.ErrIdentityNameMissing
	}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/identities/%s", f.Url, id), bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}
	token, err := f.createAuthToken(identity, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("authorization", token)

	if len(caName) > 0 {
		uri := httpReq.URL.Query()
		uri.Add("ca", caName)
		httpReq.URL.RawQuery = uri.Encode()
	}

	httpClient := &http.Client{Transport: f.getTransport()}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		result := new(caGetIdentity)
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
		if !result.Success {
			return nil, concatErrors(result.Errors)
		}
		if len(result.Errors) > 0 {
			return nil, concatErrors(result.Errors)
		}
		return &result.Result, nil
	}
	return nil, fmt.Errorf("non 200 response: %v message is: %s", resp.StatusCode, string(body))
}

// ListAllIdentities get list of all identities
func (f *PalletCAClient) GetIdentities(identity *Identity, caName string) (*CAListAllIdentitesResponse, error) {
	if identity == nil {
		return nil, config.ErrCertificateEmpty
	}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/identities", f.Url), bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}
	token, err := f.createAuthToken(identity, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("authorization", token)

	if len(caName) > 0 {
		uri := httpReq.URL.Query()
		uri.Add("ca", caName)
		httpReq.URL.RawQuery = uri.Encode()
	}

	httpClient := &http.Client{Transport: f.getTransport()}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		result := new(caListAllIdentities)
		if err := json.Unmarshal(body, result); err != nil {
			return nil, err
		}
		if !result.Success {
			return nil, concatErrors(result.Errors)
		}
		if len(result.Errors) > 0 {
			return nil, concatErrors(result.Errors)
		}
		return &result.Result, nil
	}
	return nil, fmt.Errorf("non 200 response: %v message is: %s", resp.StatusCode, string(body))
}
