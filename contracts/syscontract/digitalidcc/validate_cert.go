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
	"crypto/x509/pkix"
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
	dagConstants "github.com/palletone/go-palletone/dag/constants"
	dagModules "github.com/palletone/go-palletone/dag/modules"
	"math/big"
	"time"
)

// This is the basic validation
func ValidateCert(issuer string, cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	if err := checkExists(cert, stub); err != nil {
		return err
	}
	if err := validateIssuer(issuer, cert, stub); err != nil {
		return err
	}
	// validate

	return nil
}

func ValidateCRLIssuer(issuer string, crl *pkix.CertificateList, stub shim.ChaincodeStubInterface) (certHolder []*dagModules.CertHolderInfo, err error) {
	// check issuer identity
	certsInfo, err := getIssuerCertsInfo(issuer, stub)
	if err != nil {
		return nil, err
	}
	certHolder = []*dagModules.CertHolderInfo{}
	for _, revokeCert := range crl.TBSCertList.RevokedCertificates {
		hasHolder := false
		for _, holder := range certsInfo {
			if revokeCert.SerialNumber.String() == holder.CertID {
				certHolder = append(certHolder, holder)
				hasHolder = true
				break
			}
		}
		if !hasHolder {
			return nil, fmt.Errorf("Issuer(%s) has no authority to revoke cert(%s)", issuer, revokeCert.SerialNumber.String())
		}
	}

	if len(certHolder) != len(crl.TBSCertList.RevokedCertificates) {
		return nil, fmt.Errorf("cert lenth is invalid")
	}
	return certsInfo, nil
}

func checkExists(cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	// check root ca
	rootCert, err := GetRootCACert(stub)
	if err != nil {
		return err
	}
	if rootCert.SerialNumber.String() == cert.SerialNumber.String() {
		return fmt.Errorf("Can not add root ca.")
	}
	// check other certificates
	key := dagConstants.CERT_BYTES_SYMBOL + cert.SerialNumber.String()
	data, err := stub.GetState(key)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		return fmt.Errorf("Cert(%s) is existing.", cert.SerialNumber.String())
	}
	return nil
}

func validateIssuer(issuer string, cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	// check with root ca holder
	rootCAHolder, err := stub.GetState("RootCAHolder")
	if err != nil {
		return err
	}
	// check in intermediate certificate
	rootCert, err := GetRootCACert(stub)
	if err != nil {
		return err
	}
	if issuer == string(rootCAHolder) {
		if cert.Issuer.String() != rootCert.Subject.String() {
			return fmt.Errorf("cert issuer is invalid, excepted %s but %s", rootCert.Subject.String(), cert.Issuer.String())
		}
	} else {
		// query certid
		certid, err := GetCertIDBySubject(cert.Issuer.String(), stub)
		if err != nil {
			return err
		}
		if certid == "" {
			return fmt.Errorf("the issuer has no verified certificate")
		}
		// query server list
		revocationTime, err := GetCertRevocationTime(issuer, certid, stub)
		if err != nil {
			return err
		}
		if revocationTime.IsZero() || revocationTime.Before(time.Now()) {
			return fmt.Errorf("Has no validate intermidate certificate. Time is %s", revocationTime.String())
		}
	}
	return nil
}

// This is the certificate chain validation
// To validate certificate chain signature
func ValidateCertChain(cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	// query root ca cert bytes
	rootCert, err := GetRootCACert(stub)
	if err != nil {
		return err
	}
	// query intermidate cert bytes
	chancerts := []*x509.Certificate{}
	if cert.Issuer.String() != rootCert.Subject.String() {
		chancerts, err = GetIntermidateCertChains(cert, rootCert.Subject.String(), stub)
		if err != nil {
			return err
		}
	}
	// package x509.VerifyOptions, Intermediates and Roots field
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	intermediates := x509.NewCertPool()
	for _, newCert := range chancerts {
		intermediates.AddCert(newCert)
	}
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}
	// user x509.Verify to verify cert chain
	if _, err := cert.Verify(opts); err != nil {
		return err
	}

	return nil
}

// Validate CRL Issuer Signature
func ValidateCRLIssuerSig(issuerAddr string, crl *pkix.CertificateList, stub shim.ChaincodeStubInterface) error {
	// check ca holder
	caHolder, err := stub.GetState("RootCAHolder")
	if err != nil {
		return err
	}
	if issuerAddr == string(caHolder) {
		rootCert, err := GetRootCACert(stub)
		if err != nil {
			return err
		}
		return rootCert.CheckCRLSignature(crl)
	}
	// query issuer cert info
	key := dagConstants.CERT_SUBJECT_SYMBOL + crl.TBSCertList.Issuer.String()
	val, err := stub.GetState(key)
	if err != nil {
		return err
	}
	certid := big.Int{}
	certid.SetBytes(val)
	// check revocation time
	key = dagConstants.CERT_SERVER_SYMBOL + issuerAddr + dagConstants.CERT_SPLIT_CH + certid.String()
	val, err = stub.GetState(key)
	if err != nil {
		return err
	}
	revocationTime := time.Time{}
	if err = revocationTime.UnmarshalBinary(val); err != nil {
		return err
	}
	if revocationTime.Before(time.Now()) {
		return fmt.Errorf("your certificate has been revocation")
	}
	// query issuer cert info
	cert, err := GetX509Cert(certid.String(), stub)
	if err != nil {
		return err
	}
	// check signature
	return cert.CheckCRLSignature(crl)
}
