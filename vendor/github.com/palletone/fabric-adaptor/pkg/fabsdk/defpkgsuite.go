/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/palletone/fabric-adaptor/pkg/core/logging/api"
	"github.com/palletone/fabric-adaptor/pkg/core/logging/modlog"
	sdkApi "github.com/palletone/fabric-adaptor/pkg/fabsdk/api"
	"github.com/palletone/fabric-adaptor/pkg/fabsdk/factory/defcore"
	"github.com/palletone/fabric-adaptor/pkg/fabsdk/factory/defmsp"
	"github.com/palletone/fabric-adaptor/pkg/fabsdk/factory/defsvc"
)

type defPkgSuite struct{}

func (ps *defPkgSuite) Core() (sdkApi.CoreProviderFactory, error) {
	return defcore.NewProviderFactory(), nil
}

func (ps *defPkgSuite) MSP() (sdkApi.MSPProviderFactory, error) {
	return defmsp.NewProviderFactory(), nil
}

func (ps *defPkgSuite) Service() (sdkApi.ServiceProviderFactory, error) {
	return defsvc.NewProviderFactory(), nil
}

func (ps *defPkgSuite) Logger() (api.LoggerProvider, error) {
	return modlog.LoggerProvider(), nil
}
