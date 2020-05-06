// +build !pprof

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/palletone/fabric-adaptor/pkg/client/channel/invoke"
)

func callQuery(cc *Client, request Request, options ...RequestOption) (Response, error) {
	return cc.InvokeHandler(invoke.NewQueryHandler(), request, options...)
}

func callExecute(cc *Client, request Request, options ...RequestOption) (Response, error) {
	return cc.InvokeHandler(invoke.NewExecuteHandler(), request, options...)
}

func callExecuteZxl(cc *Client, request Request, options ...RequestOption) (Response, error) {//Zxl add
	return cc.InvokeHandler(invoke.NewSelectAndEndorseHandlerZxl(), request, options...)
}

func callExecuteBroadcastZxl(cc *Client, request Request,
	options ...RequestOption) (Response, error) {//Zxl add
	return cc.InvokeHandler(invoke.NewExecuteBroadcastHandler(), request, options...)
}
func callExecuteBroadcastFirstZxl(cc *Client, request Request,
	options ...RequestOption) (Response, error) {//Zxl add
	return cc.InvokeHandler(invoke.NewExecuteBroadcastFirstHandler(), request, options...)
}
func callExecuteSignFirstZxl(cc *Client, request Request, options ...RequestOption) (Response, error) {//Zxl add
	return cc.InvokeHandler(invoke.NewCommitTxSignHandler(), request, options...)
}
func callExecuteBroadcastSecondZxl(cc *Client, request Request,
	options ...RequestOption) (Response, error) {//Zxl add
	return cc.InvokeHandler(invoke.NewExecuteBroadcastSecondHandler(), request, options...)
}