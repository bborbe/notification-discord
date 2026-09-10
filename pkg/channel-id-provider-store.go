// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/golang/glog"

	"github.com/bborbe/notification/discord"
)

func NewChannelIDProviderStore(
	channelStore ChannelStore,
	channelIDProvider ChannelIDProvider,
	channelExistsChecker ChannelExistsChecker,
) ChannelIDProvider {
	return ChannelIDProviderFunc(
		func(ctx context.Context, channelName discord.ChannelName) (*ChannelID, error) {
			channelID, err := channelStore.Get(ctx, channelName)
			if err == nil {
				glog.V(4).Infof("found channelName(%s) in store", channelName)
				exists, err := channelExistsChecker.Exists(ctx, *channelID)
				if err != nil {
					return nil, errors.Wrapf(ctx, err, "exists failed")
				}
				if exists {
					glog.V(4).Infof("channelName(%s) exists => return", channelName)
					return channelID, nil
				}
				glog.V(2).Infof("channelName(%s) does not exist => create", channelName)
			}
			channelID, err = channelIDProvider.FindOrCreate(ctx, channelName)
			if err != nil {
				return nil, errors.Wrapf(ctx, err, "find or create channel(%s) failed", channelName)
			}
			glog.V(4).Infof("found channelID(%s) for channelName(%s)", channelID, channelName)
			if err := channelStore.Add(ctx, channelName, *channelID); err != nil {
				return nil, errors.Wrapf(ctx, err, "add channel failed")
			}
			glog.V(4).Infof("channelName(%s) added to store", channelName)
			return channelID, nil
		},
	)
}
