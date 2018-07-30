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

package dagconfig

var (
	SConfig Sconfig
)

var DefaultConfig = Config{
	//DbPath: "db/leveldb",
	DbPath: "../../cmd/gptn/leveldb",
	DbName: "palletone",

	// txpool
	UnitTxSize: 1024 * 1024,

	// utxo
	UtxoIndex: true,
}

// global configuration of dag modules
type Config struct {
	DbPath    string
	DbName    string
	DbCache   int
	DbHandles int

	// cache
	CacheSource string

	//redis
	RedisAddr   string
	RedisPwd    string
	RedisPrefix string
	RedisDb     int

	// txpool
	UnitTxSize float64

	// utxo
	UtxoIndex bool
}

type Sconfig struct {
	Blight bool
}
