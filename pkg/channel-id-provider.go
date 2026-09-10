// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/golang/glog"

	"github.com/bborbe/notification/discord"
)

type ChannelIDProvider interface {
	FindOrCreate(ctx context.Context, name discord.ChannelName) (*ChannelID, error)
}

type ChannelIDProviderFunc func(ctx context.Context, name discord.ChannelName) (*ChannelID, error)

func (c ChannelIDProviderFunc) FindOrCreate(
	ctx context.Context,
	name discord.ChannelName,
) (*ChannelID, error) {
	return c(ctx, name)
}

func NewChannelIDProvider(
	channelIDFinder ChannelIDFinder,
	channelCreator ChannelCreator,
) ChannelIDProvider {
	return ChannelIDProviderFunc(
		func(ctx context.Context, name discord.ChannelName) (*ChannelID, error) {
			channels, err := channelIDFinder.FindByName(ctx, name)
			if err != nil {
				return nil, errors.Wrapf(ctx, err, "find or create channel '%s' failed", name)
			}
			glog.V(3).Infof("found %d channels for name '%s'", len(channels), name)
			switch len(channels) {
			case 0:
				channel, err := channelCreator.CreateChannel(ctx, name)
				if err != nil {
					return nil, errors.Wrapf(ctx, err, "create channel(%s) failed", name)
				}
				glog.V(3).Infof("created channel(%s) for name '%s'", channel.ID, name)
				return channel.ID.Ptr(), nil
			case 1:
				channelID := channels[0].ID
				glog.V(3).Infof("found channel(%s) for name '%s'", channelID, name)
				return channelID.Ptr(), nil
			default:
				return nil, errors.Wrapf(
					ctx,
					MultipleChannelFoundError,
					"found multiple channels for name '%s'",
					name,
				)
			}
		},
	)
}
