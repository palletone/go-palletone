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

package common

//import (
//	"sync"
//
//	"github.com/palletone/go-palletone/dag/dagconfig"
//	"github.com/palletone/go-palletone/dag/modules"
//	"github.com/palletone/go-palletone/dag/storage"
//	"github.com/coocood/freecache"
//)
//
//func NewDag() *modules.Dag {
//	if storage.Dbconn == nil {
//		storage.ReNewDbConn(dagconfig.Config.DbPath)
//	}
//	genesis := modules.NewGenesis()
//	return &modules.Dag{Cache: freecache.NewCache(200 * 1024 * 1024),
//		Db:          storage.Dbconn,
//		GenesisUnit: genesis,
//		Mutex:       sync.RWMutex,
//	}
//
//}
