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

package rwset

type RWSetBuilder struct {
	pubRwBuilderMap map[string]*nsPubRwBuilder
}

type nsPubRwBuilder struct {
	namespace         string
	readMap           map[string]*KVRead
	writeMap          map[string]*KVWrite
}

func NewRWSetBuilder() *RWSetBuilder {
	return &RWSetBuilder{make(map[string]*nsPubRwBuilder)}
}

func (b *RWSetBuilder) AddToReadSet(ns string, key string, version *Version) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	nsPubRwBuilder.readMap[key] = NewKVRead(key, version)
}

func (b *RWSetBuilder) AddToWriteSet(ns string, key string, value []byte) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	nsPubRwBuilder.writeMap[key] = newKVWrite(key, value)
}

func (b *RWSetBuilder) getOrCreateNsPubRwBuilder(ns string) *nsPubRwBuilder {
	nsPubRwBuilder, ok := b.pubRwBuilderMap[ns]
	if !ok {
		nsPubRwBuilder = newNsPubRwBuilder(ns)
		b.pubRwBuilderMap[ns] = nsPubRwBuilder
	}
	return nsPubRwBuilder
}

func newNsPubRwBuilder(namespace string) *nsPubRwBuilder {
	return &nsPubRwBuilder{
		namespace,
		make(map[string]*KVRead),
		make(map[string]*KVWrite),
	}
}