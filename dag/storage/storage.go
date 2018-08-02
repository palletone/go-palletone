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

package storage

import (
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

// value will serialize to rlp encoding bytes
func Store(key string, value interface{}) error {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	val, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}

	_, err = Dbconn.Get([]byte(key))
	if err != nil {
		if err == errors.ErrNotFound {
			if err := Dbconn.Put([]byte(key), val); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err = Dbconn.Delete([]byte(key)); err != nil {
			return err
		}
		if err := Dbconn.Put([]byte(key), val); err != nil {
			return err
		}
	}

	return nil
}

func StoreBytes(key []byte, value interface{}) error {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	val, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}

	_, err = Dbconn.Get(key)
	if err != nil {
		if err == errors.ErrNotFound {
			if err := Dbconn.Put(key, val); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err = Dbconn.Delete(key); err != nil {
			return err
		}
		if err := Dbconn.Put(key, val); err != nil {
			return err
		}
	}

	return nil
}

func StoreString(key, value string) error {
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DefaultConfig.DbPath)
	}
	return Dbconn.Put(util.ToByte(key), util.ToByte(value))
}
