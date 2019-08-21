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

import "sync"

type sigShareSet struct {
	opLock    sync.Mutex
	dataLock  sync.Mutex
	sigShares [][]byte
}

func newSigShareSet(cap int) *sigShareSet {
	if cap > 0 {
		return &sigShareSet{
			sigShares: make([][]byte, 0, cap),
		}
	} else {
		return &sigShareSet{
			sigShares: make([][]byte, 0),
		}
	}
}

func (s *sigShareSet) len() int {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	return len(s.sigShares)
}

func (s *sigShareSet) lock() {
	s.opLock.Lock()
}

func (s *sigShareSet) unlock() {
	s.opLock.Unlock()
}

func (s *sigShareSet) append(sigShare []byte) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	s.sigShares = append(s.sigShares, sigShare)
}

func (s *sigShareSet) popSigShares() (sigShares [][]byte) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	sigShares = make([][]byte, 0, len(s.sigShares))
	sigShares = append(sigShares, s.sigShares...)

	s.sigShares = make([][]byte, 0)

	return
}
