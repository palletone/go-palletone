// This file is part of go-palletone
// go-palletone is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-palletone is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-palletone. If not, see <http://www.gnu.org/licenses/>.
//
// @author PalletOne DevTeam dev@pallet.one
// @date 2018

package core

import (
	"github.com/palletone/go-palletone/common/event"
)

type ConsensusEngine interface {
	Engine() int
	Stop()
	SubscribeCeEvent(chan<- ConsensusEvent) event.Subscription
}

type ConsensusEvent struct {
	Ce string
}

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	MediatorSlot  int      `json:"mediatorSlot"`
	MediatorCount int      `json:"mediatorCount"`
	MediatorList  []string `json:"mediatorList"`
	MediatorCycle int      `json:"mediatorCycle"`
	DepositRate   float64  `json:"depositRate"`
}

type Genesis struct {
	Height       string       `json:"height"`
	Version      string       `json:"version"`
	TokenAmount  uint64       `json:"tokenAmount"`
	TokenDecimal int          `json:"tokenDecimal"`
	ChainID      int          `json:"chainId"`
	TokenHolder  string       `json:"tokenHolder"`
	SystemConfig SystemConfig `json:"systemConfig"`
}
