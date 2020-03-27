/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/palletone/fabric-adaptor/pkg/client/common/discovery/greylist"
	"github.com/palletone/fabric-adaptor/pkg/common/providers/context"
	"github.com/palletone/fabric-adaptor/pkg/common/providers/fab"
)

func newClient(channelContext context.Channel, membership fab.ChannelMembership, eventService fab.EventService, greylistProvider *greylist.Filter) Client {
	channelClient := Client{
		membership:   membership,
		eventService: eventService,
		greylist:     greylistProvider,
		context:      channelContext,
		metrics:      channelContext.GetMetrics(),
	}
	return channelClient
}
