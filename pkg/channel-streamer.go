// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bwmarrin/discordgo"
	"github.com/golang/glog"
)

type ChannelStreamer interface {
	Stream(ctx context.Context, ch chan<- Channel) error
}

func NewChannelStreamer(
	session *discordgo.Session,
	serverID ServerID,
) ChannelStreamer {
	return &channelStreamer{
		session:  session,
		serverID: serverID,
	}
}

type channelStreamer struct {
	session  *discordgo.Session
	serverID ServerID
}

func (c *channelStreamer) Stream(ctx context.Context, ch chan<- Channel) error {
	channels, err := c.session.GuildChannels(c.serverID.String())
	if err != nil {
		return errors.Wrapf(ctx, err, "get channels failed")
	}
	for _, channel := range channels {
		switch channel.Type {
		case discordgo.ChannelTypeGuildText:
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- ChannelFromDiscordChannel(*channel):
			}
		default:
			glog.V(4).Infof("channel %s is no text channel => skip", channel.Name)
		}
	}
	glog.V(3).Infof("stream channels of server %s completed", c.serverID)
	return nil
}
