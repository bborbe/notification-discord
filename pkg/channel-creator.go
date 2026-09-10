// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bwmarrin/discordgo"

	"github.com/bborbe/notification/discord"
)

//counterfeiter:generate -o mocks/channel-creator.go --fake-name ChannelCreator . ChannelCreator
type ChannelCreator interface {
	CreateChannel(ctx context.Context, name discord.ChannelName) (*Channel, error)
}

func NewChannelCreator(
	session *discordgo.Session,
	serverID ServerID,
) ChannelCreator {
	return &channelCreator{
		session:  session,
		serverID: serverID,
	}
}

type channelCreator struct {
	session  *discordgo.Session
	serverID ServerID
}

func (c *channelCreator) CreateChannel(
	ctx context.Context,
	name discord.ChannelName,
) (*Channel, error) {
	channel, err := c.session.GuildChannelCreate(
		c.serverID.String(),
		name.String(),
		discordgo.ChannelTypeGuildText,
	)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create channel '%s' failed", name)
	}
	result := Channel{
		ID:   ChannelID(channel.ID),
		Name: name,
	}
	if err := result.Validate(ctx); err != nil {
		return nil, errors.Wrapf(ctx, err, "create channel return no ID")
	}
	return result.Ptr(), nil
}
