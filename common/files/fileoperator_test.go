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
package files

import (
	"testing"
)

func TestMakeDirAndFileThenRemove(t *testing.T) {
	paths := []string{
		"log2/log2_1/log2_2/log.log",
		"log1/log1_1/log.log",
		"test.log",
	}
	for _, p := range paths {
		MakeDirAndFile(p)
		if !IsExist(p) {
			t.Error("MakeDirAndFile error")
		}
	}
	for _, p := range paths {
		RemoveFileAndEmptyFolder(p)
		if IsExist(p) {
			t.Error("RemoveFileAndEmptyFolder error")
		}
	}

}
