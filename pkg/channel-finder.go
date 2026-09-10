// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/bborbe/collection"
	"github.com/bborbe/errors"
	"github.com/golang/glog"

	"github.com/bborbe/notification/discord"
)

var ChannelNotFoundError = stderrors.New("channel not found")

var MultipleChannelFoundError = stderrors.New("found multiple channels")

//counterfeiter:generate -o mocks/channel-id-finder.go --fake-name ChannelIDFinder . ChannelIDFinder
type ChannelIDFinder interface {
	FindByName(ctx context.Context, name discord.ChannelName) (Channels, error)
}

func NewChannelIDFinder(
	channelStreamer ChannelStreamer,
) ChannelIDFinder {
	return &channelFinder{
		channelStreamer: channelStreamer,
	}
}

type channelFinder struct {
	channelStreamer ChannelStreamer
}

func (c *channelFinder) FindByName(
	ctx context.Context,
	name discord.ChannelName,
) (Channels, error) {
	glog.V(3).Infof("find channel with name %s started", name)
	var results Channels
	err := collection.ChannelFnMap(
		ctx,
		c.channelStreamer.Stream,
		func(ctx context.Context, channel Channel) error {
			channelNameStr := strings.ToLower(channel.Name.String())
			nameStr := strings.ToLower(name.String())
			match := channelNameStr == nameStr
			glog.V(4).Infof("compare channel '%s' == '%s' => %v", channelNameStr, nameStr, match)
			if match == false {
				return nil
			}
			results = append(results, channel)
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "find channel failed")
	}
	glog.V(3).Infof("found %d channel with name %s completed", len(results), name)
	return results, nil
}
