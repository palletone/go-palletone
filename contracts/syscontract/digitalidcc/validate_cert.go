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
 *  * @date 2018-2019
 *
 */

package digitalidcc

import (
	"crypto/x509"
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
)

func ValidateCert(issuer string, cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	if err := checkExists(cert.SerialNumber.String(), stub); err != nil {
		return err
	}
	if err := validateIssuer(issuer, cert, stub); err != nil {
		return err
	}

	return nil
}

func checkExists(certid string, stub shim.ChaincodeStubInterface) error {
	key := CERT_ID + certid
	data, err := stub.GetState(key)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		return fmt.Errorf("Cert(%s) is existing.", certid)
	}
	return nil
}

func validateIssuer(issuer string, cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	// check with root ca holder
	rootCAHolder, err := stub.GetSystemConfig("RootCaHolder")
	if err != nil {
		return err
	}
	// check in server list
	if issuer != rootCAHolder {
		// query server list
		certids, err := queryCertsIDs(CERT_SERVER_SYMBOL, issuer, stub)
		if err != nil {
			return err
		}
		if len(certids) <= 0 {
			return fmt.Errorf("Has no validate intermidate certificate")
		}
	}
	return nil
}
