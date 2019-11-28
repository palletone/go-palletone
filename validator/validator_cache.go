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

package validator

import (
	"encoding/json"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
)

//如果已经验证通过的对象，那么不需要重复做全部验证
type ValidatorCache struct {
	cache palletcache.ICache
}

var expireSeconds = 600
var prefixTx = []byte("VT")
var prefixUnit = []byte("VU")
var prefixHeader = []byte("VH")

func NewValidatorCache(cache palletcache.ICache) *ValidatorCache {
	return &ValidatorCache{cache: cache}
}
func (s *ValidatorCache) AddTxValidateResult(txId common.Hash, validateResult []*modules.Addition) {
	data, err := json.Marshal(validateResult)
	if err != nil {
		log.Errorf("Json marsal struct fail,error:%s", err.Error())
	}
	s.cache.Set(append(prefixTx, txId.Bytes()...), data, expireSeconds)
}
func (s *ValidatorCache) HasTxValidateResult(txId common.Hash) (bool, []*modules.Addition) {
	data, err := s.cache.Get(append(prefixTx, txId.Bytes()...))
	if err != nil {
		return false, nil
	}

	result := []*modules.Addition{}
	json.Unmarshal(data, &result)
	log.Debugf("Validate cache has tx hash:%s", txId.String())
	return true, result
}

func (s *ValidatorCache) AddUnitValidateResult(unitHash common.Hash, code ValidationCode) {

	s.cache.Set(append(prefixUnit, unitHash.Bytes()...), []byte{byte(code)}, expireSeconds)
}
func (s *ValidatorCache) HasUnitValidateResult(unitHash common.Hash) (bool, ValidationCode) {
	data, err := s.cache.Get(append(prefixUnit, unitHash.Bytes()...))
	if err != nil {
		return false, TxValidationCode_NOT_VALIDATED
	}
	log.Debugf("Validate cache has unit hash:%s", unitHash.String())
	return true, ValidationCode(data[0])
}

func (s *ValidatorCache) AddHeaderValidateResult(unitHash common.Hash, code ValidationCode) {

	s.cache.Set(append(prefixHeader, unitHash.Bytes()...), []byte{byte(code)}, expireSeconds)
}
func (s *ValidatorCache) HasHeaderValidateResult(unitHash common.Hash) (bool, ValidationCode) {
	data, err := s.cache.Get(append(prefixHeader, unitHash.Bytes()...))
	if err != nil {
		return false, TxValidationCode_NOT_VALIDATED
	}
	log.Debugf("Validate cache has header hash:%s", unitHash.String())
	return true, ValidationCode(data[0])
}
