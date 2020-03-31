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
package config

import "errors"

var (
	ErrInvalidAlgorithmFamily = errors.New("invalid algorithm family")
	ErrInvalidAlgorithm       = errors.New("invalid algorithm for ECDSA")
	ErrInvalidHash            = errors.New("invalid hash algorithm")
	ErrInvalidKeyType         = errors.New("invalid key type is provided")
	ErrEnrollmentIDMissing    = errors.New("enrollment id is empty")
	ErrAffiliationMissing     = errors.New("affiliation is missing")
	ErrTypeMissing            = errors.New("type is missing")
	ErrCertificateEmpty       = errors.New("certificate cannot be nil")
	ErrIdentityNameMissing    = errors.New("identity must have  name")
)
