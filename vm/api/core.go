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

package api

import (
	"io"
	"golang.org/x/net/context"
	"github.com/palletone/go-palletone/vm/ccintf"
)

type BuildSpecFactory func() (io.Reader, error)
type PrelaunchFunc func() error

//VM is an abstract virtual image for supporting arbitrary virual machines
type VM interface {
	Deploy(ctxt context.Context, ccid ccintf.CCID, args []string, env []string, reader io.Reader) error
	Start(ctxt context.Context, ccid ccintf.CCID, args []string, env []string, filesToUpload map[string][]byte, builder BuildSpecFactory, preLaunchFunc PrelaunchFunc) error
	Stop(ctxt context.Context, ccid ccintf.CCID, timeout uint, dontkill bool, dontremove bool) error
	Destroy(ctxt context.Context, ccid ccintf.CCID, force bool, noprune bool) error
	GetVMName(ccID ccintf.CCID, format func(string) (string, error)) (string, error)
}
