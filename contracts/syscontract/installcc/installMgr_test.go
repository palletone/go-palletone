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

package installcc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallMgr_getTemplateId(t *testing.T) {
	tplId := getTemplateId("install", "", "v1.0")
	t.Logf("%x", tplId)
	tplId2 := getTemplateId("install", "", "v2.0")
	assert.NotEqual(t, tplId, tplId2)
}
