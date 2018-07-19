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

package mediatorplugin

import (
	"github.com/palletone/go-palletone/core/node"
	"github.com/palletone/go-palletone/common"
)

type MediatorPlugin struct {
	node *node.Node
	// Enable VerifiedUnit production, even if the chain is stale.
	// 新开启一个区块链时，必须设为true
	productionEnabled bool
	// Mediator`s account and passphrase controlled by this node
	mediators map[common.Address]string
}
