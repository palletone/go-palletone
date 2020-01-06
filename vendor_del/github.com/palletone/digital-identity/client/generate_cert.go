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
	"os"
)

type CaGenInfo struct {
	EnrolmentID string `json:"enrolmentid"`
	Name        string `json:"name"`
	Data        string `json:"data"`
	ECert       bool   `json:"ecert"`
	Type        string `json:"type"`
	Affiliation string `json:"affiliation"`
	Key         interface{}
}

func NewCaGenInfo(address string, name string, data string, ecert bool, ty string, affiliation string, key interface{}) *CaGenInfo {
	return &CaGenInfo{
		EnrolmentID: address,
		Name:        name,
		Data:        data,
		ECert:       ecert,
		Type:        ty,
		Affiliation: affiliation,
		Key:         key,
	}
}

//You must register your administrator certificate first
func (c *CaGenInfo) EnrollAdmin() (*Identity, error) {
	gopath := os.Getenv("GOPATH")
	path := gopath + "/src/github.com/palletone/digital-identity/config"

	cacli, err := InitCASDK(path, "caconfig.yaml")
	if err != nil {
		return nil, err
	}

	enrollRequest := CaEnrollmentRequest{EnrollmentID: cacli.Admin, Secret: cacli.Adminpw}

	id, _, err := Enroll(CA, enrollRequest, c.Key)
	if err != nil {
		return nil, err

	}
	return id, nil
}

func (c *CaGenInfo) Enrolluser() ([]byte, error) {
	id, _ := c.EnrollAdmin()
	attr := []CaRegisterAttribute{{
		Name:  c.Name,
		Value: c.Data,
		ECert: c.ECert,
	},
	}
	rr := CARegistrationRequest{
		EnrolmentID: c.EnrolmentID,
		Affiliation: c.Affiliation,
		Type:        c.Type,
		Attrs:       attr,
	}
	certpem, err := Register(CA, id, &rr, c.Key)

	if err != nil {
		return nil, err
	}
	return certpem, nil

}

func (c *CaGenInfo) Revoke(enrollmentid, reason string) ([]byte, error) {
	id, err := c.EnrollAdmin()
	if err != nil {
		return nil, err
	}
	req := CARevocationRequest{EnrollmentID: enrollmentid, Reason: reason, GenCRL: true}
	crlPem, err := Revoke(CA, id, &req)
	if err != nil {
		return nil, err
	}
	return crlPem, nil
}

func (c *CaGenInfo) GetIndentity(enrollmentid, caname string) (*CAGetIdentityResponse, error) {
	id, _ := c.EnrollAdmin()
	var idresp CAGetIdentityResponse
	idresp, err := GetIndentity(CA, id, enrollmentid, caname)
	if err != nil {
		return &CAGetIdentityResponse{}, err
	}
	return &idresp, nil

}

func (c *CaGenInfo) GetIndentities() (*CAListAllIdentitesResponse, error) {
	id, _ := c.EnrollAdmin()
	var idresps CAListAllIdentitesResponse
	idresps, err := GetIndentities(CA, id)
	if err != nil {
		return &CAListAllIdentitesResponse{}, err
	}
	return &idresps, nil

}

func (c *CaGenInfo) GetCaCertificateChain(caName string) (*CAGetCertResponse, error) {
	id, _ := c.EnrollAdmin()
	var certChain CAGetCertResponse
	certChain, err := GetCertificateChain(CA, id, caName)
	if err != nil {
		return &CAGetCertResponse{}, err
	}
	return &certChain, nil
}
