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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package golang

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_splitEnvPath(t *testing.T) {
	paths := splitEnvPaths("foo" + string(os.PathListSeparator) + "bar" + string(os.PathListSeparator) + "baz")
	assert.Equal(t, len(paths), 3)
}

func Test_getGoEnv(t *testing.T) {
	goenv, err := getGoEnv()
	assert.NoError(t, err)

	_, ok := goenv["GOPATH"]
	assert.Equal(t, ok, true)

	_, ok = goenv["GOROOT"]
	assert.Equal(t, ok, true)
}
