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
 *  * @author PalletOne core developer  <dev@pallet.one>
 *  * @date 2018
 *
 */

package storage

import (
	"errors"
	"fmt"
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"

	"log"
	"sort"
	"strings"
	"time"
)













//保存了对合约写集、Config、Asset信息
type StateDatabase struct {
	db ptndb.Database
}

func NewStateDatabase(db ptndb.Database) *StateDatabase {
	return &StateDatabase{db: db}
}

type StateDb interface {
	GetConfig(name []byte) []byte
	GetPrefix(prefix []byte) map[string][]byte
	SaveConfig(confs []modules.PayloadMapStruct, stateVersion *modules.StateVersion) error
	SaveAssetInfo(assetInfo *modules.AssetInfo) error
	GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error)
	SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	SaveContractTemplate(templateId []byte, bytecode []byte, version []byte) error
	SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error
	DeleteState(key []byte) error
	GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string)
	GetContractState(id string, field string) (*modules.StateVersion, []byte)
	GetTplAllState(id []byte) map[modules.ContractReadSet][]byte
	GetContractAllState(id []byte) map[modules.ContractReadSet][]byte
	GetTplState(id []byte, field string) (*modules.StateVersion, []byte)
	GetContract(id common.Hash) (*modules.Contract, error)
	SaveVote(id []byte, voteData interface{}) error
	SaveMediatorsList(Candidates MediatorCandidates) error
	GetMediatorsList() (MediatorCandidates, error)
	GetActiveMediators(n int) common.Addresses
	GetGlobalProperty() (GlobalPropertyStore, error)
	SaveGlobalProperty(globalProperty GlobalPropertyStore) error
	GetDynamicGlobalProperty() (DynamicGlobalProperty, error)
	SaveDynamicGlobalProperty(DynamicGlobalProperty DynamicGlobalProperty) error
	GetMediatorSchedule() (MediatorScheduleStore, error)
	SaveMediatorSchedule(MediatorSchedule MediatorScheduleStore) error
}

// ######################### SAVE IMPL START ###########################

func (statedb *StateDatabase) SaveAssetInfo(assetInfo *modules.AssetInfo) error {
	key := assetInfo.Tokey()
	return StoreBytes(statedb.db, key, assetInfo)
}

func (statedb *StateDatabase) SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error {
	return SaveContractState(statedb.db, CONTRACT_TPL, id, name, value, version)
}
func (statedb *StateDatabase) SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion) error {
	return SaveContractState(statedb.db, CONTRACT_STATE_PREFIX, id, name, value, version)
}
func (statedb *StateDatabase) DeleteState(key []byte) error {
	return statedb.db.Delete(key)
}

func (statedb *StateDatabase) SaveContractTemplate(templateId []byte, bytecode []byte, version []byte) error {
	key := append(CONTRACT_TPL, templateId...)
	key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	key = append(key, []byte(modules.FIELD_TPL_BYTECODE)...)
	key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	key = append(key, version...)
	if err := StoreBytes(statedb.db, key, bytecode); err != nil {
		return err
	}
	return nil
}

/**
保存合约属性信息
To save contract
*/
func SaveContractState(db ptndb.Database, prefix []byte, id []byte, field string, value interface{}, version *modules.StateVersion) error {
	key := []byte{}
	key = append(prefix, id...)
	key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	key = append(key, []byte(field)...)
	key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	key = append(key, version.Bytes()...)

	if err := StoreBytes(db, key, value); err != nil {
		log.Println("Save contract template", "error", err.Error())
		return err
	}
	return nil
}

// ######################### SAVE IMPL END ###########################

// ######################### GET IMPL START ###########################

/**
获取模板所有属性
To get contract or contract template all fields and return
*/
func (statedb *StateDatabase) GetTplAllState(id []byte) map[modules.ContractReadSet][]byte {
	// key format: [PREFIX][ID]_[field]_[version]
	key := append(CONTRACT_TPL, id...)
	key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) <= 0 {
		return nil
	}
	allState := map[modules.ContractReadSet][]byte{}
	for k, v := range data {
		sKey := strings.Split(k, "^*^")
		if len(sKey) != 3 {
			continue
		}
		var version modules.StateVersion
		if !version.ParseStringKey(k) {
			continue
		}
		rdSet := modules.ContractReadSet{
			Key:   sKey[1],
			Value: &version,
		}
		allState[rdSet] = v
	}
	return allState
}

/**
获取合约（或模板）所有属性
To get contract or contract template all fields and return
*/
func (statedb *StateDatabase) GetContractAllState(id []byte) map[modules.ContractReadSet][]byte {
	// key format: [PREFIX][ID]_[field]_[version]
	key := fmt.Sprintf("%s%s^*^", CONTRACT_STATE_PREFIX, hexutil.Encode(id))
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) <= 0 {
		return nil
	}
	allState := map[modules.ContractReadSet][]byte{}
	for k, v := range data {
		sKey := strings.Split(k, "^*^")
		if len(sKey) != 3 {
			continue
		}
		var version modules.StateVersion
		if !version.ParseStringKey(key) {
			continue
		}
		rdSet := modules.ContractReadSet{
			Key:   sKey[1],
			Value: &version,
		}
		allState[rdSet] = v
	}
	return allState
}

/**
获取合约（或模板）某一个属性
To get contract or contract template one field
*/
func (statedb *StateDatabase) GetTplState(id []byte, field string) (*modules.StateVersion, []byte) {
	//key := fmt.Sprintf("%s%s^*^%s^*^", CONTRACT_TPL, hexutil.Encode(id[:]), field)
	key := append(CONTRACT_TPL, id...)
	key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	key = append(key, []byte(field)...)
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) != 1 {
		return nil, nil
	}
	for k, v := range data {
		var version modules.StateVersion
		if !version.ParseStringKey(k) {
			return nil, nil
		}
		return &version, v
	}
	return nil, nil
}

/**
获取合约（或模板）某一个属性
To get contract or contract template one field
*/
func (statedb *StateDatabase) GetContractState(id string, field string) (*modules.StateVersion, []byte) {
	key := fmt.Sprintf("%s%s^*^%s^*^", CONTRACT_STATE_PREFIX, id, field)
	data := getprefix(statedb.db, []byte(key))
	if data == nil || len(data) != 1 {
		return nil, nil
	}
	for k, v := range data {
		var version modules.StateVersion
		if !version.ParseStringKey(k) {
			return nil, nil
		}
		return &version, v
	}
	log.Println("11111111")
	return nil, nil
}
func (statedb *StateDatabase) GetAssetInfo(assetId *modules.Asset) (*modules.AssetInfo, error) {
	key := append(modules.ASSET_INFO_PREFIX, assetId.AssetId.String()...)
	data, err := statedb.db.Get(key)
	if err != nil {
		return nil, err
	}

	var assetInfo modules.AssetInfo
	err = rlp.DecodeBytes(data, &assetInfo)

	if err != nil {
		return nil, err
	}
	return &assetInfo, nil
}

// get prefix: return maps
func (db *StateDatabase) GetPrefix(prefix []byte) map[string][]byte {
	return getprefix(db.db, prefix)
}

// GetContract can get a Contract by the contract hash
func (statedb *StateDatabase) GetContract(id common.Hash) (*modules.Contract, error) {
	if common.EmptyHash(id) {
		return nil, errors.New("the filed not defined")
	}
	con_bytes, err := statedb.db.Get(append(CONTRACT_PTEFIX, id[:]...))
	if err != nil {
		log.Println("err:", err)
		return nil, err
	}
	contract := new(modules.Contract)
	err = rlp.DecodeBytes(con_bytes, contract)
	if err != nil {
		log.Println("err:", err)
		return nil, err
	}
	return contract, nil
}

// <<<<< Fengyiran
//get mediatorCandidates ==> sort ==> RETURN specified number of Addresses
func (statedb *StateDatabase) GetActiveMediators(n int) common.Addresses {
	Candidates, err := statedb.GetMediatorsList()
	if err != nil {
		return nil
	}
	return Candidates.GetHeadAddress(n)
}

func (statedb *StateDatabase) GetMediatorsList() (MediatorCandidates, error) {
	Candidates, err := GetDecodedComplexData(statedb.db, MEDIATOR_CANDIDATE_PREFIX, MediatorCandidates{})
	return Candidates.(MediatorCandidates), err
}
func (statedb *StateDatabase) SaveMediatorsList(Candidates MediatorCandidates) error {
	key := MEDIATOR_CANDIDATE_PREFIX
	value := Candidates
	return ErrorLogHandler(StoreBytes(statedb.db, key, value), "SaveMediatorsList")
}

func (statedb *StateDatabase) GetMediatorSchedule() ( MediatorScheduleStore, error) {
	MediatorSchedule, err := GetDecodedComplexData(statedb.db, MEDIATOR_SCHEME_PREFIX,  MediatorScheduleStore{})
	return MediatorSchedule.( MediatorScheduleStore), err
}
func (statedb *StateDatabase) SaveMediatorSchedule(MediatorSchedule  MediatorScheduleStore) error {
	key := MEDIATOR_SCHEME_PREFIX
	value := MediatorSchedule
	return ErrorLogHandler(StoreBytes(statedb.db, key, value), "SaveMediatorsList")
}

func (statedb *StateDatabase) GetGlobalProperty() ( GlobalPropertyStore, error) {
	gp, err := GetDecodedComplexData(statedb.db, GLOBALPROPERTY_PREFIX,  GlobalPropertyStore{})
	return gp.( GlobalPropertyStore), err
}
func (statedb *StateDatabase) SaveGlobalProperty(globalProperty  GlobalPropertyStore) error {
	key := GLOBALPROPERTY_PREFIX
	value := globalProperty
	return ErrorLogHandler(StoreBytes(statedb.db, key, value), "SaveGlobalProperty")
}

func (statedb *StateDatabase) GetDynamicGlobalProperty() ( DynamicGlobalProperty, error) {
	dgp, err := GetDecodedComplexData(statedb.db, DYNAMIC_GLOBALPROPERTY_PREFIX,  DynamicGlobalProperty{})
	return dgp.( DynamicGlobalProperty), err
}
func (statedb *StateDatabase) SaveDynamicGlobalProperty(DynamicGlobalProperty  DynamicGlobalProperty) error {
	key := DYNAMIC_GLOBALPROPERTY_PREFIX
	value := DynamicGlobalProperty
	return ErrorLogHandler(StoreBytes(statedb.db, key, value), "SaveDynamicGlobalProperty")
}

func (statedb *StateDatabase) SaveVote(id []byte, vote interface{}) error {
	key := KeyConnector(VOTE_PREFIX, id)
	value := vote
	return ErrorLogHandler(StoreBytes(statedb.db, key, value), "SaveVote")
}

func GetDecodedComplexData(db ptndb.Database, key []byte, dataType interface{}) (interface{}, error) {
	valByte, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	return rlp.GetDecodedFromBytes(valByte, dataType)
}
func ErrorLogHandler(err error, errType string) error {
	if err != nil {
		log.Println(errType, "error", err.Error())
		return err
	}
	return nil
}
func KeyConnector(keys ...[]byte) []byte {
	var res []byte
	for _, key := range keys {
		res = append(res, key...)
	}
	return res
}

type MediatorCandidates []MediatorCandidate
type MediatorCandidate struct {
	Address    common.Address
	VoteNumber VoteNumber
}
type VoteNumber uint64
type StateDBConfig [] StateConfig
type StateConfig struct {
	Prefix []byte
	suffix []byte
}

func (sc StateConfig) HasSuffix() bool {
	return sc.suffix != nil
}
func (ms MediatorCandidates) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms MediatorCandidates) Len() int           { return len(ms) }
func (ms MediatorCandidates) Less(i, j int) bool { return ms[i].VoteNumber > ms[j].VoteNumber }
func (ms MediatorCandidates) GetHeadAddress(n int) common.Addresses {
	if n < 21 {
		log.Println("less mediator number", "error", )
		return nil
	}
	var res common.Addresses
	sort.Sort(ms)
	for i := 0; i < n; i++ {
		res = append(res, ms[i].Address)
	}
	return res
}

// Fengyiran >>>>>

/**
获取合约模板
To get contract template
*/
func (statedb *StateDatabase) GetContractTpl(templateID []byte) (version *modules.StateVersion, bytecode []byte, name string, path string) {
	key := append(CONTRACT_TPL, templateID...)
	key = append(key, []byte(modules.FIELD_SPLIT_STR)...)
	key = append(key, []byte(modules.FIELD_TPL_BYTECODE)...)
	data := statedb.GetPrefix(key)

	if len(data) == 1 {
		for _, v := range data {
			if err := rlp.DecodeBytes(v, &bytecode); err != nil {
				fmt.Println("GetContractTpl when get bytecode", "error", err.Error(), "codeing:", v, "val:", bytecode)
				return
			}
		}
	}

	version, nameByte := statedb.GetTplState(templateID, modules.FIELD_TPL_NAME)
	if nameByte == nil {
		return
	}
	if err := rlp.DecodeBytes(nameByte, &name); err != nil {
		log.Println("GetContractTpl when get name", "error", err.Error())
		return
	}

	_, pathByte := statedb.GetTplState(templateID, modules.FIELD_TPL_PATH)
	if err := rlp.DecodeBytes(pathByte, &path); err != nil {
		log.Println("GetContractTpl when get path", "error", err.Error())
		return
	}
	return
}



// ######################### GET IMPL END ###########################


//----------------from global_property.go
// 全局属性的结构体定义
type GlobalProperty struct {
	ChainParameters core.ChainParameters // 区块链网络参数
	statedb  StateDb
	ActiveMediators map[common.Address]core.Mediator // 当前活跃mediator集合；每个维护间隔更新一次
}

// 动态全局属性的结构体定义
type DynamicGlobalProperty struct {
	LastVerifiedUnitNum uint64 // 最近的验证单元编号(数量)

	LastVerifiedUnitHash common.Hash // 最近的验证单元hash

	//	LastVerifiedUnit *v.VerifiedUnit	// 最近生产的验证单元

	LastVerifiedUnitTime int64 // 最近的验证单元时间

	//	CurrentMediator *Mediator // 当前生产验证单元的mediator, 用于判断是否连续同一个mediator生产验证单元

	//	NextMaintenanceTime time.Time // 下一次系统维护时间

	// 当前的绝对时间槽数量，== 从创世开始所有的时间槽数量 == verifiedUnitNum + 丢失的槽数量
	CurrentASlot uint64

	/**
	在过去的128个见证单元生产slots中miss的数量。
	The count of verifiedUnit production slots that were missed in the past 128 verifiedUnits
	用于计算mediator的参与率。used to compute mediator participation.
	*/
	//	RecentSlotsFilled float32
}

func (gp *GlobalProperty) GetActiveMediatorCount() int {
	return len(gp.ActiveMediators)
}

func (gp *GlobalProperty) GetCurThreshold() int {
	aSize := gp.GetActiveMediatorCount()
	offset := (core.PalletOne100Percent - core.PalletOneIrreversibleThreshold) * aSize /
		core.PalletOne100Percent

	return aSize - offset
}

func (gp *GlobalProperty) GetActiveMediatorInitPubs() []kyber.Point {
	aSize := gp.GetActiveMediatorCount()
	pubs := make([]kyber.Point, aSize, aSize)

	meds := gp.GetActiveMediators()
	for i, add := range meds {
		med := gp.GetActiveMediator(add)

		pubs[i] = med.InitPartPub
	}

	return pubs
}

func (gp *GlobalProperty) IsActiveMediator(add common.Address) bool {
	_, ok := gp.ActiveMediators[add]

	return ok
}

func (gp *GlobalProperty) GetActiveMediator(add common.Address) *core.Mediator {
	if !gp.IsActiveMediator(add) {
		log.Fatal(fmt.Sprintf("%v is not active mediator!", add.Str()))
		return nil
	}

	med, _ := gp.ActiveMediators[add]

	return &med
}

func (gp *GlobalProperty) GetActiveMediatorAddr(index int) common.Address {
	if index < 0 || index > gp.GetActiveMediatorCount()-1 {
		log.Fatal(fmt.Sprintf("%v is out of the bounds of active mediator list!", index))
	}

	meds := gp.GetActiveMediators()

	return meds[index]
}

func (gp *GlobalProperty) GetActiveMediatorNode(index int) *discover.Node {
	ma := gp.GetActiveMediatorAddr(index)
	med := gp.GetActiveMediator(ma)

	return med.Node
}

// GetActiveMediators, return the list of active mediators, and the order of the list from small to large
func (gp *GlobalProperty) GetActiveMediators() []common.Address {
	mediatorNumber := gp.GetActiveMediatorCount()
	return gp.statedb.GetActiveMediators(mediatorNumber)
}

func (gp *GlobalProperty) GetInitActiveMediators() []common.Address {
	mediators := make([]common.Address, 0, gp.GetActiveMediatorCount())

	for _, m := range gp.ActiveMediators {
		mediators = append(mediators, m.Address)
	}

	//sortAddress(mediators)

	return mediators
}

// re:Yiran
// obsolete:statedb.GetActiveMediators() already sort address.
func sortAddress(adds []common.Address) {
	aSize := len(adds)
	addStrs := make([]string, aSize, aSize)
	for i, add := range adds {
		addStrs[i] = add.Str()
	}

	sort.Strings(addStrs)

	for i, addStr := range addStrs {
		adds[i], _ = common.StringToAddress(addStr)
	}
}

func (gp *GlobalProperty) GetActiveMediatorNodes() map[string]*discover.Node {
	nodes := make(map[string]*discover.Node)

	meds := gp.GetActiveMediators()
	for _, add := range meds {
		med := gp.GetActiveMediator(add)
		node := med.Node

		nodes[node.ID.TerminalString()] = node
	}

	return nodes
}

func NewGlobalProp(statedb StateDb) *GlobalProperty {
	return &GlobalProperty{
		ChainParameters: core.NewChainParams(),
		ActiveMediators: map[common.Address]core.Mediator{},
		statedb:statedb,
	}
}

func NewDynGlobalProp() *DynamicGlobalProperty {
	return &DynamicGlobalProperty{
		LastVerifiedUnitNum:  0,
		LastVerifiedUnitHash: common.Hash{},
		CurrentASlot:         0,
	}
}

func InitGlobalProp(genesis *core.Genesis) *GlobalProperty {
	log.Println("initialize global property...")

	// Create global properties
	gp := &GlobalProperty{
		ChainParameters: core.NewChainParams(),
		ActiveMediators: map[common.Address]core.Mediator{},
	}

	log.Println("initialize chain parameters...")
	gp.ChainParameters = genesis.InitialParameters

	log.Println("Set active mediators...")
	// Set active mediators
	for i := uint16(0); i < genesis.InitialActiveMediators; i++ {
		medInfo := genesis.InitialMediatorCandidates[i]
		md := core.InfoToMediator(&medInfo)

		gp.ActiveMediators[md.Address] = md
	}

	return gp
}

func InitDynGlobalProp(genesis *core.Genesis, genesisUnitHash common.Hash) *DynamicGlobalProperty {
	log.Println("initialize dynamic global property...")

	// Create dynamic global properties
	dgp := NewDynGlobalProp()
	dgp.LastVerifiedUnitTime = genesis.InitialTimestamp
	dgp.LastVerifiedUnitHash = genesisUnitHash

	return dgp
}


//struct for store
type GlobalPropertyStore struct {
	ChainParameters core.ChainParameters

	ActiveMediators []core.MediatorInfo
}

func getGPT(gp *GlobalProperty) GlobalPropertyStore {
	ams := make([]core.MediatorInfo, 0)

	for _, med := range gp.ActiveMediators {
		medInfo := core.MediatorToInfo(&med)
		ams = append(ams, medInfo)
	}

	gpt := GlobalPropertyStore{
		ChainParameters: gp.ChainParameters,
		ActiveMediators: ams,
	}

	return gpt
}

func getGP(gpt *GlobalPropertyStore,StateDb StateDb) *GlobalProperty {
	ams := make(map[common.Address]core.Mediator, 0)
	for _, medInfo := range gpt.ActiveMediators {
		med := core.InfoToMediator(&medInfo)
		ams[med.Address] = med
	}

	gp := NewGlobalProp(StateDb)
	gp.ChainParameters = gpt.ChainParameters
	gp.ActiveMediators = ams

	return gp
}

func StoreGlobalProp (db  StateDb, gp *GlobalProperty) error {
	gpt := getGPT(gp)
	return db.SaveGlobalProperty(gpt)
}

func StoreDynGlobalProp(db  StateDb, dgp *DynamicGlobalProperty) error {
	return db.SaveDynamicGlobalProperty(*dgp)
}

func RetrieveGlobalProp(db  StateDb) (*GlobalProperty, error) {
	gpt, err := db.GetGlobalProperty()
	gp := getGP(&gpt,db)
	return gp, err
}

func RetrieveDynGlobalProp(db  StateDb) (*DynamicGlobalProperty, error) {
	dgp, err := db.GetDynamicGlobalProperty()
	return &dgp, err
}
//----------------from mediator_schedule.go
// Mediator调度顺序结构体
type MediatorSchedule struct {
	CurrentShuffledMediators []core.Mediator
	statedb  StateDb
}
// re:Yiran
// This function should only be called at the initGenesis()
func InitMediatorSchl(gp *GlobalProperty, dgp *DynamicGlobalProperty) *MediatorSchedule {
	log.Println("initialize mediator schedule...")
	ms := &MediatorSchedule{
		CurrentShuffledMediators: []core.Mediator{},
	}

	aSize := uint64(len(gp.ActiveMediators))
	if aSize == 0 {
		log.Println("The current number of active mediators is 0!")
	}

	// Create witness scheduler
	ms.CurrentShuffledMediators = make([]core.Mediator, aSize, aSize)
	meds := gp.GetInitActiveMediators()
	for i, add := range meds {
		med := gp.GetActiveMediator(add)
		ms.CurrentShuffledMediators[i] = *med
	}

	//	ms.UpdateMediatorSchedule(gp, dgp)

	return ms
}

func NewMediatorSchl(statedb StateDb) *MediatorSchedule {
	return &MediatorSchedule{
		CurrentShuffledMediators: []core.Mediator{},
		statedb:statedb,
	}
}

// 洗牌算法，更新mediator的调度顺序
func (ms *MediatorSchedule) UpdateMediatorSchedule(gp *GlobalProperty, dgp *DynamicGlobalProperty) {
	aSize := uint64(len(gp.ActiveMediators))
	if aSize == 0 {
		log.Println("The current number of active mediators is 0!")
		return
	}

	// 1. 判断是否到达洗牌时刻
	if dgp.LastVerifiedUnitNum%aSize != 0 {
		return
	}

	// 2. 清除CurrentShuffledMediators原来的空间，重新分配空间
	ms.CurrentShuffledMediators = make([]core.Mediator, aSize, aSize)

	// 3. 初始化数据
	meds := gp.GetActiveMediators()
	for i, add := range meds {
		med := gp.GetActiveMediator(add)
		ms.CurrentShuffledMediators[i] = *med
	}

	// 4. 打乱证人的调度顺序
	nowHi := uint64(dgp.LastVerifiedUnitTime << 32)
	for i := uint64(0); i < aSize; i++ {
		// 高性能随机生成器(High performance random generator)
		// 原理请参考 http://xorshift.di.unimi.it/
		k := nowHi + uint64(i)*2685821657736338717
		k ^= k >> 12
		k ^= k << 25
		k ^= k >> 27
		k *= 2685821657736338717

		jmax := aSize - i
		j := i + k%jmax

		// 进行N次随机交换
		ms.CurrentShuffledMediators[i], ms.CurrentShuffledMediators[j] =
			ms.CurrentShuffledMediators[j], ms.CurrentShuffledMediators[i]
	}
}

/**
@brief 获取指定的未来slotNum对应的调度mediator来生产见证单元.
Get the mediator scheduled for uint verification in a slot.

slotNum总是对应于未来的时间。
slotNum always corresponds to a time in the future.

如果slotNum == 1，则返回下一个调度Mediator。
If slotNum == 1, return the next scheduled mediator.

如果slotNum == 2，则返回下下一个调度Mediator。
If slotNum == 2, return the next scheduled mediator after 1 verified uint gap.
*/
func (ms *MediatorSchedule) GetScheduledMediator(dgp *DynamicGlobalProperty, slotNum uint32) *core.Mediator {
	currentASlot := dgp.CurrentASlot + uint64(slotNum)
	csmLen := len(ms.CurrentShuffledMediators)
	if csmLen == 0 {
		log.Println("The current number of shuffled mediators is 0!")
		return nil
	}

	// 由于创世单元不是有mediator生产，所以这里需要减1
	index := (currentASlot - 1) % uint64(csmLen)
	return &ms.CurrentShuffledMediators[index]
}

/**
计算在过去的128个见证单元生产slots中miss的百分比，不包括当前验证单元。
Calculate the percent of verifiedUnit production slots that were missed in the past 128 verifiedUnits,
not including the current verifiedUnit.
*/
//func MediatorParticipationRate(dgp *d.DynamicGlobalProperty) float32 {
//	return dgp.RecentSlotsFilled / 128.0
//}

/**
@brief 获取给定的未来第slotNum个slot开始的时间。
Get the time at which the given slot occurs.

如果slotNum == 0，则返回time.Unix(0,0)。
If slotNum == 0, return time.Unix(0,0).

如果slotNum == N 且 N > 0，则返回大于verifiedUnitTime的第N个单元验证间隔的对齐时间
If slotNum == N for N > 0, return the Nth next unit-interval-aligned time greater than head_block_time().
*/
func GetSlotTime(gp *GlobalProperty, dgp *DynamicGlobalProperty, slotNum uint32) time.Time {
	if slotNum == 0 {
		return time.Unix(0, 0)
	}

	interval := gp.ChainParameters.MediatorInterval

	// 本条件是用来生产第一个unit
	if dgp.LastVerifiedUnitNum == 0 {
		/**
		注：第一个验证单元在genesisTime加上一个验证单元间隔
		n.b. first verifiedUnit is at genesisTime plus one verifiedUnitInterval
		*/
		genesisTime := dgp.LastVerifiedUnitTime
		return time.Unix(genesisTime+int64(slotNum)*int64(interval), 0)
	}

	// 最近的验证单元的绝对slot
	var verifiedUnitAbsSlot = dgp.LastVerifiedUnitTime / int64(interval)
	// 最近的时间槽起始时间
	verifiedUnitSlotTime := time.Unix(verifiedUnitAbsSlot*int64(interval), 0)

	// 在此处添加区块链网络参数修改维护的所需要的slot

	/**
	如果是维护周期的话，加上维护间隔时间
	如果不是，就直接加上验证单元的slot时间
	*/
	// "slot 1" is verifiedUnitSlotTime,
	// plus maintenance interval if last uint is a maintenance verifiedUnit
	// plus verifiedUnit interval if last uint is not a maintenance verifiedUnit
	return verifiedUnitSlotTime.Add(time.Second * time.Duration(slotNum) * time.Duration(interval))
}

/**
获取在给定时间或之前出现的最近一个slot。 Get the last slot which occurs AT or BEFORE the given time.
*/
func GetSlotAtTime(gp *GlobalProperty, dgp *DynamicGlobalProperty, when time.Time) uint32 {
	/**
	返回值是所有满足 GetSlotTime（N）<= when 中最大的N
	The return value is the greatest value N such that GetSlotTime( N ) <= when.
	如果都不满足，则返回 0
	If no such N exists, return 0.
	*/
	firstSlotTime := GetSlotTime(gp, dgp, 1)

	if when.Before(firstSlotTime) {
		return 0
	}

	diffSecs := when.Unix() - firstSlotTime.Unix()
	interval := int64(gp.ChainParameters.MediatorInterval)

	return uint32(diffSecs/interval) + 1
}



type MediatorScheduleStore struct {
	CurrentShuffledMediators []core.MediatorInfo
}

func getMST(ms *MediatorSchedule) MediatorScheduleStore {
	csm := make([]core.MediatorInfo, 0)

	for _, med := range ms.CurrentShuffledMediators {
		medInfo := core.MediatorToInfo(&med)
		csm = append(csm, medInfo)
	}

	mst := MediatorScheduleStore{
		CurrentShuffledMediators: csm,
	}

	return mst
}

func getMS(mst *MediatorScheduleStore, statedb  StateDb) *MediatorSchedule {
	csm := make([]core.Mediator, 0)

	for _, medInfo := range mst.CurrentShuffledMediators {
		med := core.InfoToMediator(&medInfo)
		csm = append(csm, med)
	}

	ms := NewMediatorSchl(statedb)
	ms.CurrentShuffledMediators = csm
	return ms
}

func StoreMediatorSchl(db  StateDb, ms *MediatorSchedule) error {
	mst := getMST(ms)
	err := db.SaveMediatorSchedule(mst)
	if err != nil {
		log.Println(fmt.Sprintf("Store mediator schedule error: %s", err))
	}
	return err
}

func RetrieveMediatorSchl(stateDb  StateDb) (*MediatorSchedule, error) {
	mst, err := stateDb.GetMediatorSchedule()
	ms := getMS(&mst,stateDb)
	return ms, err
}