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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package state

import (
	"github.com/palletone/go-palletone/common"
	"time"
)

// contract template
type ContractTplObj struct {
	TplID        string    `json:"address"`       // contract template address
	CodeHash     string    `json:"codeHash"`      // contract code hash
	Code         []byte    `json:"code"`          // contract bytecode
	CreationDate time.Time `json:"creation_date"` // contract template create time
}

// instance of contract
type ContractObj struct {
	TplID         string                 `json:"tpl_id"`        // contract template address
	Address       common.Address         `json:"address"`       // the contract instance address
	Params        map[string]interface{} `json:"params"`        // contract params status
	Status        string                 `json:"status"`        // the contract status, like 'good', 'bad', 'close' etc.
	CreationDate  time.Time              `json:"creation_date"` // contract template create time
	DestroyedDate time.Time              `json:"destroyed_date"`
}

type ContractAccount struct {
	Address   common.Address         `json:"address"`   // user accounts address
	Contracts map[string]ContractObj `json:"contracts"` // user contracts
}
