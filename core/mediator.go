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
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dedis/kyber"
	"github.com/dedis/kyber/pairing/bn256"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p/discover"
)

var Suite = bn256.NewSuiteG2()

func GenInitPair() (kyber.Scalar, kyber.Point) {
	sc := Suite.Scalar().Pick(Suite.RandomStream())

	return sc, Suite.Point().Mul(sc, nil)
}

// mediator 结构体 和具体的账户模型有关
type Mediator struct {
	Address    common.Address
	InitPubKey kyber.Point
	Node       *discover.Node
	*MediatorInfoExpand
}

type MediatorInfoExpand struct {
	Url                  string
	TotalMissed          uint64
	LastConfirmedUnitNum uint32
	TotalVotes           uint64
}

func NewMediatorBase() *MediatorInfoExpand {
	return &MediatorInfoExpand{
		Url:                  "",
		TotalMissed:          0,
		LastConfirmedUnitNum: 0,
		TotalVotes:           0,
	}
}

func NewMediator() *Mediator {
	return &Mediator{
		MediatorInfoExpand: NewMediatorBase(),
	}
}

func StrToMedNode(medNode string) (*discover.Node, error) {
	node, err := discover.ParseNode(medNode)
	if err != nil {
		err = fmt.Errorf("invalid mediator node \"%v\" : %v", medNode, err)
		return nil, err
	}

	return node, nil
}

func StrToMedAdd(addStr string) common.Address {
	address := strings.TrimSpace(addStr)
	address = strings.Trim(address, "\"")

	addr, err := common.StringToAddress(address)
	// addrType, err := addr.Validate()
	if err != nil || addr.GetType() != common.PublicKeyHash {
		log.Error(fmt.Sprintf("Invalid mediator account address \"%v\" : %v", address, err))
	}

	return addr
}

func StrToScalar(secStr string) (kyber.Scalar, error) {
	secB := base58.Decode(secStr)
	sec := Suite.Scalar()

	err := sec.UnmarshalBinary(secB)
	if err != nil {
		return nil, err
	}

	return sec, nil
}

func StrToPoint(pubStr string) (kyber.Point, error) {
	pubB := base58.Decode(pubStr)
	pub := Suite.Point()

	err := pub.UnmarshalBinary(pubB)
	if err != nil {
		err = fmt.Errorf("invalid init mediator public key \"%v\" : %v", pubStr, err)
		return nil, err
	}

	return pub, nil
}
