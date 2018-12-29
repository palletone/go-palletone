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

package errors

import (
	"errors"
)

// common error
var (
	ErrSetEmpty       = errors.New("dag: set is empty")
	ErrDagNotFound    = errors.New("dag: not found")
	ErrNotFound       = New("leveldb: not found")
	ErrNumberNotFound = New("dag: header's number not found")
	ErrUtxoNotFound   = New("utxo: not found")
	ErrUnitExist      = New("unit: exist")

	ErrUnknownAncestor = errors.New("unknown ancestor")
	ErrPrunedAncestor  = errors.New("pruned ancestor")
	ErrFutureBlock     = errors.New("block in the future")
	ErrInvalidNumber   = errors.New("invalid block number")
)

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}
