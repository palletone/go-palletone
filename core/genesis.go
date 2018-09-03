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
 */

package core

import (
	"encoding/base64"
	"fmt"

	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common/log"
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	DepositRate float64 `json:"depositRate"`
}

type Genesis struct {
	Version                   string                   `json:"version"`
	Alias                     string                   `json:"alias"`
	TokenAmount               uint64                   `json:"tokenAmount"`
	TokenDecimal              uint32                   `json:"tokenDecimal"`
	DecimalUnit               string                   `json:"decimal_unit"`
	ChainID                   uint64                   `json:"chainId"`
	TokenHolder               string                   `json:"tokenHolder"`
	InitialParameters         ChainParameters          `json:"initialParameters"`
	ImmutableParameters       ImmutableChainParameters `json:"immutableChainParameters"`
	InitialTimestamp          int64                    `json:"initialTimestamp"`
	InitialActiveMediators    uint16                   `json:"initialActiveMediators"`
	InitialMediatorCandidates []MediatorInfo           `json:"initialMediatorCandidates"`
	SystemConfig              SystemConfig             `json:"systemConfig"`
}

func (g *Genesis) GetTokenAmount() uint64 {
	return g.TokenAmount
}

type MediatorInfo struct {
	Address,
	InitPartPub,
	Node string
}

// author Albert·Gou
func ScalarToStr(sec kyber.Scalar) string {
	secB, err := sec.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return base64.RawURLEncoding.EncodeToString(secB)
}

// author Albert·Gou
func PointToStr(pub kyber.Point) string {
	pubB, err := pub.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return base64.RawURLEncoding.EncodeToString(pubB)
}

func MediatorToInfo(m *Mediator) MediatorInfo {
	return MediatorInfo{
		Address:     m.Address.Str(),
		InitPartPub: PointToStr(m.InitPartPub),
		Node:        m.Node.String(),
	}
}
