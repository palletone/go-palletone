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
 *
 */

package storage

import (
	"fmt"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
)

const (
	globalPropDBKey    = "GlobalProperty"
	dynGlobalPropDBKey = "DynamicGlobalProperty"
)

// modified by Yiran
type PropertyDatabase struct {
	db            ptndb.Database
	GlobalProp    *modules.GlobalProperty
	DynGlobalProp *modules.DynamicGlobalProperty
	MediatorSchl  *modules.MediatorSchedule
}
type PropertyDb interface {
	StoreGlobalProp(gp *modules.GlobalProperty) error
	StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error
	RetrieveGlobalProp() (*modules.GlobalProperty, error)
	RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error)
	StoreMediatorSchl(ms *modules.MediatorSchedule) error
	RetrieveMediatorSchl() (*modules.MediatorSchedule, error)
	GetGlobalProp() *modules.GlobalProperty
	GetDynGlobalProp() *modules.DynamicGlobalProperty
	GetMediatorSchl() *modules.MediatorSchedule
}

// modified by Yiran
// initialize PropertyDB , and retrieve gp,dgp,mc from PropertyDb.
func NewPropertyDb(db ptndb.Database) (*PropertyDatabase,error) {
	pdb := &PropertyDatabase{db: db}
	gp, err := pdb.RetrieveGlobalProp()
	if err != nil {
		log.Error("RetrieveGlobalProp Error")
		return nil,err
	}

	dgp, err := pdb.RetrieveDynGlobalProp()
	if err != nil {
		log.Error("RetrieveDynGlobalProp Error")
		return nil,err
	}

	ms, err := pdb.RetrieveMediatorSchl()
	if err != nil {
		log.Error("RetrieveMediatorSchl Error")
		return nil,err
	}
	pdb.GlobalProp = gp
	pdb.DynGlobalProp = dgp
	pdb.MediatorSchl = ms
	return pdb,nil
}

func NewPropertyDb4GenesisInit(db ptndb.Database) (*PropertyDatabase) {
	return &PropertyDatabase{db: db}
}

// modified by Yiran
func (propdb *PropertyDatabase) GetGlobalProp() *modules.GlobalProperty {
	return propdb.GlobalProp
}

// modified by Yiran
func (propdb *PropertyDatabase) GetDynGlobalProp() *modules.DynamicGlobalProperty {
	return propdb.DynGlobalProp
}

// modified by Yiran
func (propdb *PropertyDatabase) GetMediatorSchl() *modules.MediatorSchedule {
	return propdb.MediatorSchl
}

type globalProperty struct {
	ChainParameters core.ChainParameters

	ActiveMediators []core.MediatorInfo
}

func getGPT(gp *modules.GlobalProperty) globalProperty {
	ams := make([]core.MediatorInfo, 0)

	for _, med := range gp.ActiveMediators {
		medInfo := core.MediatorToInfo(&med)
		ams = append(ams, medInfo)
	}

	gpt := globalProperty{
		ChainParameters: gp.ChainParameters,
		ActiveMediators: ams,
	}

	return gpt
}

func getGP(gpt *globalProperty) *modules.GlobalProperty {
	ams := make(map[common.Address]core.Mediator, 0)
	for _, medInfo := range gpt.ActiveMediators {
		med := core.InfoToMediator(&medInfo)
		ams[med.Address] = med
	}

	gp := modules.NewGlobalProp()
	gp.ChainParameters = gpt.ChainParameters
	gp.ActiveMediators = ams

	return gp
}

func (propdb *PropertyDatabase) StoreGlobalProp(gp *modules.GlobalProperty) error {

	gpt := getGPT(gp)

	err := Store(propdb.db, globalPropDBKey, gpt)

	if err != nil {
		log.Error(fmt.Sprintf("Store global properties error:%s", err))
	}

	return err
}

func (propdb *PropertyDatabase) StoreDynGlobalProp(dgp *modules.DynamicGlobalProperty) error {

	err := Store(propdb.db, dynGlobalPropDBKey, *dgp)
	if err != nil {
		//log.Error(fmt.Sprintf("Store dynamic global properties error: %s", err))
	}

	return err
}

func (propdb *PropertyDatabase) RetrieveGlobalProp() (*modules.GlobalProperty, error) {
	gpt := new(globalProperty)

	err := Retrieve(propdb.db, globalPropDBKey, gpt)
	if err != nil {
		//log.Error(fmt.Sprintf("Retrieve global properties error: %s", err))
	}

	gp := getGP(gpt)

	return gp, err
}

func (propdb *PropertyDatabase) RetrieveDynGlobalProp() (*modules.DynamicGlobalProperty, error) {
	dgp := modules.NewDynGlobalProp()

	err := Retrieve(propdb.db, dynGlobalPropDBKey, dgp)
	if err != nil {
		//log.Error(fmt.Sprintf("Retrieve dynamic global properties error: %s", err))
	}

	return dgp, err
}
//@Yiran
//var (
//	MEDIATORVOTE_PREFIX = []byte("01")
//	COMMONVOTE_PREFIX   = []byte("00")
//
//	MEDIATORTERMINTERVAL = 3000
//)
//func (propdb *PropertyDatabase) UpdateActiveMediators () error{
//	term := unit
//	activeMediators, err := propdb.GetActiveMediators(,MEDIATORTERMINTERVAL)
//	if err != nil {
//		return ErrorLogHandler(err,"GetActiveMediators")
//	}
//
//}
//func (propdb *PropertyDatabase) GetActiveMediators(term []byte) ([]common.Address, error) {
//	key := KeyConnector(MEDIATOR_CANDIDATE_PREFIX,term)
//	// 1. Load Addresses of MediatorCandidates
//	addresses := make([]common.Address, 0)
//	ErrorLogHandler(Retrieve(propdb.db, string(key), addresses),"RetrieveMediatorCandidatesAddress")
//	// 2. Load VoteNumber of each MediatorCandidates
//	for _, address := range(addresses) {
//		tempKey := KeyConnector(key,address[:])
//		Retrieve
//	}
//
//}

//@Yiran This function checks that a transaction contains a action which creates a vote.
func IsVoteInitiationTx(transactionIndex []byte) error {
	//TODO
	return nil
}

//@Yiran this function connect multiple []byte keys to single []byte.
func KeyConnector(keys ...[]byte) []byte {
	var res []byte
	for _, key := range keys {
		res = append(res, key...)
	}
	return res
}

//@Yiran print error if exist.
func ErrorLogHandler(err error, errType string) error {
	if err != nil {
		println(errType, "error", err.Error())
		return err
	}
	return nil
}
//@Yiran
type VoteBox struct {
	Candidates []Candidate
	Voter []common.Address
}
func (box * VoteBox) Sort() {
	//TODO
}
func (box * VoteBox) AddToBoxIfNotVoted (voter common.Address,vote common.Address) {
	//TODO
	//for addr := range box.voter {
	//	if addr == voter{
	//		return
	//	}
	//}

}

func NewVoteBox () *VoteBox {
	return &VoteBox{
		Candidates:make([]Candidate,0),
		Voter:make([]common.Address,0),
	}
}
//@Yiran
type Candidate struct {
	Address    common.Address
	VoteNumber uint64
}