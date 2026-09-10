// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/validation"
	"github.com/bwmarrin/discordgo"

	"github.com/bborbe/notification/discord"
)

type Channels []Channel

type Channel struct {
	ID   ChannelID
	Name discord.ChannelName
}

func (c Channel) Validate(ctx context.Context) error {
	return validation.All{
		validation.Name("ID", c.ID),
		validation.Name("Name", c.Name),
	}.Validate(ctx)
}

func (c Channel) Ptr() *Channel {
	return &c
}

func ChannelFromDiscordChannel(channel discordgo.Channel) Channel {
	return Channel{
		ID:   ChannelID(channel.ID),
		Name: discord.ChannelName(channel.Name),
	}
}
