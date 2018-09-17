/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package core

const (
	DefaultAlias               = "PTN"
	DefaultMediatorInterval    = 5
	DefaultMediatorCount       = 21
	DefaultTokenAmount         = 100000000000000000
	DefaultTokenDecimal        = 8
	DefaultDepositRate         = 0.02
	DefaultTokenHolder         = "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ"
	DefaultPassword            = "password"
	DefaultMinMediatorCount    = 11
	DefaultMinMediatorInterval = 1

	/* percentage fields are fixed point with a denominator of 10,000 */
	PalletOne100Percent            = 10000
	PalletOne1Percent              = PalletOne100Percent / 100
	PalletOneIrreversibleThreshold = 70 * PalletOne1Percent
)
