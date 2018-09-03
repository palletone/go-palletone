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

package ccprovider

import (
	"fmt"
	"sync"
)

// ccInfoCacheImpl implements in-memory cache for ChaincodeData
// needed by endorser to verify if the local instantiation policy
// matches the instantiation policy on a channel before honoring
// an invoke
type ccInfoCacheImpl struct {
	sync.RWMutex

	cache        map[string]*ChaincodeData
	cacheSupport CCCacheSupport
}

// NewCCInfoCache returns a new cache on top of the supplied CCInfoProvider instance
func NewCCInfoCache(cs CCCacheSupport) *ccInfoCacheImpl {
	return &ccInfoCacheImpl{
		cache:        make(map[string]*ChaincodeData),
		cacheSupport: cs,
	}
}

func (c *ccInfoCacheImpl) GetChaincodeData(ccname string, ccversion string) (*ChaincodeData, error) {
	// c.cache is guaranteed to be non-nil

	key := ccname + "/" + ccversion

	c.RLock()
	ccdata, in := c.cache[key]
	c.RUnlock()

	if !in {
		var err error

		// the chaincode data is not in the cache
		// try to look it up from the file system
		ccpack, err := c.cacheSupport.GetChaincode(ccname, ccversion)
		if err != nil || ccpack == nil {
			return nil, fmt.Errorf("cannot retrieve package for chaincode %s/%s, error %s", ccname, ccversion, err)
		}

		// we have a non-nil ChaincodeData, put it in the cache
		c.Lock()
		ccdata = ccpack.GetChaincodeData()
		c.cache[key] = ccdata
		c.Unlock()
	}

	return ccdata, nil
}
