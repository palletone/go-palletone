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
 *  * @date 2018-2019
 *
 */

package modules

import "testing"

func TestUniqueId_StringFriendly(t *testing.T) {
	uid := &UniqueId{0x28, 0x5a, 0x59, 0x29}
	t.Logf("%x, %s", uid.Bytes(), uid.StringFriendly(UniqueIdType_Sequence))
	t.Logf("%x, %s", uid.Bytes(), uid.StringFriendly(UniqueIdType_Uuid))
	t.Logf("%x, %s", uid.Bytes(), uid.StringFriendly(UniqueIdType_UserDefine))
	t.Logf("%x, %s", uid.Bytes(), uid.StringFriendly(UniqueIdType_Ascii))

}
