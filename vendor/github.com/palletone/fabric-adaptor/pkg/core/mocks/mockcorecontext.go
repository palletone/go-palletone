/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/palletone/fabric-adaptor/pkg/common/providers/core"
	"github.com/palletone/fabric-adaptor/pkg/common/providers/msp"
)

// MockCoreContext is a mock core context
type MockCoreContext struct {
	MockConfig            core.CryptoSuiteConfig
	MockCryptoSuite       core.CryptoSuite
	MockUserStore         msp.UserStore
	MockSigningManager    core.SigningManager
	MockCryptoSuiteConfig core.CryptoSuiteConfig
}

// CryptoSuite ...
func (m *MockCoreContext) CryptoSuite() core.CryptoSuite {
	return m.MockCryptoSuite
}

// SigningManager ...
func (m *MockCoreContext) SigningManager() core.SigningManager {
	return m.MockSigningManager
}

//CryptoSuiteConfig ...
func (m *MockCoreContext) CryptoSuiteConfig() core.CryptoSuiteConfig {
	return m.MockCryptoSuiteConfig
}
