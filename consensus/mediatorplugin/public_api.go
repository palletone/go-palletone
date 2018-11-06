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

import "github.com/palletone/go-palletone/core"

type PublicMediatorAPI struct {
	*MediatorPlugin
}

func NewPublicMediatorAPI(mp *MediatorPlugin) *PublicMediatorAPI {
	return &PublicMediatorAPI{mp}
}

func (a *PublicMediatorAPI) List() []string {
	addStrs := make([]string, 0)
	mas := a.dag.GetMediators()

	for address, _ := range mas {
		addStrs = append(addStrs, address.Str())
	}

	return addStrs
}

func (a *PublicMediatorAPI) Schedule() []string {
	addStrs := make([]string, 0)
	ms := a.dag.MediatorSchedule()

	for _, medAdd := range ms {
		addStrs = append(addStrs, medAdd.Str())
	}

	return addStrs
}

type InitDKSResult struct {
	PrivateKey string
	PublicKey  string
}

func (a *PublicMediatorAPI) GetInitDKS() (res InitDKSResult) {
	sec, pub := GenInitPair(a.suite)

	res.PrivateKey = core.ScalarToStr(sec)
	res.PublicKey = core.PointToStr(pub)

	return
}
