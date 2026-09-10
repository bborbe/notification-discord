// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bwmarrin/discordgo"
	"github.com/golang/glog"

	"github.com/bborbe/notification/discord"
)

type MessageSender interface {
	Send(ctx context.Context, channelID ChannelID, message discord.Message) error
}

type MessageSenderFunc func(ctx context.Context, channelID ChannelID, message discord.Message) error

func (m MessageSenderFunc) Send(
	ctx context.Context,
	channelID ChannelID,
	message discord.Message,
) error {
	return m(ctx, channelID, message)
}

func NewMessageSender(session *discordgo.Session) MessageSender {
	return MessageSenderFunc(
		func(ctx context.Context, channelID ChannelID, message discord.Message) error {
			glog.V(3).Infof("send message to channel(%s) started", channelID)
			messageSend, err := session.ChannelMessageSend(channelID.String(), message.String())
			if err != nil {
				return errors.Wrapf(ctx, err, "send message to channel(%s) failed", channelID)
			}
			glog.V(3).
				Infof("send message to channel(%s) with %s completed", channelID, messageSend.ID)
			return nil
		},
	)
}
