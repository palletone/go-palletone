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


package ledger

import (
	"fmt"

	"github.com/palletone/go-palletone/core/vmContractPub/ledger"
)

type MockQueryExecutor struct {
	// State keeps all namespaces
	State map[string]map[string][]byte
}

func NewMockQueryExecutor(state map[string]map[string][]byte) *MockQueryExecutor {
	return &MockQueryExecutor{
		State: state,
	}
}

func (m *MockQueryExecutor) GetState(namespace string, key string) ([]byte, error) {
	ns := m.State[namespace]
	if ns == nil {
		return nil, fmt.Errorf("Could not retrieve namespace %s", namespace)
	}

	return ns[key], nil
}

func (m *MockQueryExecutor) GetStateMultipleKeys(namespace string, keys []string) ([][]byte, error) {
	return nil, nil

}

func (m *MockQueryExecutor) GetStateRangeScanIterator(namespace string, startKey string, endKey string) (ledger.ResultsIterator, error) {
	return nil, nil

}

func (m *MockQueryExecutor) ExecuteQuery(namespace, query string) (ledger.ResultsIterator, error) {
	return nil, nil
}

func (m *MockQueryExecutor) GetPrivateData(namespace, collection, key string) ([]byte, error) {
	return nil, nil
}

func (m *MockQueryExecutor) GetPrivateDataMultipleKeys(namespace, collection string, keys []string) ([][]byte, error) {
	return nil, nil
}

func (m *MockQueryExecutor) GetPrivateDataRangeScanIterator(namespace, collection, startKey, endKey string) (ledger.ResultsIterator, error) {
	return nil, nil
}

func (m *MockQueryExecutor) ExecuteQueryOnPrivateData(namespace, collection, query string) (ledger.ResultsIterator, error) {
	return nil, nil
}

func (m *MockQueryExecutor) Done() {
}
